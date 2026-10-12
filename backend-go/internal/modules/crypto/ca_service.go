package crypto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// CAService 只负责 CA 相关业务。
type CAService struct {
	db       *gorm.DB
	caller   *CoreCaller
	files    *FileStore
	parser   *CertParser
	keys     *KeyCrypto
	layout   *PathLayout
	whitebox *WhiteboxClient
}

func NewCAService(
	db *gorm.DB, caller *CoreCaller, files *FileStore,
	parser *CertParser, keys *KeyCrypto,
	layout *PathLayout,
	whitebox *WhiteboxClient,
) *CAService {
	return &CAService{
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

// List 返回 CA 列表（不含已软删）。
//
// ★ 运行时填充：
//   - HasPrivateKey：KeyRef / KeyPath 非空；
//   - Domain       ：从 CertPath 推导（<CertRoot>/ca/<domain>/<pubkey_sm3>.cert.pem）；
//   - NotBefore / NotAfter：解析证书获得（"2006-01-02 15:04:05" 北京时间）；
//   - Algorithm    ：★ 用证书解析出的 PublicKeyAlgorithm 覆盖（如 "ECC(256)"），
//                     确保与 GetDetail 返回一致；
//                     仅内存覆盖，不写回 DB。
func (s *CAService) List(page, pageSize int) (map[string]interface{}, error) {
	var cas []models.CA
	result, err := paginateQuery(
		s.db.Model(&models.CA{}).Where("status <> ?", "DELETED"),
		&cas, page, pageSize,
	)
	if err != nil {
		return nil, err
	}
	for i := range cas {
		s.fillCADerivedFields(&cas[i])
	}
	result["items"] = cas
	return result, nil
}

// Get 查询 CA 元数据。
func (s *CAService) Get(caID string) (*models.CA, error) {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询 CA 失败", 500, nil)
	}
	s.fillCADerivedFields(&ca)
	return &ca, nil
}

// GetDetail 解析 CA 证书，返回结构化字段。
func (s *CAService) GetDetail(caID string) (*CADetail, error) {
	ca, err := s.Get(caID)
	if err != nil {
		return nil, err
	}
	abs, err := s.resolveCAPath(ca.CertPath)
	if err != nil {
		return nil, err
	}
	return s.parser.Parse(abs)
}

// fillCADerivedFields 计算 CA 的非持久化展示字段。
//
// 计算项：
//   - HasPrivateKey：KeyRef / KeyPath 非空；
//   - Domain       ：从 CertPath 推导；
//   - NotBefore / NotAfter / Algorithm：从证书解析获得。
//
// ★ Algorithm 覆盖规则：
//   - 解析成功且 PublicKeyAlgorithm 非空 → 覆盖为解析结果（如 "ECC(256)"）；
//   - 解析失败或字段为空 → 保留 DB 原值（如 "SM2"）；
//   - 仅内存修改，不回写 DB。
func (s *CAService) fillCADerivedFields(ca *models.CA) {
	if ca == nil {
		return
	}

	// 私钥
	ca.HasPrivateKey = ca.KeyRef != "" || ca.KeyPath != ""

	// 域
	ca.Domain = extractDomainFromCertPath(ca.CertPath)

	// 从证书解析：生效/失效时间、公钥算法
	if abs, err := s.resolveCAPath(ca.CertPath); err == nil {
		if d, err := s.parser.Parse(abs); err == nil {
			ca.NotBefore = d.NotBefore
			ca.NotAfter = d.NotAfter
			if strings.TrimSpace(d.PublicKeyAlgorithm) != "" {
				// ★ 仅内存覆盖：让列表返回的 public_key_algorithm
				//   与详情（getCADetail）保持一致。
				ca.Algorithm = d.PublicKeyAlgorithm
			}
		}
	}
}

// =============================================================================
// ★ 域管理
// =============================================================================

// ListDomains 列出所有可用域。
func (s *CAService) ListDomains() ([]string, error) {
	if s.layout == nil {
		return []string{"safe"}, nil
	}
	return s.layout.ListDomains()
}

// CreateDomain 创建新域。
func (s *CAService) CreateDomain(name string) error {
	if s.layout == nil {
		return exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
	}
	return s.layout.CreateDomain(name)
}

// DeleteDomain 删除域。
//
// 约束：
//   - safe 禁止删除；
//   - 域目录必须为空（不含任何文件）。
func (s *CAService) DeleteDomain(name string) error {
	if s.layout == nil {
		return exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
	}
	return s.layout.DeleteDomain(name)
}

// =============================================================================
// 路径解析
// =============================================================================

// resolveCAPath 解析 CA 证书文件绝对路径。
func (s *CAService) resolveCAPath(certPath string) (string, error) {
	if certPath == "" {
		return "", exception.New(exception.CodeInternalError, "CA 证书路径为空", 500, nil)
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
		fmt.Sprintf("CA 证书文件不存在：%s", certPath),
		404, nil,
	)
}

// ReadCACert 读取 CA 证书内容。
func (s *CAService) ReadCACert(ca *models.CA) ([]byte, error) {
	if ca == nil {
		return nil, exception.New(exception.CodeInternalError, "CA 为空", 500, nil)
	}
	abs, err := s.resolveCAPath(ca.CertPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(abs)
}

// underCertRoot 校验路径位于 CertRoot 下。
func (s *CAService) underCertRoot(p string) bool {
	if s.layout == nil || s.layout.CertRoot == "" {
		return false
	}
	abs := filepath.Clean(p)
	root := filepath.Clean(s.layout.CertRoot)
	return abs == root || strings.HasPrefix(abs, root+string(os.PathSeparator))
}

// =============================================================================
// 删除（硬删除）
// =============================================================================

// Delete 硬删除 CA：DB 记录 + 磁盘文件全部清除。
func (s *CAService) Delete(caID string) error {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("查询 CA 失败: %v", err), 500, nil)
	}

	var refCount int64
	_ = s.db.Model(&models.Certificate{}).
		Where("ca_id = ?", caID).
		Count(&refCount).Error

	files := s.collectCAFiles(&ca)

	if err := s.db.Unscoped().
		Where("ca_id = ?", caID).
		Delete(&models.CA{}).Error; err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("删除 CA 记录失败: %v", err), 500, nil)
	}

	removed := 0
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			if !os.IsNotExist(err) {
				log.Warn().Str("file", f).Err(err).
					Msg("Delete: 删除文件失败（忽略，可能残留孤儿文件）")
			}
		} else {
			removed++
		}
	}

	log.Info().
		Str("ca_id", caID).
		Int("files_removed", removed).
		Int("files_total", len(files)).
		Int64("referenced_certs", refCount).
		Msg("CA 已硬删除（目录保留，引用证书保留）")
	return nil
}

// collectCAFiles 收集 CA 关联的所有磁盘文件（绝对路径，去重）。
func (s *CAService) collectCAFiles(ca *models.CA) []string {
	if ca == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var out []string

	add := func(p string) {
		if p == "" {
			return
		}
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(s.layout.CertRoot, abs)
		}
		abs = filepath.Clean(abs)
		if !s.underCertRoot(abs) {
			return
		}
		if _, ok := seen[abs]; ok {
			return
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	}

	certAbs := ""
	if ca.CertPath != "" {
		if abs, err := s.resolveCAPath(ca.CertPath); err == nil {
			certAbs = abs
			add(abs)
		} else {
			abs := ca.CertPath
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(s.layout.CertRoot, abs)
			}
			certAbs = filepath.Clean(abs)
			add(certAbs)
		}
	}

	if certAbs != "" && strings.HasSuffix(certAbs, ".cert.pem") {
		base := strings.TrimSuffix(certAbs, ".cert.pem")
		add(base + ".key.pem")

		for _, link := range findRehashLinksInDir(filepath.Dir(certAbs), filepath.Base(certAbs)) {
			add(link)
		}
	}

	if ca.KeyPath != "" {
		add(ca.KeyPath)
	}

	if ca.ChainPath != nil && *ca.ChainPath != "" {
		add(*ca.ChainPath)
	}

	return out
}

// findRehashLinksInDir 查找 dir 下所有指向 targetBase 的 <hash>.0 软链。
func findRehashLinksInDir(dir, targetBase string) []string {
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
		if linkTarget == targetBase {
			links = append(links, full)
		}
	}
	return links
}

// extractDomainFromCertPath 从证书路径中提取域目录名。
//
// 典型输入：
//
//	<CertRoot>/ca/safe/a2426eed....cert.pem   → "safe"
//	<CertRoot>/ca/biz-01/xxx.cert.pem         → "biz-01"
//	relative/ca/safe/xxx.cert.pem             → "safe"
//	xxx.cert.pem                              → ""
func extractDomainFromCertPath(certPath string) string {
	if certPath == "" {
		return ""
	}
	clean := strings.ReplaceAll(filepath.Clean(certPath), "\\", "/")
	parts := strings.Split(clean, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "ca" && i+1 < len(parts)-1 {
			domain := parts[i+1]
			if domain == "" {
				continue
			}
			if strings.Contains(domain, ".pem") || strings.Contains(domain, ".cert") {
				continue
			}
			return domain
		}
	}
	return ""
}
