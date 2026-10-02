package crypto

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// CSRService P10 领域 Service。
type CSRService struct {
	db     *gorm.DB
	caller *CoreCaller
	files  *FileStore
	parser *CertParser
	keys   *KeyCrypto
}

// NewCSRService 创建 P10 Service。
func NewCSRService(
	db *gorm.DB,
	caller *CoreCaller,
	files *FileStore,
	parser *CertParser,
	keys *KeyCrypto,
) *CSRService {
	return &CSRService{
		db:     db,
		caller: caller,
		files:  files,
		parser: parser,
		keys:   keys,
	}
}

// List P10 列表。
func (s *CSRService) List(page, pageSize int) (map[string]interface{}, error) {
	var csrs []models.CSR
	q := s.db.Model(&models.CSR{}).Where("status <> ?", "DELETED")
	return paginateQuery(q, &csrs, page, pageSize)
}

// Get P10 元数据。
func (s *CSRService) Get(csrID string) (*models.CSR, error) {
	var csr models.CSR
	if err := s.db.Where("csr_id = ?", csrID).First(&csr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeNotFound, "P10 不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询 P10 失败", 500, nil)
	}
	return &csr, nil
}

// Import 导入 P10。
//
// ★ 支持带私钥导入：
//   如果客户上传了 CSR 私钥（KeyPEM），平台会把私钥加密保存到 core，
//   并把 key_ref 记录在 platform_csr 表里。
//   之后签发的签名证书会继承这个 key_ref，从而支持"从签名证书解析"模式。
//
// 使用场景：
//   1. 客户只上传 CSR（不带私钥）→ KeyRef = nil，只能用手动输入模式
//   2. 客户上传 CSR + 私钥 → KeyRef 非空，支持从签名证书自动解密
func (s *CSRService) Import(
	ctx context.Context, req *ImportCsrRequest,
) (*models.CSR, error) {
	csrPEM := strings.TrimSpace(req.CSRPEM)
	if csrPEM == "" || !strings.Contains(csrPEM, "-----BEGIN CERTIFICATE REQUEST-----") {
		return nil, exception.New(
			exception.CodeParamInvalid, "P10 不是有效的 PEM 格式", 400, nil,
		)
	}

	id := newShortID()
	csrID := "csr-" + id
	csrRel := fmt.Sprintf("data/csr/imported-%s.csr", id)

	if err := s.files.WriteCoreFile(csrRel, []byte(csrPEM), 0640); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = s.files.SafeRemove(csrRel)
		}
	}()

	subject, algorithm, err := s.parseCSRInfo(csrRel)
	if err != nil {
		return nil, err
	}

	// ★ 私钥处理：如果客户上传了私钥，加密保存到 core 并记录 key_ref
	var keyRef *string
	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(
				exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil,
			)
		}

		keyRel := fmt.Sprintf("tmp/import-csr-key-%s.pem", id)
		if err := s.files.WriteCoreFile(keyRel, []byte(req.KeyPEM), 0600); err != nil {
			return nil, err
		}
		defer s.files.SafeRemove(keyRel)

		keyAbs, _ := s.files.guard.Resolve(keyRel, "key_path")

		// 若私钥加密，先解密到明文
		if bin := s.parser.OpensslBin(); bin != "" {
			encrypted, _ := s.keys.IsEncrypted(bin, keyAbs)
			if encrypted {
				if strings.TrimSpace(req.KeyPassword) == "" {
					return nil, exception.New(
						exception.CodeParamInvalid,
						"私钥已加密，请提供私钥密码", 400, nil,
					)
				}
				pwRel := fmt.Sprintf("tmp/import-csr-pass-%s", id)
				plainRel := fmt.Sprintf("tmp/import-csr-plain-%s.pem", id)
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

		// 调 core key.manage import 加密保存
		ref, err := s.caller.ImportKey(ctx, keyRel, algorithm)
		if err != nil {
			return nil, err
		}
		keyRef = &ref
	}

	now := time.Now().UTC()
	csr := &models.CSR{
		CSRID:     csrID,
		SubjectCN: extractCNFromSubject(subject),
		Algorithm: algorithm,
		CSRPath:   csrRel,
		KeyRef:    keyRef,
		Status:    "NEW",
		CreatedAt: now,
	}
	if err := s.db.Create(csr).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("P10 落库失败: %v", err), 500, nil,
		)
	}
	success = true
	return csr, nil
}

// parseCSRInfo 解析 CSR 主题与公钥算法。
func (s *CSRService) parseCSRInfo(csrRel string) (string, string, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return "", "", exception.New(
			exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil,
		)
	}
	abs, err := s.files.guard.Resolve(csrRel, "csr_path")
	if err != nil {
		return "", "", err
	}

	subjOut, err := RunOpenSSL(bin, "req", "-in", abs, "-noout", "-subject", "-nameopt", "RFC2253")
	if err != nil {
		return "", "", exception.New(
			exception.CodeParamInvalid, "CSR 解析失败", 400, nil,
		)
	}
	subject := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(subjOut), "subject="))

	textOut, _ := RunOpenSSL(bin, "req", "-in", abs, "-noout", "-text")
	algorithm := ""
	switch {
	case strings.Contains(textOut, "SM2"):
		algorithm = "SM2"
	case strings.Contains(textOut, "ML-DSA"):
		algorithm = "ML-DSA"
	case strings.Contains(textOut, "rsaEncryption") || strings.Contains(textOut, "RSA"):
		algorithm = "RSA"
	case strings.Contains(textOut, "id-ecPublicKey"):
		algorithm = "ECC"
	}

	return subject, algorithm, nil
}

// Download 下载 P10 文件。
func (s *CSRService) Download(csrID string) ([]byte, error) {
	csr, err := s.Get(csrID)
	if err != nil {
		return nil, err
	}
	return s.files.ReadCoreFile(csr.CSRPath)
}

// DownloadKey 下载 P10 对应的私钥，可选 AES-256-CBC 口令加密。
func (s *CSRService) DownloadKey(
	csrID string, encrypt bool, password string,
) ([]byte, error) {
	csr, err := s.Get(csrID)
	if err != nil {
		return nil, err
	}
	if csr.KeyRef == nil || *csr.KeyRef == "" {
		return nil, exception.New(
			exception.CodeParamInvalid, "该 P10 未关联私钥", 400, nil,
		)
	}

	plain, err := s.caller.ExportKey(context.Background(), *csr.KeyRef)
	if err != nil {
		return nil, err
	}

	if !encrypt {
		return plain, nil
	}

	if strings.TrimSpace(password) == "" {
		return nil, exception.New(
			exception.CodeParamInvalid, "加密口令不能为空", 400, nil,
		)
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(
			exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil,
		)
	}
	encrypted, err := encryptPrivateKeyPEM(bin, plain, password)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("加密私钥失败: %v", err), 500, nil,
		)
	}
	return encrypted, nil
}

// Delete 删除 P10（硬删除）。
func (s *CSRService) Delete(csrID string) error {
	if strings.TrimSpace(csrID) == "" {
		return exception.New(exception.CodeParamInvalid, "csr_id 不能为空", 400, nil)
	}

	var csr models.CSR
	if err := s.db.Where("csr_id = ?", csrID).First(&csr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return exception.New(exception.CodeNotFound, "P10 不存在或已删除", 404, nil)
		}
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("查询 P10 失败: %v", err), 500, nil,
		)
	}

	res := s.db.Where("csr_id = ?", csrID).Delete(&models.CSR{})
	if res.Error != nil {
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("删除 P10 失败: %v", res.Error), 500, nil,
		)
	}
	if res.RowsAffected == 0 {
		return exception.New(exception.CodeNotFound, "P10 不存在或已删除", 404, nil)
	}

	if csr.CSRPath != "" {
		_ = s.files.SafeRemove(csr.CSRPath)
	}
	return nil
}
