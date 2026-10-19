// backend-go/internal/modules/ssl_proxy/constants.go
package ssl_proxy

import "os"

const (
	ModeClient = "client"
	ModeServer = "server"
)

const (
	ProtocolHTTP  = "http"
	ProtocolPQC   = "pqc"
	ProtocolGMSSL = "gmssl"
)

const (
	BackendSchemeHTTP  = "http"
	BackendSchemeHTTPS = "https"
)

const (
	StatusStopped  = "STOPPED"
	StatusStarting = "STARTING"
	StatusRunning  = "RUNNING"
	StatusStopping = "STOPPING"
	StatusFailed   = "FAILED"
)

var (
	DeployRoot    = envOr("PLATS_DEPLOY_ROOT", "/opt/plats_tool")
	CoreRoot      = DeployRoot + "/core"
	TengineBin    = CoreRoot + "/sbin/nginx"
	MimeTypesPath = CoreRoot + "/conf/mime.types"
	TemplateDir   = CoreRoot + "/conf/template"

	ConfRoot = CoreRoot + "/conf"
	PidRoot  = CoreRoot + "/pid"
	TmpRoot  = CoreRoot + "/tmp"

	UserConfRoot = DeployRoot + "/user_data/conf/core"
	LogRoot      = DeployRoot + "/user_data/logs/ssl_proxy"
	CertRoot     = DeployRoot + "/user_data/cert"
)

// ---------- 默认 SSL 参数 ----------
const (
	// ★ DefaultSSLProtocols 用于 ssl_protocols 指令的默认值。
	//   nginx 只接受 TLSv1 / TLSv1.1 / TLSv1.2 / TLSv1.3 / SSLv2 / SSLv3。
	DefaultSSLProtocols = "TLSv1.2 TLSv1.3"

	// ★ DefaultGMSSLProtocols 仅用于前端展示 / 文档 / API 选项标识。
	//   **绝不写入 nginx.conf 的 ssl_protocols**（nginx 会报 invalid value）。
	//   国密 NTLS 通道由 nginx 指令 `enable_ntls on;` 开启，与 ssl_protocols 无关。
	DefaultGMSSLProtocols = "NTLSv1.1"

	DefaultPQCProtocols = "TLSv1.3"

	DefaultSSLCiphers = "HIGH:!RC4:!3DES:!aDSS:!aNULL:!kPSK:!kSRP:!MD5:@STRENGTH:+SHA1:+kRSA:ECC-SM4-SM3:ECDHE-SM4-SM3:ECC-SM4-GCM-SM3:ECDHE-SM4-GCM-SM3"

	DefaultPQCConfCmd = "Groups X25519MLKEM768:X25519"

	DefaultSSLSessionCache   = "shared:SSL:10m"
	DefaultSSLSessionTimeout = "10m"
)

const (
	SSLCiphersModeDefault = "default"
	SSLCiphersModeCustom  = "custom"
	SSLCiphersModeInput   = "input"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
