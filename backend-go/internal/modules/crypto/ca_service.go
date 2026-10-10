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
//
// 依赖：
//   - db       : 数据库
//   - caller   : core 调用（解析旧模式的 keystore 引用等）
//   - files    : 受控文件读写（core/data 与 core/tmp）
//   - parser   : 证书解析（铜锁 openssl）
//   - keys     : 私钥加密检测 / 解密
//   - layout   : 证书路径布局（CertRoot / CARoot / ServerRoot）
//   - whitebox : 白盒 SM4 工具（导入私钥白盒加密、使用私钥前解密）
//
// 说明：
//   - CA 目录 <CARoot>/<domain>（如 ca/safe/）是校验兜底目录，必须始终存在；
//     Delete 只删文件，不删目录。
//   - 私钥有两种模式：
//       旧模式：KeyRef 指向 core keystore；
//       新模式：KeyPath 指向白盒密文文件 <pubkey_sm3>.key.pem。
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
func (s *CAService) List(page, pageSize int) (map[string]interface{}, error) {
	var cas []models.CA
	return paginateQuery(
		s.db.Model(&models.CA{}).Where("status <> ?", "DELETED"),
		&cas, page, pageSize,
	)
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

// =============================================================================
// 路径解析
// =============================================================================

// resolveCAPath 解析 CA 证书文件绝对路径。
//
// 优先级：
//  1. 绝对路径，且文件存在
//  2. 按 CertRoot 解析（新规范）
//  3. 按 CoreRoot 解析（旧数据）
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
//
// ★ 语义说明（与引用无关）：
//   - CA 与证书之间只有**签名关系**，不是外键依赖；
//   - 删除 CA 后，已签发证书仍然存在，但校验链会失败；
//   - 这是合理的业务语义，因此**不做引用检查**，直接删除。
//
// ★ 目录保留：<CARoot>/<domain>（如 ca/safe/）是校验兜底目录，必须始终存在。
//   本方法只删除文件，不删除目录。
//
// 执行顺序：
//  1. 收集磁盘文件（必须在 CertRoot 下）：
//     - <pubkey_sm3>.cert.pem     证书（CertPath）
//     - <pubkey_sm3>.key.pem      白盒私钥（KeyPath，新模式）
//     - <hash>.0                  指向本 CA 的 rehash 软链
//     - ca.ChainPath（旧数据兼容）
//  2. 先删 DB（Unscoped，物理删除）；
//  3. 再删磁盘文件（失败仅告警）。
//
// 附注：删除后，仍引用该 CA 的证书记录会形成"悬空 ca_id"。
//   前端展示这类证书时，可以提示"CA 已删除"，或按业务策略自动把
//   这些证书状态标记为 INVALID（当前不自动标记，保持证书原状）。
func (s *CAService) Delete(caID string) error {
	var ca models.CA
	if err := s.db.Where("ca_id = ?", caID).First(&ca).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeNotFound, "CA 不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("查询 CA 失败: %v", err), 500, nil)
	}

	// ★ 不再做引用检查（CA 与证书只有签名关系，无外键依赖）
	// 仅统计引用数用于日志审计
	var refCount int64
	_ = s.db.Model(&models.Certificate{}).
		Where("ca_id = ?", caID).
		Count(&refCount).Error

	// 收集磁盘文件
	files := s.collectCAFiles(&ca)

	// 先删 DB（Unscoped 跳过软删钩子）
	if err := s.db.Unscoped().
		Where("ca_id = ?", caID).
		Delete(&models.CA{}).Error; err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("删除 CA 记录失败: %v", err), 500, nil)
	}

	// 再删磁盘文件（失败仅告警）
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

	// ★ 目录保留，不做任何删除（ca/safe/ 必须始终存在）

	log.Info().
		Str("ca_id", caID).
		Int("files_removed", removed).
		Int("files_total", len(files)).
		Int64("referenced_certs", refCount).
		Msg("CA 已硬删除（目录保留，引用证书保留）")
	return nil
}

// collectCAFiles 收集 CA 关联的所有磁盘文件（绝对路径，去重）。
//
// 来源：
//  1. ca.CertPath（.cert.pem）
//  2. 从 CertPath 推导 .key.pem、rehash 软链 <hash>.0
//  3. ca.KeyPath（新模式）
//  4. ca.ChainPath（旧数据兼容）
//
// 所有路径必须位于 CertRoot 下，否则跳过（防越权删除）。
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

	// 1. 证书本体
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

	// 2. 从 CertPath 推导同目录的其他文件
	if certAbs != "" && strings.HasSuffix(certAbs, ".cert.pem") {
		base := strings.TrimSuffix(certAbs, ".cert.pem")
		add(base + ".key.pem")

		for _, link := range findRehashLinksInDir(filepath.Dir(certAbs), filepath.Base(certAbs)) {
			add(link)
		}
	}

	// 3. KeyPath 字段（新模式）
	if ca.KeyPath != "" {
		add(ca.KeyPath)
	}

	// 4. ChainPath 字段（旧数据兼容）
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
