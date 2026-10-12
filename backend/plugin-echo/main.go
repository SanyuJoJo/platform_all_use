package main
import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "github.com/hashicorp/go-plugin"
    "github.com/yourorg/backend/contract"
)
// EchoPlugin 示例：把输入原样返回，可加前缀
type EchoPlugin struct {
    prefix string
}
func (p *EchoPlugin) Name() string    { return "echo" }
func (p *EchoPlugin) Version() string { return "1.0.0" }
func (p *EchoPlugin) Initialize(_ context.Context, cfg map[string]any) error {
    if v, ok := cfg["prefix"].(string); ok {
        p.prefix = v
    } else {
        p.prefix = "[echo] "
    }
    log.Printf("EchoPlugin 初始化完成 prefix=%q", p.prefix)
    return nil
}
func (p *EchoPlugin) Execute(_ context.Context, input any) (any, error) {
    return fmt.Sprintf("%s%v", p.prefix, input), nil
}
func (p *EchoPlugin) Shutdown(_ context.Context) error {
    log.Println("EchoPlugin 关闭")
    return nil
}
func main() {
    impl := &EchoPlugin{}
    // 独立启动模式（HTTP）
    if os.Getenv("PLUGIN_STANDALONE") == "true" {
        runStandalone(impl)
        return
    }
    // 托管模式（go-plugin RPC）
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: contract.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            contract.PluginKey: &contract.PluginRPC{Impl: impl},
        },
    })
}
// runStandalone 以 HTTP 独立服务运行，遵循 contract/http_protocol.go 中的协议
func runStandalone(p *EchoPlugin) {
    _ = p.Initialize(context.Background(), map[string]any{})
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        _ = json.NewEncoder(w).Encode(contract.HTTPHealthResponse{
            Status:  "ok",
            Name:    p.Name(),
            Version: p.Version(),
        })
    })
    mux.HandleFunc("/initialize", func(w http.ResponseWriter, r *http.Request) {
        var req contract.HTTPInitializeRequest
        _ = json.NewDecoder(r.Body).Decode(&req)
        if err := p.Initialize(r.Context(), req.Config); err != nil {
            _ = json.NewEncoder(w).Encode(contract.HTTPGenericResponse{Error: err.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(contract.HTTPGenericResponse{})
    })
    mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
        var req contract.HTTPExecuteRequest
        _ = json.NewDecoder(r.Body).Decode(&req)
        out, err := p.Execute(r.Context(), req.Input)
        resp := contract.HTTPExecuteResponse{Output: out}
        if err != nil {
            resp.Error = err.Error()
        }
        _ = json.NewEncoder(w).Encode(resp)
    })
    mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
        _ = p.Shutdown(r.Context())
        _ = json.NewEncoder(w).Encode(contract.HTTPGenericResponse{})
    })
    port := os.Getenv("PLUGIN_PORT")
    if port == "" {
        port = "9100"
    }
    srv := &http.Server{Addr: ":" + port, Handler: mux}
    go func() {
        log.Printf("EchoPlugin 独立 HTTP 监听 :%s", port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("HTTP 异常: %v", err)
        }
    }()
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(ctx)
    _ = p.Shutdown(ctx)
}
