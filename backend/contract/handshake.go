package contract
import "github.com/hashicorp/go-plugin"
const (
    // PluginKey 是 go-plugin 插件映射中的键名，主应用与插件必须一致
    PluginKey = "backend_plugin"
    // ProtocolVersion 是 go-plugin 协议版本，破坏性变更时递增
    ProtocolVersion = 1
    // MagicCookieKey / MagicCookieValue 用于 go-plugin 握手校验
    MagicCookieKey   = "BACKEND_PLUGIN"
    MagicCookieValue = "go-plugin-backend-v1"
)
// HandshakeConfig go-plugin 握手配置，主应用与插件共享
var HandshakeConfig = plugin.HandshakeConfig{
    ProtocolVersion:  ProtocolVersion,
    MagicCookieKey:   MagicCookieKey,
    MagicCookieValue: MagicCookieValue,
}
