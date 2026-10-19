package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// ImportCARequest 导入 CA 请求。
type ImportCARequest struct {
	CertPEM     string `json:"cert_pem" binding:"required"`
	KeyPEM      string `json:"key_pem,omitempty"`
	KeyPassword string `json:"key_password,omitempty"`
	KeyRef      string `json:"key_ref,omitempty"`
	// Domain 指定域目录，默认 "safe"。
	Domain string `json:"domain,omitempty"`
}

// Import 导入 CA。
//
// ★ 证书路径改造后的落盘规范：
//
//	<CertRoot>/ca/<domain>/<pubkey_sm3>.cert.pem     证书（0640）
//	<CertRoot>/ca/<domain>/<pubkey_sm3>.key.pem      白盒私钥（0600，若提供）
//	<CertRoot>/ca/<domain>/<subject_hash>.0          rehash 软链
//	<CertRoot>/ca/<domain>/bundle.crt                全量根证书（0644，自动重建）
//
// ★ 私钥三种模式：
//   - KeyRef 非空 → 引用 core keystore（旧模式），DB 只记 KeyRef
//   - KeyPEM 非空 → 白盒加密落盘（新模式），DB 只记 KeyPath
//   - 都为空    → 无关联私钥（只能验证，不能签发）
func (s *CAService) Import(ctx context.Context, req *ImportCARequest) (*models.CA, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" {
		return nil, exception.New(exception.CodeParamInvalid, "证书文件不能为空", 400, nil)
	}
	if !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(exception.CodeParamInvalid, "证书文件不是有效的 PEM 格式", 400, nil)
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	// 1. 计算证书公钥 SM3（用于文件命名）
	pubSM3, err := PubkeySM3FromCertPEM(bin, certPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("计算证书公钥 SM3 失败：%v", err), 400, nil)
	}

	// 2. 确定域目录（默认 safe）
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = "safe"
	}
	caDir, err := s.layout.CADir(domain)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(caDir, 0750); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建 CA 目录失败：%v", err), 500, nil)
	}

	// 3. 写证书到规范路径
	certAbs := filepath.Join(caDir, pubSM3+".cert.pem")
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

	// 4. 解析证书
	detail, err := s.parser.Parse(certAbs)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败: %v", err), 400, nil)
	}
	subject := detail.Subject
	serial := detail.Serial
	algorithm := detail.PublicKeyAlgorithm

	// 5. 生成 rehash 软链
	if err := s.ensureRehash(certAbs); err != nil {
		log.Warn().
			Str("cert", certAbs).
			Err(err).
			Msg("ensureRehash 失败（忽略）")
	}

	// 6. 处理私钥
	keyRef, keyPath, err := s.resolveImportedKeyNew(ctx, req, caDir, pubSM3)
	if err != nil {
		return nil, err
	}

	// 7. 落库
	now := timeNowUTC()
	caID := "ca-" + pubSM3[:12]

	// ★ 冲突检查：同一 pubkey 的 CA 已存在时拒绝，但**仍触发 bundle 重建**
	//   （保证历史数据 / 手工补充的证书自动进 bundle）
	var existing models.CA
	if e := s.db.Where("ca_id = ?", caID).First(&existing).Error; e == nil {
		if err := s.RebuildDomainBundle(domain); err != nil {
			log.Warn().Err(err).Str("domain", domain).
				Msg("Import: 已存在 CA 触发 bundle 重建失败（不阻断）")
		}
		return nil, exception.New(exception.CodeParamInvalid,
			"该 CA 已存在（相同公钥 SM3）", 409, nil)
	}

	ca := &models.CA{
		CAID:         caID,
		SubjectCN:    extractCNFromSubject(subject),
		Algorithm:    algorithm,
		CertPath:     certAbs,
		KeyRef:       keyRef,
		KeyPath:      keyPath,
		ValidityDays: 0,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if serial != "" {
		ca.Serial = &serial
	}
	if err := s.db.Create(ca).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写库失败: %v", err), 500, nil)
	}
	success = true

	// ★ 重建域 bundle（新增 CA 自动进全量根证书）
	if err := s.RebuildDomainBundle(domain); err != nil {
		log.Warn().Err(err).Str("domain", domain).
			Msg("Import: 重建 bundle 失败（不阻断）")
	}

	log.Info().
		Str("ca_id", caID).
		Str("domain", domain).
		Str("cert_path", certAbs).
		Bool("has_key_ref", keyRef != "").
		Bool("has_key_path", keyPath != "").
		Msg("CA 导入完成")

	return ca, nil
}

// resolveImportedKeyNew 处理导入私钥。
func (s *CAService) resolveImportedKeyNew(
	ctx context.Context, req *ImportCARequest, caDir, pubSM3 string,
) (string, string, error) {
	if ref := strings.TrimSpace(req.KeyRef); ref != "" {
		return ref, "", nil
	}

	keyPEM := strings.TrimSpace(req.KeyPEM)
	if keyPEM == "" {
		return "", "", nil
	}
	if !strings.Contains(keyPEM, "-----BEGIN") {
		return "", "", exception.New(exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil)
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return "", "", exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	plainKeyPEM := []byte(keyPEM)
	tmpDir, _ := os.MkdirTemp("", "import-ca-key-")
	defer os.RemoveAll(tmpDir)

	tmpKey := filepath.Join(tmpDir, "key.pem")
	if err := os.WriteFile(tmpKey, []byte(keyPEM), 0600); err != nil {
		return "", "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
	}

	if s.keys != nil {
		if encrypted, _ := s.keys.IsEncrypted(bin, tmpKey); encrypted {
			if strings.TrimSpace(req.KeyPassword) == "" {
				return "", "", exception.New(exception.CodeParamInvalid,
					"私钥已加密，请提供私钥密码", 400, nil)
			}
			pwPath := filepath.Join(tmpDir, "pw.txt")
			plainPath := filepath.Join(tmpDir, "plain.key")
			if err := os.WriteFile(pwPath, []byte(req.KeyPassword), 0600); err != nil {
				return "", "", exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
			}
			if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
				return "", "", exception.New(exception.CodeParamInvalid,
					"私钥密码错误或解密失败", 400, nil)
			}
			data, err := os.ReadFile(plainPath)
			if err != nil {
				return "", "", exception.New(exception.CodeInternalError, "读解密私钥失败", 500, nil)
			}
			plainKeyPEM = data
		}
	}

	if s.whitebox == nil {
		return "", "", exception.New(exception.CodeInternalError,
			"whitebox 未初始化", 500, nil)
	}

	keyAbs := filepath.Join(caDir, pubSM3+".key.pem")
	plainTmp := filepath.Join(caDir, ".import.key.plain.tmp")
	if err := os.WriteFile(plainTmp, plainKeyPEM, 0600); err != nil {
		return "", "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("写临时明文私钥失败：%v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	if err := s.whitebox.EncryptFile(ctx, plainTmp, keyAbs); err != nil {
		return "", "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
	}
	_ = os.Chmod(keyAbs, 0600)

	return "", keyAbs, nil
}

func (s *CAService) ensureRehash(certAbs string) error {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return fmt.Errorf("openssl 不可用")
	}
	dir := filepath.Dir(certAbs)
	base := filepath.Base(certAbs)

	out, err := RunOpenSSL(bin, "x509", "-in", certAbs, "-noout", "-subject_hash")
	if err != nil {
		return fmt.Errorf("subject_hash: %w", err)
	}
	hash := strings.TrimSpace(out)
	if hash == "" {
		return fmt.Errorf("empty subject_hash")
	}

	link := filepath.Join(dir, hash+".0")
	_ = os.Remove(link)
	if err := os.Symlink(base, link); err != nil {
		return fmt.Errorf("symlink %s: %w", link, err)
	}
	return nil
}
