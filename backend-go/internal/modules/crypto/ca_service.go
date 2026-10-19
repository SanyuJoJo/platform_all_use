package crypto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
// ★ 域 bundle 维护
// =============================================================================

// RebuildDomainBundle 重建域内的"全量根证书"文件 <CARoot>/<domain>/bundle.crt。
//
// ★ 数据来源：**磁盘文件系统**，而非 DB
//   - 目录下所有 *.cert.pem（跳过 .0 rehash 软链、.key.pem、bundle.crt 自身）；
//   - 按文件名排序，保证多次重建的输出稳定；
//   - 与 DB 记录解耦：即使 DB 记录与磁盘文件不一致（如手工导入的 CA），
//     bundle.crt 也忠实反映磁盘现状，与 openssl -CApath 语义一致。
//
// ★ 空域处理：写 0 字节文件，表示"该域无 CA"。
//
// ★ 调用时机：
//   - CA Import 成功后（新增 CA 自动进 bundle）；
//   - CA Import 撞到"已存在"时（幂等重建，保证历史数据兼容）；
//   - CA Delete 成功后（删除 CA 自动从 bundle 移除）；
//   - 后端启动时可遍历所有域做一次全量重建。
func (s *CAService) RebuildDomainBundle(domain string) error {
	if s.layout == nil {
		return fmt.Errorf("PathLayout 未初始化")
	}
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain 为空")
	}

	dir := filepath.Join(s.layout.CARoot, domain)
	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("域目录不存在: %s", dir)
		}
		return fmt.Errorf("访问域目录失败: %s: %w", dir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("域路径不是目录: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取域目录失败: %w", err)
	}
	// 按文件名字典序排序，保证 bundle 内容稳定可 diff
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var (
		buf   []byte
		count int
	)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// 排除 bundle.crt 自身
		if name == "bundle.crt" {
			continue
		}
		// 只收集 .cert.pem
		if !strings.HasSuffix(name, ".cert.pem") {
			continue
		}
		// 跳过 rehash 软链（.0）
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}
		abs := filepath.Join(dir, name)
		data, err := os.ReadFile(abs)
		if err != nil {
			log.Warn().Str("file", abs).Err(err).
				Msg("RebuildDomainBundle: 读取 CA 证书失败，跳过")
			continue
		}
		if len(data) == 0 {
			continue
		}
		// 相邻证书之间保证换行分隔，避免 PEM 头尾粘连
		if len(buf) > 0 && buf[len(buf)-1] != '\n' {
			buf = append(buf, '\n')
		}
		buf = append(buf, data...)
		count++
	}

	bundlePath := filepath.Join(dir, "bundle.crt")
	if err := os.WriteFile(bundlePath, buf, 0644); err != nil {
		return fmt.Errorf("写入 bundle.crt 失败: %w", err)
	}

	log.Info().
		Str("domain", domain).
		Str("bundle", bundlePath).
		Int("certs", count).
		Int("bytes", len(buf)).
		Msg("CA 域 bundle 已重建")
	return nil
}

// =============================================================================
// ★ 路径查询（供 SSL 代理使用）
// =============================================================================

// GetCAPath 返回 CA 证书的绝对路径，供 nginx ssl_client_certificate 直接引用。
func (s *CAService) GetCAPath(caID string) (string, error) {
	ca, err := s.Get(caID)
	if err != nil {
		return "", err
	}
	return s.resolveCAPath(ca.CertPath)
}

// =============================================================================
// 列表 / 详情
// =============================================================================

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

func (s *CAService) fillCADerivedFields(ca *models.CA) {
	if ca == nil {
		return
	}
	ca.HasPrivateKey = ca.KeyRef != "" || ca.KeyPath != ""
	ca.Domain = extractDomainFromCertPath(ca.CertPath)
	if abs, err := s.resolveCAPath(ca.CertPath); err == nil {
		if d, err := s.parser.Parse(abs); err == nil {
			ca.NotBefore = d.NotBefore
			ca.NotAfter = d.NotAfter
			if strings.TrimSpace(d.PublicKeyAlgorithm) != "" {
				ca.Algorithm = d.PublicKeyAlgorithm
			}
		}
	}
}

// =============================================================================
// 域管理
// =============================================================================

func (s *CAService) ListDomains() ([]string, error) {
	if s.layout == nil {
		return []string{"safe"}, nil
	}
	return s.layout.ListDomains()
}

func (s *CAService) CreateDomain(name string) error {
	if s.layout == nil {
		return exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
	}
	return s.layout.CreateDomain(name)
}

func (s *CAService) DeleteDomain(name string) error {
	if s.layout == nil {
		return exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
	}
	return s.layout.DeleteDomain(name)
}

// =============================================================================
// 路径解析
// =============================================================================

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

func (s *CAService) underCertRoot(p string) bool {
	if s.layout == nil || s.layout.CertRoot == "" {
		return false
	}
	abs := filepath.Clean(p)
	root := filepath.Clean(s.layout.CertRoot)
	return abs == root || strings.HasPrefix(abs, root+string(os.PathSeparator))
}

// =============================================================================
// 删除
// =============================================================================

// Delete 硬删除 CA：DB 记录 + 磁盘文件全部清除，随后重建域 bundle。
func (s *CAService) Delete(caID string) error {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("查询 CA 失败: %v", err), 500, nil)
	}

	// ★ 先提取 domain（删除 DB 记录后仍可用，但提前取更清晰）
	domain := extractDomainFromCertPath(ca.CertPath)

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

	// ★ 重建域 bundle（磁盘证书变了，bundle 必须同步刷新）
	if domain != "" {
		if err := s.RebuildDomainBundle(domain); err != nil {
			log.Warn().Err(err).Str("domain", domain).
				Msg("Delete: 重建 bundle 失败（不阻断）")
		}
	}

	log.Info().
		Str("ca_id", caID).
		Str("domain", domain).
		Int("files_removed", removed).
		Int("files_total", len(files)).
		Int64("referenced_certs", refCount).
		Msg("CA 已硬删除（目录保留，引用证书保留）")
	return nil
}

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
