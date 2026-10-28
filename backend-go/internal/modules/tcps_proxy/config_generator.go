// backend-go/internal/modules/tcps_proxy/config_generator.go
package tcps_proxy
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)
// ConfigGenerator KV 配置生成器。
type ConfigGenerator struct {
	baseDir string
}
func NewConfigGenerator(baseDir string) *ConfigGenerator {
	if baseDir == "" {
		baseDir = UserTCPSConfRoot()
	}
	return &ConfigGenerator{baseDir: baseDir}
}
// UserTCPSConfRoot 返回 KV 配置的根目录。
func UserTCPSConfRoot() string {
	return filepath.Join(DeployRoot(), "user_data", "conf", "core")
}
func (g *ConfigGenerator) WriteKV(p *PlatformTCPSProxy) (string, error) {
	dir := filepath.Join(g.baseDir, p.ModeDir())
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", newErr(ErrConfigGenerate, "创建 KV 目录失败", err.Error())
	}
	path := filepath.Join(dir, p.ServiceID+".kv")
	if err := os.WriteFile(path, []byte(renderKV(p)), 0640); err != nil {
		return "", newErr(ErrConfigGenerate, "写入 KV 文件失败", err.Error())
	}
	return path, nil
}
func (g *ConfigGenerator) RemoveKV(p *PlatformTCPSProxy) error {
	path := filepath.Join(g.baseDir, p.ModeDir(), p.ServiceID+".kv")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
// ModeDir client → tcps_client，server → tcps_server。
func (p *PlatformTCPSProxy) ModeDir() string {
	if p.Mode == ModeClient {
		return "tcps_client"
	}
	return "tcps_server"
}
func renderKV(p *PlatformTCPSProxy) string {
	kv := map[string]string{
		"service_id":          p.ServiceID,
		"service_name":        p.ServiceName,
		"mode":                p.Mode,
		"protocol":            p.Protocol,
		"listen_ip":           p.ListenIP,
		"listen_port":         fmt.Sprintf("%d", p.ListenPort),
		"listen_ssl":          boolStr(p.IsListenSSL()),
		"backend_ip":          p.BackendIP,
		"backend_port":        fmt.Sprintf("%d", p.BackendPort),
		"backend_tls":         boolStr(p.IsBackendTLS()),
		"ssl_server_name":     p.SSLServerName,
		"cert_id":             p.CertID,
		"enc_cert_id":         p.EncCertID,
		"intl_cert_id":        p.IntlCertID,
		"ca_cert_id":          p.CACertID,
		"ca_domain":           p.CADomain,
		"client_cert_id":      p.ClientCertID,
		"client_enc_cert_id":  p.ClientEncCertID,
		"intl_client_cert_id": p.IntlClientCertID,
		"client_key_ref":      p.ClientKeyRef,
		"ssl_protocols":       p.SSLProtocols,
		"ssl_ciphers":         p.SSLCiphers,
		"ssl_conf_command":    p.SSLConfCommand,
		"ssl_ciphers_mode":          p.SSLCiphersMode,
		"ssl_session_cache_enabled": boolStr(p.SSLSessionCacheEnabled),
		"ssl_verify_client":         boolStr(p.SSLVerifyClient),
		"ssl_session_tickets":       boolStr(p.SSLSessionTickets),
		"status":       p.Status,
		"pid":          fmt.Sprintf("%d", p.PID),
		"config_nginx": p.ConfigNginxPath,
		"log_dir":      p.LogDir,
		"updated_at":   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "# TCPS Proxy KV — %s\n", p.ServiceID)
	fmt.Fprintf(&b, "# 修改后请在界面点击\"保存\"生效\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, kv[k])
	}
	return b.String()
}
func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
