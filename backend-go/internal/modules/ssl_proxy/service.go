// backend-go/internal/modules/ssl_proxy/service.go
package ssl_proxy

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	kv          *ConfigGenerator
	checker     *PortChecker
	pm          *ProcessManager
	gen         *NginxGenerator
	certCatalog *CertCatalog
}

func NewService(db *gorm.DB, resolver CertResolver) *Service {
	return &Service{
		db:          db,
		kv:          NewConfigGenerator(UserConfRoot),
		checker:     NewPortChecker(db),
		pm:          NewProcessManager(db, resolver),
		gen:         NewNginxGenerator(resolver),
		certCatalog: NewCertCatalog(db),
	}
}

func (s *Service) List(q *ProxyListQuery) ([]ProxyListItem, int64, error) {
	tx := s.db.Model(&PlatformSSLProxy{})

	if q.Mode != "" {
		if q.Mode != ModeClient && q.Mode != ModeServer {
			log.Warn().Str("mode", q.Mode).
				Msg("ssl_proxy: 列表查询 mode 参数非法，返回空")
			return []ProxyListItem{}, 0, nil
		}
		tx = tx.Where("mode = ?", q.Mode)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []PlatformSSLProxy
	if err := tx.Order("id DESC").
		Offset((q.Page - 1) * q.Size).Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]ProxyListItem, 0, len(rows))
	for i := range rows {
		items = append(items, toListItem(&rows[i]))
	}
	return items, total, nil
}

func (s *Service) Get(serviceID string) (*PlatformSSLProxy, error) {
	var p PlatformSSLProxy
	if err := s.db.Where("service_id = ?", serviceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newErr(ErrNotFound, "实例不存在", serviceID)
		}
		return nil, err
	}
	return &p, nil
}

func (s *Service) Create(req *ProxyCreateReq) (*PlatformSSLProxy, error) {
	if err := validateAddress("后端地址", req.BackendIP); err != nil {
		return nil, err
	}
	if err := validateAddress("监听地址", req.ListenIP); err != nil {
		return nil, err
	}

	req.ServiceName = strings.TrimSpace(req.ServiceName)

	log.Info().
		Str("mode", req.Mode).
		Str("service_name", req.ServiceName).
		Str("protocol", req.Protocol).
		Str("backend_scheme", req.BackendScheme).
		Msg("ssl_proxy: create 开始")

	if err := s.checkServiceNameUnique(req.ServiceName, req.Mode, ""); err != nil {
		return nil, err
	}

	pc := s.checker.Check(&PortCheckReq{
		ListenIP:   req.ListenIP,
		ListenPort: req.ListenPort,
	})
	if !pc.Available {
		return nil, newErr(ErrPortConflict, pc.Message, pc.OccupiedBy)
	}

	applyCreateDefaults(req)

	if err := validateSchemeProtocol(req.Mode, req.Protocol, req.BackendScheme); err != nil {
		return nil, err
	}

	// 服务端新建：强制 listen_ssl=true + 必须配证书
	if req.Mode == ModeServer {
		t := true
		req.ListenSSL = &t

		if err := validateServerCertRequired(
			req.Mode, req.Protocol, true,
			req.CertID, req.EncCertID, req.IntlCertID,
		); err != nil {
			return nil, err
		}
	}

	// SSL 详细配置校验（当前不阻断）
	if err := validateSSLSettings(
		req.Mode, req.Protocol,
		boolPtrVal(req.ListenSSL, false),
		req.CACertID,
		boolPtrVal(req.SSLVerifyClient, true),
	); err != nil {
		return nil, err
	}

	listenSSL := false
	if req.Mode == ModeServer {
		listenSSL = req.ListenSSL == nil || *req.ListenSSL
	}

	p := &PlatformSSLProxy{
		ServiceID:        generateServiceID(req.Mode),
		ServiceName:      req.ServiceName,
		Mode:             req.Mode,
		Protocol:         req.Protocol,
		ListenIP:         req.ListenIP,
		ListenPort:       req.ListenPort,
		BackendIP:        req.BackendIP,
		BackendPort:      req.BackendPort,
		BackendScheme:    req.BackendScheme,
		SSLServerName:    req.SSLServerName,
		ListenSSL:        listenSSL,
		CertID:           req.CertID,
		EncCertID:        req.EncCertID,
		CACertID:         req.CACertID,
		CADomain:         strings.TrimSpace(req.CADomain), // ★ 新增
		ClientCertID:     req.ClientCertID,
		ClientEncCertID:  req.ClientEncCertID,
		ClientKeyRef:     req.ClientKeyRef,
		IntlCertID:       req.IntlCertID,
		IntlClientCertID: req.IntlClientCertID,

		SSLProtocols:   req.SSLProtocols,
		SSLCiphers:     req.SSLCiphers,
		SSLConfCommand: req.SSLConfCommand,

		SSLCiphersMode:         normalizeCiphersMode(req.SSLCiphersMode),
		SSLSessionCacheEnabled: boolPtrVal(req.SSLSessionCacheEnabled, false),
		SSLVerifyClient:        boolPtrVal(req.SSLVerifyClient, req.Mode == ModeServer),
		SSLSessionTickets:      boolPtrVal(req.SSLSessionTickets, false),

		Status: StatusStopped,
		LogDir: fmt.Sprintf("%s/%s", LogRoot, ""),
	}
	if err := s.db.Create(p).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, newErr(ErrNameConflict, "服务名称已存在", req.ServiceName)
		}
		return nil, err
	}

	kvPath, _ := s.kv.WriteKV(p)
	_ = s.db.Model(p).Update("config_kv_path", kvPath).Error
	p.ConfigKVPath = kvPath
	return p, nil
}

func (s *Service) Update(serviceID string, req *ProxyUpdateReq) (*PlatformSSLProxy, error) {
	p, err := s.Get(serviceID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("service_id", serviceID).
		Str("mode", p.Mode).
		Str("old_protocol", p.Protocol).
		Str("new_protocol", req.Protocol).
		Str("new_backend_scheme", req.BackendScheme).
		Str("ca_domain", req.CADomain).
		Msg("ssl_proxy: update 开始")

	if err := validateAddress("后端地址", req.BackendIP); err != nil {
		return nil, err
	}
	if err := validateAddress("监听地址", req.ListenIP); err != nil {
		return nil, err
	}

	pc := s.checker.Check(&PortCheckReq{
		ServiceID:  serviceID,
		ListenIP:   req.ListenIP,
		ListenPort: req.ListenPort,
	})
	if !pc.Available {
		return nil, newErr(ErrPortConflict, pc.Message, pc.OccupiedBy)
	}

	newProtocol := p.Protocol
	if req.Protocol != "" {
		newProtocol = req.Protocol
	}

	applyUpdateDefaults(req, p.Mode, newProtocol, p)

	if err := validateSchemeProtocol(p.Mode, newProtocol, req.BackendScheme); err != nil {
		log.Warn().
			Str("service_id", serviceID).
			Str("mode", p.Mode).
			Str("protocol", newProtocol).
			Str("backend_scheme", req.BackendScheme).
			Err(err).
			Msg("ssl_proxy: 协议/scheme 一致性校验失败")
		return nil, err
	}

	finalListenSSL := p.ListenSSL
	if p.Mode == ModeServer {
		if req.ListenSSL != nil {
			finalListenSSL = *req.ListenSSL
		}
	} else {
		finalListenSSL = false
	}

	if err := validateServerCertRequired(
		p.Mode, newProtocol, finalListenSSL,
		req.CertID, req.EncCertID, req.IntlCertID,
	); err != nil {
		return nil, err
	}

	finalVerifyClient := p.SSLVerifyClient
	if req.SSLVerifyClient != nil {
		finalVerifyClient = *req.SSLVerifyClient
	}
	if err := validateSSLSettings(
		p.Mode, newProtocol, finalListenSSL,
		req.CACertID, finalVerifyClient,
	); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"protocol":            newProtocol,
		"listen_ip":           req.ListenIP,
		"listen_port":         req.ListenPort,
		"backend_ip":          req.BackendIP,
		"backend_port":        req.BackendPort,
		"backend_scheme":      req.BackendScheme,
		"ssl_server_name":     req.SSLServerName,
		"cert_id":             req.CertID,
		"enc_cert_id":         req.EncCertID,
		"ca_cert_id":          req.CACertID,
		"ca_domain":           strings.TrimSpace(req.CADomain), // ★ 新增
		"client_cert_id":      req.ClientCertID,
		"client_enc_cert_id":  req.ClientEncCertID,
		"client_key_ref":      req.ClientKeyRef,
		"intl_cert_id":        req.IntlCertID,
		"intl_client_cert_id": req.IntlClientCertID,

		"ssl_protocols":    req.SSLProtocols,
		"ssl_ciphers":      req.SSLCiphers,
		"ssl_conf_command": req.SSLConfCommand,

		"ssl_ciphers_mode": normalizeCiphersMode(req.SSLCiphersMode),
	}

	if p.Mode == ModeServer {
		updates["listen_ssl"] = finalListenSSL
		updates["ssl_verify_client"] = finalVerifyClient
	} else {
		updates["listen_ssl"] = false
		updates["ssl_verify_client"] = false
	}

	if req.SSLSessionCacheEnabled != nil {
		updates["ssl_session_cache_enabled"] = *req.SSLSessionCacheEnabled
	}
	if req.SSLSessionTickets != nil && p.Mode == ModeServer {
		updates["ssl_session_tickets"] = *req.SSLSessionTickets
	}

	if err := s.db.Model(&PlatformSSLProxy{}).
		Where("service_id = ?", serviceID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	p, err = s.Get(serviceID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("service_id", serviceID).
		Str("saved_protocol", p.Protocol).
		Str("saved_backend_scheme", p.BackendScheme).
		Str("saved_ca_domain", p.CADomain).
		Bool("saved_listen_ssl", p.ListenSSL).
		Msg("ssl_proxy: update 完成")

	kvPath, _ := s.kv.WriteKV(p)
	_ = s.db.Model(p).Update("config_kv_path", kvPath).Error
	p.ConfigKVPath = kvPath
	return p, nil
}

func (s *Service) Delete(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	if p.IsRunning() {
		return newErr(ErrRunningConflict, "实例运行中，禁止删除", serviceID)
	}
	if err := s.db.Delete(&PlatformSSLProxy{}, "service_id = ?", serviceID).Error; err != nil {
		return err
	}
	_ = s.kv.RemoveKV(p)
	return nil
}

func (s *Service) Start(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	return s.pm.Start(p)
}

func (s *Service) Stop(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	return s.pm.Stop(p)
}

func (s *Service) Restart(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	return s.pm.Restart(p)
}

func (s *Service) GetConfig(serviceID string) (kv, nginx string, err error) {
	p, err := s.Get(serviceID)
	if err != nil {
		return "", "", err
	}
	if p.ConfigKVPath != "" {
		b, e := os.ReadFile(p.ConfigKVPath)
		if e == nil {
			kv = string(b)
		}
	}
	if p.ConfigNginxPath != "" {
		b, e := os.ReadFile(p.ConfigNginxPath)
		if e == nil {
			nginx = string(b)
		}
	} else {
		conf, _, e := s.gen.Render(p)
		if e == nil {
			b, _ := os.ReadFile(conf)
			nginx = string(b)
		}
	}
	return kv, nginx, nil
}

func (s *Service) ListAvailableCerts(purpose, protocol string) ([]CertBrief, error) {
	return s.certCatalog.List(purpose, protocol)
}

// ---------- 内部工具 ----------

func (s *Service) checkServiceNameUnique(name, mode, excludeServiceID string) error {
	if name == "" {
		return newErr(ErrInvalidParam, "服务名称不能为空", "")
	}
	if mode != ModeClient && mode != ModeServer {
		return newErr(ErrInvalidParam, "非法 mode", mode)
	}
	q := s.db.Model(&PlatformSSLProxy{}).
		Where("service_name = ? AND mode = ?", name, mode)
	if excludeServiceID != "" {
		q = q.Where("service_id <> ?", excludeServiceID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return newErr(ErrNameConflict, "服务名称已存在", name)
	}
	return nil
}

func validateAddress(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return newErr(ErrInvalidParam, field+"不能为空", "")
	}
	return nil
}

func validateSchemeProtocol(mode, protocol, scheme string) error {
	if mode != ModeClient {
		return nil
	}
	if protocol == ProtocolGMSSL && scheme != BackendSchemeHTTPS {
		return newErr(ErrInvalidParam,
			"客户端国密协议必须开启后端 SSL（backend_scheme=https）",
			fmt.Sprintf("protocol=%s, backend_scheme=%s", protocol, scheme))
	}
	return nil
}

func validateServerCertRequired(
	mode, protocol string, listenSSL bool,
	certID, encCertID, intlCertID string,
) error {
	if mode != ModeServer {
		return nil
	}
	if !listenSSL {
		return nil
	}

	switch protocol {
	case ProtocolHTTP, ProtocolPQC:
		if certID == "" && intlCertID == "" {
			return newErr(ErrInvalidParam,
				"开启 SSL 通信时必须选择服务器证书",
				"cert_id 必填")
		}
	case ProtocolGMSSL:
		if certID == "" {
			return newErr(ErrInvalidParam,
				"开启 SSL 通信时必须选择国密签名证书",
				"cert_id 必填")
		}
		if encCertID == "" {
			return newErr(ErrInvalidParam,
				"开启 SSL 通信时必须选择国密加密证书",
				"enc_cert_id 必填")
		}
	}
	return nil
}

// validateSSLSettings SSL 详细配置校验。
//
// ★ 当前策略：不阻断。
//   - 前端默认开启双向认证；
//   - 用户可能先保存再补 CA，此处不因 CA 空而拦截；
//   - 无 CA 时渲染 ssl_verify_client on; 但无 ssl_client_certificate，
//     nginx -t 会失败，实例启动失败 —— 由用户补齐 CA 后自动恢复。
func validateSSLSettings(
	mode, protocol string, listenSSL bool,
	caCertID string, verifyClient bool,
) error {
	_ = mode
	_ = protocol
	_ = listenSSL
	_ = caCertID
	_ = verifyClient
	return nil
}

func boolPtrVal(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func normalizeCiphersMode(m string) string {
	switch m {
	case SSLCiphersModeDefault, SSLCiphersModeCustom, SSLCiphersModeInput:
		return m
	default:
		return SSLCiphersModeDefault
	}
}

func applyCreateDefaults(req *ProxyCreateReq) {
	if req.BackendScheme == "" {
		if req.Mode == ModeClient {
			req.BackendScheme = BackendSchemeHTTPS
		} else {
			req.BackendScheme = BackendSchemeHTTP
		}
	}
	if req.Mode == ModeClient && req.Protocol == ProtocolGMSSL {
		req.BackendScheme = BackendSchemeHTTPS
	}
	if req.Mode == ModeServer {
		if req.ListenSSL == nil {
			t := true
			req.ListenSSL = &t
		}
	} else {
		req.ListenSSL = nil
	}
	if req.SSLProtocols == "" {
		req.SSLProtocols = defaultProtocolsFor(req.Protocol)
	}
	if req.SSLCiphers == "" {
		req.SSLCiphers = DefaultSSLCiphers
	}
	if req.SSLConfCommand == "" && req.Protocol == ProtocolPQC {
		req.SSLConfCommand = DefaultPQCConfCmd
	}

	if req.SSLCiphersMode == "" {
		req.SSLCiphersMode = SSLCiphersModeDefault
	}
	if req.SSLSessionCacheEnabled == nil {
		f := false
		req.SSLSessionCacheEnabled = &f
	}
	if req.Mode == ModeServer {
		if req.SSLVerifyClient == nil {
			t := true
			req.SSLVerifyClient = &t
		}
		if req.SSLSessionTickets == nil {
			f := false
			req.SSLSessionTickets = &f
		}
	} else {
		req.SSLVerifyClient = nil
		req.SSLSessionTickets = nil
	}

	// ★ ca_domain 只在详情页配置，新建时不设默认值（保持空字符串）
}

func applyUpdateDefaults(req *ProxyUpdateReq, mode, protocol string, cur *PlatformSSLProxy) {
	if req.BackendScheme == "" {
		if mode == ModeClient {
			req.BackendScheme = BackendSchemeHTTPS
		} else {
			req.BackendScheme = BackendSchemeHTTP
		}
	}
	if mode != ModeServer {
		req.ListenSSL = nil
	}
	if req.SSLProtocols == "" {
		req.SSLProtocols = defaultProtocolsFor(protocol)
	}
	if req.SSLCiphers == "" {
		req.SSLCiphers = DefaultSSLCiphers
	}
	if req.SSLConfCommand == "" && protocol == ProtocolPQC {
		req.SSLConfCommand = DefaultPQCConfCmd
	}
	if req.SSLCiphersMode == "" {
		req.SSLCiphersMode = cur.SSLCiphersMode
		if req.SSLCiphersMode == "" {
			req.SSLCiphersMode = SSLCiphersModeDefault
		}
	}
	// ★ ca_domain：用户显式提交空串即"清空"，不在这里做兜底
	//   如果前端未传该字段（缺省 nil 语义），也不会被覆盖，因为
	//   Update 里始终以 req.CADomain 为准。
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "duplicate key")
}

func toListItem(p *PlatformSSLProxy) ProxyListItem {
	return ProxyListItem{
		ID:            p.ID,
		ServiceID:     p.ServiceID,
		ServiceName:   p.ServiceName,
		Mode:          p.Mode,
		Protocol:      p.Protocol,
		ListenAddress: fmt.Sprintf("%s:%d", p.ListenIP, p.ListenPort),
		BackendAddr:   fmt.Sprintf("%s://%s:%d", p.BackendScheme, p.BackendIP, p.BackendPort),
		CertRef:       firstNonEmpty(p.CertID, p.ClientCertID),
		Status:        p.Status,
		PID:           p.PID,
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),

		CertID:           p.CertID,
		EncCertID:        p.EncCertID,
		ClientCertID:     p.ClientCertID,
		ClientEncCertID:  p.ClientEncCertID,
		IntlCertID:       p.IntlCertID,
		IntlClientCertID: p.IntlClientCertID,
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func generateServiceID(mode string) string {
	return fmt.Sprintf("proxy-%s-%d", mode, time.Now().UnixNano())
}
