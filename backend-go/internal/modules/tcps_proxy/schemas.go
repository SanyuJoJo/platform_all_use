// backend-go/internal/modules/tcps_proxy/schemas.go
package tcps_proxy
type ProxyCreateReq struct {
	ServiceName   string `json:"service_name" binding:"required,min=1,max=255"`
	Mode          string `json:"mode" binding:"required,oneof=client server"`
	Protocol      string `json:"protocol" binding:"required,oneof=tcp tls gmssl pqc"`
	ListenIP      string `json:"listen_ip" binding:"required"`
	ListenPort    int    `json:"listen_port" binding:"required,min=1,max=65535"`
	BackendIP     string `json:"backend_ip" binding:"required"`
	BackendPort   int    `json:"backend_port" binding:"required,min=1,max=65535"`
	BackendTLS    *bool  `json:"backend_tls,omitempty"`
	SSLServerName string `json:"ssl_server_name"`
	ListenSSL     *bool  `json:"listen_ssl,omitempty"`
	CertID           string `json:"cert_id"`
	EncCertID        string `json:"enc_cert_id"`
	IntlCertID       string `json:"intl_cert_id"`
	CACertID         string `json:"ca_cert_id"`
	CADomain         string `json:"ca_domain"`
	ClientCertID     string `json:"client_cert_id"`
	ClientEncCertID  string `json:"client_enc_cert_id"`
	IntlClientCertID string `json:"intl_client_cert_id"`
	ClientKeyRef     string `json:"client_key_ref"`
	SSLProtocols   string `json:"ssl_protocols"`
	SSLCiphers     string `json:"ssl_ciphers"`
	SSLConfCommand string `json:"ssl_conf_command"`
	SSLCiphersMode         string `json:"ssl_ciphers_mode"`
	SSLSessionCacheEnabled *bool  `json:"ssl_session_cache_enabled,omitempty"`
	SSLVerifyClient        *bool  `json:"ssl_verify_client,omitempty"`
	SSLSessionTickets      *bool  `json:"ssl_session_tickets,omitempty"`
}
type ProxyUpdateReq struct {
	Protocol      string `json:"protocol" binding:"omitempty,oneof=tcp tls gmssl pqc"`
	ListenIP      string `json:"listen_ip" binding:"required"`
	ListenPort    int    `json:"listen_port" binding:"required,min=1,max=65535"`
	BackendIP     string `json:"backend_ip" binding:"required"`
	BackendPort   int    `json:"backend_port" binding:"required,min=1,max=65535"`
	BackendTLS    *bool  `json:"backend_tls,omitempty"`
	SSLServerName string `json:"ssl_server_name"`
	ListenSSL     *bool  `json:"listen_ssl,omitempty"`
	CertID           string `json:"cert_id"`
	EncCertID        string `json:"enc_cert_id"`
	IntlCertID       string `json:"intl_cert_id"`
	CACertID         string `json:"ca_cert_id"`
	CADomain         string `json:"ca_domain"`
	ClientCertID     string `json:"client_cert_id"`
	ClientEncCertID  string `json:"client_enc_cert_id"`
	IntlClientCertID string `json:"intl_client_cert_id"`
	ClientKeyRef     string `json:"client_key_ref"`
	SSLProtocols   string `json:"ssl_protocols"`
	SSLCiphers     string `json:"ssl_ciphers"`
	SSLConfCommand string `json:"ssl_conf_command"`
	SSLCiphersMode         string `json:"ssl_ciphers_mode"`
	SSLSessionCacheEnabled *bool  `json:"ssl_session_cache_enabled,omitempty"`
	SSLVerifyClient        *bool  `json:"ssl_verify_client,omitempty"`
	SSLSessionTickets      *bool  `json:"ssl_session_tickets,omitempty"`
}
type ProxyListQuery struct {
	Mode   string `form:"mode"`
	Status string `form:"status"`
	Page   int    `form:"page,default=1"`
	Size   int    `form:"size,default=20"`
}
type ProxyListItem struct {
	ID            int64  `json:"id"`
	ServiceID     string `json:"service_id"`
	ServiceName   string `json:"service_name"`
	Mode          string `json:"mode"`
	Protocol      string `json:"protocol"`
	ListenAddress string `json:"listen_address"`
	BackendAddr   string `json:"backend_address"`
	CertRef       string `json:"cert_ref"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	UpdatedAt     string `json:"updated_at"`
	CertID           string `json:"cert_id,omitempty"`
	EncCertID        string `json:"enc_cert_id,omitempty"`
	ClientCertID     string `json:"client_cert_id,omitempty"`
	ClientEncCertID  string `json:"client_enc_cert_id,omitempty"`
	IntlCertID       string `json:"intl_cert_id,omitempty"`
	IntlClientCertID string `json:"intl_client_cert_id,omitempty"`
}
type PortCheckReq struct {
	ServiceID  string `json:"service_id"`
	ListenIP   string `json:"listen_ip" binding:"required"`
	ListenPort int    `json:"listen_port" binding:"required,min=1,max=65535"`
}
type PortCheckResult struct {
	Available     bool   `json:"available"`
	ConflictType  string `json:"conflict_type,omitempty"`
	OccupiedBy    string `json:"occupied_by,omitempty"`
	SuggestedPort int    `json:"suggested_port,omitempty"`
	Message       string `json:"message"`
}
type BatchReq struct {
	ServiceIDs []string `json:"service_ids" binding:"required,min=1"`
}
type BatchItemResult struct {
	ServiceID string  `json:"service_id"`
	Success   bool    `json:"success"`
	Code      ErrCode `json:"code"`
	Message   string  `json:"message"`
}
type BatchResp struct {
	Total     int               `json:"total"`
	Succeeded int               `json:"succeeded"`
	Failed    int               `json:"failed"`
	Results   []BatchItemResult `json:"results"`
}
type Resp struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}
type AvailableCertsQuery struct {
	Purpose  string `form:"purpose"  binding:"required,oneof=sign enc"`
	Protocol string `form:"protocol" binding:"required,oneof=tcp tls gmssl pqc"`
}
type CertBrief struct {
	CertID        string `json:"cert_id"`
	SubjectCN     string `json:"subject_cn"`
	Algorithm     string `json:"algorithm"`
	CertType      string `json:"cert_type"`
	NotAfter      string `json:"not_after"`
	HasPrivateKey bool   `json:"has_private_key"`
}
type AvailableCertsResp struct {
	Items []CertBrief `json:"items"`
}
