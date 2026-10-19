package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

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
// ★ 路径查询 / 私钥加密（供 SSL 代理使用）
// =============================================================================

// GetCertPath 返回证书的绝对路径，供 nginx ssl_certificate 直接引用。
//
// 证书是公开数据，SSL 代理不需要复制到自己的目录，直接引用一期磁盘文件即可。
// 返回前校验文件存在性，避免 nginx -t 阶段因文件缺失失败。
func (s *CertService) GetCertPath(certID string) (string, error) {
	c, err := s.Get(certID)
	if err != nil {
		return "", err
	}
	abs, err := s.resolveCertPath(c.CertPath)
	if err != nil {
		return "", err
	}
	return abs, nil
}

// SealPrivateKey 就地加密一个明文私钥文件（白盒 SM4）。
//
// 调用时机：SSL 代理启动 nginx 成功后，立刻把临时落盘的明文私钥加密，
// 保证私钥在磁盘上始终处于白盒密文状态（除极短暂的启动窗口）。
func (s *CertService) SealPrivateKey(path string) error {
	if s.whitebox == nil {
		return exception.New(exception.CodeInternalError,
			"whitebox 未初始化", 500, nil)
	}
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return exception.New(exception.CodeNotFound,
				"私钥文件不存在: "+path, 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("访问私钥文件失败: %v", err), 500, nil)
	}
	if st.IsDir() || st.Size() == 0 {
		return exception.New(exception.CodeParamInvalid,
			"私钥文件非法（空或目录）: "+path, 400, nil)
	}
	if err := s.whitebox.EncryptInPlace(context.Background(), path); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("加密私钥失败: %v", err), 500, nil)
	}
	log.Debug().
		Str("path", path).
		Msg("CertService: 私钥已就地加密")
	return nil
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
	result, err := paginateQuery(q, &certs, page, pageSize)
	if err != nil {
		return nil, err
	}
	for i := range certs {
		certs[i].HasPrivateKey =
			(certs[i].KeyRef != nil && *certs[i].KeyRef != "") ||
				(certs[i].KeyPath != nil && *certs[i].KeyPath != "")
	}
	result["items"] = certs
	return result, nil
}

func (s *CertService) Get(certID string) (*models.Certificate, error) {
	var c models.Certificate
	if err := s.db.Where("cert_id = ?", certID).First(&c).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询证书失败", 500, nil)
	}
	c.HasPrivateKey =
		(c.KeyRef != nil && *c.KeyRef != "") ||
			(c.KeyPath != nil && *c.KeyPath != "")
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
// 删除
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
