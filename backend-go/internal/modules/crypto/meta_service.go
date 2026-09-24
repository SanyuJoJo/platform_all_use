package crypto

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// MetaService 提供 CA/证书/CSR/CRL/密钥的列表、详情、导入、导出、删除。
type MetaService struct {
	db       *gorm.DB
	adapter  *CoreAdapter
	coreRoot string
}

// NewMetaService 创建元数据查询与 CA 管理服务。
func NewMetaService(db *gorm.DB, cfg *config.Config) *MetaService {
	adapter := NewCoreAdapter(
		cfg.CoreDispatchPath,
		cfg.CoreTimeoutMs,
		cfg.CoreMaxConcurrency,
	)
	coreRoot := deriveCoreRoot(cfg.CoreDispatchPath)

	log.Info().
		Str("dispatch_path", cfg.CoreDispatchPath).
		Str("core_root", coreRoot).
		Msg("MetaService 初始化")

	return &MetaService{
		db:       db,
		adapter:  adapter,
		coreRoot: coreRoot,
	}
}

// deriveCoreRoot 从 dispatch.sh 路径推导 core 根目录。
//
// 处理三种情况：
//   - 空字符串：返回 "."；
//   - 绝对路径：取两次 Dir；
//   - 相对路径：转绝对后处理，失败则回退。
func deriveCoreRoot(dispatchPath string) string {
	if strings.TrimSpace(dispatchPath) == "" {
		return "."
	}
	abs := dispatchPath
	if !filepath.IsAbs(abs) {
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
	}
	// <coreRoot>/sbin/dispatch.sh → <coreRoot>
	return filepath.Dir(filepath.Dir(abs))
}

// -----------------------------------------------------------------------------
// CA 查询
// -----------------------------------------------------------------------------

func (s *MetaService) ListCAs(page, pageSize int) (map[string]interface{}, error) {
	q := s.db.Model(&models.CA{}).Where("status <> ?", "DELETED")
	var cas []models.CA
	return paginateQuery(q, &cas, page, pageSize)
}

func (s *MetaService) GetCA(caID string) (*models.CA, error) {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		log.Error().Err(err).Str("ca_id", caID).Msg("GetCA 查询失败")
		return nil, exception.New(exception.CodeInternalError, "查询 CA 失败", 500, nil)
	}
	return &ca, nil
}

// CADetail CA 证书解析详情。
type CADetail struct {
	Version            string `json:"version"`
	Serial             string `json:"serial"`
	Issuer             string `json:"issuer"`
	Subject            string `json:"subject"`
	Fingerprint        string `json:"fingerprint"`
	NotBefore          string `json:"not_before"`
	NotAfter           string `json:"not_after"`
	PublicKeyAlgorithm string `json:"public_key_algorithm"`
	SignatureAlgorithm string `json:"signature_algorithm"`
	SignatureValue     string `json:"signature_value"`
	PublicKeyValue     string `json:"public_key_value"`
}

func (s *MetaService) GetCADetail(caID string) (*CADetail, error) {
	ca, err := s.GetCA(caID)
	if err != nil {
		return nil, err
	}
	certPEM, err := s.readCoreFile(ca.CertPath)
	if err != nil {
		return nil, err
	}

	// 优先用铜锁 openssl 解析（支持 SM2 / Falcon / ML-DSA）
	if opensslBin := s.getTongsuoOpenssl(); opensslBin != "" {
		certAbs := ca.CertPath
		if !filepath.IsAbs(certAbs) {
			certAbs = filepath.Join(s.coreRoot, ca.CertPath)
		}
		if detail, err := parseCertWithOpenSSL(opensslBin, certAbs); err == nil && detail.Subject != "" {
			return detail, nil
		}
	}
	return parseCADetailFromPEM(certPEM)
}

func (s *MetaService) getTongsuoOpenssl() string {
	candidates := []string{
		filepath.Join(s.coreRoot, "run/bin/openssl"),
		filepath.Join(s.coreRoot, "libs/bin/tongsuo/bin/openssl"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() && st.Mode()&0111 != 0 {
			return p
		}
	}
	if p, err := exec.LookPath("openssl"); err == nil {
		return p
	}
	return ""
}

func runOpenSSLFull(bin string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func runOpenSSL(bin string, args ...string) (string, error) {
	stdout, _, err := runOpenSSLFull(bin, args...)
	return stdout, err
}

func parseCertWithOpenSSL(bin, certPath string) (*CADetail, error) {
	textOut, err := runOpenSSL(
		bin, "x509", "-in", certPath, "-noout", "-text", "-nameopt", "RFC2253",
	)
	if err != nil {
		return nil, fmt.Errorf("openssl x509 -text failed: %w", err)
	}
	detail := &CADetail{}
	parseCertText(textOut, detail)

	if fpOut, err := runOpenSSL(
		bin, "x509", "-in", certPath, "-noout", "-fingerprint", "-sha256",
	); err == nil {
		fp := strings.TrimSpace(fpOut)
		if idx := strings.Index(fp, "="); idx >= 0 {
			fp = strings.TrimSpace(fp[idx+1:])
		}
		detail.Fingerprint = cleanHex(fp)
	}
	return detail, nil
}

func parseCertText(text string, detail *CADetail) {
	lines := strings.Split(text, "\n")

	var (
		sigAlgCount   int
		inPubKey      bool
		inSigValue    bool
		inSerial      bool
		pubKeyLines   []string
		sigValueLines []string
		serialLines   []string
		pubAlg        string
	)

	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if inSigValue {
			if isHexLine(trimmed) {
				sigValueLines = append(sigValueLines, cleanHex(trimmed))
			}
			continue
		}
		if inPubKey {
			if isHexLine(trimmed) {
				pubKeyLines = append(pubKeyLines, cleanHex(trimmed))
				continue
			}
			if strings.HasPrefix(trimmed, "ASN1 OID") ||
				strings.HasPrefix(trimmed, "NIST CURVE") ||
				strings.HasPrefix(trimmed, "pub:") ||
				strings.HasPrefix(trimmed, "Public-Key:") {
				continue
			}
			inPubKey = false
		}
		if inSerial {
			if isHexLine(trimmed) {
				serialLines = append(serialLines, cleanHex(trimmed))
				continue
			}
			inSerial = false
			if detail.Serial == "" {
				detail.Serial = strings.Join(serialLines, "")
			}
		}

		switch {
		case strings.HasPrefix(trimmed, "Version:"):
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "Version:"))
			if idx := strings.Index(v, " "); idx > 0 {
				v = v[:idx]
			}
			if v != "" {
				detail.Version = "v" + v
			}
		case strings.HasPrefix(trimmed, "Serial Number:"):
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "Serial Number:"))
			if rest != "" && isHexLine(rest) {
				serialLines = append(serialLines, cleanHex(rest))
			}
			inSerial = true
		case strings.HasPrefix(trimmed, "Signature Algorithm:"):
			sigAlgCount++
			if sigAlgCount == 1 && detail.SignatureAlgorithm == "" {
				detail.SignatureAlgorithm = strings.TrimSpace(
					strings.TrimPrefix(trimmed, "Signature Algorithm:"))
			} else if sigAlgCount >= 2 {
				inSigValue = true
			}
		case strings.HasPrefix(trimmed, "Issuer:"):
			detail.Issuer = strings.TrimSpace(strings.TrimPrefix(trimmed, "Issuer:"))
		case strings.HasPrefix(trimmed, "Subject:"):
			detail.Subject = strings.TrimSpace(strings.TrimPrefix(trimmed, "Subject:"))
		case strings.HasPrefix(trimmed, "Not Before:"):
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "Not Before:"))
			detail.NotBefore = normalizeTime(v)
		case strings.HasPrefix(trimmed, "Not After"):
			v := strings.TrimSpace(strings.TrimPrefix(trimmed, "Not After"))
			v = strings.TrimSpace(strings.TrimPrefix(v, ":"))
			detail.NotAfter = normalizeTime(v)
		case strings.HasPrefix(trimmed, "Public Key Algorithm:"):
			pubAlg = strings.TrimSpace(strings.TrimPrefix(trimmed, "Public Key Algorithm:"))
		case strings.HasPrefix(trimmed, "ASN1 OID:"):
			oid := strings.TrimSpace(strings.TrimPrefix(trimmed, "ASN1 OID:"))
			if oid != "" && (pubAlg == "id-ecPublicKey" || pubAlg == "") {
				pubAlg = oid
			}
		case strings.HasPrefix(trimmed, "pub:"):
			inPubKey = true
		case strings.HasPrefix(trimmed, "Public-Key:"):
			inPubKey = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "Public-Key:"))
			if rest != "" && isHexLine(rest) {
				pubKeyLines = append(pubKeyLines, cleanHex(rest))
			}
		case strings.HasPrefix(trimmed, "Public Key:"):
			inPubKey = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "Public Key:"))
			if rest != "" && isHexLine(rest) {
				pubKeyLines = append(pubKeyLines, cleanHex(rest))
			}
		case strings.HasPrefix(trimmed, "SHA256 Fingerprint="):
			fp := strings.TrimSpace(strings.TrimPrefix(trimmed, "SHA256 Fingerprint="))
			detail.Fingerprint = cleanHex(fp)
		}
	}

	if inSerial && detail.Serial == "" {
		detail.Serial = strings.Join(serialLines, "")
	}
	if detail.Serial == "" && len(serialLines) > 0 {
		detail.Serial = strings.Join(serialLines, "")
	}
	if detail.PublicKeyAlgorithm == "" {
		detail.PublicKeyAlgorithm = pubAlg
	}
	if len(pubKeyLines) > 0 {
		detail.PublicKeyValue = strings.Join(pubKeyLines, "")
	}
	if len(sigValueLines) > 0 {
		detail.SignatureValue = strings.Join(sigValueLines, "")
	}
}

func isHexLine(s string) bool {
	if len(s) < 2 {
		return false
	}
	s = strings.TrimSuffix(s, ":")
	if s == "" {
		return false
	}
	for _, c := range s {
		if c == ':' {
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func cleanHex(s string) string {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	return strings.TrimSpace(s)
}

func normalizeTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"Jan _2 15:04:05 2006 GMT",
		"Jan 2 15:04:05 2006 GMT",
		"Jan _2 15:04:05 2006 MST",
		"Jan 2 15:04:05 2006 MST",
		"2006-01-02 15:04:05 MST",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Format("2006-01-02 15:04:05")
		}
	}
	return s
}

func parseCADetailFromPEM(certPEM []byte) (*CADetail, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, exception.New(exception.CodeParamInvalid, "PEM 解析失败", 400, nil)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败（可能是不支持的算法）: %v", err), 400, nil,
		)
	}
	fp := sha256.Sum256(cert.Raw)
	return &CADetail{
		Version:            fmt.Sprintf("v%d", cert.Version),
		Serial:             cert.SerialNumber.Text(16),
		Issuer:             cert.Issuer.String(),
		Subject:            cert.Subject.String(),
		Fingerprint:        hex.EncodeToString(fp[:]),
		NotBefore:          cert.NotBefore.UTC().Format("2006-01-02 15:04:05"),
		NotAfter:           cert.NotAfter.UTC().Format("2006-01-02 15:04:05"),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		SignatureValue:     hex.EncodeToString(cert.Signature),
		PublicKeyValue:     hex.EncodeToString(cert.RawSubjectPublicKeyInfo),
	}, nil
}

// -----------------------------------------------------------------------------
// 私钥密码检测与解密
// -----------------------------------------------------------------------------

func isPrivateKeyEncrypted(opensslBin, keyPath string) (bool, error) {
	_, _, err := runOpenSSLFull(opensslBin, "pkey", "-in", keyPath, "-noout")
	if err == nil {
		return false, nil
	}
	return true, nil
}

func decryptKeyWithPasswordFile(
	opensslBin, keyPath, passwordFile, outputPath string,
) error {
	_, _, err := runOpenSSLFull(
		opensslBin, "pkey",
		"-in", keyPath,
		"-passin", "file:"+passwordFile,
		"-out", outputPath,
	)
	return err
}

// -----------------------------------------------------------------------------
// CA 导入
// -----------------------------------------------------------------------------

type ImportCARequest struct {
	CertPEM     string `json:"cert_pem"`
	KeyPEM      string `json:"key_pem,omitempty"`
	KeyPassword string `json:"key_password,omitempty"`
	KeyRef      string `json:"key_ref,omitempty"`
}

func (s *MetaService) ImportCA(
	ctx context.Context, req *ImportCARequest,
) (*models.CA, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" {
		return nil, exception.New(exception.CodeParamInvalid, "证书文件不能为空", 400, nil)
	}
	if !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(exception.CodeParamInvalid, "证书文件不是有效的 PEM 格式", 400, nil)
	}

	idStr := strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	caID := "ca-" + idStr

	certRel := fmt.Sprintf("data/ca/imported-%s.pem", idStr)
	certAbs := filepath.Join(s.coreRoot, certRel)
	if err := os.MkdirAll(filepath.Dir(certAbs), 0750); err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("创建目录失败: %v", err), 500, nil,
		)
	}
	if err := os.WriteFile(certAbs, []byte(certPEM), 0640); err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("写证书失败: %v", err), 500, nil,
		)
	}
	success := false
	defer func() {
		if !success {
			_ = os.Remove(certAbs)
		}
	}()

	parsed, err := s.parseCertViaCore(ctx, certRel)
	if err != nil {
		return nil, err
	}
	subject := getString(parsed, "subject")
	serial := getString(parsed, "serial")
	algorithm := getString(parsed, "public_key_algorithm")

	keyRef := strings.TrimSpace(req.KeyRef)
	keyPEM := strings.TrimSpace(req.KeyPEM)

	if keyRef == "" && keyPEM != "" {
		if !strings.Contains(keyPEM, "-----BEGIN") {
			return nil, exception.New(exception.CodeParamInvalid, "私钥文件不是有效的 PEM 格式", 400, nil)
		}

		keyRel := fmt.Sprintf("tmp/import-%s.key", idStr)
		keyAbs := filepath.Join(s.coreRoot, keyRel)
		if err := os.MkdirAll(filepath.Dir(keyAbs), 0700); err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("创建临时目录失败: %v", err), 500, nil,
			)
		}
		if err := os.WriteFile(keyAbs, []byte(keyPEM), 0600); err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("写临时私钥失败: %v", err), 500, nil,
			)
		}
		defer os.Remove(keyAbs)

		opensslBin := s.getTongsuoOpenssl()
		if opensslBin == "" {
			return nil, exception.New(
				exception.CodeInternalError,
				"铜锁 openssl 不可用，无法检测私钥密码", 500, nil,
			)
		}

		encrypted, _ := isPrivateKeyEncrypted(opensslBin, keyAbs)
		if encrypted {
			if strings.TrimSpace(req.KeyPassword) == "" {
				return nil, exception.New(
					exception.CodeParamInvalid,
					"私钥已加密，请提供私钥密码", 400, nil,
				)
			}
			pwRel := fmt.Sprintf("tmp/import-%s.pw", idStr)
			pwAbs := filepath.Join(s.coreRoot, pwRel)
			if err := os.WriteFile(pwAbs, []byte(req.KeyPassword), 0600); err != nil {
				return nil, exception.New(
					exception.CodeInternalError,
					fmt.Sprintf("写密码文件失败: %v", err), 500, nil,
				)
			}
			defer os.Remove(pwAbs)

			plainRel := fmt.Sprintf("tmp/import-%s.plain.key", idStr)
			plainAbs := filepath.Join(s.coreRoot, plainRel)
			defer os.Remove(plainAbs)

			if err := decryptKeyWithPasswordFile(opensslBin, keyAbs, pwAbs, plainAbs); err != nil {
				return nil, exception.New(
					exception.CodeParamInvalid,
					"私钥密码错误或解密失败", 400, nil,
				)
			}
			plainData, err := os.ReadFile(plainAbs)
			if err != nil {
				return nil, exception.New(
					exception.CodeInternalError,
					fmt.Sprintf("读取解密结果失败: %v", err), 500, nil,
				)
			}
			if err := os.WriteFile(keyAbs, plainData, 0600); err != nil {
				return nil, exception.New(
					exception.CodeInternalError,
					fmt.Sprintf("写明文私钥失败: %v", err), 500, nil,
				)
			}
		}

		keyRef, err = s.importKeyViaCore(ctx, keyRel, algorithm)
		if err != nil {
			return nil, err
		}
	}

	now := timeNowUTC()
	ca := &models.CA{
		CAID:         caID,
		SubjectCN:    extractCNFromSubject(subject),
		Algorithm:    algorithm,
		CertPath:     certRel,
		KeyRef:       keyRef,
		ValidityDays: 0,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if serial != "" {
		ca.Serial = &serial
	}
	if err := s.db.Create(ca).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("写库失败: %v", err), 500, nil,
		)
	}
	success = true
	return ca, nil
}

// -----------------------------------------------------------------------------
// CA 导出
// -----------------------------------------------------------------------------

type ExportFormat string

const (
	ExportFormatCert   ExportFormat = "cert"
	ExportFormatKey    ExportFormat = "key"
	ExportFormatPKCS12 ExportFormat = "pkcs12"
)

type ExportResult struct {
	Data        []byte
	Filename    string
	ContentType string
}

func (s *MetaService) ExportCA(
	ctx context.Context, caID string, format ExportFormat, password string,
) (*ExportResult, error) {
	ca, err := s.GetCA(caID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("ca_id", caID).
		Str("cert_path", ca.CertPath).
		Str("key_ref", ca.KeyRef).
		Str("core_root", s.coreRoot).
		Str("format", string(format)).
		Msg("ExportCA 开始")

	switch format {
	case ExportFormatCert:
		data, err := s.readCoreFile(ca.CertPath)
		if err != nil {
			log.Error().Err(err).Str("ca_id", caID).Str("path", ca.CertPath).Msg("导出证书失败")
			return nil, err
		}
		return &ExportResult{
			Data:        data,
			Filename:    caID + ".pem",
			ContentType: "application/x-pem-file",
		}, nil

	case ExportFormatKey:
		if ca.KeyRef == "" {
			return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥", 400, nil)
		}
		data, err := s.exportKeyViaCore(ctx, ca.KeyRef)
		if err != nil {
			log.Error().Err(err).Str("ca_id", caID).Str("key_ref", ca.KeyRef).Msg("导出私钥失败")
			return nil, err
		}
		return &ExportResult{
			Data:        data,
			Filename:    caID + ".key.pem",
			ContentType: "application/x-pem-file",
		}, nil

	case ExportFormatPKCS12:
		if ca.KeyRef == "" {
			return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥，无法导出 PKCS#12", 400, nil)
		}
		if strings.TrimSpace(password) == "" {
			return nil, exception.New(exception.CodeParamInvalid, "PKCS#12 导出需要密码", 400, nil)
		}
		data, err := s.exportPKCS12ViaCore(ctx, ca.CertPath, ca.KeyRef, password)
		if err != nil {
			log.Error().Err(err).Str("ca_id", caID).Msg("导出 PKCS#12 失败")
			return nil, err
		}
		return &ExportResult{
			Data:        data,
			Filename:    caID + ".p12",
			ContentType: "application/x-pkcs12",
		}, nil
	}

	return nil, exception.New(
		exception.CodeParamInvalid,
		fmt.Sprintf("不支持的导出格式: %s", format), 400, nil,
	)
}

func (s *MetaService) DeleteCA(caID string) error {
	res := s.db.Model(&models.CA{}).
		Where("ca_id = ? AND status <> ?", caID, "DELETED").
		Updates(map[string]interface{}{
			"status":     "DELETED",
			"updated_at": timeNowUTC(),
		})
	if res.Error != nil {
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("删除 CA 失败: %v", res.Error), 500, nil,
		)
	}
	if res.RowsAffected == 0 {
		return exception.New(exception.CodeNotFound, "CA 不存在或已删除", 404, nil)
	}
	return nil
}

// -----------------------------------------------------------------------------
// 证书 / CSR / CRL / 密钥 查询
// -----------------------------------------------------------------------------

func (s *MetaService) ListCerts(page, pageSize int, certType, caID string) (map[string]interface{}, error) {
	q := s.db.Model(&models.Certificate{})
	if certType != "" {
		q = q.Where("cert_type = ?", certType)
	}
	if caID != "" {
		q = q.Where("ca_id = ?", caID)
	}
	var certs []models.Certificate
	return paginateQuery(q, &certs, page, pageSize)
}

func (s *MetaService) GetCert(certID string) (*models.Certificate, error) {
	var c models.Certificate
	if err := s.db.Where("cert_id = ?", certID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询证书失败", 500, nil)
	}
	return &c, nil
}

func (s *MetaService) ListCSRs(page, pageSize int) (map[string]interface{}, error) {
	var csrs []models.CSR
	return paginateQuery(s.db.Model(&models.CSR{}), &csrs, page, pageSize)
}

func (s *MetaService) ListCRLs(page, pageSize int, caID string) (map[string]interface{}, error) {
	q := s.db.Model(&models.CRL{})
	if caID != "" {
		q = q.Where("ca_id = ?", caID)
	}
	var crls []models.CRL
	return paginateQuery(q, &crls, page, pageSize)
}

func (s *MetaService) ListKeys(page, pageSize int, algorithm, state string) (map[string]interface{}, error) {
	q := s.db.Model(&models.KeyMeta{})
	if algorithm != "" {
		q = q.Where("algorithm = ?", algorithm)
	}
	if state != "" {
		q = q.Where("state = ?", state)
	}
	var keys []models.KeyMeta
	return paginateQuery(q, &keys, page, pageSize)
}

// -----------------------------------------------------------------------------
// 通用分页
// -----------------------------------------------------------------------------

func paginate(db *gorm.DB, model interface{}, page, pageSize int, where func(*gorm.DB) *gorm.DB) (map[string]interface{}, error) {
	q := db.Model(model)
	if where != nil {
		q = where(q)
	}
	return paginateQuery(q, model, page, pageSize)
}

func paginateQuery(q *gorm.DB, slicePtr interface{}, page, pageSize int) (map[string]interface{}, error) {
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if err := q.Session(&gorm.Session{}).
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(slicePtr).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	pages := 0
	if pageSize > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return map[string]interface{}{
		"items":     slicePtr,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     pages,
	}, nil
}

// -----------------------------------------------------------------------------
// core 调用辅助
// -----------------------------------------------------------------------------

func (s *MetaService) callCore(
	ctx context.Context, op string, params map[string]interface{},
) (*CoreResponse, error) {
	req := &CoreRequest{
		SchemaVersion: "1.0",
		OperationID:   op,
		RequestID:     uuid.NewString(),
		Actor:         CoreActor{Type: "platform-backend", ID: "system"},
		Params:        params,
	}
	resp, _, err := s.adapter.Call(ctx, op, req)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("调用 core %s 失败: %v", op, err), 500, nil,
		)
	}
	if resp.Code != "OK" {
		m := MapCoreError(resp.Code)
		return nil, exception.New(
			exceptionCodeFromHTTP(m.HTTPStatus),
			fmt.Sprintf("[%s] %s", resp.Code, resp.Message),
			m.HTTPStatus, nil,
		)
	}
	return resp, nil
}

func (s *MetaService) parseCertViaCore(
	ctx context.Context, certRel string,
) (map[string]interface{}, error) {
	resp, err := s.callCore(ctx, "cert.parse", map[string]interface{}{
		"cert_path": certRel,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (s *MetaService) importKeyViaCore(
	ctx context.Context, keyRel, algorithm string,
) (string, error) {
	if algorithm == "" {
		algorithm = "SM2"
	}
	resp, err := s.callCore(ctx, "key.manage", map[string]interface{}{
		"action":    "import",
		"key_path":  keyRel,
		"algorithm": algorithm,
	})
	if err != nil {
		return "", err
	}
	keyRef := getString(resp.Data, "key_ref")
	if keyRef == "" {
		return "", exception.New(
			exception.CodeInternalError, "key.manage 未返回 key_ref", 500, nil,
		)
	}
	return keyRef, nil
}

func (s *MetaService) exportKeyViaCore(
	ctx context.Context, keyRef string,
) ([]byte, error) {
	idStr := strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	exportRel := fmt.Sprintf("tmp/export-%s.key.pem", idStr)
	exportAbs := filepath.Join(s.coreRoot, exportRel)

	_, err := s.callCore(ctx, "key.manage", map[string]interface{}{
		"action":             "export",
		"key_ref":            keyRef,
		"export_path":        exportRel,
		"allow_plain_export": true,
	})
	if err != nil {
		return nil, err
	}
	defer os.Remove(exportAbs)

	return s.readCoreFile(exportRel)
}

func (s *MetaService) exportPKCS12ViaCore(
	ctx context.Context, certRel, keyRef, password string,
) ([]byte, error) {
	idStr := strings.ReplaceAll(uuid.NewString(), "-", "")[:16]

	pwRel := fmt.Sprintf("tmp/export-%s.pw", idStr)
	pwAbs := filepath.Join(s.coreRoot, pwRel)
	if err := os.WriteFile(pwAbs, []byte(password), 0600); err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("写密码文件失败: %v", err), 500, nil,
		)
	}
	defer os.Remove(pwAbs)

	targetRel := fmt.Sprintf("tmp/export-%s.p12", idStr)
	targetAbs := filepath.Join(s.coreRoot, targetRel)
	defer os.Remove(targetAbs)

	_, err := s.callCore(ctx, "cert.convert", map[string]interface{}{
		"source_format": "PEM",
		"target_format": "PKCS12",
		"source_path":   certRel,
		"target_path":   targetRel,
		"password_file": pwRel,
		"key_ref":       keyRef,
	})
	if err != nil {
		return nil, err
	}
	return s.readCoreFile(targetRel)
}

// readCoreFile 读取 core 受控路径下的文件。
//
// 错误分类：
//   - 路径不在受控目录 → 400
//   - 文件不存在 → 404
//   - 权限不足 → 403（返回 500 会掩盖真实原因）
//   - 路径是目录 → 400
//   - 其他 IO 错误 → 500
func (s *MetaService) readCoreFile(relPath string) ([]byte, error) {
	if strings.TrimSpace(relPath) == "" {
		return nil, exception.New(
			exception.CodeParamInvalid, "文件路径为空", 400, nil,
		)
	}
	if !isUnderCorePath(s.coreRoot, relPath) {
		log.Warn().
			Str("core_root", s.coreRoot).
			Str("path", relPath).
			Msg("路径不在受控目录")
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("路径不在受控目录: %s", relPath), 400, nil,
		)
	}

	full := relPath
	if !filepath.IsAbs(full) {
		full = filepath.Join(s.coreRoot, relPath)
	}

	// 先判断是否目录
	if st, err := os.Stat(full); err == nil && st.IsDir() {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("路径是目录不是文件: %s", full), 400, nil,
		)
	}

	data, err := os.ReadFile(full)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			log.Warn().Str("path", full).Msg("文件不存在")
			return nil, exception.New(
				exception.CodeNotFound,
				fmt.Sprintf("文件不存在: %s", relPath), 404, nil,
			)
		case os.IsPermission(err):
			log.Error().Str("path", full).Msg("读取文件权限不足")
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("读取文件权限不足: %s（请检查后端进程用户对 core 目录的读权限）", relPath),
				500, nil,
			)
		default:
			log.Error().Err(err).Str("path", full).Msg("读取文件失败")
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("读取文件失败: %v", err), 500, nil,
			)
		}
	}
	return data, nil
}

// -----------------------------------------------------------------------------
// 内部辅助
// -----------------------------------------------------------------------------

func extractCNFromSubject(subject string) string {
	s := strings.TrimSpace(subject)
	if s == "" {
		return "imported-ca"
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		upper := strings.ToUpper(part)
		if strings.HasPrefix(upper, "CN=") {
			return part[3:]
		}
	}
	return s
}

func isUnderCorePath(coreRoot, p string) bool {
	if p == "" || coreRoot == "" {
		return false
	}
	abs := p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(coreRoot, p)
	}
	abs = filepath.Clean(abs)
	rootAbs := coreRoot
	if !filepath.IsAbs(rootAbs) {
		if a, err := filepath.Abs(rootAbs); err == nil {
			rootAbs = a
		}
	}
	rootAbs = filepath.Clean(rootAbs)

	dataDir := filepath.Join(rootAbs, "data")
	tmpDir := filepath.Join(rootAbs, "tmp")
	return strings.HasPrefix(abs, dataDir+string(os.PathSeparator)) ||
		abs == dataDir ||
		strings.HasPrefix(abs, tmpDir+string(os.PathSeparator)) ||
		abs == tmpDir
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}

func exceptionCodeFromHTTP(status int) int {
	switch status {
	case http.StatusBadRequest, http.StatusForbidden:
		return exception.CodeParamInvalid
	case http.StatusNotFound:
		return exception.CodeNotFound
	case http.StatusUnprocessableEntity:
		return exception.CodeValidationFail
	default:
		return exception.CodeInternalError
	}
}
