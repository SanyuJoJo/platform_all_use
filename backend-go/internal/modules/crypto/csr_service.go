package crypto

import (
	"bytes"
	"context"
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
// ★ CFCA attribute 判定
// =============================================================================

var cfcaTmpPubOIDBytes = []byte{
	0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x01, 0x09, 0x3f,
}

func (s *CSRService) detectCFCAAttribute(csrAbs string) bool {
	data, err := os.ReadFile(csrAbs)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(data)
	if block == nil || len(block.Bytes) == 0 {
		return false
	}
	return bytes.Contains(block.Bytes, cfcaTmpPubOIDBytes)
}

// =============================================================================
// 列表 / 详情
// =============================================================================

func (s *CSRService) List(page, pageSize int) (map[string]interface{}, error) {
	var csrs []models.CSR
	q := s.db.Model(&models.CSR{}).Where("status <> ?", "DELETED")
	result, err := paginateQuery(q, &csrs, page, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range csrs {
		s.fillCSRDerivedFields(&csrs[i])
	}
	result["items"] = csrs
	return result, nil
}

func (s *CSRService) Get(csrID string) (*models.CSR, error) {
	var csr models.CSR
	if err := s.db.Where("csr_id = ?", csrID).First(&csr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeNotFound, "P10 不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询 P10 失败", 500, nil)
	}
	s.fillCSRDerivedFields(&csr)
	return &csr, nil
}

// fillCSRDerivedFields 计算 CSR 的非持久化展示字段。
//
// ★ 填充项：
//
//	HasPrivateKey       : KeyRef 或磁盘 <csr_dir>/<pubSM3>.key.pem 存在；
//	IsCFCA              : CSR attributes 段是否携带 CFCA 临时公钥 OID；
//	Subject             : 从 CSR 文件解析完整 DN（RFC2253 规范化）；
//	PublicKeyAlgorithm  : 解析展示名（"ECC(256)" / "Falcon(1024)" …）；
//	                      —— 与 Algorithm（DB 规范 ID）完全独立；
//	Algorithm           : **仅在为空时**兜底为规范 ID（老数据）；
//	NotBefore/NotAfter  : CreatedAt（北京时间）/ 空。
func (s *CSRService) fillCSRDerivedFields(csr *models.CSR) {
	if csr == nil {
		return
	}

	// HasPrivateKey
	has := csr.KeyRef != nil && *csr.KeyRef != ""
	if !has {
		if keyAbs, err := s.deriveWhiteboxKeyPath(csr); err == nil {
			if st, err := os.Stat(keyAbs); err == nil && !st.IsDir() && st.Size() > 0 {
				has = true
			}
		}
	}
	csr.HasPrivateKey = has

	// IsCFCA
	if csr.CSRPath != "" {
		csr.IsCFCA = s.detectCFCAAttribute(csr.CSRPath)
	}

	// ★ Subject / PublicKeyAlgorithm / Algorithm 兜底
	if csr.CSRPath != "" {
		bin := s.parser.OpensslBin()
		if bin != "" {
			// 1. Subject
			if csr.Subject == "" {
				if out, err := RunOpenSSL(bin, "req", "-in", csr.CSRPath, "-noout", "-subject", "-nameopt", "RFC2253"); err == nil {
					subj := strings.TrimSpace(out)
					subj = strings.TrimSpace(strings.TrimPrefix(subj, "subject="))
					if subj != "" {
						csr.Subject = normalizeDNOrder(subj)
					}
				}
			}

			// 2. PublicKeyAlgorithm 展示名 + Algorithm 规范 ID 兜底（共用一次 openssl req -text）
			if strings.TrimSpace(csr.PublicKeyAlgorithm) == "" || strings.TrimSpace(csr.Algorithm) == "" {
				if textOut, err := RunOpenSSL(bin, "req", "-in", csr.CSRPath, "-noout", "-text"); err == nil {
					if strings.TrimSpace(csr.PublicKeyAlgorithm) == "" {
						if algName := extractPubAlgFromCSRText(textOut); algName != "" {
							csr.PublicKeyAlgorithm = normalizePublicKeyAlg(algName)
						}
					}
					if strings.TrimSpace(csr.Algorithm) == "" {
						if id := detectAlgorithmFromCSRText(textOut); id != "" {
							csr.Algorithm = id
						}
					}
				}
			}

			// 3. 万一 Algorithm 仍为空，用展示名反查一次兜底
			if strings.TrimSpace(csr.Algorithm) == "" && csr.PublicKeyAlgorithm != "" {
				if id := displayNameToID(csr.PublicKeyAlgorithm, ""); id != "" {
					csr.Algorithm = id
				}
			}
		}
	}

	// NotBefore / NotAfter
	if !csr.CreatedAt.IsZero() {
		csr.NotBefore = csr.CreatedAt.In(cstZone).Format("2006-01-02 15:04:05")
	} else {
		csr.NotBefore = ""
	}
	csr.NotAfter = ""
}

// =============================================================================
// 导入 P10
// =============================================================================

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

	pubSM3, err := PubkeySM3FromCSRPEM(bin, csrPEM)
	if err != nil {
		return nil, exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("计算 CSR 公钥 SM3 失败：%v", err), 400, nil,
		)
	}

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

	subject, algorithm, err := s.parseCSRInfo(csrAbs)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(
				exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil,
			)
		}

		plainKeyPEM := []byte(req.KeyPEM)

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

	now := time.Now().UTC()
	csr := &models.CSR{
		CSRID:     "csr-" + dirNo + "-" + pubSM3[:12],
		SubjectCN: extractCNFromSubject(subject),
		Algorithm: algorithm,   // ★ 规范 ID（detectAlgorithmFromCSRText 返回）
		CSRPath:   csrAbs,
		KeyRef:    nil,
		Status:    "NEW",
		CreatedAt: now,
	}
	s.fillCSRDerivedFields(csr)

	if err := s.db.Create(csr).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("P10 落库失败: %v", err), 500, nil,
		)
	}
	success = true
	return csr, nil
}

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
	algorithm := detectAlgorithmFromCSRText(textOut)

	return subject, algorithm, nil
}

// =============================================================================
// 算法识别
// =============================================================================

// detectAlgorithmFromCSRText 从 `openssl req -text` 输出中识别算法，
// 返回 AlgorithmRegistry 中的规范 ID（如 SM2 / RSA2048 / MLDSA65 /
// falcon512 / slh-dsa-*）。
func detectAlgorithmFromCSRText(textOut string) string {
	lower := strings.ToLower(textOut)
	if lower == "" {
		return ""
	}

	// SLH-DSA（优先匹配，最具体）
	for _, h := range []string{"sha2", "shake"} {
		for _, lvl := range []string{"128", "192", "256"} {
			for _, spd := range []string{"s", "f"} {
				id := "slh-dsa-" + h + "-" + lvl + spd
				if strings.Contains(lower, id) {
					return id
				}
			}
		}
	}

	// ML-DSA
	switch {
	case strings.Contains(lower, "ml-dsa-44") ||
		strings.Contains(lower, "mldsa44") ||
		strings.Contains(lower, "dilithium2"):
		return "MLDSA44"
	case strings.Contains(lower, "ml-dsa-65") ||
		strings.Contains(lower, "mldsa65") ||
		strings.Contains(lower, "dilithium3"):
		return "MLDSA65"
	case strings.Contains(lower, "ml-dsa-87") ||
		strings.Contains(lower, "mldsa87") ||
		strings.Contains(lower, "dilithium5"):
		return "MLDSA87"
	}

	// Falcon
	switch {
	case strings.Contains(lower, "falcon-512") ||
		strings.Contains(lower, "falcon512") ||
		strings.Contains(lower, "falcon(512)"):
		return "falcon512"
	case strings.Contains(lower, "falcon-1024") ||
		strings.Contains(lower, "falcon1024") ||
		strings.Contains(lower, "falcon(1024)"):
		return "falcon1024"
	}

	// SM2 / ECC
	if strings.Contains(lower, "sm2") ||
		strings.Contains(lower, "id-ecpublickey") ||
		strings.Contains(lower, "prime256v1") ||
		strings.Contains(lower, "secp256r1") {
		return "SM2"
	}

	// RSA
	if strings.Contains(lower, "rsa") {
		return detectRSAKeySize(lower)
	}

	return ""
}

func detectRSAKeySize(lowerText string) string {
	switch {
	case strings.Contains(lowerText, "(4096 bit)"):
		return "RSA4096"
	case strings.Contains(lowerText, "(2048 bit)"):
		return "RSA2048"
	case strings.Contains(lowerText, "(1024 bit)"):
		return "RSA1024"
	}
	return "RSA2048"
}

// =============================================================================
// 下载 / 删除
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

	csrDir := ""
	if csr.CSRPath != "" {
		csrDir = filepath.Dir(csr.CSRPath)
		_ = os.Remove(csr.CSRPath)
		if keyAbs, err := s.deriveWhiteboxKeyPath(&csr); err == nil {
			_ = os.Remove(keyAbs)
		}
	}

	s.tryRemoveEmptyServerDir(csrDir)

	return nil
}

func (s *CSRService) tryRemoveEmptyServerDir(dir string) {
	if dir == "" || s.layout == nil {
		return
	}
	abs := filepath.Clean(dir)
	root := filepath.Clean(s.layout.CertRoot)

	if abs == root {
		return
	}
	if !strings.HasPrefix(abs, root+string(os.PathSeparator)) {
		return
	}

	base := filepath.Base(abs)
	parent := filepath.Dir(abs)

	if parent != filepath.Clean(s.layout.ServerRoot) || !dirNoRe.MatchString(base) {
		return
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return
	}
	if len(entries) > 0 {
		return
	}

	if err := os.Remove(abs); err != nil {
		log.Warn().Str("dir", abs).Err(err).
			Msg("CSR Delete: 删除空目录失败（忽略）")
		return
	}
	log.Info().Str("dir", abs).Msg("CSR Delete: 已删除空 server 目录")
}
