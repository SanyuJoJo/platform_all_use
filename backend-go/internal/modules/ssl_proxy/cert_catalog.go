// backend-go/internal/modules/ssl_proxy/cert_catalog.go
package ssl_proxy

import (
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/models"
)

// CertCatalog 证书目录：提供"可用证书列表"，供前端下拉框使用。
//
// 数据源：platform_certificate。
//
// ★ 国密判定依据（不改一期证书模块，纯读 DB 字段）：
//
//	SM2 与 ECDSA 的 algorithm 都被一期模块归一化为 "ECC(256)"，
//	唯一可靠区分字段是 signature_algorithm：
//
//	  SM2 证书   → "SM2SigningwithSM3"
//	  ECDSA 证书 → "ECDSAWithSHA256" / "ECDSAWithSHA384" / ...
//
// ★ 用途（purpose）判定：
//
//	cert_type 可能取值：
//	  "signature"              → 签名
//	  "encryption"             → 加密
//	  "signature,encryption"   → 两者
//	  "dual_sign"              → 国密双证-签名
//	  "dual_enc"               → 国密双证-加密
//	  "imported"               → 导入
//	  "unknown" / 空           → 无法判定（两个下拉框都显示）
//
// unknown / 空 视为"两者都可以"，避免用户导入的证书因缺 EKU 被隐藏。
type CertCatalog struct {
	db *gorm.DB
}

// NewCertCatalog 构造。
func NewCertCatalog(db *gorm.DB) *CertCatalog {
	return &CertCatalog{db: db}
}

// List 按 purpose + protocol 查询可用证书。
//
//	purpose  : sign（签名）| enc（加密）
//	protocol : http | pqc | gmssl
func (c *CertCatalog) List(purpose, protocol string) ([]CertBrief, error) {
	if purpose != "sign" && purpose != "enc" {
		return []CertBrief{}, nil
	}
	if protocol != ProtocolGMSSL && protocol != ProtocolHTTP && protocol != ProtocolPQC {
		return []CertBrief{}, nil
	}

	var rows []models.Certificate
	if err := c.db.Model(&models.Certificate{}).
		Where("status <> ?", "DELETED").
		Order("id DESC").
		Limit(2000).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]CertBrief, 0)
	for i := range rows {
		m := &rows[i]
		if !matchProtocol(m, protocol) {
			continue
		}
		if !matchPurpose(m, purpose) {
			continue
		}
		hasKey := (m.KeyRef != nil && *m.KeyRef != "") ||
			(m.KeyPath != nil && *m.KeyPath != "")
		notAfter := ""
		if !m.NotAfter.IsZero() {
			notAfter = m.NotAfter.Format("2006-01-02 15:04:05")
		}
		out = append(out, CertBrief{
			CertID:        m.CertID,
			SubjectCN:     m.SubjectCN,
			Algorithm:     m.Algorithm,
			CertType:      m.CertType,
			NotAfter:      notAfter,
			HasPrivateKey: hasKey,
		})
	}

	log.Debug().
		Str("purpose", purpose).
		Str("protocol", protocol).
		Int("total", len(rows)).
		Int("filtered", len(out)).
		Msg("CertCatalog.List")
	return out, nil
}

// matchProtocol 判定证书是否匹配协议。
func matchProtocol(m *models.Certificate, protocol string) bool {
	isGM := isGMCert(m)
	if protocol == ProtocolGMSSL {
		return isGM
	}
	return !isGM
}

// isGMCert 判定是否国密证书（纯 DB 字段判定）。
//
// 判定条件（任意一条命中即国密）：
//   - signature_algorithm 含 SM2 / SM3（最可靠：SM2 用 SM3 摘要）
//   - algorithm 含 SM2（兼容未来可能保留 SM2 标识的数据）
//   - cert_type 前缀 dual_（国密双证）
func isGMCert(m *models.Certificate) bool {
	sig := strings.ToUpper(m.SignatureAlgorithm)
	if strings.Contains(sig, "SM2") || strings.Contains(sig, "SM3") {
		return true
	}
	alg := strings.ToUpper(m.Algorithm)
	if strings.Contains(alg, "SM2") {
		return true
	}
	if strings.HasPrefix(m.CertType, "dual_") {
		return true
	}
	return false
}

// matchPurpose 判定证书是否匹配用途。
//
// 规则：
//   - "unknown" / "" / "imported" → 两个下拉框都显示（用户导入时可能没填 EKU）
//   - purpose=sign → cert_type 含 "sign" 或 "signature"
//   - purpose=enc  → cert_type 含 "enc" 或 "encryption"
//
// 注意：国密双证 cert_type 是 "dual_sign" / "dual_enc"，内含 sign / enc 子串，
// 天然被上面的子串匹配命中。
func matchPurpose(m *models.Certificate, purpose string) bool {
	ct := strings.ToLower(strings.TrimSpace(m.CertType))

	// 无法判定用途的证书：两个下拉框都显示
	if ct == "" || ct == "unknown" || ct == "imported" {
		return true
	}

	switch purpose {
	case "sign":
		return strings.Contains(ct, "sign")
	case "enc":
		return strings.Contains(ct, "enc")
	}
	return false
}

// DebugDump 打印最近 N 条记录的关键字段（仅调试用）。
func (c *CertCatalog) DebugDump(limit int) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.Certificate
	if err := c.db.Model(&models.Certificate{}).
		Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		log.Error().Err(err).Msg("DebugDump 查询失败")
		return
	}
	log.Info().Int("count", len(rows)).Msg("===== CertCatalog DebugDump =====")
	for i := range rows {
		m := &rows[i]
		hasKey := (m.KeyRef != nil && *m.KeyRef != "") ||
			(m.KeyPath != nil && *m.KeyPath != "")
		log.Info().
			Str("cert_id", m.CertID).
			Str("algorithm", m.Algorithm).
			Str("signature_algorithm", m.SignatureAlgorithm).
			Str("cert_type", m.CertType).
			Str("subject_cn", m.SubjectCN).
			Bool("is_gm", isGMCert(m)).
			Bool("has_key", hasKey).
			Msg("cert_row")
	}
}
