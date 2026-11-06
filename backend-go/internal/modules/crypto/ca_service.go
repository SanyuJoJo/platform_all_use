package crypto

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

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
	layout *PathLayout, whitebox *WhiteboxClient,
) *CAService {
	return &CAService{
		db: db, caller: caller, files: files, parser: parser,
		keys: keys, layout: layout, whitebox: whitebox,
	}
}

// =============================================================================
// 域 bundle 维护
// =============================================================================

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
		if name == "bundle.crt" || !strings.HasSuffix(name, ".cert.pem") {
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}
		abs := filepath.Join(dir, name)
		data, err := os.ReadFile(abs)
		if err != nil || len(data) == 0 {
			continue
		}
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
// 路径查询
// =============================================================================

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

// fillCADerivedFields 填充 CA 的非持久化展示字段。
//
// ★ 契约：
//
//	ca.Algorithm           = DB 规范 ID（"SM2" / "RSA2048" / "falcon1024" …）
//	                          —— 创建时写入，**绝不允许被解析结果覆盖**。
//	ca.PublicKeyAlgorithm  = 证书解析展示名（"ECC(256)" / "Falcon(1024)" …）
//	                          —— 非持久化（gorm:"-"），**永不被 algorithm 影响**。
//	ca.ValidityDays        = 若 DB 中为 0，用 NotAfter - NotBefore 回填。
//
// ★ 只在 algorithm 为空（老数据 / 导入未填）时用解析结果兜底。
func (s *CAService) fillCADerivedFields(ca *models.CA) {
	if ca == nil {
		return
	}
	ca.HasPrivateKey = ca.KeyRef != "" || ca.KeyPath != ""
	ca.Domain = extractDomainFromCertPath(ca.CertPath)

	abs, err := s.resolveCAPath(ca.CertPath)
	if err != nil {
		return
	}
	d, err := s.parser.Parse(abs)
	if err != nil {
		return
	}

	ca.NotBefore = d.NotBefore
	ca.NotAfter = d.NotAfter

	// ★ 展示名字段：永远用解析结果填充（与 algorithm 完全独立）
	if strings.TrimSpace(d.PublicKeyAlgorithm) != "" {
		ca.PublicKeyAlgorithm = d.PublicKeyAlgorithm
	}

	// ★ 规范 ID：仅在 algorithm 为空时兜底（老数据 / 导入未填）
	if strings.TrimSpace(ca.Algorithm) == "" {
		if id := displayNameToID(d.PublicKeyAlgorithm, ""); id != "" {
			ca.Algorithm = id
		}
	}

	// ★ validity_days 回填
	if ca.ValidityDays == 0 && d.NotBefore != "" && d.NotAfter != "" {
		if t1, e1 := time.ParseInLocation("2006-01-02 15:04:05", d.NotBefore, cstZone); e1 == nil {
			if t2, e2 := time.ParseInLocation("2006-01-02 15:04:05", d.NotAfter, cstZone); e2 == nil {
				days := int(t2.Sub(t1).Hours() / 24)
				if days > 0 {
					ca.ValidityDays = days
				}
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
		fmt.Sprintf("CA 证书文件不存在：%s", certPath), 404, nil,
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

func (s *CAService) Delete(caID string) error {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("查询 CA 失败: %v", err), 500, nil)
	}

	domain := extractDomainFromCertPath(ca.CertPath)

	var refCount int64
	_ = s.db.Model(&models.Certificate{}).Where("ca_id = ?", caID).Count(&refCount).Error

	files := s.collectCAFiles(&ca)

	if err := s.db.Unscoped().Where("ca_id = ?", caID).Delete(&models.CA{}).Error; err != nil {
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
