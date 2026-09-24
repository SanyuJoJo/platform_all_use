package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// CertService 证书领域 Service。
type CertService struct {
	db     *gorm.DB
	caller *CoreCaller
	files  *FileStore
	parser *CertParser
	keys   *KeyCrypto
}

// NewCertService 创建证书领域 Service。
func NewCertService(
	db *gorm.DB, caller *CoreCaller, files *FileStore,
	parser *CertParser, keys *KeyCrypto,
) *CertService {
	return &CertService{db: db, caller: caller, files: files, parser: parser, keys: keys}
}

// List 证书列表。
func (s *CertService) List(page, pageSize int, certType, caID string) (map[string]interface{}, error) {
	q := s.db.Model(&models.Certificate{}).Where("status <> ?", "DELETED")
	if certType != "" {
		q = q.Where("cert_type LIKE ?", "%"+certType+"%")
	}
	if caID != "" {
		q = q.Where("ca_id = ?", caID)
	}
	var certs []models.Certificate
	return paginateQuery(q, &certs, page, pageSize)
}

// Get 证书元数据。
func (s *CertService) Get(certID string) (*models.Certificate, error) {
	var c models.Certificate
	if err := s.db.Where("cert_id = ?", certID).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询证书失败", 500, nil)
	}
	return &c, nil
}

// GetDetail 解析证书详情。
func (s *CertService) GetDetail(certID string) (*CertDetail, error) {
	c, err := s.Get(certID)
	if err != nil {
		return nil, err
	}
	abs, err := s.files.guard.Resolve(c.CertPath, "cert_path")
	if err != nil {
		return nil, err
	}
	return s.parser.Parse(abs)
}

// Sign 申请/签发证书（无 CSR）。
func (s *CertService) Sign(ctx context.Context, req *SignCertRequest) (*models.Certificate, error) {
	if len(req.CertTypes) == 0 {
		return nil, exception.New(exception.CodeParamInvalid, "cert_types 不能为空", 400, nil)
	}
	allowed := map[string]struct{}{
		"server": {}, "client": {}, "signature": {}, "encryption": {},
	}
	for _, t := range req.CertTypes {
		if _, ok := allowed[t]; !ok {
			return nil, exception.New(
				exception.CodeParamInvalid,
				fmt.Sprintf("不支持的 cert_type: %s", t), 400, nil,
			)
		}
	}

	subj := make(map[string]interface{}, len(req.Subject))
	for k, v := range req.Subject {
		subj[k] = v
	}
	if err := ValidateSubject(subj); err != nil {
		return nil, err
	}

	if req.ValidityDays <= 0 {
		req.ValidityDays = 365
	}

	params := map[string]interface{}{
		"ca_source":     req.CASource,
		"cert_types":    req.CertTypes,
		"validity_days": req.ValidityDays,
		"algorithm":     req.Algorithm,
		"subject":       req.Subject,
	}
	if req.Algorithm == "" {
		params["algorithm"] = "SM2"
	}
	if req.KeyParams != nil {
		params["key_params"] = req.KeyParams
	}
	if len(req.SAN) > 0 {
		params["san"] = req.SAN
	}

	switch req.CASource {
	case "local":
		if req.CAID == "" {
			return nil, exception.New(
				exception.CodeParamInvalid, "本地 CA 需要 ca_id", 400, nil,
			)
		}
		var ca models.CA
		if err := s.db.Where("ca_id = ? AND status <> ?", req.CAID, "DELETED").
			First(&ca).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, exception.New(
					exception.CodeNotFound, "CA 不存在", 404, nil,
				)
			}
			return nil, exception.New(
				exception.CodeInternalError, "查询 CA 失败", 500, nil,
			)
		}
		if ca.KeyRef == "" {
			return nil, exception.New(
				exception.CodeParamInvalid, "该 CA 未关联私钥，无法签发", 400, nil,
			)
		}
		params["ca_id"] = ca.CAID
		params["ca_cert_path"] = ca.CertPath
		params["ca_key_ref"] = ca.KeyRef

	case "manual":
		if strings.TrimSpace(req.CACertPEM) == "" ||
			strings.TrimSpace(req.CAKeyPEM) == "" {
			return nil, exception.New(
				exception.CodeParamInvalid,
				"手动 CA 需要证书和私钥", 400, nil,
			)
		}
		if !strings.Contains(req.CACertPEM, "-----BEGIN CERTIFICATE-----") {
			return nil, exception.New(
				exception.CodeParamInvalid,
				"CA 证书不是有效 PEM 格式", 400, nil,
			)
		}
		if !strings.Contains(req.CAKeyPEM, "-----BEGIN") {
			return nil, exception.New(
				exception.CodeParamInvalid,
				"CA 私钥不是有效 PEM 格式", 400, nil,
			)
		}

		id := newShortID()
		certRel := fmt.Sprintf("tmp/manual-ca-cert-%s.pem", id)
		keyRel := fmt.Sprintf("tmp/manual-ca-key-%s.pem", id)
		if err := s.files.WriteCoreFile(certRel, []byte(req.CACertPEM), 0600); err != nil {
			return nil, err
		}
		defer s.files.SafeRemove(certRel)
		if err := s.files.WriteCoreFile(keyRel, []byte(req.CAKeyPEM), 0600); err != nil {
			return nil, err
		}
		defer s.files.SafeRemove(keyRel)

		params["ca_cert_path"] = certRel
		params["ca_key_path"] = keyRel

		if strings.TrimSpace(req.CAKeyPassword) != "" {
			pwRel := fmt.Sprintf("tmp/manual-ca-pass-%s", id)
			if err := s.files.WriteCoreFile(pwRel, []byte(req.CAKeyPassword), 0600); err != nil {
				return nil, err
			}
			defer s.files.SafeRemove(pwRel)
			params["ca_key_password_file"] = pwRel
		}

	default:
		return nil, exception.New(
			exception.CodeParamInvalid,
			"ca_source 必须是 local 或 manual", 400, nil,
		)
	}

	resp, err := s.caller.Run(ctx, "cert.sign", params)
	if err != nil {
		return nil, err
	}

	certID := getString(resp.Data, "cert_id")
	certPath := getString(resp.Data, "cert_path")
	chainPath := getString(resp.Data, "chain_path")
	serial := getString(resp.Data, "serial")
	keyRef := getString(resp.Data, "key_ref")
	subject := getString(resp.Data, "subject")
	issuer := getString(resp.Data, "issuer")

	if certID == "" || certPath == "" {
		return nil, exception.New(
			exception.CodeInternalError,
			"core cert.sign 未返回 cert_id/cert_path", 500, nil,
		)
	}

	abs, _ := s.files.guard.Resolve(certPath, "cert_path")
	detail, _ := s.parser.Parse(abs)
	if subject == "" && detail != nil {
		subject = detail.Subject
	}
	if issuer == "" && detail != nil {
		issuer = detail.Issuer
	}
	if serial == "" && detail != nil {
		serial = detail.Serial
	}

	certTypeJoined := strings.Join(req.CertTypes, ",")

	now := time.Now().UTC()
	cert := &models.Certificate{
		CertID:    certID,
		CertType:  certTypeJoined,
		Serial:    serial,
		Subject:   subject,
		SubjectCN: extractCNFromSubject(subject),
		Issuer:    issuer,
		IssuerCN:  extractCNFromSubject(issuer),
		CAID:      req.CAID,
		Algorithm: req.Algorithm,
		CertPath:  certPath,
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, req.ValidityDays),
		CreatedAt: now,
	}
	if cert.Algorithm == "" {
		cert.Algorithm = "SM2"
	}
	if detail != nil {
		cert.Fingerprint = detail.Fingerprint
		cert.PublicKeyAlgorithm = detail.PublicKeyAlgorithm
		cert.SignatureAlgorithm = detail.SignatureAlgorithm
	}
	if chainPath != "" {
		cert.ChainPath = &chainPath
	}
	if keyRef != "" {
		cert.KeyRef = &keyRef
	}

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil,
		)
	}
	return cert, nil
}

// Import 导入证书（可携带私钥）。
//
// 证书类型推断优先级：
//  1. KeyUsage 含 digitalSignature / nonRepudiation → 签名
//  2. KeyUsage 含 keyEncipherment / keyAgreement  → 加密
//  3. 两者都有 → "signature,encryption"
//  4. KeyUsage 为空时用 EKU 兜底（serverAuth / clientAuth / codeSigning → 签名）
//  5. 仍为空时用公钥算法兜底（ML-KEM → 加密；SM2/RSA/ECC/ML-DSA/SLH-DSA → 签名）
//  6. 全部无 → "imported"
func (s *CertService) Import(ctx context.Context, req *ImportCertRequest) (*models.Certificate, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" || !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(
			exception.CodeParamInvalid,
			"证书不是有效的 PEM 格式", 400, nil,
		)
	}

	id := newShortID()
	certID := "cert-" + id
	certRel := fmt.Sprintf("data/certs/imported-%s.pem", id)

	if err := s.files.WriteCoreFile(certRel, []byte(certPEM), 0640); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = s.files.SafeRemove(certRel)
		}
	}()

	abs, err := s.files.guard.Resolve(certRel, "cert_path")
	if err != nil {
		return nil, err
	}
	detail, err := s.parser.Parse(abs)
	if err != nil {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败: %v", err), 400, nil,
		)
	}

	var keyRef *string
	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(
				exception.CodeParamInvalid,
				"私钥不是有效的 PEM 格式", 400, nil,
			)
		}
		keyRel := fmt.Sprintf("tmp/import-cert-key-%s.pem", id)
		if err := s.files.WriteCoreFile(keyRel, []byte(req.KeyPEM), 0600); err != nil {
			return nil, err
		}
		defer s.files.SafeRemove(keyRel)

		keyAbs, _ := s.files.guard.Resolve(keyRel, "key_path")

		if bin := s.parser.OpensslBin(); bin != "" {
			encrypted, _ := s.keys.IsEncrypted(bin, keyAbs)
			if encrypted {
				if strings.TrimSpace(req.KeyPassword) == "" {
					return nil, exception.New(
						exception.CodeParamInvalid,
						"私钥已加密，请提供私钥密码", 400, nil,
					)
				}
				pwRel := fmt.Sprintf("tmp/import-cert-pass-%s", id)
				plainRel := fmt.Sprintf("tmp/import-cert-plain-%s.pem", id)
				if err := s.files.WriteCoreFile(pwRel, []byte(req.KeyPassword), 0600); err != nil {
					return nil, err
				}
				defer s.files.SafeRemove(pwRel)
				defer s.files.SafeRemove(plainRel)

				pwAbs, _ := s.files.guard.Resolve(pwRel, "path")
				plainAbs, _ := s.files.guard.Resolve(plainRel, "path")

				if err := s.keys.DecryptWithPasswordFile(bin, keyAbs, pwAbs, plainAbs); err != nil {
					return nil, exception.New(
						exception.CodeParamInvalid,
						"私钥密码错误或解密失败", 400, nil,
					)
				}
				plainData, err := s.files.ReadCoreFile(plainRel)
				if err != nil {
					return nil, err
				}
				if err := s.files.WriteCoreFile(keyRel, plainData, 0600); err != nil {
					return nil, err
				}
			}
		}

		ref, err := s.caller.ImportKey(ctx, keyRel, detail.PublicKeyAlgorithm)
		if err != nil {
			return nil, err
		}
		keyRef = &ref
	} else if req.KeyRef != "" {
		keyRef = &req.KeyRef
	}

	inferredType := inferCertType(
		detail.KeyUsage,
		detail.ExtendedKeyUsage,
		detail.PublicKeyAlgorithm,
	)

	now := time.Now().UTC()
	cert := &models.Certificate{
		CertID:             certID,
		CertType:           inferredType,
		Serial:             detail.Serial,
		Subject:            detail.Subject,
		SubjectCN:          extractCNFromSubject(detail.Subject),
		Issuer:             detail.Issuer,
		IssuerCN:           extractCNFromSubject(detail.Issuer),
		Algorithm:          detail.PublicKeyAlgorithm,
		Fingerprint:        detail.Fingerprint,
		PublicKeyAlgorithm: detail.PublicKeyAlgorithm,
		SignatureAlgorithm: detail.SignatureAlgorithm,
		NotBefore:          now,
		NotAfter:           now,
		CertPath:           certRel,
		KeyRef:             keyRef,
		Status:             "VALID",
		CreatedAt:          now,
	}

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil,
		)
	}
	success = true
	return cert, nil
}

// inferCertType 根据 KeyUsage / ExtendedKeyUsage / PublicKeyAlgorithm 推断证书类型。
//
// 规则：
//  1. KeyUsage 明确：
//     - 有 digitalSignature 或 nonRepudiation → 签名
//     - 有 keyEncipherment 或 keyAgreement   → 加密
//     - 两者都有                              → "signature,encryption"
//  2. KeyUsage 无法判断，用 EKU 兜底：
//     - serverAuth / clientAuth / codeSigning → 签名
//  3. 仍无法判断，用公钥算法兜底：
//     - ML-KEM                → 加密
//     - SM2/RSA/ECC/ML-DSA/SLH-DSA → 签名
//  4. 全部无 → "imported"
func inferCertType(keyUsage, eku, pubAlg string) string {
	kuLower := strings.ToLower(keyUsage)
	hasSig := strings.Contains(kuLower, "digital signature") ||
		strings.Contains(kuLower, "non repudiation")
	hasEnc := strings.Contains(kuLower, "key encipherment") ||
		strings.Contains(kuLower, "key agreement")

	switch {
	case hasSig && hasEnc:
		return "signature,encryption"
	case hasSig:
		return "signature"
	case hasEnc:
		return "encryption"
	}

	// KeyUsage 无法判断，用 EKU 兜底
	ekuLower := strings.ToLower(eku)
	if strings.Contains(ekuLower, "server authentication") ||
		strings.Contains(ekuLower, "client authentication") ||
		strings.Contains(ekuLower, "code signing") {
		return "signature"
	}

	// 用公钥算法兜底
	pubLower := strings.ToLower(pubAlg)
	switch {
	case strings.Contains(pubLower, "ml-kem"):
		return "encryption"
	case strings.Contains(pubLower, "ml-dsa"),
		strings.Contains(pubLower, "slh-dsa"),
		strings.Contains(pubLower, "sm2"),
		strings.Contains(pubLower, "rsa"),
		strings.Contains(pubLower, "ecdsa"),
		strings.Contains(pubLower, "ecc"),
		strings.Contains(pubLower, "ecpublickey"):
		return "signature"
	}

	return "imported"
}

// Export 导出证书。
func (s *CertService) Export(
	ctx context.Context, certID, exportType, password string,
) (*ExportResult, error) {
	c, err := s.Get(certID)
	if err != nil {
		return nil, err
	}

	switch exportType {
	case "cert":
		data, err := s.files.ReadCoreFile(c.CertPath)
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Data:        data,
			Filename:    certID + ".pem",
			ContentType: "application/x-pem-file",
		}, nil

	case "key":
		if c.KeyRef == nil || *c.KeyRef == "" {
			return nil, exception.New(
				exception.CodeParamInvalid, "该证书未关联私钥", 400, nil,
			)
		}
		plain, err := s.caller.ExportKey(ctx, *c.KeyRef)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(password) != "" {
			bin := s.parser.OpensslBin()
			if bin == "" {
				return nil, exception.New(
					exception.CodeInternalError,
					"铜锁 openssl 不可用，无法加密私钥", 500, nil,
				)
			}
			encrypted, err := encryptPrivateKeyPEM(bin, plain, password)
			if err != nil {
				return nil, exception.New(
					exception.CodeInternalError,
					fmt.Sprintf("加密私钥失败: %v", err), 500, nil,
				)
			}
			plain = encrypted
		}
		return &ExportResult{
			Data:        plain,
			Filename:    certID + ".key.pem",
			ContentType: "application/x-pem-file",
		}, nil

	case "pkcs12":
		if c.KeyRef == nil || *c.KeyRef == "" {
			return nil, exception.New(
				exception.CodeParamInvalid, "该证书未关联私钥", 400, nil,
			)
		}
		if len(password) < 6 {
			return nil, exception.New(
				exception.CodeParamInvalid, "PKCS#12 密码至少 6 位", 400, nil,
			)
		}
		data, err := s.caller.ExportPKCS12(ctx, c.CertPath, *c.KeyRef, password)
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Data:        data,
			Filename:    certID + ".p12",
			ContentType: "application/x-pkcs12",
		}, nil

	default:
		return nil, exception.New(
			exception.CodeParamInvalid,
			"不支持的导出类型: "+exportType, 400, nil,
		)
	}
}

// Delete 删除证书（软删除）。
func (s *CertService) Delete(certID string) error {
	res := s.db.Model(&models.Certificate{}).
		Where("cert_id = ? AND status <> ?", certID, "DELETED").
		Updates(map[string]interface{}{"status": "DELETED"})
	if res.Error != nil {
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("删除证书失败: %v", res.Error), 500, nil,
		)
	}
	if res.RowsAffected == 0 {
		return exception.New(exception.CodeNotFound, "证书不存在或已删除", 404, nil)
	}
	return nil
}

// encryptPrivateKeyPEM 用 AES-256-CBC 加密 PEM 私钥。
func encryptPrivateKeyPEM(opensslBin string, plainPEM []byte, password string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cert-key-enc-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	plainPath := filepath.Join(tmpDir, "plain.key")
	encPath := filepath.Join(tmpDir, "enc.key")
	pwPath := filepath.Join(tmpDir, "pw.txt")

	if err := os.WriteFile(plainPath, plainPEM, 0600); err != nil {
		return nil, fmt.Errorf("写明文私钥失败: %w", err)
	}
	if err := os.WriteFile(pwPath, []byte(password), 0600); err != nil {
		return nil, fmt.Errorf("写口令文件失败: %w", err)
	}

	_, stderr, err := RunOpenSSLFull(
		opensslBin, "pkey",
		"-in", plainPath,
		"-aes-256-cbc",
		"-passout", "file:"+pwPath,
		"-out", encPath,
	)
	if err != nil {
		return nil, fmt.Errorf("openssl pkey 加密失败: %s", strings.TrimSpace(stderr))
	}

	enc, err := os.ReadFile(encPath)
	if err != nil {
		return nil, fmt.Errorf("读取加密结果失败: %w", err)
	}
	return enc, nil
}
