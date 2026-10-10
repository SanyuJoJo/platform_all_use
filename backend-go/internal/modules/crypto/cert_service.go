package crypto

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// =============================================================================
// 类型定义
// =============================================================================

var sm4OID = asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 104}

type algIdentifier struct {
	Algorithm asn1.ObjectIdentifier
}

type gmEnvelope struct {
	KeyAlg       algIdentifier
	SymKeyCipher asn1.RawValue
	EncPubPoint  asn1.BitString
	EncKeyCipher asn1.BitString
}

type legacyEnvelope struct {
	Version             string `json:"version"`
	Algorithm           string `json:"algorithm"`
	SignCertID          string `json:"sign_cert_id,omitempty"`
	EncCertID           string `json:"enc_cert_id,omitempty"`
	SignAlg             string `json:"sign_alg"`
	EncAlg              string `json:"enc_alg"`
	SymmetricKeyCipher  string `json:"symmetric_key_cipher"`
	IV                  string `json:"iv"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

// CertService 证书领域 Service。
type CertService struct {
	db       *gorm.DB
	caller   *CoreCaller
	files    *FileStore
	parser   *CertParser
	keys     *KeyCrypto
	layout   *PathLayout
	whitebox *WhiteboxClient
}

func NewCertService(
	db *gorm.DB, caller *CoreCaller, files *FileStore,
	parser *CertParser, keys *KeyCrypto,
	layout *PathLayout, whitebox *WhiteboxClient,
) *CertService {
	return &CertService{
		db: db, caller: caller, files: files,
		parser: parser, keys: keys,
		layout: layout, whitebox: whitebox,
	}
}

// =============================================================================
// 列表 / 详情
// =============================================================================

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
	abs, err := s.resolveCertPath(c.CertPath)
	if err != nil {
		return nil, err
	}
	return s.parser.Parse(abs)
}

// =============================================================================
// 路径解析
// =============================================================================

// resolveCertPath 解析证书绝对路径。
//
// 优先级：
//  1. 绝对路径，且文件存在
//  2. 按 CertRoot 解析（新规范）
//  3. 按 CoreRoot 解析（旧数据）
func (s *CertService) resolveCertPath(certPath string) (string, error) {
	if certPath == "" {
		return "", exception.New(exception.CodeInternalError, "证书路径为空", 500, nil)
	}
	if filepath.IsAbs(certPath) {
		if st, err := os.Stat(certPath); err == nil && !st.IsDir() {
			return certPath, nil
		}
	}
	if s.layout != nil && s.layout.CertRoot != "" {
		p := filepath.Join(s.layout.CertRoot, certPath)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	if s.caller != nil {
		p := s.caller.ResolveCorePath(certPath)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", exception.New(
		exception.CodeNotFound,
		fmt.Sprintf("证书文件不存在：%s", certPath),
		404, nil,
	)
}

// resolveCSRAbsPath 解析 CSR 绝对路径。
//
// 优先级：
//  1. 绝对路径，且文件存在（新规范：<CertRoot>/server/<dir_no>/<pubkey_sm3>.req.csr）
//  2. 按 CertRoot 解析
//  3. 按 CoreRoot 解析（core/tmp/ 或 core/data/，旧数据）
func (s *CertService) resolveCSRAbsPath(csrPath string) (string, error) {
	if csrPath == "" {
		return "", exception.New(exception.CodeInternalError, "CSR 路径为空", 500, nil)
	}
	if filepath.IsAbs(csrPath) {
		if st, err := os.Stat(csrPath); err == nil && !st.IsDir() {
			return csrPath, nil
		}
	}
	if s.layout != nil && s.layout.CertRoot != "" {
		p := filepath.Join(s.layout.CertRoot, csrPath)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	if s.caller != nil {
		p := s.caller.ResolveCorePath(csrPath)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", exception.New(exception.CodeNotFound,
		fmt.Sprintf("CSR 文件不存在：%s", csrPath), 404, nil)
}

// =============================================================================
// 签发证书（普通）
// =============================================================================

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

	var cleanupRels []string
	defer func() {
		for _, rel := range cleanupRels {
			_ = s.files.SafeRemove(rel)
		}
	}()

	if req.CertMode == "dual" {
		return s.signDualCert(ctx, req, &cleanupRels)
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

	if err := s.resolveCASource(ctx, req, params, &cleanupRels); err != nil {
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

	if err := s.attachOutputLayout(params, req.Domain, ""); err != nil {
		return nil, err
	}

	resp, err := s.caller.Run(ctx, "cert.sign", params)
	if err != nil {
		return nil, err
	}

	cert, err := s.persistCertFromCore(req, resp, params)
	if err != nil {
		return nil, err
	}

	result := &SignCertResult{Certificate: cert}
	if data, err := os.ReadFile(cert.CertPath); err == nil {
		result.CertPEM = string(data)
	} else if pem, err := s.files.ReadCoreFile(cert.CertPath); err == nil {
		result.CertPEM = string(pem)
	}
	if req.ReturnKey && cert.KeyRef != nil && *cert.KeyRef != "" {
		if keyPEM, err := s.exportKeyPEM(ctx, *cert.KeyRef, req.KeyExportPassword); err == nil {
			result.KeyPEM = string(keyPEM)
		}
	}
	return result, nil
}

// attachOutputLayout 为 core 调用附加路径约定参数。
func (s *CertService) attachOutputLayout(
	params map[string]interface{}, domain, existingDirNo string,
) error {
	if s.layout == nil {
		return exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
	}
	dirNo := existingDirNo
	if dirNo == "" {
		var err error
		dirNo, err = s.layout.AllocServerDir()
		if err != nil {
			return exception.New(exception.CodeInternalError,
				fmt.Sprintf("分配 server 目录失败：%v", err), 500, nil)
		}
	}
	dir, err := s.layout.ServerDir(dirNo)
	if err != nil {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("server 目录非法：%v", err), 400, nil)
	}
	params["dir_no"] = dirNo
	params["output_dir"] = dir
	params["output_layout"] = "server_pubkey_sm3"
	params["cert_root"] = s.layout.CertRoot
	if domain != "" {
		params["domain"] = domain
	}
	return nil
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

// resolveCASource 解析 CA 来源并准备参数。
//
// ★ 两种 CA 模式统一走 manual 转换：
//   - 旧模式（KeyRef 非空）：从 keystore 导出私钥 → 拷 core/tmp → 切 manual
//   - 新模式（KeyPath 非空）：解密白盒私钥 → 拷 core/tmp → 切 manual
//
// 原因：core 的 pathguard 只允许 coreRoot/data 与 coreRoot/tmp，
//       而 CA 证书位于 <CertRoot>/ca/<domain>/，必须转换。
//
// cleanup 用于注册需要在请求结束后清理的 core 相对路径。
func (s *CertService) resolveCASource(
	ctx context.Context, req *SignCertRequest, params map[string]interface{},
	cleanup *[]string,
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

		params["ca_id"] = ca.CAID

		// 旧模式：keystore 引用 → 导出 + 拷 core/tmp
		if ca.KeyRef != "" {
			return s.prepareManualCAFromKeystore(ctx, &ca, params, cleanup)
		}
		// 新模式：白盒密文 → 解密 + 拷 core/tmp
		if ca.KeyPath != "" {
			return s.prepareManualCAFromWhitebox(ctx, &ca, params, cleanup)
		}
		return exception.New(exception.CodeParamInvalid,
			"该 CA 未关联私钥，无法签发", 400, nil)

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
		if cleanup != nil {
			*cleanup = append(*cleanup, certRel, keyRel)
		}
		params["ca_cert_path"] = certRel
		params["ca_key_path"] = keyRel
		if strings.TrimSpace(req.CAKeyPassword) != "" {
			pwRel := fmt.Sprintf("tmp/manual-ca-pass-%s", id)
			if err := s.files.WriteCoreFile(pwRel, []byte(req.CAKeyPassword), 0600); err != nil {
				return err
			}
			if cleanup != nil {
				*cleanup = append(*cleanup, pwRel)
			}
			params["ca_key_password_file"] = pwRel
		}
		return nil

	default:
		return exception.New(exception.CodeParamInvalid, "ca_source 必须是 local 或 manual", 400, nil)
	}
}

// prepareManualCAFromKeystore 把 core keystore 中的 CA 私钥导出，
// 连同 CA 证书一起拷到 core/tmp/，并把 params 切换为 manual 模式。
//
// 适用场景：**旧模式 CA**（DB 中 ca.KeyRef 非空、KeyPath 空）。
//
// 核心问题：
//   - 旧模式 CA 的 CertPath 在新规范下是绝对路径（<CertRoot>/ca/<domain>/...），
//     core 的 pathguard 只允许 coreRoot/data 和 coreRoot/tmp；
//   - 所以必须先把证书与私钥拷到 core/tmp/ 再喂给 core；
//   - 私钥通过 core 的 key.manage export 拿到明文（allow_plain_export=true）。
func (s *CertService) prepareManualCAFromKeystore(
	ctx context.Context, ca *models.CA, params map[string]interface{},
	cleanup *[]string,
) error {
	if ca == nil || ca.KeyRef == "" {
		return exception.New(exception.CodeInternalError, "CA keyRef 为空", 500, nil)
	}

	// 1. 从 core keystore 导出私钥明文
	plainKey, err := s.caller.ExportKey(ctx, ca.KeyRef)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("导出 CA 私钥失败：%v", err), 500, nil)
	}
	if len(plainKey) == 0 {
		return exception.New(exception.CodeInternalError, "导出的 CA 私钥为空", 500, nil)
	}

	id := newShortID()

	// 2. 拷私钥到 core/tmp/
	coreKeyRel := fmt.Sprintf("tmp/ca-sign-key-%s.pem", id)
	if err := s.files.WriteCoreFile(coreKeyRel, plainKey, 0600); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("写 core 临时私钥失败：%v", err), 500, nil)
	}
	if cleanup != nil {
		*cleanup = append(*cleanup, coreKeyRel)
	}

	// 3. 拷 CA 证书到 core/tmp/
	certAbs, err := s.resolveCACertPathForSign(ca.CertPath)
	if err != nil {
		return err
	}
	certData, err := os.ReadFile(certAbs)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取 CA 证书失败：%v", err), 500, nil)
	}
	coreCertRel := fmt.Sprintf("tmp/ca-sign-cert-%s.pem", id)
	if err := s.files.WriteCoreFile(coreCertRel, certData, 0640); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("写 core 临时证书失败：%v", err), 500, nil)
	}
	if cleanup != nil {
		*cleanup = append(*cleanup, coreCertRel)
	}

	// 4. 切 manual 模式
	params["ca_source"] = "manual"
	params["ca_cert_path"] = coreCertRel
	params["ca_key_path"] = coreKeyRel

	log.Info().
		Str("ca_id", ca.CAID).
		Str("core_cert_rel", coreCertRel).
		Str("core_key_rel", coreKeyRel).
		Msg("CA keystore 私钥已导出并转入 core/tmp（manual 模式）")
	return nil
}

// prepareManualCAFromWhitebox 把白盒 CA 私钥解密并转入 core/tmp/，
// 使 params 切换为 manual 模式。
//
// 适用场景：**新模式 CA**（DB 中 ca.KeyPath 非空、KeyRef 空）。
func (s *CertService) prepareManualCAFromWhitebox(
	ctx context.Context, ca *models.CA, params map[string]interface{},
	cleanup *[]string,
) error {
	keyAbs := ca.KeyPath
	if !filepath.IsAbs(keyAbs) {
		keyAbs = filepath.Join(s.layout.CertRoot, keyAbs)
	}
	if _, err := os.Stat(keyAbs); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("CA 白盒私钥文件不存在：%s", keyAbs), 500, nil)
	}

	plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("解密 CA 白盒私钥失败：%v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	plainData, err := os.ReadFile(plainTmp)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取解密私钥失败：%v", err), 500, nil)
	}

	id := newShortID()

	coreKeyRel := fmt.Sprintf("tmp/ca-sign-key-%s.pem", id)
	if err := s.files.WriteCoreFile(coreKeyRel, plainData, 0600); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("写 core 临时私钥失败：%v", err), 500, nil)
	}
	if cleanup != nil {
		*cleanup = append(*cleanup, coreKeyRel)
	}

	certAbs, err := s.resolveCACertPathForSign(ca.CertPath)
	if err != nil {
		return err
	}
	certData, err := os.ReadFile(certAbs)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取 CA 证书失败：%v", err), 500, nil)
	}
	coreCertRel := fmt.Sprintf("tmp/ca-sign-cert-%s.pem", id)
	if err := s.files.WriteCoreFile(coreCertRel, certData, 0640); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("写 core 临时证书失败：%v", err), 500, nil)
	}
	if cleanup != nil {
		*cleanup = append(*cleanup, coreCertRel)
	}

	params["ca_source"] = "manual"
	params["ca_cert_path"] = coreCertRel
	params["ca_key_path"] = coreKeyRel

	log.Info().
		Str("ca_id", ca.CAID).
		Str("core_cert_rel", coreCertRel).
		Str("core_key_rel", coreKeyRel).
		Msg("CA 白盒私钥已解密并转入 core/tmp（manual 模式）")
	return nil
}

func (s *CertService) resolveCACertPathForSign(p string) (string, error) {
	if p == "" {
		return "", exception.New(exception.CodeInternalError, "CA 证书路径为空", 500, nil)
	}
	if filepath.IsAbs(p) {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if s.layout != nil && s.layout.CertRoot != "" {
		abs := filepath.Join(s.layout.CertRoot, p)
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	if s.caller != nil {
		abs := s.caller.ResolveCorePath(p)
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	return "", exception.New(exception.CodeNotFound,
		fmt.Sprintf("CA 证书文件不存在：%s", p), 404, nil)
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
		params["csr_source"] = "existing"
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
		params["csr_source"] = "upload"
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

// parseCSRSubjectMap ★ 使用 resolveCSRAbsPath，兼容 <CertRoot>/server/... 下 CSR
func (s *CertService) parseCSRSubjectMap(csrRel string) (map[string]string, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, fmt.Errorf("openssl not available")
	}
	abs, err := s.resolveCSRAbsPath(csrRel)
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

// extractCNFromParams 从 params["subject"] 里提取 CN。
func extractCNFromParams(params map[string]interface{}) string {
	raw, ok := params["subject"]
	if !ok || raw == nil {
		return ""
	}
	switch m := raw.(type) {
	case map[string]string:
		for _, k := range []string{"CN", "cn", "CommonName", "common_name"} {
			if v, ok := m[k]; ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	case map[string]interface{}:
		for _, k := range []string{"CN", "cn", "CommonName", "common_name"} {
			if v, ok := m[k].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

// persistCertFromCore ★ SubjectCN 多来源兜底
func (s *CertService) persistCertFromCore(
	req *SignCertRequest, resp *CoreResponse, params map[string]interface{},
) (*models.Certificate, error) {
	certID := getString(resp.Data, "cert_id")
	certPath := getString(resp.Data, "cert_path")
	keyPath := getString(resp.Data, "key_path")
	serial := getString(resp.Data, "serial")
	keyRef := getString(resp.Data, "key_ref")
	subject := getString(resp.Data, "subject")
	issuer := getString(resp.Data, "issuer")
	pubSM3 := getString(resp.Data, "pubkey_sm3")
	dirNo := getString(resp.Data, "dir_no")

	if dirNo == "" {
		if v, ok := params["dir_no"].(string); ok {
			dirNo = v
		}
	}
	if certPath == "" && pubSM3 != "" && dirNo != "" {
		_, cPath, _, _, _, err := s.layout.ServerPaths(dirNo, pubSM3)
		if err == nil {
			certPath = cPath
		}
	}
	if certPath == "" && req.CSRSource == "existing" {
		if csrPath, ok := params["csr_path"].(string); ok && csrPath != "" && pubSM3 != "" {
			csrDir := filepath.Dir(csrPath)
			certPath = filepath.Join(csrDir, pubSM3+".cert.pem")
		}
	}
	if certID == "" || certPath == "" {
		return nil, exception.New(exception.CodeInternalError,
			"core cert.sign 未返回 cert_id/cert_path", 500, nil)
	}

	abs, _ := s.resolveCertPath(certPath)
	var detail *CertDetail
	if abs != "" {
		detail, _ = s.parser.Parse(abs)
	}

	if subject == "" {
		if cn := extractCNFromParams(params); cn != "" {
			subject = "CN=" + cn
		}
	}
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
		SubjectCN: cnFromSubject(subject),
		Issuer:    issuer,
		IssuerCN:  cnFromSubject(issuer),
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
	if keyPath != "" {
		cert.KeyPath = &keyPath
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

// =============================================================================
// 国密双证签发
// =============================================================================

func (s *CertService) signDualCert(
	ctx context.Context, req *SignCertRequest, cleanup *[]string,
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

	if err := s.resolveCSRSource(req, params); err != nil {
		return nil, err
	}
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
	if err := s.resolveCASource(ctx, req, params, cleanup); err != nil {
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

	// ★ 用 resolveCSRAbsPath，兼容 <CertRoot>/server/... 下的 CSR
	csrAbs, err := s.resolveCSRAbsPath(csrPath)
	if err != nil {
		return nil, err
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

	signDirNo := ""
	if rel, err := filepath.Rel(s.layout.ServerRoot, filepath.Dir(csrAbs)); err == nil {
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		if len(parts) >= 1 && len(parts[0]) == 4 {
			signDirNo = parts[0]
		}
	}
	if signDirNo == "" {
		if err := s.attachOutputLayout(params, req.Domain, ""); err != nil {
			return nil, err
		}
	} else {
		signDir, _ := s.layout.ServerDir(signDirNo)
		params["sign_dir_no"] = signDirNo
		params["sign_output_dir"] = signDir
	}

	encDirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("分配加密证书目录失败：%v", err), 500, nil)
	}
	encDir, _ := s.layout.ServerDir(encDirNo)
	params["enc_dir_no"] = encDirNo
	params["enc_output_dir"] = encDir
	params["output_layout"] = "server_pubkey_sm3"
	params["cert_root"] = s.layout.CertRoot
	if req.Domain != "" {
		params["domain"] = req.Domain
	}

	resp, err := s.caller.Run(ctx, "dual_cert.create", params)
	if err != nil {
		return nil, err
	}

	signCert, err := s.persistDualCertFromCore(req, resp, csrKeyRef, params)
	if err != nil {
		return nil, err
	}

	result := &SignCertResult{Certificate: signCert}

	signCertPath := getString(resp.Data, "sign_cert_path")
	if signCertPath != "" {
		if data, err := os.ReadFile(signCertPath); err == nil {
			result.SignCertPEM = string(data)
		} else if pem, err := s.files.ReadCoreFile(signCertPath); err == nil {
			result.SignCertPEM = string(pem)
		}
	}
	encCertPath := getString(resp.Data, "enc_cert_path")
	if encCertPath != "" {
		if data, err := os.ReadFile(encCertPath); err == nil {
			result.EncCertPEM = string(data)
		} else if pem, err := s.files.ReadCoreFile(encCertPath); err == nil {
			result.EncCertPEM = string(pem)
		}
	}

	if result.SignCertPEM != "" {
		signCertPubPEM, err := extractPubKeyFromCertPEM(bin, result.SignCertPEM)
		if err != nil {
			log.Warn().Err(err).
				Str("csr_path", csrPath).
				Str("sign_cert_path", signCertPath).
				Msg("从签名证书提取公钥失败（不阻断签发）")
		} else if !pubKeyEqual(csrPubPEM, signCertPubPEM) {
			log.Warn().
				Str("csr_path", csrPath).
				Str("sign_cert_path", signCertPath).
				Str("csr_pub_sha256", pubKeySHA256(csrPubPEM)).
				Str("sign_cert_pub_sha256", pubKeySHA256(signCertPubPEM)).
				Msg("签名证书公钥指纹与 P10 不一致（已忽略，不阻断签发）")
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

	var encKeyAbsPath string
	if len(encKeyPEM) > 0 && encCertPath != "" {
		encKeyAbsPath, err = s.persistEncKeyWhitebox(ctx, encDirNo, encDir, encCertPath, encKeyPEM)
		if err != nil {
			return nil, err
		}
		var encCert models.Certificate
		if e := s.db.Where("cert_id LIKE ?", "dual-enc-%").
			Order("id DESC").First(&encCert).Error; e == nil {
			encCert.KeyPath = &encKeyAbsPath
			_ = s.db.Save(&encCert).Error
		}
	}

	if len(encKeyPEM) > 0 && result.EncCertPEM != "" {
		envelopeStr, err := s.buildEnvelopeForDual(
			csrPubPEM,
			[]byte(result.EncCertPEM),
			encKeyPEM,
		)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("构建数字信封失败: %v", err), 500, nil)
		}
		result.EncryptedEnvelope = envelopeStr

		if env, _, err := s.decodeEnvelopeAny(envelopeStr); err == nil && env != nil {
			record := &models.EnvelopedKeyRecord{
				SignCertID:          signCert.CertID,
				EncCertID:           getString(resp.Data, "enc_cert_id"),
				Format:              "pkcs10",
				Algorithm:           "SM2+SM4-ECB",
				SignAlg:             "SM3withSM2",
				EncAlg:              "SM4-ECB",
				SymmetricKeyCipher:  base64.StdEncoding.EncodeToString(env.SymKeyCipher.FullBytes),
				IV:                  "",
				EncryptedPrivateKey: base64.StdEncoding.EncodeToString(env.EncKeyCipher.Bytes),
				CreatedAt:           time.Now().UTC(),
			}
			_ = s.db.Create(record).Error
		}
	}

	return result, nil
}

func (s *CertService) persistEncKeyWhitebox(
	ctx context.Context, encDirNo, encDir, encCertPath string, encKeyPEM []byte,
) (string, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return "", exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}
	encCertPEM, err := os.ReadFile(encCertPath)
	if err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取加密证书失败：%v", err), 500, nil)
	}
	encPubSM3, err := PubkeySM3FromCertPEM(bin, string(encCertPEM))
	if err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("计算加密证书公钥 SM3 失败：%v", err), 500, nil)
	}

	encKeyPath := filepath.Join(encDir, encPubSM3+".key.pem")

	plainTmp := filepath.Join(encDir, ".enc.key.plain.tmp")
	if err := os.WriteFile(plainTmp, encKeyPEM, 0600); err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	if err := s.whitebox.EncryptFile(ctx, plainTmp, encKeyPath); err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
	}
	_ = os.Chmod(encKeyPath, 0600)

	log.Info().
		Str("dir_no", encDirNo).
		Str("enc_pub_sm3", encPubSM3).
		Str("key_path", encKeyPath).
		Msg("加密私钥白盒加密落盘完成")
	return encKeyPath, nil
}

func (s *CertService) persistDualCertFromCore(
	req *SignCertRequest, resp *CoreResponse, csrKeyRef string,
	params map[string]interface{},
) (*models.Certificate, error) {
	signCertPath := getString(resp.Data, "sign_cert_path")
	encCertPath := getString(resp.Data, "enc_cert_path")
	signPubSM3 := getString(resp.Data, "sign_pubkey_sm3")
	encPubSM3 := getString(resp.Data, "enc_pubkey_sm3")

	signDirNo := getString(resp.Data, "sign_dir_no")
	if signDirNo == "" {
		if v, ok := params["sign_dir_no"].(string); ok {
			signDirNo = v
		}
	}
	encDirNo := getString(resp.Data, "enc_dir_no")
	if encDirNo == "" {
		if v, ok := params["enc_dir_no"].(string); ok {
			encDirNo = v
		}
	}

	if signCertPath == "" && signPubSM3 != "" && signDirNo != "" {
		_, cPath, _, _, _, err := s.layout.ServerPaths(signDirNo, signPubSM3)
		if err == nil {
			signCertPath = cPath
		}
	}
	if encCertPath == "" && encPubSM3 != "" && encDirNo != "" {
		_, cPath, _, _, _, err := s.layout.ServerPaths(encDirNo, encPubSM3)
		if err == nil {
			encCertPath = cPath
		}
	}
	if signCertPath == "" {
		return nil, exception.New(exception.CodeInternalError,
			"core dual_cert.create 未返回 sign_cert_path", 500, nil)
	}

	now := time.Now().UTC()
	baseID := newShortID()

	signAbs, _ := s.resolveCertPath(signCertPath)
	var signDetail *CertDetail
	if signAbs != "" {
		signDetail, _ = s.parser.Parse(signAbs)
	}

	signSubject, signIssuer, signSerial := "", "", ""
	if signDetail != nil {
		signSubject = signDetail.Subject
		signIssuer = signDetail.Issuer
		signSerial = signDetail.Serial
	}
	if signSubject == "" {
		if cn := extractCNFromParams(params); cn != "" {
			signSubject = "CN=" + cn
		}
	}

	signCert := &models.Certificate{
		CertID:    "dual-sign-" + baseID,
		CertType:  "dual_sign",
		Serial:    signSerial,
		Subject:   signSubject,
		SubjectCN: cnFromSubject(signSubject),
		Issuer:    signIssuer,
		IssuerCN:  cnFromSubject(signIssuer),
		CAID:      req.CAID,
		Algorithm: "SM2",
		CertPath:  signCertPath,
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, req.ValidityDays),
		CreatedAt: now,
	}
	if csrKeyRef != "" {
		signCert.KeyRef = &csrKeyRef
	}
	if signDetail != nil {
		signCert.Fingerprint = signDetail.Fingerprint
		signCert.PublicKeyAlgorithm = signDetail.PublicKeyAlgorithm
		signCert.SignatureAlgorithm = signDetail.SignatureAlgorithm
	}
	if err := s.db.Create(signCert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("双证签名证书落库失败: %v", err), 500, nil)
	}

	if encCertPath != "" {
		encAbs, _ := s.resolveCertPath(encCertPath)
		var encDetail *CertDetail
		if encAbs != "" {
			encDetail, _ = s.parser.Parse(encAbs)
		}

		encSubject, encIssuer, encSerial := "", "", ""
		if encDetail != nil {
			encSubject = encDetail.Subject
			encIssuer = encDetail.Issuer
			encSerial = encDetail.Serial
		}
		if encSubject == "" {
			if cn := extractCNFromParams(params); cn != "" {
				encSubject = "CN=" + cn
			}
		}

		encCert := &models.Certificate{
			CertID:    "dual-enc-" + baseID,
			CertType:  "dual_enc",
			Serial:    encSerial,
			Subject:   encSubject,
			SubjectCN: cnFromSubject(encSubject),
			Issuer:    encIssuer,
			IssuerCN:  cnFromSubject(encIssuer),
			CAID:      req.CAID,
			Algorithm: "SM2",
			CertPath:  encCertPath,
			Status:    "VALID",
			NotBefore: now,
			NotAfter:  now.AddDate(0, 0, req.ValidityDays),
			CreatedAt: now,
		}
		if encDetail != nil {
			encCert.Fingerprint = encDetail.Fingerprint
			encCert.PublicKeyAlgorithm = encDetail.PublicKeyAlgorithm
			encCert.SignatureAlgorithm = encDetail.SignatureAlgorithm
		}
		_ = s.db.Create(encCert).Error
	}

	return signCert, nil
}

// =============================================================================
// 数字信封：构建
// =============================================================================

func (s *CertService) buildEnvelopeForDual(
	csrPubPEM []byte, encCertPEM []byte, encKeyPEM []byte,
) (string, error) {
	der, err := s.buildEncryptedEnvelopeDER(csrPubPEM, encCertPEM, encKeyPEM)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

func (s *CertService) buildEncryptedEnvelopeDER(
	csrPubPEM []byte, encCertPEM []byte, encKeyPEM []byte,
) ([]byte, error) {
	if len(bytes.TrimSpace(csrPubPEM)) == 0 {
		return nil, fmt.Errorf("加密公钥为空，必须使用 P10 公钥")
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, fmt.Errorf("铜锁 openssl 不可用")
	}

	encPubPoint, err := extractEncCertECPoint(bin, encCertPEM)
	if err != nil {
		return nil, fmt.Errorf("提取加密证书公钥点失败: %w", err)
	}
	if len(encPubPoint) != 65 || encPubPoint[0] != 0x04 {
		return nil, fmt.Errorf("加密证书公钥点格式异常（len=%d, first=%02x）",
			len(encPubPoint), encPubPoint[0])
	}

	rawEncKey, err := extractRawECPrivateKey(bin, encKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("提取加密私钥裸字节失败: %w", err)
	}
	if len(rawEncKey) != 32 {
		return nil, fmt.Errorf("加密私钥裸字节长度异常：%d", len(rawEncKey))
	}

	sm4Key := make([]byte, 16)
	if _, err := rand.Read(sm4Key); err != nil {
		return nil, fmt.Errorf("生成 SM4 密钥失败: %w", err)
	}

	symKeyCipherASN1, err := sm2EncryptWithPubKey(bin, csrPubPEM, sm4Key)
	if err != nil {
		return nil, fmt.Errorf("SM2 加密对称密钥失败: %w", err)
	}
	if len(symKeyCipherASN1) == 0 || symKeyCipherASN1[0] != 0x30 {
		return nil, fmt.Errorf("SM2 密文不是 SEQUENCE（首字节 %02x）", symKeyCipherASN1[0])
	}

	encKeyCipher, err := sm4EncryptECBNoPadding(bin, sm4Key, rawEncKey)
	if err != nil {
		return nil, fmt.Errorf("SM4 加密加密私钥失败: %w", err)
	}

	env := gmEnvelope{
		KeyAlg:       algIdentifier{Algorithm: sm4OID},
		SymKeyCipher: asn1.RawValue{FullBytes: symKeyCipherASN1},
		EncPubPoint: asn1.BitString{
			Bytes: encPubPoint, BitLength: len(encPubPoint) * 8,
		},
		EncKeyCipher: asn1.BitString{
			Bytes: encKeyCipher, BitLength: len(encKeyCipher) * 8,
		},
	}
	der, err := asn1.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("序列化信封失败: %w", err)
	}
	return der, nil
}

// =============================================================================
// 数字信封：解析
// =============================================================================

func (s *CertService) decodeEnvelopeAny(raw string) (*gmEnvelope, *legacyEnvelope, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil, fmt.Errorf("空字符串")
	}
	var decoded []byte
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil {
		decoded = b
	} else {
		decoded = []byte(raw)
	}
	if len(decoded) > 0 && decoded[0] == 0x30 {
		var env gmEnvelope
		if _, err := asn1.Unmarshal(decoded, &env); err == nil &&
			len(env.SymKeyCipher.FullBytes) > 0 &&
			len(env.EncKeyCipher.Bytes) > 0 {
			return &env, nil, nil
		}
	}
	if len(decoded) > 0 && decoded[0] == '{' {
		var leg legacyEnvelope
		if err := json.Unmarshal(decoded, &leg); err == nil &&
			leg.SymmetricKeyCipher != "" && leg.EncryptedPrivateKey != "" {
			return nil, &leg, nil
		}
	}
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil &&
		len(b) > 0 && b[0] == '{' {
		var leg legacyEnvelope
		if err := json.Unmarshal(b, &leg); err == nil &&
			leg.SymmetricKeyCipher != "" && leg.EncryptedPrivateKey != "" {
			return nil, &leg, nil
		}
	}
	return nil, nil, fmt.Errorf("不是有效的加密数字信封")
}

// =============================================================================
// 查询信封信息
// =============================================================================

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
	env, leg, err := s.decodeEnvelopeAny(req.EncryptedEnvelope)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密的数字信封解析失败: %v", err), 400, nil)
	}

	var signKeyPEM string
	var signKeyPassword string
	switch mode {
	case "cert":
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

	plainKey, err := s.decryptEnvelopeAuto(req.EncryptedEnvelope, signKeyPEM, signKeyPassword)
	if err != nil {
		return nil, err
	}
	resp := &QueryEnvelopeResponse{
		Format:          format,
		Mode:            mode,
		DecryptedKeyPEM: string(plainKey),
	}
	if env != nil {
		resp.SymmetricKeyCipher = base64.StdEncoding.EncodeToString(env.SymKeyCipher.FullBytes)
		resp.EncryptedPrivateKey = base64.StdEncoding.EncodeToString(env.EncKeyCipher.Bytes)
	} else if leg != nil {
		resp.SymmetricKeyCipher = leg.SymmetricKeyCipher
		resp.IV = leg.IV
		resp.EncryptedPrivateKey = leg.EncryptedPrivateKey
	}
	return resp, nil
}

func (s *CertService) decryptEnvelopeAuto(
	raw string, signKeyPEM, signKeyPassword string,
) ([]byte, error) {
	env, leg, err := s.decodeEnvelopeAny(raw)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密的数字信封解析失败: %v", err), 400, nil)
	}
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

	switch {
	case env != nil:
		return s.decryptNewEnvelope(bin, tmpDir, plainKeyPath, env)
	case leg != nil:
		return s.decryptLegacyEnvelope(bin, plainKeyPath, leg)
	}
	return nil, exception.New(exception.CodeInternalError, "信封格式未知", 500, nil)
}

func (s *CertService) decryptNewEnvelope(
	bin, tmpDir, plainKeyPath string, env *gmEnvelope,
) ([]byte, error) {
	if len(env.SymKeyCipher.FullBytes) == 0 {
		return nil, exception.New(exception.CodeInternalError, "SM2 密文为空", 500, nil)
	}
	cipherPath := filepath.Join(tmpDir, "sm2.cipher")
	if err := os.WriteFile(cipherPath, env.SymKeyCipher.FullBytes, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写 SM2 密文失败: %v", err), 500, nil)
	}
	sm4KeyPath := filepath.Join(tmpDir, "sm4.key")
	_, stderr, err := RunOpenSSLFull(bin, "pkeyutl", "-decrypt",
		"-inkey", plainKeyPath, "-in", cipherPath, "-out", sm4KeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("SM2 解密对称密钥失败: %s", strings.TrimSpace(stderr)), 400, nil)
	}
	sm4Key, err := os.ReadFile(sm4KeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("读 SM4 密钥失败: %v", err), 500, nil)
	}
	if len(sm4Key) != 16 {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4 密钥长度异常：%d", len(sm4Key)), 500, nil)
	}
	encKeyPath := filepath.Join(tmpDir, "enc.key.cipher")
	if err := os.WriteFile(encKeyPath, env.EncKeyCipher.Bytes, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写加密私钥密文失败: %v", err), 500, nil)
	}
	decKeyPath := filepath.Join(tmpDir, "enc.key.plain")
	keyHex := hex.EncodeToString(sm4Key)
	_, stderr, err = RunOpenSSLFull(bin, "enc", "-sm4-ecb",
		"-d", "-K", keyHex, "-nopad",
		"-in", encKeyPath, "-out", decKeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4-ECB 解密失败: %s", strings.TrimSpace(stderr)), 500, nil)
	}
	return os.ReadFile(decKeyPath)
}

func (s *CertService) decryptLegacyEnvelope(
	bin, plainKeyPath string, leg *legacyEnvelope,
) ([]byte, error) {
	symCipher, err := base64.StdEncoding.DecodeString(leg.SymmetricKeyCipher)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("对称密钥密文 base64 解码失败: %v", err), 500, nil)
	}
	tmpDir, _ := os.MkdirTemp("", "legacy-dec-")
	defer os.RemoveAll(tmpDir)

	cipherPath := filepath.Join(tmpDir, "sm2.cipher")
	if err := os.WriteFile(cipherPath, symCipher, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写 SM2 密文失败", 500, nil)
	}
	sm4KeyPath := filepath.Join(tmpDir, "sm4.key")
	_, stderr, err := RunOpenSSLFull(bin, "pkeyutl", "-decrypt",
		"-inkey", plainKeyPath, "-in", cipherPath, "-out", sm4KeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("SM2 解密对称密钥失败: %s", strings.TrimSpace(stderr)), 400, nil)
	}
	sm4Key, _ := os.ReadFile(sm4KeyPath)
	if len(sm4Key) != 16 {
		return nil, exception.New(exception.CodeInternalError, "SM4 密钥长度异常", 500, nil)
	}
	iv, err := base64.StdEncoding.DecodeString(leg.IV)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("IV base64 解码失败: %v", err), 500, nil)
	}
	encKeyCipher, err := base64.StdEncoding.DecodeString(leg.EncryptedPrivateKey)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("加密私钥密文 base64 解码失败: %v", err), 500, nil)
	}
	encKeyPath := filepath.Join(tmpDir, "enc.key.cipher")
	decKeyPath := filepath.Join(tmpDir, "enc.key.plain")
	if err := os.WriteFile(encKeyPath, encKeyCipher, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写加密私钥密文失败", 500, nil)
	}
	keyHex := hex.EncodeToString(sm4Key)
	ivHex := hex.EncodeToString(iv)
	_, stderr, err = RunOpenSSLFull(bin, "enc", "-sm4-cbc",
		"-d", "-K", keyHex, "-iv", ivHex, "-in", encKeyPath, "-out", decKeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4-CBC 解密失败: %s", strings.TrimSpace(stderr)), 500, nil)
	}
	return os.ReadFile(decKeyPath)
}

// =============================================================================
// SM2 / SM4 底层原语
// =============================================================================

func sm2EncryptWithPubKey(opensslBin string, pubPEM, plaintext []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "sm2-enc-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	pubPath := filepath.Join(tmpDir, "pub.pem")
	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")

	if err := os.WriteFile(pubPath, pubPEM, 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(inPath, plaintext, 0600); err != nil {
		return nil, err
	}
	_, stderr, err := RunOpenSSLFull(opensslBin, "pkeyutl", "-encrypt",
		"-pubin", "-inkey", pubPath,
		"-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl pkeyutl -encrypt 失败: %s", strings.TrimSpace(stderr))
	}
	return os.ReadFile(outPath)
}

func sm4EncryptECBNoPadding(opensslBin string, key, plaintext []byte) ([]byte, error) {
	if len(plaintext)%16 != 0 {
		return nil, fmt.Errorf("SM4-ECB/NoPadding 明文长度必须是 16 的倍数，当前 %d", len(plaintext))
	}
	tmpDir, err := os.MkdirTemp("", "sm4-ecb-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")
	if err := os.WriteFile(inPath, plaintext, 0600); err != nil {
		return nil, err
	}
	keyHex := hex.EncodeToString(key)
	_, stderr, err := RunOpenSSLFull(opensslBin, "enc", "-sm4-ecb",
		"-e", "-K", keyHex, "-nopad",
		"-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl enc -sm4-ecb 失败: %s", strings.TrimSpace(stderr))
	}
	return os.ReadFile(outPath)
}

func extractEncCertECPoint(opensslBin string, certPEM []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "enc-point-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	certPath := filepath.Join(tmpDir, "cert.pem")
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		return nil, err
	}
	pubPEM, err := RunOpenSSL(opensslBin, "x509", "-in", certPath, "-noout", "-pubkey")
	if err != nil {
		return nil, fmt.Errorf("openssl x509 -pubkey 失败: %w", err)
	}
	pubPath := filepath.Join(tmpDir, "pub.pem")
	if err := os.WriteFile(pubPath, []byte(pubPEM), 0600); err != nil {
		return nil, err
	}
	derOut, stderr, err := RunOpenSSLFull(opensslBin, "ec", "-pubin",
		"-in", pubPath, "-conv_form", "uncompressed", "-outform", "DER")
	if err != nil {
		return nil, fmt.Errorf("openssl ec -conv_form 失败: %s", strings.TrimSpace(stderr))
	}
	var spki struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}
	if _, err := asn1.Unmarshal([]byte(derOut), &spki); err != nil {
		return nil, fmt.Errorf("解析 SPKI 失败: %w", err)
	}
	return spki.PublicKey.Bytes, nil
}

func extractRawECPrivateKey(opensslBin string, keyPEM []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "raw-ec-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	keyPath := filepath.Join(tmpDir, "key.pem")
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, err
	}
	out, stderr, err := RunOpenSSLFull(opensslBin, "ec", "-in", keyPath, "-noout", "-text")
	if err != nil {
		out, stderr, err = RunOpenSSLFull(opensslBin, "pkey", "-in", keyPath, "-noout", "-text")
		if err != nil {
			return nil, fmt.Errorf("提取裸私钥失败: %s", strings.TrimSpace(stderr))
		}
	}
	lines := strings.Split(out, "\n")
	var inPriv bool
	var hexBuf strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "priv:") {
			inPriv = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "priv:"))
			if rest != "" {
				hexBuf.WriteString(rest)
			}
			continue
		}
		if inPriv {
			if trimmed == "" || strings.HasPrefix(trimmed, "pub:") {
				break
			}
			hexBuf.WriteString(trimmed)
		}
	}
	hexStr := hexBuf.String()
	for _, r := range []string{":", " ", "\n", "\r", "\t"} {
		hexStr = strings.ReplaceAll(hexStr, r, "")
	}
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("解析私钥 hex 失败: %w", err)
	}
	for len(raw) > 32 && raw[0] == 0 {
		raw = raw[1:]
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("私钥长度异常：%d（应为 32）", len(raw))
	}
	return raw, nil
}

// =============================================================================
// 公钥提取 / 比对
// =============================================================================

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

func pubKeyEqual(a, b []byte) bool {
	da, err1 := decodePEMToDER(a)
	db, err2 := decodePEMToDER(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return bytes.Equal(da, db)
}

func decodePEMToDER(pemBytes []byte) ([]byte, error) {
	s := string(pemBytes)
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	block, _ := pem.Decode([]byte(strings.TrimSpace(s)))
	if block == nil {
		return nil, fmt.Errorf("不是有效的 PEM 块")
	}
	return block.Bytes, nil
}

func pubKeySHA256(pemBytes []byte) string {
	der, err := decodePEMToDER(pemBytes)
	if err != nil {
		return "invalid-pem"
	}
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}

// =============================================================================
// 导入单证书
// =============================================================================

func (s *CertService) Import(
	ctx context.Context, req *ImportCertRequest,
) (*models.Certificate, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" || !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(exception.CodeParamInvalid, "证书不是有效的 PEM 格式", 400, nil)
	}

	dirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("分配 server 目录失败：%v", err), 500, nil)
	}
	dir, _ := s.layout.ServerDir(dirNo)

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}
	pubSM3, err := PubkeySM3FromCertPEM(bin, certPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("计算证书公钥 SM3 失败：%v", err), 400, nil)
	}
	certAbs := filepath.Join(dir, pubSM3+".cert.pem")
	if err := os.WriteFile(certAbs, []byte(certPEM), 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写证书失败：%v", err), 500, nil)
	}
	success := false
	defer func() {
		if !success {
			_ = os.Remove(certAbs)
		}
	}()

	detail, err := s.parser.Parse(certAbs)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败: %v", err), 400, nil)
	}

	var keyRef *string
	var keyPath string

	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil)
		}
		plainKeyPEM := []byte(req.KeyPEM)
		if s.keys != nil {
			tmpDir, _ := os.MkdirTemp("", "import-key-")
			defer os.RemoveAll(tmpDir)
			tmpKey := filepath.Join(tmpDir, "key.pem")
			if err := os.WriteFile(tmpKey, []byte(req.KeyPEM), 0600); err != nil {
				return nil, exception.New(exception.CodeInternalError, "写临时私钥失败", 500, nil)
			}
			if encrypted, _ := s.keys.IsEncrypted(bin, tmpKey); encrypted {
				if strings.TrimSpace(req.KeyPassword) == "" {
					return nil, exception.New(exception.CodeParamInvalid,
						"私钥已加密，请提供私钥密码", 400, nil)
				}
				pwPath := filepath.Join(tmpDir, "pw.txt")
				plainPath := filepath.Join(tmpDir, "plain.key")
				if err := os.WriteFile(pwPath, []byte(req.KeyPassword), 0600); err != nil {
					return nil, exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
				}
				if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
					return nil, exception.New(exception.CodeParamInvalid,
						"私钥密码错误或解密失败", 400, nil)
				}
				data, err := os.ReadFile(plainPath)
				if err != nil {
					return nil, exception.New(exception.CodeInternalError, "读解密私钥失败", 500, nil)
				}
				plainKeyPEM = data
			}
		}

		keyPath = filepath.Join(dir, pubSM3+".key.pem")

		plainTmp := filepath.Join(dir, ".import.key.plain.tmp")
		if err := os.WriteFile(plainTmp, plainKeyPEM, 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
		}
		defer os.Remove(plainTmp)

		if err := s.whitebox.EncryptFile(ctx, plainTmp, keyPath); err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
		}
		_ = os.Chmod(keyPath, 0600)
	} else if req.KeyRef != "" {
		keyRef = &req.KeyRef
	}

	inferredType := inferCertType(detail.KeyUsage, detail.ExtendedKeyUsage, detail.PublicKeyAlgorithm)

	now := time.Now().UTC()
	cert := &models.Certificate{
		CertID:             "cert-" + dirNo + "-" + pubSM3[:12],
		CertType:           inferredType,
		Serial:             detail.Serial,
		Subject:            detail.Subject,
		SubjectCN:          cnFromSubject(detail.Subject),
		Issuer:             detail.Issuer,
		IssuerCN:           cnFromSubject(detail.Issuer),
		Algorithm:          detail.PublicKeyAlgorithm,
		Fingerprint:        detail.Fingerprint,
		PublicKeyAlgorithm: detail.PublicKeyAlgorithm,
		SignatureAlgorithm: detail.SignatureAlgorithm,
		NotBefore:          now,
		NotAfter:           now,
		CertPath:           certAbs,
		KeyRef:             keyRef,
		Status:             "VALID",
		CreatedAt:          now,
	}
	if keyPath != "" {
		cert.KeyPath = &keyPath
	}
	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil)
	}
	success = true
	return cert, nil
}

// ImportDualCert 国密双证导入。
func (s *CertService) ImportDualCert(
	ctx context.Context, req *ImportDualCertRequest,
) (*models.Certificate, error) {
	if req == nil ||
		strings.TrimSpace(req.CSRPubSM3) == "" ||
		strings.TrimSpace(req.SignCertPEM) == "" ||
		strings.TrimSpace(req.EncCertPEM) == "" ||
		strings.TrimSpace(req.EncKeyPEM) == "" {
		return nil, exception.New(exception.CodeParamInvalid, "双证导入参数不完整", 400, nil)
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	csrDir, err := s.layout.FindCSRDirByPubkeySM3(req.CSRPubSM3)
	if err != nil {
		return nil, err
	}

	signPubSM3, err := PubkeySM3FromCertPEM(bin, req.SignCertPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("签名证书公钥 SM3 计算失败：%v", err), 400, nil)
	}
	signCertAbs := filepath.Join(csrDir, signPubSM3+".cert.pem")
	if err := os.WriteFile(signCertAbs, []byte(req.SignCertPEM), 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写签名证书失败：%v", err), 500, nil)
	}

	encDirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("分配加密证书目录失败：%v", err), 500, nil)
	}
	encDir, _ := s.layout.ServerDir(encDirNo)

	encPubSM3, err := PubkeySM3FromCertPEM(bin, req.EncCertPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密证书公钥 SM3 计算失败：%v", err), 400, nil)
	}
	encCertAbs := filepath.Join(encDir, encPubSM3+".cert.pem")
	if err := os.WriteFile(encCertAbs, []byte(req.EncCertPEM), 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写加密证书失败：%v", err), 500, nil)
	}

	plainKeyPEM := []byte(req.EncKeyPEM)
	if strings.TrimSpace(req.EncKeyPassword) != "" && s.keys != nil {
		tmpDir, _ := os.MkdirTemp("", "import-dual-key-")
		defer os.RemoveAll(tmpDir)
		tmpKey := filepath.Join(tmpDir, "key.pem")
		pwPath := filepath.Join(tmpDir, "pw.txt")
		plainPath := filepath.Join(tmpDir, "plain.key")
		if err := os.WriteFile(tmpKey, []byte(req.EncKeyPEM), 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError, "写临时私钥失败", 500, nil)
		}
		if err := os.WriteFile(pwPath, []byte(req.EncKeyPassword), 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
		}
		if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
			return nil, exception.New(exception.CodeParamInvalid,
				"加密私钥口令错误或解密失败", 400, nil)
		}
		data, err := os.ReadFile(plainPath)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError, "读解密私钥失败", 500, nil)
		}
		plainKeyPEM = data
	}

	encKeyAbs := filepath.Join(encDir, encPubSM3+".key.pem")

	plainTmp := filepath.Join(encDir, ".import.dual.key.plain.tmp")
	if err := os.WriteFile(plainTmp, plainKeyPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	if err := s.whitebox.EncryptFile(ctx, plainTmp, encKeyAbs); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
	}
	_ = os.Chmod(encKeyAbs, 0600)

	now := time.Now().UTC()
	encDetail, _ := s.parser.Parse(encCertAbs)
	cert := &models.Certificate{
		CertID:    "dual-enc-" + encDirNo + "-" + encPubSM3[:12],
		CertType:  "dual_enc",
		CertPath:  encCertAbs,
		Algorithm: "SM2",
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now,
		CreatedAt: now,
	}
	if encDetail != nil {
		cert.Serial = encDetail.Serial
		cert.Subject = encDetail.Subject
		cert.SubjectCN = cnFromSubject(encDetail.Subject)
		cert.Issuer = encDetail.Issuer
		cert.IssuerCN = cnFromSubject(encDetail.Issuer)
		cert.Fingerprint = encDetail.Fingerprint
		cert.PublicKeyAlgorithm = encDetail.PublicKeyAlgorithm
		cert.SignatureAlgorithm = encDetail.SignatureAlgorithm
	}
	cert.KeyPath = &encKeyAbs

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("双证加密证书落库失败：%v", err), 500, nil)
	}
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

// =============================================================================
// 导出
// =============================================================================

func (s *CertService) Export(
	ctx context.Context, certID, exportType, password string,
) (*ExportResult, error) {
	c, err := s.Get(certID)
	if err != nil {
		return nil, err
	}
	switch exportType {
	case "cert":
		data, err := os.ReadFile(c.CertPath)
		if err != nil {
			data, err = s.files.ReadCoreFile(c.CertPath)
			if err != nil {
				return nil, err
			}
		}
		return &ExportResult{
			Data:        data,
			Filename:    certID + ".pem",
			ContentType: "application/x-pem-file",
		}, nil

	case "key":
		keyPEM, err := s.exportKeyFromStorage(ctx, c)
		if err != nil {
			if c.KeyRef != nil && *c.KeyRef != "" {
				keyPEM, err = s.caller.ExportKey(ctx, *c.KeyRef)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		if strings.TrimSpace(password) != "" {
			bin := s.parser.OpensslBin()
			if bin == "" {
				return nil, exception.New(exception.CodeInternalError,
					"铜锁 openssl 不可用，无法加密私钥", 500, nil)
			}
			encrypted, err := encryptPrivateKeyPEM(bin, keyPEM, password)
			if err != nil {
				return nil, exception.New(exception.CodeInternalError,
					fmt.Sprintf("加密私钥失败: %v", err), 500, nil)
			}
			keyPEM = encrypted
		}
		return &ExportResult{
			Data:        keyPEM,
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

func (s *CertService) exportKeyFromStorage(
	ctx context.Context, c *models.Certificate,
) ([]byte, error) {
	keyAbs := s.resolveCertKeyPath(c)
	if keyAbs == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"该证书未关联白盒私钥", 400, nil)
	}
	if _, err := os.Stat(keyAbs); err != nil {
		return nil, exception.New(exception.CodeNotFound, "白盒私钥文件不存在", 404, nil)
	}
	plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = s.whitebox.ReEncrypt(ctx, plainTmp, keyAbs)
		_ = os.Remove(plainTmp)
	}()
	return os.ReadFile(plainTmp)
}

func (s *CertService) resolveCertKeyPath(c *models.Certificate) string {
	if c == nil {
		return ""
	}
	if c.KeyPath != nil && *c.KeyPath != "" {
		p := *c.KeyPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(s.layout.CertRoot, p)
		}
		return p
	}
	if c.ChainPath != nil && *c.ChainPath != "" {
		p := *c.ChainPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(s.layout.CertRoot, p)
		}
		return p
	}
	return ""
}

// =============================================================================
// 删除（硬删除）
// =============================================================================

func (s *CertService) Delete(certID string) error {
	var c models.Certificate
	if err := s.db.Where("cert_id = ?", certID).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("查询证书失败: %v", err), 500, nil)
	}

	var files []string
	if c.CertPath != "" {
		if abs := s.absUnderCertRoot(c.CertPath); abs != "" {
			files = append(files, abs)
		}
	}
	if kp := s.resolveCertKeyPath(&c); kp != "" && s.underCertRoot(kp) {
		files = append(files, kp)
	}
	if c.CertPath != "" {
		if abs := s.absUnderCertRoot(c.CertPath); abs != "" {
			files = append(files, findRehashLinks(abs)...)
		}
	}

	if err := s.db.Unscoped().
		Where("cert_id = ?", certID).
		Delete(&models.Certificate{}).Error; err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("删除证书记录失败: %v", err), 500, nil)
	}

	removed := 0
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			if !os.IsNotExist(err) {
				log.Warn().Str("file", f).Err(err).
					Msg("Delete: 删除文件失败（忽略）")
			}
		} else {
			removed++
		}
	}

	if c.CertPath != "" {
		if abs := s.absUnderCertRoot(c.CertPath); abs != "" {
			s.tryRemoveEmptyDir(filepath.Dir(abs))
		}
	}

	log.Info().
		Str("cert_id", certID).
		Int("files_removed", removed).
		Int("files_total", len(files)).
		Msg("证书已硬删除")
	return nil
}

func (s *CertService) tryRemoveEmptyDir(dir string) {
	if dir == "" || s.layout == nil {
		return
	}
	abs := filepath.Clean(dir)
	root := filepath.Clean(s.layout.CertRoot)
	if abs == root {
		return
	}
	if !s.underCertRoot(abs) {
		return
	}

	base := filepath.Base(abs)
	parent := filepath.Dir(abs)

	isServerDir := parent == filepath.Clean(s.layout.ServerRoot) && dirNoRe.MatchString(base)
	isCADir := parent == filepath.Clean(s.layout.CARoot) &&
		base != "" && !strings.HasPrefix(base, ".")

	if !isServerDir && !isCADir {
		return
	}

	entries, err := os.ReadDir(abs)
	if err != nil || len(entries) > 0 {
		return
	}

	if err := os.Remove(abs); err != nil {
		log.Warn().Str("dir", abs).Err(err).Msg("Delete: 删除空目录失败（忽略）")
		return
	}
	log.Info().Str("dir", abs).Msg("Delete: 已删除空目录")
}

func findRehashLinks(target string) []string {
	dir := filepath.Dir(target)
	base := filepath.Base(target)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var links []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".0") || len(name) != 10 {
			continue
		}
		full := filepath.Join(dir, name)
		info, err := os.Lstat(full)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		linkTarget, err := os.Readlink(full)
		if err != nil {
			continue
		}
		if linkTarget == base || linkTarget == target {
			links = append(links, full)
		}
	}
	return links
}

func (s *CertService) absUnderCertRoot(p string) string {
	if p == "" {
		return ""
	}
	abs := p
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(s.layout.CertRoot, p)
	}
	if s.underCertRoot(abs) {
		return abs
	}
	if s.caller != nil {
		abs2 := s.caller.ResolveCorePath(p)
		if s.underCertRoot(abs2) {
			return abs2
		}
	}
	return ""
}

func (s *CertService) underCertRoot(p string) bool {
	if s.layout == nil || s.layout.CertRoot == "" {
		return false
	}
	abs := filepath.Clean(p)
	root := filepath.Clean(s.layout.CertRoot)
	return abs == root || strings.HasPrefix(abs, root+string(os.PathSeparator))
}

// =============================================================================
// 私钥导出 / 加密辅助
// =============================================================================

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
	return os.ReadFile(encPath)
}
