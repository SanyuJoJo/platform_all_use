package crypto
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)
// ImportCARequest 导入 CA 请求。
type ImportCARequest struct {
	CertPEM     string `json:"cert_pem"`
	KeyPEM      string `json:"key_pem,omitempty"`
	KeyPassword string `json:"key_password,omitempty"`
	KeyRef      string `json:"key_ref,omitempty"`
}
// Import 导入 CA。
//
// 流程：
//  1. 写证书到 data/ca/imported-<id>.pem
//  2. core cert.parse 拿算法/主题/序列号
//  3. 若有私钥：写临时文件 → 检测加密 → 按需解密 → core key.manage import
//  4. 落库
func (s *CAService) Import(ctx context.Context, req *ImportCARequest) (*models.CA, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" {
		return nil, exception.New(exception.CodeParamInvalid, "证书文件不能为空", 400, nil)
	}
	if !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(exception.CodeParamInvalid, "证书文件不是有效的 PEM 格式", 400, nil)
	}
	id := newShortID()
	caID := "ca-" + id
	certRel := fmt.Sprintf("data/ca/imported-%s.pem", id)
	if err := s.files.WriteCoreFile(certRel, []byte(certPEM), 0640); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = s.files.SafeRemove(certRel)
		}
	}()
	parsed, err := s.caller.ParseCert(ctx, certRel)
	if err != nil {
		return nil, err
	}
	subject := getString(parsed, "subject")
	serial := getString(parsed, "serial")
	algorithm := getString(parsed, "public_key_algorithm")
	keyRef, err := s.resolveImportedKey(ctx, req, id, algorithm)
	if err != nil {
		return nil, err
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
		return nil, exception.New(exception.CodeInternalError, "写库失败", 500, nil)
	}
	success = true
	return ca, nil
}
// resolveImportedKey 处理私钥：写临时文件 → 检测加密 → 解密 → core import。
func (s *CAService) resolveImportedKey(
	ctx context.Context, req *ImportCARequest, id, algorithm string,
) (string, error) {
	if ref := strings.TrimSpace(req.KeyRef); ref != "" {
		return ref, nil
	}
	keyPEM := strings.TrimSpace(req.KeyPEM)
	if keyPEM == "" {
		return "", nil
	}
	if !strings.Contains(keyPEM, "-----BEGIN") {
		return "", exception.New(exception.CodeParamInvalid, "私钥文件不是有效的 PEM 格式", 400, nil)
	}
	keyRel := fmt.Sprintf("tmp/import-%s.key", id)
	if err := s.files.WriteCoreFile(keyRel, []byte(keyPEM), 0600); err != nil {
		return "", err
	}
	defer s.files.SafeRemove(keyRel)
	openssl := s.parser.OpensslBin()
	if openssl == "" {
		return "", exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}
	encrypted, _ := s.keys.IsEncrypted(openssl, filepath.Join(s.files.coreRoot, keyRel))
	if encrypted {
		if strings.TrimSpace(req.KeyPassword) == "" {
			return "", exception.New(exception.CodeParamInvalid, "私钥已加密，请提供私钥密码", 400, nil)
		}
		pwRel := fmt.Sprintf("tmp/import-%s.pw", id)
		plainRel := fmt.Sprintf("tmp/import-%s.plain.key", id)
		if err := s.files.WriteCoreFile(pwRel, []byte(req.KeyPassword), 0600); err != nil {
			return "", err
		}
		defer s.files.SafeRemove(pwRel)
		defer s.files.SafeRemove(plainRel)
		pwAbs := filepath.Join(s.files.coreRoot, pwRel)
		keyAbs := filepath.Join(s.files.coreRoot, keyRel)
		plainAbs := filepath.Join(s.files.coreRoot, plainRel)
		if err := s.keys.DecryptWithPasswordFile(openssl, keyAbs, pwAbs, plainAbs); err != nil {
			return "", exception.New(exception.CodeParamInvalid, "私钥密码错误或解密失败", 400, nil)
		}
		plainData, err := os.ReadFile(plainAbs)
		if err != nil {
			return "", exception.New(exception.CodeInternalError, "读取解密结果失败", 500, nil)
		}
		if err := os.WriteFile(keyAbs, plainData, 0600); err != nil {
			return "", exception.New(exception.CodeInternalError, "写明文私钥失败", 500, nil)
		}
	}
	return s.caller.ImportKey(ctx, keyRel, algorithm)
}
