// backend-go/internal/modules/ssl_proxy/nginx_generator.go
package ssl_proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type NginxGenerator struct {
	TemplateDir  string
	CertResolver CertResolver
	ConfRoot     string
	PidRoot      string
	TmpRoot      string
	LogRoot      string
}

func NewNginxGenerator(resolver CertResolver) *NginxGenerator {
	return &NginxGenerator{
		TemplateDir:  TemplateDir,
		CertResolver: resolver,
		ConfRoot:     ConfRoot,
		PidRoot:      PidRoot,
		TmpRoot:      TmpRoot,
		LogRoot:      LogRoot,
	}
}

// Render 渲染 nginx.conf。
//
// ★ 本版语义（按用户要求）：
//   - 客户端 TLS 场景下，验证四件套 + CA 加载全部开启：
//     proxy_ssl_server_name / proxy_ssl_name / proxy_ssl_trusted_certificate /
//     proxy_ssl_verify / proxy_ssl_verify_depth
//   - proxy_ssl_name 值：
//       用户填了 SSL Server Name → 用它
//       用户没填 → 输出 ''（空字符串，用户已修改 nginx 适配）
//   - 不做自动探测、不读本地证书、不用 backend_ip 兜底
func (g *NginxGenerator) Render(p *PlatformSSLProxy) (confPath, pidPath string, err error) {
	confDir := filepath.Join(g.ConfRoot, p.ServiceID)
	pidDir := filepath.Join(g.PidRoot, p.ServiceID)
	tmpDir := filepath.Join(g.TmpRoot, p.ServiceID)
	logDir := filepath.Join(g.LogRoot, p.ServiceID)

	for _, d := range []struct {
		path string
		mode os.FileMode
	}{
		{confDir, 0700},
		{pidDir, 0700},
		{tmpDir, 0755},
		{logDir, 0750},
	} {
		if err := os.MkdirAll(d.path, d.mode); err != nil {
			return "", "", newErr(ErrConfigGenerate, "创建目录失败", d.path+": "+err.Error())
		}
	}

	confPath = filepath.Join(confDir, "nginx.conf")
	pidPath = filepath.Join(pidDir, "nginx.pid")

	certs, err := g.CertResolver.Resolve(p, confDir)
	if err != nil {
		return "", "", err
	}

	// 直接用用户填的值，不做兜底
	resolvedSNI := strings.TrimSpace(p.SSLServerName)

	vars := buildVars(p, certs, confDir, pidDir, tmpDir, logDir, pidPath, resolvedSNI)

	var tmplFile string
	if p.Mode == ModeClient {
		tmplFile = filepath.Join(g.TemplateDir, "nchttp.conf.tmpl")
	} else {
		tmplFile = filepath.Join(g.TemplateDir, "nshttp.conf.tmpl")
	}
	raw, err := os.ReadFile(tmplFile)
	if err != nil {
		return "", "", newErr(ErrConfigGenerate, "读取模板失败", tmplFile+": "+err.Error())
	}

	rendered := g.renderLines(string(raw), p, vars, resolvedSNI)

	if err := os.WriteFile(confPath, []byte(rendered), 0640); err != nil {
		return "", "", newErr(ErrConfigGenerate, "写入 nginx.conf 失败", err.Error())
	}
	return confPath, pidPath, nil
}

func (g *NginxGenerator) renderLines(
	raw string, p *PlatformSSLProxy, vars map[string]string, resolvedSNI string,
) string {
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")

		if strings.HasPrefix(trimmed, "#@") {
			indent := line[:len(line)-len(trimmed)]
			cmdLine := strings.TrimSpace(strings.TrimPrefix(trimmed, "#@"))
			cmdName := firstToken(cmdLine)

			rendered := renderVars(cmdLine, vars)
			if shouldEnable(cmdName, p, resolvedSNI) {
				out = append(out, indent+rendered)
			} else {
				out = append(out, indent+"# "+rendered)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			out = append(out, renderVars(line, vars))
			continue
		}

		out = append(out, renderVars(line, vars))
	}
	return strings.Join(out, "\n")
}

func firstToken(s string) string {
	s = strings.TrimSpace(s)
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			return s[:i]
		}
	}
	return s
}

var varRe = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func renderVars(line string, vars map[string]string) string {
	return varRe.ReplaceAllStringFunc(line, func(m string) string {
		key := strings.TrimSpace(strings.Trim(m, "{}"))
		if v, ok := vars[key]; ok {
			return v
		}
		return ""
	})
}

// buildVars 构建模板变量。
//
// ★ 变量职责划分：
//   - ssl_server_name        原始值（不兜底），用于日志/调试
//   - server_name_value      给 server_name 用，空时兜底 "_"（nginx 要求非空）
//   - proxy_ssl_name_expr    给 proxy_ssl_name 用，空时输出 "''"
//                            （用户已改 nginx 支持空字符串）
func buildVars(
	p *PlatformSSLProxy,
	c *ResolvedCertPaths,
	confDir, pidDir, tmpDir, logDir, pidPath string,
	resolvedSNI string,
) map[string]string {
	listenSuffix := ""
	if p.Mode == ModeServer && p.ListenSSL {
		listenSuffix = " ssl"
	}

	sslProtocols := sanitizeSSLProtocols(p.SSLProtocols, p.Protocol)

	// proxy_ssl_name 表达式：有值原样输出，无值输出 ''
	proxySSLNameExpr := resolvedSNI
	if proxySSLNameExpr == "" {
		proxySSLNameExpr = "''"
	}

	v := map[string]string{
		"service_id":        p.ServiceID,
		"service_name":      p.ServiceName,
		"mode":              p.Mode,
		"protocol":          p.Protocol,
		"listen_ip":         p.ListenIP,
		"listen_port":       fmt.Sprintf("%d", p.ListenPort),
		"listen_ssl_suffix": listenSuffix,
		"backend_ip":        p.BackendIP,
		"backend_port":      fmt.Sprintf("%d", p.BackendPort),
		"backend_scheme":    p.BackendScheme,

		// ★ 三变量并存，各司其职
		"ssl_server_name":      resolvedSNI,                            // 原始值
		"server_name_value":    defStr(resolvedSNI, "_"),               // server_name 用
		"proxy_ssl_name_expr":  proxySSLNameExpr,                       // proxy_ssl_name 用

		"ssl_protocols":    sslProtocols,
		"ssl_ciphers":      defaultIfEmpty(p.SSLCiphers, DefaultSSLCiphers),
		"ssl_conf_command": p.SSLConfCommand,

		"pid_path":        pidPath,
		"conf_dir":        confDir,
		"pid_dir":         pidDir,
		"tmp_dir":         tmpDir,
		"log_dir":         logDir,
		"mime_types_path": MimeTypesPath,

		"cert_path":        c.ServerCertPath,
		"key_path":         c.ServerKeyPath,
		"ca_cert_path":     c.CAPath,
		"client_cert_path": c.ClientCertPath,
		"client_key_path":  c.ClientKeyPath,

		"server_sign_cert": c.ServerSignCert,
		"server_sign_key":  c.ServerSignKey,
		"server_enc_cert":  c.ServerEncCert,
		"server_enc_key":   c.ServerEncKey,
		"client_sign_cert": c.ClientSignCert,
		"client_sign_key":  c.ClientSignKey,
		"client_enc_cert":  c.ClientEncCert,
		"client_enc_key":   c.ClientEncKey,

		"proxy_ssl_session_reuse_value": boolOnOff(p.SSLSessionCacheEnabled),
		"ssl_session_cache_value":       sslSessionCacheValue(p.SSLSessionCacheEnabled),
		"ssl_session_timeout_value":     DefaultSSLSessionTimeout,
		"ssl_session_tickets_value":     boolOnOff(p.SSLSessionTickets),
		"ssl_verify_client_value":       boolOnOff(p.SSLVerifyClient),
	}

	if p.Protocol == ProtocolPQC && p.SSLConfCommand == "" {
		v["ssl_conf_command"] = DefaultPQCConfCmd
	}
	return v
}

func sanitizeSSLProtocols(raw, protocol string) string {
	parts := strings.Fields(raw)
	kept := make([]string, 0, len(parts))
	for _, token := range parts {
		if strings.EqualFold(token, "NTLSv1.1") {
			continue
		}
		kept = append(kept, token)
	}
	if len(kept) == 0 {
		return defaultProtocolsFor(protocol)
	}
	return strings.Join(kept, " ")
}

func defaultProtocolsFor(protocol string) string {
	switch protocol {
	case ProtocolPQC:
		return DefaultPQCProtocols
	default:
		return DefaultSSLProtocols
	}
}

func boolOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func sslSessionCacheValue(enabled bool) string {
	if enabled {
		return DefaultSSLSessionCache
	}
	return "off"
}

// isClientTLSTraffic 客户端 + 走 TLS。
func isClientTLSTraffic(p *PlatformSSLProxy) bool {
	return p.Mode == ModeClient && p.BackendScheme == BackendSchemeHTTPS
}

// hasCAConfigured 是否配置了 CA（ca_domain 默认 safe，正常恒为 true）。
func hasCAConfigured(p *PlatformSSLProxy) bool {
	return p.CACertID != "" || p.CADomain != ""
}

// shouldEnable 指令启用条件。
//
// ★ 客户端 TLS 场景下（本版按用户要求）：
//   - proxy_ssl_server_name / proxy_ssl_name / proxy_ssl_trusted_certificate /
//     proxy_ssl_verify / proxy_ssl_verify_depth 全部启用
//   - 前提：hasCA（ca_domain 默认 safe 保证）
//   - proxy_ssl_name 值为空时输出 ''（用户已修改 nginx 适配）
func shouldEnable(cmd string, p *PlatformSSLProxy, resolvedSNI string) bool {
	clientTLS := isClientTLSTraffic(p)
	clientIntlTLS := clientTLS && (p.Protocol == ProtocolHTTP || p.Protocol == ProtocolPQC)
	clientGMSSL := clientTLS && p.Protocol == ProtocolGMSSL

	hasCA := hasCAConfigured(p)

	switch cmd {
	// ---------- 客户端 TLS ----------

	case "proxy_ssl_protocols", "proxy_ssl_ciphers":
		return clientIntlTLS

	// ★ SNI 相关：全部启用（用户要求）
	case "proxy_ssl_server_name", "proxy_ssl_name":
		return clientTLS && hasCA

	case "proxy_ssl_session_reuse":
		return clientIntlTLS && p.SSLSessionCacheEnabled

	case "proxy_ssl_certificate", "proxy_ssl_certificate_key":
		return clientIntlTLS && p.ClientCertID != ""

	case "proxy_enable_ntls":
		return clientGMSSL

	case "proxy_ssl_sign_certificate", "proxy_ssl_sign_certificate_key",
		"proxy_ssl_enc_certificate", "proxy_ssl_enc_certificate_key":
		return clientGMSSL && (p.ClientCertID != "" || p.ClientEncCertID != "")

	// ★ CA 加载：hasCA 就启用
	case "proxy_ssl_trusted_certificate":
		return clientTLS && hasCA

	// ★ verify：hasCA 就启用（与 proxy_ssl_name 成对）
	case "proxy_ssl_verify", "proxy_ssl_verify_depth":
		return clientTLS && hasCA

	case "proxy_ssl_conf_command":
		return clientIntlTLS && (p.SSLConfCommand != "" || p.Protocol == ProtocolPQC)

	// ---------- 服务端（保持不变） ----------

	case "ssl_certificate", "ssl_certificate_key":
		if p.Mode != ModeServer || !p.ListenSSL {
			return false
		}
		if p.Protocol == ProtocolHTTP || p.Protocol == ProtocolPQC {
			return true
		}
		return p.Protocol == ProtocolGMSSL && p.IntlCertID != ""

	case "enable_ntls",
		"ssl_sign_certificate", "ssl_sign_certificate_key",
		"ssl_enc_certificate", "ssl_enc_certificate_key":
		return p.Mode == ModeServer && p.ListenSSL && p.Protocol == ProtocolGMSSL

	case "ssl_protocols", "ssl_ciphers", "ssl_prefer_server_ciphers":
		return p.Mode == ModeServer && p.ListenSSL

	case "ssl_session_cache", "ssl_session_timeout":
		return p.Mode == ModeServer && p.ListenSSL && p.SSLSessionCacheEnabled

	case "ssl_session_tickets":
		return p.Mode == ModeServer && p.ListenSSL

	case "ssl_conf_command":
		return p.Mode == ModeServer && p.ListenSSL &&
			(p.SSLConfCommand != "" || p.Protocol == ProtocolPQC)

	case "ssl_verify_client":
		return p.Mode == ModeServer && p.ListenSSL && p.SSLVerifyClient

	case "ssl_client_certificate":
		return p.Mode == ModeServer && p.ListenSSL &&
			p.SSLVerifyClient &&
			(p.CADomain != "" || p.CACertID != "")
	}
	return false
}

func defStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func defaultIfEmpty(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
