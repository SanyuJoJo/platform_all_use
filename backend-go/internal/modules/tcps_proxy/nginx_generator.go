// backend-go/internal/modules/tcps_proxy/nginx_generator.go
package tcps_proxy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// NginxGenerator nginx stream 配置生成器。
type NginxGenerator struct {
	TemplateDir  string
	CertResolver CertResolver
	ConfRoot     string
	PidRoot      string
	TmpRoot      string
	LogRoot      string
}

// NewNginxGenerator 构造生成器。
func NewNginxGenerator(resolver CertResolver) *NginxGenerator {
	return &NginxGenerator{
		TemplateDir:  TemplateDir(),
		CertResolver: resolver,
		ConfRoot:     ConfRoot(),
		PidRoot:      PidRoot(),
		TmpRoot:      TmpRoot(),
		LogRoot:      LogRoot(),
	}
}

// Render 渲染 nginx.conf。
//
// ★ 本版语义（与 HTTP 同步）：
//   - 客户端 TLS 场景下，验证五件套全部开启：
//     proxy_ssl_server_name / proxy_ssl_name / proxy_ssl_trusted_certificate /
//     proxy_ssl_verify / proxy_ssl_verify_depth
//   - proxy_ssl_name 值：用户填了 SSL Server Name → 用它；没填 → 输出 ''
//   - 不做自动探测、不读本地证书 CN、不用 backend_ip 兜底
func (g *NginxGenerator) Render(p *PlatformTCPSProxy) (confPath, pidPath string, err error) {
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

	// 直接用用户填的 SSLServerName，不做兜底
	resolvedSNI := strings.TrimSpace(p.SSLServerName)

	vars := buildVarsTCPS(p, certs, confDir, pidDir, tmpDir, logDir, pidPath, resolvedSNI)

	tmplFile := filepath.Join(g.TemplateDir, templateFor(p.Mode))
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

// templateFor 按 mode 返回模板文件名。
func templateFor(mode string) string {
	if mode == ModeClient {
		return "tcps_client.conf.tmpl"
	}
	return "tcps_server.conf.tmpl"
}

func (g *NginxGenerator) renderLines(
	raw string, p *PlatformTCPSProxy, vars map[string]string, resolvedSNI string,
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
			if shouldEnableTCPS(cmdName, p, resolvedSNI) {
				out = append(out, indent+rendered)
			} else {
				out = append(out, indent+"# "+rendered)
			}
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

// buildVarsTCPS 构建模板变量。
//
// ★ stream 层没有 server_name 指令，不需要 server_name_value。
//   proxy_ssl_name 的变量：
//   - proxy_ssl_name_expr：空值时输出 ''，非空时输出用户值
//   - ssl_server_name    ：保留原始值（调试用）
func buildVarsTCPS(
	p *PlatformTCPSProxy,
	c *ResolvedCertPaths,
	confDir, pidDir, tmpDir, logDir, pidPath string,
	resolvedSNI string,
) map[string]string {
	listenSuffix := ""
	if p.IsListenSSL() {
		listenSuffix = " ssl"
	}

	sslProtocols := sanitizeSSLProtocols(p.SSLProtocols, p.Protocol)

	// 空值输出 ''
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
		// ★ 两个变量并存：原始值 + 空值用 '' 的表达式
		"ssl_server_name":     resolvedSNI,
		"proxy_ssl_name_expr": proxySSLNameExpr,
		"ssl_protocols":       sslProtocols,
		"ssl_ciphers":         defaultIfEmpty(p.SSLCiphers, DefaultSSLCiphers),
		"ssl_conf_command":    p.SSLConfCommand,

		"pid_path": pidPath,
		"conf_dir": confDir,
		"pid_dir":  pidDir,
		"tmp_dir":  tmpDir,
		"log_dir":  logDir,

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

		"ssl_session_cache_value":   sslSessionCacheValue(p.SSLSessionCacheEnabled),
		"ssl_session_timeout_value": DefaultSSLSessionTimeout,
		"ssl_session_tickets_value": boolOnOff(p.SSLSessionTickets),
		"ssl_verify_client_value":   boolOnOff(p.SSLVerifyClient),
		"proxy_ssl_on_value":        boolOnOff(p.IsBackendTLS()),
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
		if protocol == ProtocolPQC {
			return DefaultPQCProtocols
		}
		return DefaultSSLProtocols
	}
	return strings.Join(kept, " ")
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

// shouldEnableTCPS 指令启用条件。
//
// ★ 客户端 TLS 场景下（与 HTTP 语义一致）：
//   - SNI / proxy_ssl_name / verify 全部按 hasCA 启用（国密/国际一视同仁）
//   - 前提：hasCA = (ca_domain 非空 || ca_cert_id 非空)
//     ca_domain 由 Service 层默认 "safe"，所以正常恒为 true
func shouldEnableTCPS(cmd string, p *PlatformTCPSProxy, resolvedSNI string) bool {
	clientSSL := p.Mode == ModeClient && p.IsBackendTLS()
	clientIntlSSL := clientSSL && (p.Protocol == ProtocolTLS || p.Protocol == ProtocolPQC)
	clientGMSSL := clientSSL && p.Protocol == ProtocolGMSSL

	hasCA := p.CACertID != "" || p.CADomain != ""

	switch cmd {
	// =====================================================================
	// 客户端模式
	// =====================================================================

	case "proxy_ssl":
		return clientSSL

	case "proxy_ssl_protocols", "proxy_ssl_ciphers":
		return clientIntlSSL

	// ★ SNI 相关：clientSSL + hasCA 就启用（国密/国际统一）
	case "proxy_ssl_server_name", "proxy_ssl_name":
		return clientSSL && hasCA

	case "proxy_ssl_certificate", "proxy_ssl_certificate_key":
		return clientIntlSSL && p.ClientCertID != ""

	case "proxy_enable_ntls":
		return clientGMSSL

	case "proxy_ssl_sign_certificate", "proxy_ssl_sign_certificate_key",
		"proxy_ssl_enc_certificate", "proxy_ssl_enc_certificate_key":
		return clientGMSSL && (p.ClientCertID != "" || p.ClientEncCertID != "")

	// ★ CA 加载：hasCA 就启用
	case "proxy_ssl_trusted_certificate":
		return clientSSL && hasCA

	// ★ verify：hasCA 就启用（与 proxy_ssl_name 成对）
	case "proxy_ssl_verify", "proxy_ssl_verify_depth":
		return clientSSL && hasCA

	case "proxy_ssl_conf_command":
		return clientIntlSSL && (p.SSLConfCommand != "" || p.Protocol == ProtocolPQC)

	// =====================================================================
	// 服务端模式（保持不变）
	// =====================================================================

	case "ssl_certificate", "ssl_certificate_key":
		if !p.IsListenSSL() {
			return false
		}
		if p.Protocol == ProtocolTLS || p.Protocol == ProtocolPQC {
			return true
		}
		return p.Protocol == ProtocolGMSSL && p.IntlCertID != ""

	case "enable_ntls",
		"ssl_sign_certificate", "ssl_sign_certificate_key",
		"ssl_enc_certificate", "ssl_enc_certificate_key":
		return p.IsListenSSL() && p.Protocol == ProtocolGMSSL

	case "ssl_protocols", "ssl_ciphers", "ssl_prefer_server_ciphers":
		return p.IsListenSSL()

	case "ssl_session_cache", "ssl_session_timeout":
		return p.IsListenSSL() && p.SSLSessionCacheEnabled

	case "ssl_session_tickets":
		return p.IsListenSSL()

	case "ssl_conf_command":
		return p.IsListenSSL() &&
			(p.SSLConfCommand != "" || p.Protocol == ProtocolPQC)

	case "ssl_verify_client":
		return p.IsListenSSL() && p.SSLVerifyClient

	case "ssl_client_certificate":
		return p.IsListenSSL() && p.SSLVerifyClient &&
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
