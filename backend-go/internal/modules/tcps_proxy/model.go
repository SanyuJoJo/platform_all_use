// backend-go/internal/modules/tcps_proxy/model.go
package tcps_proxy
import "time"
// PlatformTCPSProxy TCPS 代理实例。
type PlatformTCPSProxy struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ServiceID   string `gorm:"column:service_id;size:64;uniqueIndex;not null" json:"service_id"`
	ServiceName string `gorm:"column:service_name;size:255;not null;uniqueIndex:idx_tcps_proxy_mode_name,priority:2" json:"service_name"`
	Mode        string `gorm:"column:mode;size:16;not null;uniqueIndex:idx_tcps_proxy_mode_name,priority:1" json:"mode"`
	Protocol    string `gorm:"column:protocol;size:16;not null" json:"protocol"`
	ListenIP   string `gorm:"column:listen_ip;size:64;not null" json:"listen_ip"`
	ListenPort int    `gorm:"column:listen_port;not null" json:"listen_port"`
	BackendIP  string `gorm:"column:backend_ip;size:64;not null" json:"backend_ip"`
	BackendPort int   `gorm:"column:backend_port;not null" json:"backend_port"`
	BackendTLS  bool  `gorm:"column:backend_tls;not null;default:false" json:"backend_tls"`
	SSLServerName string `gorm:"column:ssl_server_name;size:255;default:''" json:"ssl_server_name"`
	ListenSSL     bool   `gorm:"column:listen_ssl;not null;default:false" json:"listen_ssl"`
	// ---- 证书字段 ----
	CertID           string `gorm:"column:cert_id;size:64;default:''" json:"cert_id"`
	EncCertID        string `gorm:"column:enc_cert_id;size:64;default:''" json:"enc_cert_id"`
	IntlCertID       string `gorm:"column:intl_cert_id;size:64;default:''" json:"intl_cert_id"`
	CACertID         string `gorm:"column:ca_cert_id;size:64;default:''" json:"ca_cert_id"`
	CADomain         string `gorm:"column:ca_domain;size:64;default:''" json:"ca_domain"`
	ClientCertID     string `gorm:"column:client_cert_id;size:64;default:''" json:"client_cert_id"`
	ClientEncCertID  string `gorm:"column:client_enc_cert_id;size:64;default:''" json:"client_enc_cert_id"`
	IntlClientCertID string `gorm:"column:intl_client_cert_id;size:64;default:''" json:"intl_client_cert_id"`
	ClientKeyRef     string `gorm:"column:client_key_ref;size:64;default:''" json:"client_key_ref"`
	// ---- SSL 参数 ----
	SSLProtocols   string `gorm:"column:ssl_protocols;size:128;default:''" json:"ssl_protocols"`
	SSLCiphers     string `gorm:"column:ssl_ciphers;size:1024;default:''" json:"ssl_ciphers"`
	SSLConfCommand string `gorm:"column:ssl_conf_command;size:512;default:''" json:"ssl_conf_command"`
	SSLCiphersMode         string `gorm:"column:ssl_ciphers_mode;size:16;not null;default:default" json:"ssl_ciphers_mode"`
	SSLSessionCacheEnabled bool   `gorm:"column:ssl_session_cache_enabled;not null;default:false" json:"ssl_session_cache_enabled"`
	SSLVerifyClient        bool   `gorm:"column:ssl_verify_client;not null;default:false" json:"ssl_verify_client"`
	SSLSessionTickets      bool   `gorm:"column:ssl_session_tickets;not null;default:false" json:"ssl_session_tickets"`
	Status          string    `gorm:"column:status;size:16;not null;default:STOPPED" json:"status"`
	PID             int       `gorm:"column:pid;not null;default:0" json:"pid"`
	ConfigKVPath    string    `gorm:"column:config_kv_path;size:255;default:''" json:"config_kv_path"`
	ConfigNginxPath string    `gorm:"column:config_nginx_path;size:255;default:''" json:"config_nginx_path"`
	LogDir          string    `gorm:"column:log_dir;size:255;default:''" json:"log_dir"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
func (PlatformTCPSProxy) TableName() string { return "platform_tcps_proxy" }
func (p *PlatformTCPSProxy) IsRunning() bool {
	return p.Status == StatusRunning || p.Status == StatusStarting
}
// IsListenSSL 服务端模式下是否 TLS 监听；客户端模式恒 false。
func (p *PlatformTCPSProxy) IsListenSSL() bool {
	if p.Mode != ModeServer {
		return false
	}
	return p.ListenSSL
}
// IsBackendTLS 客户端模式下是否 TLS 连接后端；服务端模式恒 false。
func (p *PlatformTCPSProxy) IsBackendTLS() bool {
	if p.Mode != ModeClient {
		return false
	}
	return p.BackendTLS
}
