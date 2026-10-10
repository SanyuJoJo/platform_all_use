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

// CSRService P10 领域 Service。
type CSRService struct {
	db       *gorm.DB
	caller   *CoreCaller
	files    *FileStore
	parser   *CertParser
	keys     *KeyCrypto
	layout   *PathLayout
	whitebox *WhiteboxClient
}

func NewCSRService(
	db *gorm.DB,
	caller *CoreCaller,
	files *FileStore,
	parser *CertParser,
	keys *KeyCrypto,
	layout *PathLayout,
	whitebox *WhiteboxClient,
) *CSRService {
	return &CSRService{
		db:       db,
		caller:   caller,
		files:    files,
		parser:   parser,
		keys:     keys,
		layout:   layout,
		whitebox: whitebox,
	}
}

// =============================================================================
// 列表 / 详情
// =============================================================================

func (s *CSRService) List(page, pageSize int) (map[string]interface{}, error) {
	var csrs []models.CSR
	q := s.db.Model(&models.CSR{}).Where("status <> ?", "DELETED")
	return paginateQuery(q, &csrs, page, pageSize)
}

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

// =============================================================================
// 导入 P10
// =============================================================================

// Import 导入 P10。
//
// ★ 证书路径改造：
//   - 分配 server/<dir_no>/ 目录；
//   - CSR 落 server/<dir_no>/<csr_pubkey_sm3>.req.csr（0640）；
//   - 私钥（若有）白盒加密落 server/<dir_no>/<csr_pubkey_sm3>.key.pem（0600）
//     （密钥内嵌二进制，无 .pass）。
func (s *CSRService) Import(
	ctx context.Context, req *ImportCsrRequest,
) (*models.CSR, error) {
	csrPEM := strings.TrimSpace(req.CSRPEM)
	if csrPEM == "" || !strings.Contains(csrPEM, "-----BEGIN CERTIFICATE REQUEST-----") {
		return nil, exception.New(
			exception.CodeParamInvalid, "P10 不是有效的 PEM 格式", 400, nil,
		)
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(
			exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil,
		)
	}

	// 1. 分配 server 目录
	dirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("分配 server 目录失败：%v", err), 500, nil,
		)
	}
	dir, err := s.layout.ServerDir(dirNo)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("server 目录非法：%v", err), 500, nil,
		)
	}

	// 2. 计算 CSR 公钥 SM3
	pubSM3, err := PubkeySM3FromCSRPEM(bin, csrPEM)
	if err != nil {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("计算 CSR 公钥 SM3 失败：%v", err), 400, nil,
		)
	}

	// 3. 写 CSR 到规范路径
	csrAbs := filepath.Join(dir, pubSM3+".req.csr")
	if err := os.WriteFile(csrAbs, []byte(csrPEM), 0640); err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("写 P10 失败：%v", err), 500, nil,
		)
	}
	success := false
	defer func() {
		if !success {
			_ = os.Remove(csrAbs)
		}
	}()

	// 4. 解析主题 / 算法
	subject, algorithm, err := s.parseCSRInfo(csrAbs)
	if err != nil {
		return nil, err
	}

	// 5. 私钥处理：白盒加密落盘（无 .pass）
	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(
				exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil,
			)
		}

		plainKeyPEM := []byte(req.KeyPEM)

		// 5.1 若私钥已加密，先解密到明文
		if s.keys != nil {
			tmpDir, _ := os.MkdirTemp("", "import-csr-key-")
			defer os.RemoveAll(tmpDir)

			tmpKey := filepath.Join(tmpDir, "key.pem")
			if err := os.WriteFile(tmpKey, []byte(req.KeyPEM), 0600); err != nil {
				return nil, exception.New(
					exception.CodeInternalError,
					fmt.Sprintf("写临时私钥失败：%v", err), 500, nil,
				)
			}

			if encrypted, _ := s.keys.IsEncrypted(bin, tmpKey); encrypted {
				if strings.TrimSpace(req.KeyPassword) == "" {
					return nil, exception.New(
						exception.CodeParamInvalid,
						"私钥已加密，请提供私钥密码", 400, nil,
					)
				}
				pwPath := filepath.Join(tmpDir, "pw.txt")
				plainPath := filepath.Join(tmpDir, "plain.key")
				if err := os.WriteFile(pwPath, []byte(req.KeyPassword), 0600); err != nil {
					return nil, exception.New(
						exception.CodeInternalError,
						fmt.Sprintf("写口令文件失败：%v", err), 500, nil,
					)
				}
				if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
					return nil, exception.New(
						exception.CodeParamInvalid,
						"私钥密码错误或解密失败", 400, nil,
					)
				}
				data, err := os.ReadFile(plainPath)
				if err != nil {
					return nil, exception.New(
						exception.CodeInternalError,
						fmt.Sprintf("读解密私钥失败：%v", err), 500, nil,
					)
				}
				plainKeyPEM = data
			}
		}

		// 5.2 白盒加密落盘（无 .pass）
		keyAbs := filepath.Join(dir, pubSM3+".key.pem")

		plainTmp := filepath.Join(dir, ".import.csr.key.plain.tmp")
		if err := os.WriteFile(plainTmp, plainKeyPEM, 0600); err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("写临时私钥失败：%v", err), 500, nil,
			)
		}
		defer os.Remove(plainTmp)

		if err := s.whitebox.EncryptFile(ctx, plainTmp, keyAbs); err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil,
			)
		}
		_ = os.Chmod(keyAbs, 0600)
	}

	// 6. 落库
	now := time.Now().UTC()
	csr := &models.CSR{
		CSRID:     "csr-" + dirNo + "-" + pubSM3[:12],
		SubjectCN: extractCNFromSubject(subject),
		Algorithm: algorithm,
		CSRPath:   csrAbs,
		KeyRef:    nil,
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
func (s *CSRService) parseCSRInfo(csrAbs string) (string, string, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return "", "", exception.New(
			exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil,
		)
	}
	if _, err := os.Stat(csrAbs); err != nil {
		return "", "", exception.New(
			exception.CodeNotFound,
			fmt.Sprintf("P10 文件不存在：%s", csrAbs), 404, nil,
		)
	}

	subjOut, err := RunOpenSSL(bin, "req", "-in", csrAbs, "-noout", "-subject", "-nameopt", "RFC2253")
	if err != nil {
		return "", "", exception.New(
			exception.CodeParamInvalid, "CSR 解析失败", 400, nil,
		)
	}
	subject := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(subjOut), "subject="))

	textOut, _ := RunOpenSSL(bin, "req", "-in", csrAbs, "-noout", "-text")
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

// =============================================================================
// 下载
// =============================================================================

func (s *CSRService) Download(csrID string) ([]byte, error) {
	csr, err := s.Get(csrID)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(csr.CSRPath)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("读取 P10 文件失败：%v", err), 500, nil,
		)
	}
	return data, nil
}

// DownloadKey 下载 P10 对应的私钥，可选 AES-256-CBC 口令加密。
//
// ★ 无 .pass：白盒密钥内嵌二进制，直接解密/重加密。
func (s *CSRService) DownloadKey(
	csrID string, encrypt bool, password string,
) ([]byte, error) {
	csr, err := s.Get(csrID)
	if err != nil {
		return nil, err
	}

	keyAbs, err := s.deriveWhiteboxKeyPath(csr)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(keyAbs); err != nil {
		return nil, exception.New(
			exception.CodeParamInvalid, "该 P10 未关联私钥", 400, nil,
		)
	}

	// ★ 无 pass 参数
	plainTmp, err := s.whitebox.DecryptToTemp(context.Background(), keyAbs, s.layout.TmpDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = s.whitebox.ReEncrypt(context.Background(), plainTmp, keyAbs)
		_ = os.Remove(plainTmp)
	}()

	plain, err := os.ReadFile(plainTmp)
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("读取解密私钥失败：%v", err), 500, nil,
		)
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

// deriveWhiteboxKeyPath 从 CSR 路径推导白盒私钥路径（无 .pass）。
func (s *CSRService) deriveWhiteboxKeyPath(csr *models.CSR) (string, error) {
	if csr == nil || csr.CSRPath == "" {
		return "", exception.New(
			exception.CodeInternalError, "P10 路径为空", 500, nil,
		)
	}
	dir := filepath.Dir(csr.CSRPath)
	base := filepath.Base(csr.CSRPath)
	if !strings.HasSuffix(base, ".req.csr") {
		return "", exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("P10 文件名不符合规范：%s", base), 500, nil,
		)
	}
	pubSM3 := strings.TrimSuffix(base, ".req.csr")
	return filepath.Join(dir, pubSM3+".key.pem"), nil
}

// =============================================================================
// 删除
// =============================================================================

// Delete 删除 P10（硬删除），同时清理白盒私钥（无 .pass）。
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

	// 清理磁盘文件：CSR + 白盒私钥
	if csr.CSRPath != "" {
		_ = os.Remove(csr.CSRPath)
		if keyAbs, err := s.deriveWhiteboxKeyPath(&csr); err == nil {
			_ = os.Remove(keyAbs)
		}
	}
	return nil
}
