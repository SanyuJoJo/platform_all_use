package crypto

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

func NewCertService(
	db *gorm.DB, caller *CoreCaller, files *FileStore,
	parser *CertParser, keys *KeyCrypto,
) *CertService {
	return &CertService{db: db, caller: caller, files: files, parser: parser, keys: keys}
}

// -----------------------------------------------------------------------------
// 列表 / 详情
// -----------------------------------------------------------------------------

func (s *CertService) List(
	page, pageSize int, certType, caID string,
) (map[string]interface{}, error) {
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

// -----------------------------------------------------------------------------
// 签发证书
// -----------------------------------------------------------------------------

func (s *CertService) Sign(
	ctx context.Context, req *SignCertRequest,
) (*SignCertResult, error) {
	if req.CertMode == "" {
		req.CertMode = "normal"
	}
	if req.ValidityDays <= 0 {
		req.ValidityDays = 365
	}
	if req.Algorithm == "" {
		req.Algorithm = "SM2"
	}

	if len(req.Subject) == 0 {
		if subj, err := s.resolveSubjectFromCSR(req); err == nil && len(subj) > 0 {
			req.Subject = subj
		}
	}

	if req.CertMode == "dual" {
		return s.signDualCert(ctx, req)
	}

	params := map[string]interface{}{
		"ca_source":     req.CASource,
		"validity_days": req.ValidityDays,
		"algorithm":     req.Algorithm,
	}
	if len(req.CertTypes) > 0 {
		params["cert_types"] = req.CertTypes
		params["cert_type"] = req.CertTypes[0]
	}
	if req.KeyParams != nil {
		params["key_params"] = req.KeyParams
	}
	if len(req.Subject) > 0 {
		params["subject"] = req.Subject
	}
	if len(req.SAN) > 0 {
		params["san"] = req.SAN
	}

	if err := s.resolveCASource(ctx, req, params); err != nil {
		return nil, err
	}
	if err := s.resolveCSRSource(req, params); err != nil {
		return nil, err
	}

	if _, hasSubj := params["subject"]; !hasSubj {
		if csrPath, ok := params["csr_path"].(string); ok && csrPath != "" {
			if subjMap, err := s.parseCSRSubjectMap(csrPath); err == nil && len(subjMap) > 0 {
				params["subject"] = subjMap
			}
		}
	}

	resp, err := s.caller.Run(ctx, "cert.sign", params)
	if err != nil {
		return nil, err
	}

	cert, err := s.persistCertFromCore(req, resp)
	if err != nil {
		return nil, err
	}

	result := &SignCertResult{Certificate: cert}

	if pem, err := s.files.ReadCoreFile(cert.CertPath); err == nil {
		result.CertPEM = string(pem)
	}

	if req.ReturnKey && cert.KeyRef != nil && *cert.KeyRef != "" {
		if keyPEM, err := s.exportKeyPEM(ctx, *cert.KeyRef, req.KeyExportPassword); err == nil {
			result.KeyPEM = string(keyPEM)
		}
	}

	return result, nil
}

func (s *CertService) resolveSubjectFromCSR(
	req *SignCertRequest,
) (map[string]string, error) {
	switch req.CSRSource {
	case "existing":
		if req.CSRID == "" {
			return nil, fmt.Errorf("missing csr_id")
		}
		var csr models.CSR
		if err := s.db.Where("csr_id = ?", req.CSRID).First(&csr).Error; err != nil {
			return nil, fmt.Errorf("query csr: %w", err)
		}
		return s.parseCSRSubjectMap(csr.CSRPath)

	case "upload":
		if strings.TrimSpace(req.CSRPEM) == "" {
			return nil, fmt.Errorf("missing csr_pem")
		}
		if !strings.Contains(req.CSRPEM, "-----BEGIN CERTIFICATE REQUEST-----") {
			return nil, fmt.Errorf("invalid csr_pem")
		}
		id := newShortID()
		csrRel := fmt.Sprintf("tmp/resolve-subj-%s.csr", id)
		if err := s.files.WriteCoreFile(csrRel, []byte(req.CSRPEM), 0600); err != nil {
			return nil, err
		}
		defer s.files.SafeRemove(csrRel)
		return s.parseCSRSubjectMap(csrRel)
	}
	return nil, fmt.Errorf("no csr source")
}

func (s *CertService) resolveCASource(
	ctx context.Context, req *SignCertRequest, params map[string]interface{},
) error {
	switch req.CASource {
	case "local":
		if req.CAID == "" {
			return exception.New(exception.CodeParamInvalid, "本地 CA 需要 ca_id", 400, nil)
		}
		var ca models.CA
		if err := s.db.Where("ca_id = ? AND status <> ?", req.CAID, "DELETED").
			First(&ca).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
			}
			return exception.New(exception.CodeInternalError, "查询 CA 失败", 500, nil)
		}
		if ca.KeyRef == "" {
			return exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥，无法签发", 400, nil)
		}
		params["ca_id"] = ca.CAID
		params["ca_cert_path"] = ca.CertPath
		params["ca_key_ref"] = ca.KeyRef

	case "manual":
		if strings.TrimSpace(req.CACertPEM) == "" ||
			strings.TrimSpace(req.CAKeyPEM) == "" {
			return exception.New(exception.CodeParamInvalid, "手动 CA 需要证书和私钥", 400, nil)
		}
		if !strings.Contains(req.CACertPEM, "-----BEGIN CERTIFICATE-----") {
			return exception.New(exception.CodeParamInvalid, "CA 证书不是有效 PEM 格式", 400, nil)
		}
		if !strings.Contains(req.CAKeyPEM, "-----BEGIN") {
			return exception.New(exception.CodeParamInvalid, "CA 私钥不是有效 PEM 格式", 400, nil)
		}

		id := newShortID()
		certRel := fmt.Sprintf("tmp/manual-ca-cert-%s.pem", id)
		keyRel := fmt.Sprintf("tmp/manual-ca-key-%s.pem", id)
		if err := s.files.WriteCoreFile(certRel, []byte(req.CACertPEM), 0600); err != nil {
			return err
		}
		if err := s.files.WriteCoreFile(keyRel, []byte(req.CAKeyPEM), 0600); err != nil {
			return err
		}
		params["ca_cert_path"] = certRel
		params["ca_key_path"] = keyRel

		if strings.TrimSpace(req.CAKeyPassword) != "" {
			pwRel := fmt.Sprintf("tmp/manual-ca-pass-%s", id)
			if err := s.files.WriteCoreFile(pwRel, []byte(req.CAKeyPassword), 0600); err != nil {
				return err
			}
			params["ca_key_password_file"] = pwRel
		}

	default:
		return exception.New(exception.CodeParamInvalid, "ca_source 必须是 local 或 manual", 400, nil)
	}
	return nil
}

func (s *CertService) resolveCSRSource(
	req *SignCertRequest, params map[string]interface{},
) error {
	switch req.CSRSource {
	case "":
		return nil

	case "existing":
		if req.CSRID == "" {
			return exception.New(exception.CodeParamInvalid, "缺少 csr_id", 400, nil)
		}
		var csr models.CSR
		if err := s.db.Where("csr_id = ?", req.CSRID).First(&csr).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return exception.New(exception.CodeNotFound, "P10 不存在", 404, nil)
			}
			return exception.New(exception.CodeInternalError, "查询 P10 失败", 500, nil)
		}
		params["csr_id"] = csr.CSRID
		params["csr_path"] = csr.CSRPath
		if csr.KeyRef != nil && *csr.KeyRef != "" {
			params["csr_key_ref"] = *csr.KeyRef
		}
		if _, exists := params["subject"]; !exists {
			if subjMap, err := s.parseCSRSubjectMap(csr.CSRPath); err == nil && len(subjMap) > 0 {
				params["subject"] = subjMap
			}
		}

	case "upload":
		if strings.TrimSpace(req.CSRPEM) == "" {
			return exception.New(exception.CodeParamInvalid, "缺少 CSR PEM", 400, nil)
		}
		if !strings.Contains(req.CSRPEM, "-----BEGIN CERTIFICATE REQUEST-----") {
			return exception.New(exception.CodeParamInvalid, "CSR PEM 格式错误", 400, nil)
		}
		id := newShortID()
		csrRel := fmt.Sprintf("tmp/upload-csr-%s.csr", id)
		if err := s.files.WriteCoreFile(csrRel, []byte(req.CSRPEM), 0600); err != nil {
			return err
		}
		params["csr_path"] = csrRel

		if strings.TrimSpace(req.CSRKeyPEM) != "" {
			if !strings.Contains(req.CSRKeyPEM, "-----BEGIN") {
				return exception.New(exception.CodeParamInvalid, "私钥 PEM 格式错误", 400, nil)
			}
			keyRel := fmt.Sprintf("tmp/upload-csr-key-%s.pem", id)
			if err := s.files.WriteCoreFile(keyRel, []byte(req.CSRKeyPEM), 0600); err != nil {
				return err
			}
			params["csr_key_path"] = keyRel

			if strings.TrimSpace(req.CSRKeyPassword) != "" {
				pwRel := fmt.Sprintf("tmp/upload-csr-pass-%s", id)
				if err := s.files.WriteCoreFile(pwRel, []byte(req.CSRKeyPassword), 0600); err != nil {
					return err
				}
				params["csr_key_password_file"] = pwRel
			}
		}

		if _, exists := params["subject"]; !exists {
			if subjMap, err := s.parseCSRSubjectMap(csrRel); err == nil && len(subjMap) > 0 {
				params["subject"] = subjMap
			}
		}

	default:
		return exception.New(exception.CodeParamInvalid, "csr_source 必须是空、existing 或 upload", 400, nil)
	}
	return nil
}

func (s *CertService) parseCSRSubjectMap(csrRel string) (map[string]string, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, fmt.Errorf("openssl not available")
	}
	abs, err := s.files.guard.Resolve(csrRel, "csr_path")
	if err != nil {
		return nil, err
	}

	out, err := RunOpenSSL(bin, "req", "-in", abs, "-noout", "-subject", "-nameopt", "RFC2253")
	if err != nil {
		return nil, fmt.Errorf("openssl req -subject failed: %w", err)
	}

	subjStr := strings.TrimSpace(out)
	subjStr = strings.TrimSpace(strings.TrimPrefix(subjStr, "subject="))
	if subjStr == "" {
		return nil, fmt.Errorf("empty CSR subject")
	}

	result := make(map[string]string)
	for _, part := range splitDN(subjStr) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(part[:idx])
		val := strings.TrimSpace(part[idx+1:])
		val = strings.ReplaceAll(val, `\,`, ",")
		val = strings.ReplaceAll(val, `\=`, "=")
		val = strings.ReplaceAll(val, `\+`, "+")
		val = strings.ReplaceAll(val, `\\`, `\`)
		result[key] = val
	}
	return result, nil
}

func splitDN(input string) []string {
	var (
		parts   []string
		current strings.Builder
		escaped bool
	)
	for _, r := range input {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			current.WriteRune(r)
			escaped = true
		case r == ',':
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

func (s *CertService) persistCertFromCore(
	req *SignCertRequest, resp *CoreResponse,
) (*models.Certificate, error) {
	certID := getString(resp.Data, "cert_id")
	certPath := getString(resp.Data, "cert_path")
	chainPath := getString(resp.Data, "chain_path")
	serial := getString(resp.Data, "serial")
	keyRef := getString(resp.Data, "key_ref")
	subject := getString(resp.Data, "subject")
	issuer := getString(resp.Data, "issuer")

	if certID == "" || certPath == "" {
		return nil, exception.New(exception.CodeInternalError, "core cert.sign 未返回 cert_id/cert_path", 500, nil)
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
	if certTypeJoined == "" {
		certTypeJoined = "server"
	}

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
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil)
	}
	return cert, nil
}

// -----------------------------------------------------------------------------
// 国密双证
// -----------------------------------------------------------------------------

// signDualCert 国密双证签发。
//
// ★ 与 core 的约定：
//   core 用 CSR 签发签名证书（签名证书公钥 = CSR 公钥）；
//   如果该 P10 有 KeyRef（客户导入 P10 时上传过私钥），
//   签名证书继承该 KeyRef，从而支持"从签名证书解析"模式。
func (s *CertService) signDualCert(
	ctx context.Context, req *SignCertRequest,
) (*SignCertResult, error) {
	if req.Algorithm != "" && req.Algorithm != "SM2" {
		return nil, exception.New(exception.CodeParamInvalid, "国密双证仅支持 SM2", 400, nil)
	}

	params := map[string]interface{}{
		"sign_algorithm": "SM2",
		"enc_algorithm":  "SM2",
		"tlcp_profile":   "GB/T 38636-2020",
		"validity_days":  req.ValidityDays,
	}

	// 传 P10 路径（csr_path 必填）
	if err := s.resolveCSRSource(req, params); err != nil {
		return nil, err
	}

	// 记录 P10 的 key_ref（如果有），用于签名证书继承
	csrKeyRef := ""
	if v, ok := params["csr_key_ref"].(string); ok {
		csrKeyRef = v
	}

	if len(req.Subject) > 0 {
		params["subject"] = req.Subject
	}
	if len(req.SAN) > 0 {
		params["san"] = req.SAN
	}

	if err := s.resolveCASource(ctx, req, params); err != nil {
		return nil, err
	}

	csrPath := ""
	if v, ok := params["csr_path"].(string); ok {
		csrPath = v
	}
	if csrPath == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"国密双证必须提供 P10（csr_path）", 400, nil)
	}

	csrAbs, err := s.files.guard.Resolve(csrPath, "csr_path")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("P10 路径不合法（%s）：%v", csrPath, err), 500, nil)
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	csrPubPEM, err := extractPubKeyFromCSR(bin, csrAbs)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("无法从 P10 提取公钥（%s）：%v", csrAbs, err), 500, nil)
	}
	if len(bytes.TrimSpace(csrPubPEM)) == 0 {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("P10 公钥为空（%s）", csrAbs), 500, nil)
	}

	resp, err := s.caller.Run(ctx, "dual_cert.create", params)
	if err != nil {
		return nil, err
	}

	// ★ 把 csrKeyRef 传给 persistDualCertFromCore，让签名证书继承它
	signCert, err := s.persistDualCertFromCore(req, resp, csrKeyRef)
	if err != nil {
		return nil, err
	}

	result := &SignCertResult{Certificate: signCert}

	signCertPath := getString(resp.Data, "sign_cert_path")
	if signCertPath != "" {
		if pem, err := s.files.ReadCoreFile(signCertPath); err == nil {
			result.SignCertPEM = string(pem)
		}
	}

	encCertPath := getString(resp.Data, "enc_cert_path")
	if encCertPath != "" {
		if pem, err := s.files.ReadCoreFile(encCertPath); err == nil {
			result.EncCertPEM = string(pem)
		}
	}

	if result.SignCertPEM != "" {
		signCertPubPEM, err := extractPubKeyFromCertPEM(bin, result.SignCertPEM)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("无法从签名证书提取公钥：%v", err), 500, nil)
		}
		if !pubKeyEqual(csrPubPEM, signCertPubPEM) {
			return nil, exception.New(exception.CodeInternalError,
				"签名证书公钥与 P10 公钥不一致，core 未按 P10 签发签名证书", 500, nil)
		}
	}

	encKeyRef := getString(resp.Data, "enc_key_ref")
	if encKeyRef == "" {
		return nil, exception.New(exception.CodeInternalError,
			"core 未返回 enc_key_ref", 500, nil)
	}
	defer func() {
		_, _ = s.caller.Run(context.Background(), "key.manage", map[string]interface{}{
			"action":  "delete",
			"key_ref": encKeyRef,
		})
	}()

	encKeyPEM, err := s.caller.ExportKey(ctx, encKeyRef)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("导出加密私钥明文失败：%v", err), 500, nil)
	}

	if len(encKeyPEM) > 0 && result.SignCertPEM != "" {
		envelopeStr, err := s.buildEncryptedEnvelope(
			csrPubPEM,
			signCert.CertID,
			getString(resp.Data, "enc_cert_id"),
			encKeyPEM,
		)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("构建数字信封失败: %v", err), 500, nil)
		}
		result.EncryptedEnvelope = envelopeStr

		if env, err := s.decodeEnvelope(envelopeStr); err == nil {
			record := &models.EnvelopedKeyRecord{
				SignCertID:          signCert.CertID,
				EncCertID:           getString(resp.Data, "enc_cert_id"),
				Format:              "pkcs10",
				Algorithm:           env.Algorithm,
				SignAlg:             env.SignAlg,
				EncAlg:              env.EncAlg,
				SymmetricKeyCipher:  env.SymmetricKeyCipher,
				IV:                  env.IV,
				EncryptedPrivateKey: env.EncryptedPrivateKey,
				CreatedAt:           time.Now().UTC(),
			}
			_ = s.db.Create(record).Error
		}
	}

	return result, nil
}

// buildEncryptedEnvelope 构建数字信封。
//
// ★ 强制使用传入的 pubPEM（即 P10 公钥）加密 SM4 密钥。
func (s *CertService) buildEncryptedEnvelope(
	pubPEM []byte,
	signCertID, encCertID string,
	encKeyPEM []byte,
) (string, error) {
	if len(bytes.TrimSpace(pubPEM)) == 0 {
		return "", fmt.Errorf("加密公钥为空，必须使用 P10 公钥")
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return "", fmt.Errorf("铜锁 openssl 不可用")
	}

	sm4Key := make([]byte, 16)
	iv := make([]byte, 16)
	if _, err := rand.Read(sm4Key); err != nil {
		return "", fmt.Errorf("生成 SM4 密钥失败: %w", err)
	}
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("生成 IV 失败: %w", err)
	}

	symKeyCipher, err := sm2EncryptWithPubKey(bin, pubPEM, sm4Key)
	if err != nil {
		return "", fmt.Errorf("SM2 加密对称密钥失败: %w", err)
	}

	encKeyCipher, err := sm4Encrypt(bin, sm4Key, iv, encKeyPEM)
	if err != nil {
		return "", fmt.Errorf("SM4 加密私钥失败: %w", err)
	}

	env := &envelopedKey{
		Version:             "1.0",
		Algorithm:           "SM2+SM4-CBC",
		SignCertID:          signCertID,
		EncCertID:           encCertID,
		SignAlg:             "SM3withSM2",
		EncAlg:              "SM4-CBC",
		SymmetricKeyCipher:  base64.StdEncoding.EncodeToString(symKeyCipher),
		IV:                  base64.StdEncoding.EncodeToString(iv),
		EncryptedPrivateKey: base64.StdEncoding.EncodeToString(encKeyCipher),
	}

	raw, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("序列化信封失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}

func (s *CertService) decodeEnvelope(raw string) (*envelopedKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("空字符串")
	}

	if strings.HasPrefix(raw, "{") {
		var env envelopedKey
		if err := json.Unmarshal([]byte(raw), &env); err == nil &&
			env.SymmetricKeyCipher != "" && env.IV != "" &&
			env.EncryptedPrivateKey != "" {
			return &env, nil
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		var env envelopedKey
		if err := json.Unmarshal(decoded, &env); err == nil &&
			env.SymmetricKeyCipher != "" && env.IV != "" &&
			env.EncryptedPrivateKey != "" {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("不是有效的加密数字信封")
}

// -----------------------------------------------------------------------------
// 查询信封信息
// -----------------------------------------------------------------------------

// QueryEnvelope 查询信封信息。
//
// 两种模式：
//   - mode=cert   ：从签名证书反查私钥。
//                   若签名证书有 KeyRef（客户导入 P10 时上传过私钥），
//                   从 core 导出私钥解密。
//                   若签名证书没有 KeyRef，返回明确错误引导用 manual 模式。
//   - mode=manual ：客户上传自己的 CSR 私钥 s_pri，Server B 临时解密。
func (s *CertService) QueryEnvelope(
	req *QueryEnvelopeRequest,
) (*QueryEnvelopeResponse, error) {
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "pkcs10"
	}
	if format != "pkcs10" {
		if format == "cfca" {
			return nil, exception.New(exception.CodeParamInvalid,
				"CFCA 证书格式暂未实现", 400, nil)
		}
		return nil, exception.New(exception.CodeParamInvalid,
			"format 必须是 pkcs10 或 cfca", 400, nil)
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "cert"
	}
	if mode != "cert" && mode != "manual" {
		return nil, exception.New(exception.CodeParamInvalid,
			"mode 必须是 cert 或 manual", 400, nil)
	}

	if strings.TrimSpace(req.EncryptedEnvelope) == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"请提供加密的数字信封", 400, nil)
	}

	env, err := s.decodeEnvelope(req.EncryptedEnvelope)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密的数字信封解析失败: %v", err), 400, nil)
	}

	var signKeyPEM string
	var signKeyPassword string

	switch mode {
	case "cert":
		// 从签名证书反查私钥
		if strings.TrimSpace(req.CertID) == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"从签名证书解析需要 cert_id", 400, nil)
		}
		var cert models.Certificate
		if err := s.db.Where("cert_id = ?", req.CertID).First(&cert).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
			}
			return nil, exception.New(exception.CodeInternalError, "查询证书失败", 500, nil)
		}
		if cert.KeyRef == nil || *cert.KeyRef == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"该签名证书未关联私钥（平台不持有客户签名私钥）。"+
					"请改用\"手动输入\"模式，上传您的 CSR 私钥 PEM", 400, nil)
		}
		plain, err := s.caller.ExportKey(context.Background(), *cert.KeyRef)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("导出签名私钥失败: %v", err), 500, nil)
		}
		signKeyPEM = string(plain)

	case "manual":
		if strings.TrimSpace(req.SignKeyPEM) == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"手动输入模式需要提供 sign_key_pem（CSR 私钥 PEM）", 400, nil)
		}
		signKeyPEM = req.SignKeyPEM
		signKeyPassword = req.SignKeyPassword
	}

	plainKey, err := s.decryptEnvelope(
		env.SymmetricKeyCipher, env.IV, env.EncryptedPrivateKey,
		signKeyPEM, signKeyPassword,
	)
	if err != nil {
		return nil, err
	}

	return &QueryEnvelopeResponse{
		Format:              format,
		Mode:                mode,
		SymmetricKeyCipher:  env.SymmetricKeyCipher,
		IV:                  env.IV,
		EncryptedPrivateKey: env.EncryptedPrivateKey,
		DecryptedKeyPEM:     string(plainKey),
	}, nil
}

func (s *CertService) decryptEnvelope(
	symKeyCipherB64, ivB64, encKeyCipherB64 string,
	signKeyPEM, signKeyPassword string,
) ([]byte, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	tmpDir, err := os.MkdirTemp("", "envelope-dec-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建临时目录失败: %v", err), 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	signKeyPath := filepath.Join(tmpDir, "sign.key")
	if err := os.WriteFile(signKeyPath, []byte(signKeyPEM), 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写签名私钥失败: %v", err), 500, nil)
	}

	plainKeyPath := signKeyPath
	if strings.TrimSpace(signKeyPassword) != "" {
		pwPath := filepath.Join(tmpDir, "pw.txt")
		plainPath := filepath.Join(tmpDir, "sign.plain.key")
		if err := os.WriteFile(pwPath, []byte(signKeyPassword), 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("写口令文件失败: %v", err), 500, nil)
		}
		_, stderr, err := RunOpenSSLFull(bin, "pkey",
			"-in", signKeyPath, "-passin", "file:"+pwPath, "-out", plainPath)
		if err != nil {
			return nil, exception.New(exception.CodeParamInvalid,
				fmt.Sprintf("签名私钥口令错误或解密失败: %s", strings.TrimSpace(stderr)), 400, nil)
		}
		plainKeyPath = plainPath
	}

	symKeyCipher, err := base64.StdEncoding.DecodeString(symKeyCipherB64)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("对称密钥密文 base64 解码失败: %v", err), 500, nil)
	}
	sm4Key, err := sm2DecryptWithPrivKey(bin, plainKeyPath, symKeyCipher)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("用签名私钥解密对称密钥失败: %v", err), 400, nil)
	}

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("IV base64 解码失败: %v", err), 500, nil)
	}
	encKeyCipher, err := base64.StdEncoding.DecodeString(encKeyCipherB64)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("加密私钥密文 base64 解码失败: %v", err), 500, nil)
	}
	plainKey, err := sm4Decrypt(bin, sm4Key, iv, encKeyCipher)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4 解密加密私钥失败: %v", err), 500, nil)
	}

	return plainKey, nil
}

// -----------------------------------------------------------------------------
// SM2 / SM4 / 证书解析辅助
// -----------------------------------------------------------------------------

func extractPubKeyFromCertPEM(opensslBin, certPEM string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cert-pub-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "cert.pem")
	if err := os.WriteFile(certPath, []byte(certPEM), 0600); err != nil {
		return nil, fmt.Errorf("写临时证书失败: %w", err)
	}

	out, err := RunOpenSSL(opensslBin, "x509", "-in", certPath, "-noout", "-pubkey")
	if err != nil {
		return nil, fmt.Errorf("openssl x509 -pubkey 失败: %w", err)
	}
	return []byte(out), nil
}

// extractPubKeyFromCSR 从 P10（CSR）文件中提取公钥 PEM。
//
// 参数 csrPath 必须是绝对路径（调用方负责 Resolve）。
func extractPubKeyFromCSR(opensslBin, csrPath string) ([]byte, error) {
	if !filepath.IsAbs(csrPath) {
		return nil, fmt.Errorf("csrPath 必须是绝对路径: %s", csrPath)
	}
	if _, err := os.Stat(csrPath); err != nil {
		return nil, fmt.Errorf("P10 文件不存在: %s: %w", csrPath, err)
	}

	out, err := RunOpenSSL(opensslBin, "req", "-in", csrPath, "-noout", "-pubkey")
	if err != nil {
		return nil, fmt.Errorf("openssl req -pubkey 失败: %w", err)
	}
	return []byte(out), nil
}

// pubKeyEqual 比较两个 PEM 公钥是否等价（忽略换行和首尾空白）。
func pubKeyEqual(a, b []byte) bool {
	na := normalizePEM(a)
	nb := normalizePEM(b)
	return na != "" && na == nb
}

func normalizePEM(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

const sm2DecryptHelperDefaultPath = "/opt/core/run/bin/sm2_decrypt_helper"

func sm2EncryptWithPubKey(opensslBin string, pubKeyPEM []byte, plaintext []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "sm2-enc-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pubPath := filepath.Join(tmpDir, "pub.pem")
	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")

	if err := os.WriteFile(pubPath, pubKeyPEM, 0600); err != nil {
		return nil, fmt.Errorf("写公钥失败: %w", err)
	}
	if err := os.WriteFile(inPath, plaintext, 0600); err != nil {
		return nil, fmt.Errorf("写明文失败: %w", err)
	}

	_, stderr, err := RunOpenSSLFull(opensslBin, "pkeyutl", "-encrypt",
		"-pubin", "-inkey", pubPath,
		"-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl pkeyutl -encrypt 失败: %s", strings.TrimSpace(stderr))
	}

	out, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读取加密结果失败: %w", err)
	}

	return stripSM2C1Header(out), nil
}

func sm2DecryptWithPrivKey(opensslBin, privKeyPath string, ciphertext []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "sm2-dec-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")
	if err := os.WriteFile(inPath, ciphertext, 0600); err != nil {
		return nil, fmt.Errorf("写密文失败: %w", err)
	}

	helperPath := strings.TrimSpace(os.Getenv("SM2_DECRYPT_HELPER"))
	if helperPath == "" {
		helperPath = sm2DecryptHelperDefaultPath
	}

	if _, err := os.Stat(helperPath); err != nil {
		return nil, fmt.Errorf(
			"sm2_decrypt_helper 不存在: %s（请先执行 make -C core/src/tool install）: %w",
			helperPath, err)
	}

	_, stderr, err := RunOpenSSLFull(helperPath,
		"--key", privKeyPath,
		"--in", inPath,
		"--out", outPath,
	)
	if err != nil {
		return nil, fmt.Errorf("sm2_decrypt_helper 失败: %s", strings.TrimSpace(stderr))
	}

	out, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读取解密结果失败: %w", err)
	}
	return out, nil
}

func stripSM2C1Header(b []byte) []byte {
	if len(b) > 0 && b[0] == 0x04 {
		return b[1:]
	}
	return b
}

func sm4Encrypt(opensslBin string, key, iv, plaintext []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "sm4-enc-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")

	if err := os.WriteFile(inPath, plaintext, 0600); err != nil {
		return nil, fmt.Errorf("写明文失败: %w", err)
	}

	keyHex := hex.EncodeToString(key)
	ivHex := hex.EncodeToString(iv)

	_, stderr, err := RunOpenSSLFull(opensslBin, "enc", "-sm4-cbc",
		"-e", "-K", keyHex, "-iv", ivHex, "-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl enc -sm4-cbc 失败: %s", strings.TrimSpace(stderr))
	}

	out, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读取加密结果失败: %w", err)
	}
	return out, nil
}

func sm4Decrypt(opensslBin string, key, iv, ciphertext []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "sm4-dec-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")
	if err := os.WriteFile(inPath, ciphertext, 0600); err != nil {
		return nil, fmt.Errorf("写密文失败: %w", err)
	}

	keyHex := hex.EncodeToString(key)
	ivHex := hex.EncodeToString(iv)
	_, stderr, err := RunOpenSSLFull(opensslBin, "enc", "-sm4-cbc",
		"-d", "-K", keyHex, "-iv", ivHex, "-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl enc -sm4-cbc -d 失败: %s", strings.TrimSpace(stderr))
	}

	out, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读取解密结果失败: %w", err)
	}
	return out, nil
}

// persistDualCertFromCore 国密双证落库。
//
// ★ 参数 csrKeyRef：
//   如果客户导入 P10 时上传过私钥，csrKeyRef 非空。
//   签名证书继承这个 KeyRef，从而支持"从签名证书解析"模式。
//   如果客户没上传私钥，csrKeyRef 为空，签名证书 KeyRef 也为 nil，
//   此时必须用"手动输入"模式解密信封。
func (s *CertService) persistDualCertFromCore(
	req *SignCertRequest, resp *CoreResponse, csrKeyRef string,
) (*models.Certificate, error) {
	signCertPath := getString(resp.Data, "sign_cert_path")
	encCertPath := getString(resp.Data, "enc_cert_path")
	chainPath := getString(resp.Data, "chain_path")

	if signCertPath == "" {
		return nil, exception.New(exception.CodeInternalError,
			"core dual_cert.create 未返回 sign_cert_path", 500, nil)
	}

	now := time.Now().UTC()
	baseID := newShortID()

	signAbs, _ := s.files.guard.Resolve(signCertPath, "cert_path")
	signDetail, _ := s.parser.Parse(signAbs)

	signSubject, signIssuer, signSerial := "", "", ""
	if signDetail != nil {
		signSubject = signDetail.Subject
		signIssuer = signDetail.Issuer
		signSerial = signDetail.Serial
	}

	signCert := &models.Certificate{
		CertID:    "dual-sign-" + baseID,
		CertType:  "dual_sign",
		Serial:    signSerial,
		Subject:   signSubject,
		SubjectCN: extractCNFromSubject(signSubject),
		Issuer:    signIssuer,
		IssuerCN:  extractCNFromSubject(signIssuer),
		CAID:      req.CAID,
		Algorithm: "SM2",
		CertPath:  signCertPath,
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, req.ValidityDays),
		CreatedAt: now,
	}

	// ★ 若 P10 有 KeyRef，签名证书继承它
	if csrKeyRef != "" {
		signCert.KeyRef = &csrKeyRef
	}

	if signDetail != nil {
		signCert.Fingerprint = signDetail.Fingerprint
		signCert.PublicKeyAlgorithm = signDetail.PublicKeyAlgorithm
		signCert.SignatureAlgorithm = signDetail.SignatureAlgorithm
	}
	if chainPath != "" {
		signCert.ChainPath = &chainPath
	}

	if err := s.db.Create(signCert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("双证签名证书落库失败: %v", err), 500, nil)
	}

	if encCertPath != "" {
		encAbs, _ := s.files.guard.Resolve(encCertPath, "cert_path")
		encDetail, _ := s.parser.Parse(encAbs)

		encSubject, encIssuer, encSerial := "", "", ""
		if encDetail != nil {
			encSubject = encDetail.Subject
			encIssuer = encDetail.Issuer
			encSerial = encDetail.Serial
		}

		encCert := &models.Certificate{
			CertID:    "dual-enc-" + baseID,
			CertType:  "dual_enc",
			Serial:    encSerial,
			Subject:   encSubject,
			SubjectCN: extractCNFromSubject(encSubject),
			Issuer:    encIssuer,
			IssuerCN:  extractCNFromSubject(encIssuer),
			CAID:      req.CAID,
			Algorithm: "SM2",
			CertPath:  encCertPath,
			Status:    "VALID",
			NotBefore: now,
			NotAfter:  now.AddDate(0, 0, req.ValidityDays),
			CreatedAt: now,
			KeyRef:    nil, // 加密私钥明文只在信封里
		}
		if encDetail != nil {
			encCert.Fingerprint = encDetail.Fingerprint
			encCert.PublicKeyAlgorithm = encDetail.PublicKeyAlgorithm
			encCert.SignatureAlgorithm = encDetail.SignatureAlgorithm
		}
		if chainPath != "" {
			encCert.ChainPath = &chainPath
		}
		_ = s.db.Create(encCert).Error
	}

	return signCert, nil
}

// -----------------------------------------------------------------------------
// 导入证书
// -----------------------------------------------------------------------------

func (s *CertService) Import(
	ctx context.Context, req *ImportCertRequest,
) (*models.Certificate, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" || !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(exception.CodeParamInvalid, "证书不是有效的 PEM 格式", 400, nil)
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
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败: %v", err), 400, nil)
	}

	var keyRef *string
	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil)
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
					return nil, exception.New(exception.CodeParamInvalid,
						"私钥已加密，请提供私钥密码", 400, nil)
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
					return nil, exception.New(exception.CodeParamInvalid,
						"私钥密码错误或解密失败", 400, nil)
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

	inferredType := inferCertType(detail.KeyUsage, detail.ExtendedKeyUsage, detail.PublicKeyAlgorithm)

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
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil)
	}
	success = true
	return cert, nil
}

func inferCertType(keyUsage, eku, pubAlg string) string {
	kuLower := strings.ToLower(keyUsage)
	hasSig := strings.Contains(kuLower, "digital signature") || strings.Contains(kuLower, "non repudiation")
	hasEnc := strings.Contains(kuLower, "key encipherment") || strings.Contains(kuLower, "key agreement")

	switch {
	case hasSig && hasEnc:
		return "signature,encryption"
	case hasSig:
		return "signature"
	case hasEnc:
		return "encryption"
	}

	ekuLower := strings.ToLower(eku)
	if strings.Contains(ekuLower, "server authentication") ||
		strings.Contains(ekuLower, "client authentication") ||
		strings.Contains(ekuLower, "code signing") {
		return "signature"
	}

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

// -----------------------------------------------------------------------------
// 导出
// -----------------------------------------------------------------------------

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
			return nil, exception.New(exception.CodeParamInvalid, "该证书未关联私钥", 400, nil)
		}
		plain, err := s.caller.ExportKey(ctx, *c.KeyRef)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(password) != "" {
			bin := s.parser.OpensslBin()
			if bin == "" {
				return nil, exception.New(exception.CodeInternalError,
					"铜锁 openssl 不可用，无法加密私钥", 500, nil)
			}
			encrypted, err := encryptPrivateKeyPEM(bin, plain, password)
			if err != nil {
				return nil, exception.New(exception.CodeInternalError,
					fmt.Sprintf("加密私钥失败: %v", err), 500, nil)
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
			return nil, exception.New(exception.CodeParamInvalid, "该证书未关联私钥", 400, nil)
		}
		if len(password) < 6 {
			return nil, exception.New(exception.CodeParamInvalid, "PKCS#12 密码至少 6 位", 400, nil)
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
		return nil, exception.New(exception.CodeParamInvalid,
			"不支持的导出类型: "+exportType, 400, nil)
	}
}

func (s *CertService) Delete(certID string) error {
	res := s.db.Model(&models.Certificate{}).
		Where("cert_id = ? AND status <> ?", certID, "DELETED").
		Updates(map[string]interface{}{"status": "DELETED"})
	if res.Error != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("删除证书失败: %v", res.Error), 500, nil)
	}
	if res.RowsAffected == 0 {
		return exception.New(exception.CodeNotFound, "证书不存在或已删除", 404, nil)
	}
	return nil
}

// -----------------------------------------------------------------------------
// 私钥辅助
// -----------------------------------------------------------------------------

func (s *CertService) exportKeyPEM(
	ctx context.Context, keyRef, password string,
) ([]byte, error) {
	plain, err := s.caller.ExportKey(ctx, keyRef)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(password) == "" {
		return plain, nil
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, fmt.Errorf("铜锁 openssl 不可用，无法加密私钥")
	}
	return encryptPrivateKeyPEM(bin, plain, password)
}

func encryptPrivateKeyPEM(
	opensslBin string, plainPEM []byte, password string,
) ([]byte, error) {
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

	_, stderr, err := RunOpenSSLFull(opensslBin, "pkey",
		"-in", plainPath, "-aes-256-cbc", "-passout", "file:"+pwPath, "-out", encPath)
	if err != nil {
		return nil, fmt.Errorf("openssl pkey 加密失败: %s", strings.TrimSpace(stderr))
	}

	enc, err := os.ReadFile(encPath)
	if err != nil {
		return nil, fmt.Errorf("读取加密结果失败: %w", err)
	}
	return enc, nil
}
