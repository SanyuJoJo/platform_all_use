package crypto

import (
	"context"
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
	// ★ 归一化：兼容旧前端残留的 server / client
	if len(req.CertTypes) > 0 {
		normalized := normalizeCertTypes(req.CertTypes)
		params["cert_types"] = normalized
		params["cert_type"] = normalized[0]
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

		if ca.KeyRef != "" {
			return s.prepareManualCAFromKeystore(ctx, &ca, params, cleanup)
		}
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
func (s *CertService) prepareManualCAFromKeystore(
	ctx context.Context, ca *models.CA, params map[string]interface{},
	cleanup *[]string,
) error {
	if ca == nil || ca.KeyRef == "" {
		return exception.New(exception.CodeInternalError, "CA keyRef 为空", 500, nil)
	}

	plainKey, err := s.caller.ExportKey(ctx, ca.KeyRef)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("导出 CA 私钥失败：%v", err), 500, nil)
	}
	if len(plainKey) == 0 {
		return exception.New(exception.CodeInternalError, "导出的 CA 私钥为空", 500, nil)
	}

	id := newShortID()

	coreKeyRel := fmt.Sprintf("tmp/ca-sign-key-%s.pem", id)
	if err := s.files.WriteCoreFile(coreKeyRel, plainKey, 0600); err != nil {
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
		Msg("CA keystore 私钥已导出并转入 core/tmp（manual 模式）")
	return nil
}

// prepareManualCAFromWhitebox 把白盒 CA 私钥解密并转入 core/tmp/，
// 使 params 切换为 manual 模式。
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

	// ★ 修改：归一化证书类型（兼容旧前端传 server / client）
	certTypeJoined := strings.Join(normalizeCertTypes(req.CertTypes), ",")

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
	// ★ 运行时填充展示字段
	cert.HasPrivateKey =
		(cert.KeyRef != nil && *cert.KeyRef != "") ||
			(cert.KeyPath != nil && *cert.KeyPath != "")

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil)
	}
	return cert, nil
}
