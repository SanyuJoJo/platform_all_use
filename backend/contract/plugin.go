// Package contract 定义主应用与所有插件之间共享的契约。
//
// 契约分为两个维度：
//   - Plugin 接口：Go 插件实现该接口，通过 go-plugin RPC 与主应用通信。
//   - HTTP 协议（见 http_protocol.go）：非 Go 语言插件按此协议实现 HTTP 服务。
//
// 插件开发方仅依赖本模块，不需依赖主应用任何内部代码。
package contract
import "context"
// Plugin 是所有 Go 插件必须实现的接口。
//
// 该接口一经发布应保持向后兼容，任何破坏性变更都必须通过升级 ProtocolVersion
// 并同时更新 HandshakeConfig 来通知主应用。
type Plugin interface {
    // Name 返回插件唯一标识（与 manifest.name 一致）
    Name() string
    // Version 返回插件版本号（与 manifest.version 一致）
    Version() string
    // Initialize 在插件加载时调用一次，用于注入配置
    Initialize(ctx context.Context, config map[string]any) error
    // Execute 执行插件核心逻辑
    Execute(ctx context.Context, input any) (any, error)
    // Shutdown 在插件停止前调用，用于优雅释放资源
    Shutdown(ctx context.Context) error
}
