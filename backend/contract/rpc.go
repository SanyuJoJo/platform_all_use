package contract
import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/rpc"
    "github.com/hashicorp/go-plugin"
)
// PluginRPC 是 go-plugin 的服务端/客户端实现。
// 插件方通过 Server() 暴露 RPC 方法；主应用通过 Client() 获取 RPC 代理。
type PluginRPC struct {
    plugin.NetRPCUnsupportedPlugin
    Impl Plugin // 服务端使用；客户端为 nil
}
// Server 由 go-plugin 调用，返回承载 RPC 方法的对象
func (p *PluginRPC) Server(_ *plugin.MuxBroker) (interface{}, error) {
    if p.Impl == nil {
        return nil, errors.New("contract: PluginRPC.Impl 未设置")
    }
    return &RPCServer{Impl: p.Impl}, nil
}
// Client 由 go-plugin 调用，返回主应用侧的 RPC 代理
func (p *PluginRPC) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
    return &RPCClient{rpc: c}, nil
}
// ——— RPC 服务端 ———
// RPCServer 承载所有 RPC 方法
type RPCServer struct {
    Impl Plugin
}
// InitArgs Initialize 的参数
type InitArgs struct {
    Config map[string]any
}
// ExecuteArgs Execute 的参数
type ExecuteArgs struct {
    Input string // JSON 序列化后的输入
}
// ExecuteReply Execute 的返回值
type ExecuteReply struct {
    Output string // JSON 序列化后的输出
    Error  string
}
func (s *RPCServer) Name(_ struct{}, resp *string) error {
    *resp = s.Impl.Name()
    return nil
}
func (s *RPCServer) Version(_ struct{}, resp *string) error {
    *resp = s.Impl.Version()
    return nil
}
func (s *RPCServer) Initialize(args InitArgs, resp *string) error {
    if err := s.Impl.Initialize(context.Background(), args.Config); err != nil {
        return err
    }
    *resp = "ok"
    return nil
}
func (s *RPCServer) Execute(args ExecuteArgs, reply *ExecuteReply) error {
    out, err := s.Impl.Execute(context.Background(), args.Input)
    if err != nil {
        reply.Error = err.Error()
        return nil
    }
    b, jerr := json.Marshal(out)
    if jerr != nil {
        reply.Output = fmt.Sprintf("%v", out)
    } else {
        reply.Output = string(b)
    }
    return nil
}
func (s *RPCServer) Shutdown(_ struct{}, resp *string) error {
    if err := s.Impl.Shutdown(context.Background()); err != nil {
        return err
    }
    *resp = "ok"
    return nil
}
// ——— RPC 客户端 ———
// RPCClient 主应用侧使用的 RPC 代理，实现 Plugin 接口
type RPCClient struct {
    rpc *rpc.Client
}
func (c *RPCClient) Name() string {
    var resp string
    _ = c.rpc.Call("Plugin.Name", struct{}{}, &resp)
    return resp
}
func (c *RPCClient) Version() string {
    var resp string
    _ = c.rpc.Call("Plugin.Version", struct{}{}, &resp)
    return resp
}
func (c *RPCClient) Initialize(ctx context.Context, config map[string]any) error {
    var resp string
    return c.rpc.Call("Plugin.Initialize", InitArgs{Config: config}, &resp)
}
func (c *RPCClient) Execute(ctx context.Context, input any) (any, error) {
    b, err := json.Marshal(input)
    if err != nil {
        return nil, err
    }
    var reply ExecuteReply
    if err := c.rpc.Call("Plugin.Execute", ExecuteArgs{Input: string(b)}, &reply); err != nil {
        return nil, err
    }
    if reply.Error != "" {
        return nil, errors.New(reply.Error)
    }
    var out any
    if err := json.Unmarshal([]byte(reply.Output), &out); err != nil {
        return reply.Output, nil
    }
    return out, nil
}
func (c *RPCClient) Shutdown(ctx context.Context) error {
    var resp string
    return c.rpc.Call("Plugin.Shutdown", struct{}{}, &resp)
}
// PluginMap 用于 go-plugin 注册，主应用与插件共用
var PluginMap = map[string]plugin.Plugin{
    PluginKey: &PluginRPC{},
}
