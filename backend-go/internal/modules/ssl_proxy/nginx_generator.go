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

	vars := buildVars(p, certs, confDir, pidDir, tmpDir, logDir, pidPath)

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

	rendered := g.renderLines(string(raw), p, vars)

	if err := os.WriteFile(confPath, []byte(rendered), 0640); err != nil {
		return "", "", newErr(ErrConfigGenerate, "写入 nginx.conf 失败", err.Error())
	}
	return confPath, pidPath, nil
}

func (g *NginxGenerator) renderLines(raw string, p *PlatformSSLProxy, vars map[string]string) string {
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")

		if strings.HasPrefix(trimmed, "#@") {
			indent := line[:len(line)-len(trimmed)]
			cmdLine := strings.TrimSpace(strings.TrimPrefix(trimmed, "#@"))
			cmdName := firstToken(cmdLine)

			rendered := renderVars(cmdLine, vars)
			if shouldEnable(cmdName, p) {
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

func buildVars(
	p *PlatformSSLProxy,
	c *ResolvedCertPaths,
	confDir, pidDir, tmpDir, logDir, pidPath string,
) map[string]string {
	listenSuffix := ""
	if p.Mode == ModeServer && p.ListenSSL {
		listenSuffix = " ssl"
	}

	// ★ ssl_protocols 经 sanitize 处理：过滤 NTLSv1.1，空则回落默认
	sslProtocols := sanitizeSSLProtocols(p.SSLProtocols, p.Protocol)

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
		"ssl_server_name":   defStr(p.SSLServerName, "_"),
		"ssl_protocols":     sslProtocols,
		"ssl_ciphers":       defaultIfEmpty(p.SSLCiphers, DefaultSSLCiphers),
		"ssl_conf_command":  p.SSLConfCommand,

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

	// PQC 场景补默认 ssl_conf_command（ssl_protocols 已由 sanitize 处理）
	if p.Protocol == ProtocolPQC && p.SSLConfCommand == "" {
		v["ssl_conf_command"] = DefaultPQCConfCmd
	}
	return v
}

// ★ sanitizeSSLProtocols 清理 ssl_protocols，保证写入 nginx.conf 的值合法。
//
// 背景：
//   - nginx 的 ssl_protocols 只接受标准 TLS 值：
//     TLSv1 / TLSv1.1 / TLSv1.2 / TLSv1.3 / SSLv2 / SSLv3；
//   - 国密 NTLS 通道由 `enable_ntls on;` 开启，与 ssl_protocols 无关；
//   - 前端"国密"选项 value 仍为 NTLSv1.1（显示名可能为 GMTLSv1.1），
//     提交后后端必须过滤掉，否则 nginx -t 报 invalid value "NTLSv1.1"。
//
// 规则：
//   - 按空白分词；
//   - 丢弃 NTLSv1.1（大小写不敏感）；
//   - 保留其他值（用户显式选择的国际协议）；
//   - 过滤后为空 → 按 protocol 回落默认：
//     pqc   → TLSv1.3
//     其他  → TLSv1.2 TLSv1.3
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

// defaultProtocolsFor 按 protocol 返回 ssl_protocols 的默认值。
//
// ★ GMSSL 不再返回 NTLSv1.1：
//   国密通道由 enable_ntls 控制，ssl_protocols 只描述国际 TLS 通道。
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

// shouldEnable 指令 → 启用条件映射。
func shouldEnable(cmd string, p *PlatformSSLProxy) bool {
	switch cmd {
	// ---------- nchttp 客户端 ----------

	case "proxy_ssl_protocols", "proxy_ssl_ciphers",
		"proxy_ssl_server_name", "proxy_ssl_name":
		return p.Protocol == ProtocolHTTP || p.Protocol == ProtocolPQC

	case "proxy_ssl_session_reuse":
		return (p.Protocol == ProtocolHTTP || p.Protocol == ProtocolPQC) &&
			p.SSLSessionCacheEnabled

	case "proxy_ssl_certificate", "proxy_ssl_certificate_key":
		return (p.Protocol == ProtocolHTTP || p.Protocol == ProtocolPQC) && p.ClientCertID != ""

	case "proxy_enable_ntls":
		return p.Protocol == ProtocolGMSSL && p.BackendScheme == BackendSchemeHTTPS

	case "proxy_ssl_sign_certificate", "proxy_ssl_sign_certificate_key",
		"proxy_ssl_enc_certificate", "proxy_ssl_enc_certificate_key":
		return p.Protocol == ProtocolGMSSL &&
			p.BackendScheme == BackendSchemeHTTPS &&
			p.ClientCertID != ""

	case "proxy_ssl_trusted_certificate", "proxy_ssl_verify", "proxy_ssl_verify_depth":
		return p.CACertID != ""

	case "proxy_ssl_conf_command":
		return p.SSLConfCommand != "" || p.Protocol == ProtocolPQC

	// ---------- nshttp 服务端 ----------

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
