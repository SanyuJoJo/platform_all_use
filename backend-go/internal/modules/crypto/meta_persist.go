package crypto

import (
	"time"

	"github.com/rs/zerolog/log"

	"backend-go/internal/models"
)

// persistMetadata 把 core 返回的元数据 + 请求参数落库到平台表。
//
// 说明：
//   - 仅处理"产生元数据"的操作（ca.create / ca.intermediate.create /
//     csr.create / cert.sign / dual_cert.create / crl.create / key.manage）；
//   - 失败仅记录日志，不影响主流程；
//   - 使用 Upsert 语义（ON CONFLICT）避免重复；
//   - ⚠️ 关键点：core 的响应 data 通常只含 ca_id/cert_path/key_ref/serial，
//     而 subject/algorithm/validity_days 等字段需要从**请求 params** 取。
func (s *Service) persistMetadata(op string, params, data map[string]interface{}) {
	if s.db == nil || data == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Str("op", op).Msg("元数据落库 panic")
		}
	}()

	switch op {
	case "ca.create", "ca.intermediate.create":
		s.persistCA(op, params, data)
	case "csr.create":
		s.persistCSR(params, data)
	case "cert.sign":
		s.persistCert(params, data)
	case "crl.create":
		s.persistCRL(params, data)
	case "key.manage":
		s.persistKey(params, data)
	}
}

// -----------------------------------------------------------------------------
// CA
// -----------------------------------------------------------------------------

func (s *Service) persistCA(op string, params, data map[string]interface{}) {
	caID := getString(data, "ca_id")
	if caID == "" {
		return
	}

	now := time.Now().UTC()
	ca := &models.CA{
		CAID:         caID,
		SubjectCN:    getNestedString(params, "subject", "CN"),
		Algorithm:    getString(params, "algorithm"),
		CertPath:     getString(data, "cert_path"),
		KeyRef:       getString(data, "key_ref"),
		ValidityDays: getInt(params, "validity_days"),
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 可选字段
	if v := getNestedString(params, "subject", "O"); v != "" {
		ca.SubjectO = &v
	}
	if v := getString(data, "chain_path"); v != "" {
		ca.ChainPath = &v
	}
	if v := getString(data, "serial"); v != "" {
		ca.Serial = &v
	}
	// ca.intermediate.create 的父 CA
	if op == "ca.intermediate.create" {
		if v := getString(params, "parent_ca_id"); v != "" {
			ca.ParentCAID = &v
		}
	}
	// key_params 原样落库
	if kp, ok := params["key_params"].(map[string]interface{}); ok {
		ca.KeyParams = models.JSONMap(kp)
	}

	err := s.db.Clauses(clauseOnConflict("ca_id", []string{
		"parent_ca_id", "subject_cn", "subject_o", "algorithm",
		"key_params", "validity_days", "cert_path", "key_ref",
		"chain_path", "serial", "status", "updated_at",
	})).Create(ca).Error
	if err != nil {
		log.Error().Err(err).Str("ca_id", caID).Msg("CA 元数据落库失败")
		return
	}
	log.Info().Str("ca_id", caID).Msg("CA 元数据已落库")
}

// -----------------------------------------------------------------------------
// CSR
// -----------------------------------------------------------------------------

func (s *Service) persistCSR(params, data map[string]interface{}) {
	csrID := getString(data, "csr_id")
	if csrID == "" {
		return
	}

	csr := &models.CSR{
		CSRID:     csrID,
		SubjectCN: getNestedString(params, "subject", "CN"),
		Algorithm: getString(params, "algorithm"),
		CSRPath:   getString(data, "csr_path"),
		Status:    "NEW",
		CreatedAt: time.Now().UTC(),
	}
	if v := getString(data, "key_ref"); v != "" {
		csr.KeyRef = &v
	}
	if err := s.db.Clauses(clauseOnConflict("csr_id", []string{
		"subject_cn", "algorithm", "csr_path", "key_ref", "status",
	})).Create(csr).Error; err != nil {
		log.Error().Err(err).Str("csr_id", csrID).Msg("CSR 元数据落库失败")
		return
	}
	log.Info().Str("csr_id", csrID).Msg("CSR 元数据已落库")
}

// -----------------------------------------------------------------------------
// Certificate
// -----------------------------------------------------------------------------

func (s *Service) persistCert(params, data map[string]interface{}) {
	certID := getString(data, "cert_id")
	if certID == "" {
		return
	}

	now := time.Now().UTC()
	cert := &models.Certificate{
		CertID:    certID,
		Serial:    getString(data, "serial"),
		SubjectCN: getNestedString(params, "subject", "CN"),
		CAID:      getString(params, "ca_id"),
		Algorithm: getString(params, "algorithm"),
		CertPath:  getString(data, "cert_path"),
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, getIntDefault(params, "validity_days", 365)),
		CreatedAt: now,
	}
	// cert_type 从 params 取，默认 server
	certType := getString(params, "cert_type")
	if certType == "" {
		certType = "server"
	}
	cert.CertType = certType
	if v := getString(data, "chain_path"); v != "" {
		cert.ChainPath = &v
	}
	if err := s.db.Clauses(clauseOnConflict("cert_id", []string{
		"serial", "subject_cn", "issuer_cn", "ca_id", "algorithm",
		"not_before", "not_after", "cert_path", "chain_path",
		"cert_type", "status",
	})).Create(cert).Error; err != nil {
		log.Error().Err(err).Str("cert_id", certID).Msg("证书元数据落库失败")
		return
	}
	log.Info().Str("cert_id", certID).Msg("证书元数据已落库")
}

// -----------------------------------------------------------------------------
// CRL
// -----------------------------------------------------------------------------

func (s *Service) persistCRL(params, data map[string]interface{}) {
	crlID := getString(data, "crl_id")
	if crlID == "" {
		return
	}

	crl := &models.CRL{
		CRLID:      crlID,
		CAID:       getString(params, "ca_id"),
		CRLPath:    getString(data, "crl_path"),
		DigestAlgo: getString(params, "digest_algorithm"),
		Status:     "ACTIVE",
		CreatedAt:  time.Now().UTC(),
	}
	if v, ok := data["revoked_count"].(float64); ok {
		crl.RevokedCount = int(v)
	}
	if err := s.db.Clauses(clauseOnConflict("crl_id", []string{
		"ca_id", "crl_path", "revoked_count", "digest_algorithm", "status",
	})).Create(crl).Error; err != nil {
		log.Error().Err(err).Str("crl_id", crlID).Msg("CRL 元数据落库失败")
		return
	}
	log.Info().Str("crl_id", crlID).Msg("CRL 元数据已落库")
}

// -----------------------------------------------------------------------------
// KeyMeta
// -----------------------------------------------------------------------------

func (s *Service) persistKey(params, data map[string]interface{}) {
	keyRef := getString(data, "key_ref")
	if keyRef == "" {
		return
	}
	// 仅对 generate 和 import 落库；export/delete 不改主表
	action := getString(params, "action")
	if action != "generate" && action != "import" {
		return
	}

	now := time.Now().UTC()
	key := &models.KeyMeta{
		KeyID:         getString(data, "key_id"),
		KeyRef:        keyRef,
		Algorithm:     getString(data, "algorithm"),
		EncryptedPath: getString(data, "encrypted_path"),
		Permission:    "0600",
		State:         "ACTIVE",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if key.KeyID == "" {
		key.KeyID = keyRef
	}
	if kp, ok := params["key_params"].(map[string]interface{}); ok {
		key.KeyParams = models.JSONMap(kp)
	}
	if err := s.db.Clauses(clauseOnConflict("key_id", []string{
		"key_ref", "algorithm", "key_params", "encrypted_path",
		"permission", "state", "updated_at",
	})).Create(key).Error; err != nil {
		log.Error().Err(err).Str("key_ref", keyRef).Msg("密钥元数据落库失败")
		return
	}
	log.Info().Str("key_ref", keyRef).Msg("密钥元数据已落库")
}

// -----------------------------------------------------------------------------
// 辅助函数
// -----------------------------------------------------------------------------

// getString 从 map 取字符串。
func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// getNestedString 从嵌套 map 取字符串，如 getNestedString(params, "subject", "CN")。
func getNestedString(m map[string]interface{}, outer, inner string) string {
	if m == nil {
		return ""
	}
	sub, ok := m[outer].(map[string]interface{})
	if !ok {
		return ""
	}
	if s, ok := sub[inner].(string); ok {
		return s
	}
	return ""
}

// getInt 从 map 取整数（JSON 解析后通常为 float64）。
func getInt(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

// getIntDefault 从 map 取整数，缺省返回 def。
func getIntDefault(m map[string]interface{}, key string, def int) int {
	v := getInt(m, key)
	if v == 0 {
		return def
	}
	return v
}
