// backend-go/internal/modules/tcps_proxy/constants.go
package tcps_proxy
const (
	ModeClient = "client"
	ModeServer = "server"
)
const (
	ProtocolTCP   = "tcp"
	ProtocolTLS   = "tls"
	ProtocolPQC   = "pqc"
	ProtocolGMSSL = "gmssl"
)
const (
	StatusStopped  = "STOPPED"
	StatusStarting = "STARTING"
	StatusRunning  = "RUNNING"
	StatusStopping = "STOPPING"
	StatusFailed   = "FAILED"
)
// tcps 专用部署路径（在 DeployRoot 之外，独立子目录）
const (
	TcpsSubDir       = "tcps"
	TcpsConfSubDir   = "conf/tcps"
	TcpsPidSubDir    = "pid/tcps"
	TcpsTmpSubDir    = "tmp/tcps"
	TcpsLogSubDir    = "logs/tcps_proxy"
)
// ---------- 默认 SSL 参数 ----------
const (
	DefaultSSLProtocols   = "TLSv1.2 TLSv1.3"
	DefaultGMSSLProtocols = "NTLSv1.1"
	DefaultPQCProtocols   = "TLSv1.3"
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
// IsTCPSProtocol 是否 TCPS 合法协议。
func IsTCPSProtocol(p string) bool {
	switch p {
	case ProtocolTCP, ProtocolTLS, ProtocolPQC, ProtocolGMSSL:
		return true
	}
	return false
}
// ProtocolNeedsSSL 协议是否需要 SSL（tcp 除外）。
func ProtocolNeedsSSL(p string) bool {
	return p != ProtocolTCP
}
// defaultProtocolsFor 按 protocol 返回默认 ssl_protocols。
//
// ★ gmssl 返回 NTLSv1.1（仅作为 UI/持久化标识，渲染 nginx.conf 时被过滤）；
//   tls / pqc 返回标准 TLS 值。
func defaultProtocolsFor(protocol string) string {
	switch protocol {
	case ProtocolPQC:
		return DefaultPQCProtocols
	case ProtocolGMSSL:
		return DefaultGMSSLProtocols
	default:
		return DefaultSSLProtocols
	}
}
