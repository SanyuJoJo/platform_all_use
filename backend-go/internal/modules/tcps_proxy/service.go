// backend-go/internal/modules/tcps_proxy/service.go
package tcps_proxy

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Service TCPS 代理业务服务。
//
// 所有对外方法都以 service_id 为入参（除 List），
// 保证 handler 层不接触内部字段。
type Service struct {
	db          *gorm.DB
	kv          *ConfigGenerator
	checker     *PortChecker
	pm          *ProcessManager
	gen         *NginxGenerator
	certCatalog *CertCatalog
}

// NewService 构造 TCPS 代理服务。
func NewService(db *gorm.DB, resolver CertResolver) *Service {
	return &Service{
		db:          db,
		kv:          NewConfigGenerator(UserTCPSConfRoot()),
		checker:     NewPortChecker(db),
		pm:          NewProcessManager(db, resolver),
		gen:         NewNginxGenerator(resolver),
		certCatalog: NewCertCatalog(db),
	}
}

// =============================================================================
// 列表 / 详情
// =============================================================================

// List 查询 TCPS 代理列表。
func (s *Service) List(q *ProxyListQuery) ([]ProxyListItem, int64, error) {
	tx := s.db.Model(&PlatformTCPSProxy{})
	if q.Mode != "" {
		if q.Mode != ModeClient && q.Mode != ModeServer {
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
	var rows []PlatformTCPSProxy
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

// Get 按 service_id 查询详情。
func (s *Service) Get(serviceID string) (*PlatformTCPSProxy, error) {
	var p PlatformTCPSProxy
	if err := s.db.Where("service_id = ?", serviceID).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, newErr(ErrNotFound, "实例不存在", serviceID)
		}
		return nil, err
	}
	return &p, nil
}

// =============================================================================
// 新建 / 更新 / 删除
// =============================================================================

// Create 新建 TCPS 代理。
func (s *Service) Create(req *ProxyCreateReq) (*PlatformTCPSProxy, error) {
	if strings.TrimSpace(req.ListenIP) == "" {
		return nil, newErr(ErrInvalidParam, "监听地址不能为空", "")
	}
	if strings.TrimSpace(req.BackendIP) == "" {
		return nil, newErr(ErrInvalidParam, "后端地址不能为空", "")
	}
	req.ServiceName = strings.TrimSpace(req.ServiceName)

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

	// 服务端：非 tcp 强制 listen_ssl=true，并校验证书
	if req.Mode == ModeServer {
		if req.Protocol == ProtocolTCP {
			f := false
			req.ListenSSL = &f
		} else {
			t := true
			req.ListenSSL = &t
			if err := validateServerCertRequired(
				req.Protocol, true,
				req.CertID, req.EncCertID, req.IntlCertID,
			); err != nil {
				return nil, err
			}
		}
	} else {
		// 客户端：listen_ssl 恒 false
		f := false
		req.ListenSSL = &f
	}

	p := &PlatformTCPSProxy{
		ServiceID:        generateServiceID(req.Mode),
		ServiceName:      req.ServiceName,
		Mode:             req.Mode,
		Protocol:         req.Protocol,
		ListenIP:         req.ListenIP,
		ListenPort:       req.ListenPort,
		BackendIP:        req.BackendIP,
		BackendPort:      req.BackendPort,
		BackendTLS:       boolPtrVal(req.BackendTLS, req.Mode == ModeClient && ProtocolNeedsSSL(req.Protocol)),
		SSLServerName:    req.SSLServerName,
		ListenSSL:        boolPtrVal(req.ListenSSL, false),

		CertID:           req.CertID,
		EncCertID:        req.EncCertID,
		IntlCertID:       req.IntlCertID,
		CACertID:         req.CACertID,
		CADomain:         strings.TrimSpace(req.CADomain),
		ClientCertID:     req.ClientCertID,
		ClientEncCertID:  req.ClientEncCertID,
		IntlClientCertID: req.IntlClientCertID,
		ClientKeyRef:     req.ClientKeyRef,

		SSLProtocols:   req.SSLProtocols,
		SSLCiphers:     req.SSLCiphers,
		SSLConfCommand: req.SSLConfCommand,

		SSLCiphersMode:         normalizeCiphersMode(req.SSLCiphersMode),
		SSLSessionCacheEnabled: boolPtrVal(req.SSLSessionCacheEnabled, false),
		SSLVerifyClient:        boolPtrVal(req.SSLVerifyClient, req.Mode == ModeServer && req.Protocol != ProtocolTCP),
		SSLSessionTickets:      boolPtrVal(req.SSLSessionTickets, false),

		Status: StatusStopped,
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

// Update 编辑 TCPS 代理。
func (s *Service) Update(serviceID string, req *ProxyUpdateReq) (*PlatformTCPSProxy, error) {
	p, err := s.Get(serviceID)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.ListenIP) == "" || strings.TrimSpace(req.BackendIP) == "" {
		return nil, newErr(ErrInvalidParam, "地址不能为空", "")
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

	finalListenSSL := p.ListenSSL
	finalBackendTLS := p.BackendTLS
	if p.Mode == ModeServer {
		if newProtocol == ProtocolTCP {
			finalListenSSL = false
		} else {
			finalListenSSL = true
			if req.ListenSSL != nil {
				finalListenSSL = *req.ListenSSL
			}
			if err := validateServerCertRequired(
				newProtocol, finalListenSSL,
				req.CertID, req.EncCertID, req.IntlCertID,
			); err != nil {
				return nil, err
			}
		}
	} else {
		finalListenSSL = false
		if newProtocol == ProtocolTCP {
			finalBackendTLS = false
		} else {
			finalBackendTLS = true
			if req.BackendTLS != nil {
				finalBackendTLS = *req.BackendTLS
			}
		}
	}

	finalVerifyClient := p.SSLVerifyClient
	if req.SSLVerifyClient != nil {
		finalVerifyClient = *req.SSLVerifyClient
	}

	updates := map[string]interface{}{
		"protocol":            newProtocol,
		"listen_ip":           req.ListenIP,
		"listen_port":         req.ListenPort,
		"backend_ip":          req.BackendIP,
		"backend_port":        req.BackendPort,
		"backend_tls":         finalBackendTLS,
		"ssl_server_name":     req.SSLServerName,
		"listen_ssl":          finalListenSSL,
		"cert_id":             req.CertID,
		"enc_cert_id":         req.EncCertID,
		"intl_cert_id":        req.IntlCertID,
		"ca_cert_id":          req.CACertID,
		"ca_domain":           strings.TrimSpace(req.CADomain),
		"client_cert_id":      req.ClientCertID,
		"client_enc_cert_id":  req.ClientEncCertID,
		"intl_client_cert_id": req.IntlClientCertID,
		"client_key_ref":      req.ClientKeyRef,

		"ssl_protocols":     req.SSLProtocols,
		"ssl_ciphers":       req.SSLCiphers,
		"ssl_conf_command":  req.SSLConfCommand,

		"ssl_ciphers_mode":  normalizeCiphersMode(req.SSLCiphersMode),
		"ssl_verify_client": finalVerifyClient,
	}

	if req.SSLSessionCacheEnabled != nil {
		updates["ssl_session_cache_enabled"] = *req.SSLSessionCacheEnabled
	}
	if req.SSLSessionTickets != nil {
		updates["ssl_session_tickets"] = *req.SSLSessionTickets
	}

	if err := s.db.Model(&PlatformTCPSProxy{}).
		Where("service_id = ?", serviceID).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	p, err = s.Get(serviceID)
	if err != nil {
		return nil, err
	}

	kvPath, _ := s.kv.WriteKV(p)
	_ = s.db.Model(p).Update("config_kv_path", kvPath).Error
	p.ConfigKVPath = kvPath
	return p, nil
}

// Delete 删除 TCPS 代理。
func (s *Service) Delete(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	if p.IsRunning() {
		return newErr(ErrRunningConflict, "实例运行中，禁止删除", serviceID)
	}
	if err := s.db.Delete(&PlatformTCPSProxy{}, "service_id = ?", serviceID).Error; err != nil {
		return err
	}
	_ = s.kv.RemoveKV(p)
	return nil
}

// =============================================================================
// 生命周期
// =============================================================================

// Start 启动实例。
func (s *Service) Start(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	return s.pm.Start(p)
}

// Stop 停止实例。
func (s *Service) Stop(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	return s.pm.Stop(p)
}

// Restart 重启实例。
func (s *Service) Restart(serviceID string) error {
	p, err := s.Get(serviceID)
	if err != nil {
		return err
	}
	return s.pm.Restart(p)
}

// =============================================================================
// 配置 / 证书 / 端口
// =============================================================================

// GetConfig 返回 KV + nginx 配置内容。
func (s *Service) GetConfig(serviceID string) (kv, nginx string, err error) {
	p, err := s.Get(serviceID)
	if err != nil {
		return "", "", err
	}
	if p.ConfigKVPath != "" {
		if b, e := os.ReadFile(p.ConfigKVPath); e == nil {
			kv = string(b)
		}
	}
	if p.ConfigNginxPath != "" {
		if b, e := os.ReadFile(p.ConfigNginxPath); e == nil {
			nginx = string(b)
		}
	} else {
		conf, _, e := s.gen.Render(p)
		if e == nil {
			if b, _ := os.ReadFile(conf); b != nil {
				nginx = string(b)
			}
		}
	}
	return kv, nginx, nil
}

// ListAvailableCerts 查询可绑定证书。
func (s *Service) ListAvailableCerts(purpose, protocol string) ([]CertBrief, error) {
	return s.certCatalog.List(purpose, protocol)
}

// CheckPort 端口检测（对外暴露，供 handler 调用）。
func (s *Service) CheckPort(req *PortCheckReq) *PortCheckResult {
	return s.checker.Check(req)
}

// =============================================================================
// 内部工具
// =============================================================================

func (s *Service) checkServiceNameUnique(name, mode, excludeServiceID string) error {
	if name == "" {
		return newErr(ErrInvalidParam, "服务名称不能为空", "")
	}
	q := s.db.Model(&PlatformTCPSProxy{}).
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

func validateServerCertRequired(protocol string, listenSSL bool, certID, encCertID, intlCertID string) error {
	if !listenSSL {
		return nil
	}
	switch protocol {
	case ProtocolTLS, ProtocolPQC:
		if certID == "" && intlCertID == "" {
			return newErr(ErrInvalidParam,
				"开启 SSL 通信时必须选择服务器证书", "cert_id 必填")
		}
	case ProtocolGMSSL:
		if certID == "" {
			return newErr(ErrInvalidParam,
				"开启 SSL 通信时必须选择国密签名证书", "cert_id 必填")
		}
		if encCertID == "" {
			return newErr(ErrInvalidParam,
				"开启 SSL 通信时必须选择国密加密证书", "enc_cert_id 必填")
		}
	}
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
	// 协议非 tcp → 强制开启 TLS
	if req.Mode == ModeClient {
		if req.Protocol == ProtocolTCP {
			f := false
			req.BackendTLS = &f
		} else {
			t := true
			req.BackendTLS = &t
		}
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
			t := req.Protocol != ProtocolTCP
			req.SSLVerifyClient = &t
		}
		if req.SSLSessionTickets == nil {
			f := false
			req.SSLSessionTickets = &f
		}
	}

	// ★ 默认信任域 safe（与 HTTP 一致）
	if strings.TrimSpace(req.CADomain) == "" {
		req.CADomain = "safe"
	}
}

func applyUpdateDefaults(req *ProxyUpdateReq, mode, protocol string, cur *PlatformTCPSProxy) {
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

	// ★ 默认信任域 safe（与 HTTP 一致）
	if strings.TrimSpace(req.CADomain) == "" {
		req.CADomain = "safe"
	}
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

func toListItem(p *PlatformTCPSProxy) ProxyListItem {
	return ProxyListItem{
		ID:            p.ID,
		ServiceID:     p.ServiceID,
		ServiceName:   p.ServiceName,
		Mode:          p.Mode,
		Protocol:      p.Protocol,
		ListenAddress: fmt.Sprintf("%s:%d", p.ListenIP, p.ListenPort),
		BackendAddr:   fmt.Sprintf("%s:%d", p.BackendIP, p.BackendPort),
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
	return fmt.Sprintf("tcps-%s-%d", mode, time.Now().UnixNano())
}

// ensureLogInit 确保日志目录已创建（用于首次创建实例时）。
func ensureLogInit(logDir string) {
	if logDir == "" {
		return
	}
	_ = os.MkdirAll(logDir, 0750)
}

// 保留 log 引用，避免 future 使用时报 unused
var _ = log.Info
