// backend-go/internal/modules/tcps_proxy/cert_catalog.go
package tcps_proxy
import (
	"strings"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/models"
)
type CertCatalog struct {
	db *gorm.DB
}
func NewCertCatalog(db *gorm.DB) *CertCatalog {
	return &CertCatalog{db: db}
}
func (c *CertCatalog) List(purpose, protocol string) ([]CertBrief, error) {
	if !IsTCPSProtocol(protocol) {
		return []CertBrief{}, nil
	}
	if purpose != "sign" && purpose != "enc" {
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
	for _, m := range rows {
		if !matchProtocolTCPS(&m, protocol) {
			continue
		}
		if !matchPurpose(&m, purpose) {
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
		Msg("CertCatalog.List(tcps)")
	return out, nil
}
// matchProtocolTCPS tcp 协议不区分国密/国际（默认全部返回）；其余按国密性判定。
func matchProtocolTCPS(m *models.Certificate, protocol string) bool {
	if protocol == ProtocolTCP {
		return true
	}
	isGM := isGMCert(m)
	if protocol == ProtocolGMSSL {
		return isGM
	}
	return !isGM
}
func isGMCert(m *models.Certificate) bool {
	sig := strings.ToUpper(m.SignatureAlgorithm)
	if strings.Contains(sig, "SM2") || strings.Contains(sig, "SM3") {
		return true
	}
	if strings.Contains(strings.ToUpper(m.Algorithm), "SM2") {
		return true
	}
	if strings.HasPrefix(m.CertType, "dual_") {
		return true
	}
	return false
}
func matchPurpose(m *models.Certificate, purpose string) bool {
	ct := strings.ToLower(m.CertType)
	if ct == "" || ct == "imported" || ct == "unknown" {
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
