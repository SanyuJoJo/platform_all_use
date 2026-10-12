package crypto

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"backend-go/internal/config"
	"backend-go/internal/exception"
)

var (
	domainRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	dirNoRe  = regexp.MustCompile(`^[0-9]{4}$`)
)

// PathLayout 证书路径布局。
//
// 目录约定：
//
//	<CertRoot>/ca/<domain>/         CA 证书目录
//	<CertRoot>/server/<dir_no>/     SERVER/CSR 证书目录
//	<CertRoot>/tmp/                 白盒私钥临时解密目录
type PathLayout struct {
	CertRoot   string
	CARoot     string
	ServerRoot string
	TmpDir     string
}

func NewPathLayout(cfg *config.Config) *PathLayout {
	return &PathLayout{
		CertRoot:   filepath.Clean(cfg.CertRoot),
		CARoot:     filepath.Clean(cfg.CARoot),
		ServerRoot: filepath.Clean(cfg.ServerRoot),
		TmpDir:     filepath.Clean(cfg.CertTmpDir),
	}
}

// EnsureRoots 创建证书根目录（幂等）。
func (l *PathLayout) EnsureRoots() error {
	for _, d := range []string{l.CertRoot, l.CARoot, l.ServerRoot} {
		if d == "" {
			continue
		}
		if err := os.MkdirAll(d, 0750); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}
	if l.TmpDir != "" {
		if err := os.MkdirAll(l.TmpDir, 0700); err != nil {
			return fmt.Errorf("mkdir %s: %w", l.TmpDir, err)
		}
	}
	return nil
}

func (l *PathLayout) safeDomain(domain string) (string, error) {
	d := strings.TrimSpace(domain)
	if d == "" || !domainRe.MatchString(d) {
		return "", exception.New(exception.CodeParamInvalid, "非法域名目录", 400, nil)
	}
	return d, nil
}

// CADir CA 目录：<CARoot>/<domain>。
func (l *PathLayout) CADir(domain string) (string, error) {
	d, err := l.safeDomain(domain)
	if err != nil {
		return "", err
	}
	return filepath.Join(l.CARoot, d), nil
}

// CAPaths 返回 CA 相关路径。
//
//	dir/cert/key/pass/rehash
func (l *PathLayout) CAPaths(domain, pubSM3 string) (
	dir, cert, key, pass, rehash string, err error,
) {
	dir, err = l.CADir(domain)
	if err != nil {
		return
	}
	cert = filepath.Join(dir, pubSM3+".cert.pem")
	key = filepath.Join(dir, pubSM3+".key.pem")
	pass = filepath.Join(dir, pubSM3+".key.pem.pass")
	rehash = filepath.Join(dir, pubSM3+".0")
	return
}

// ServerDir server 目录：<ServerRoot>/<dir_no>。
func (l *PathLayout) ServerDir(dirNo string) (string, error) {
	if !dirNoRe.MatchString(dirNo) {
		return "", exception.New(exception.CodeParamInvalid, "非法 server 编号", 400, nil)
	}
	return filepath.Join(l.ServerRoot, dirNo), nil
}

// ServerPaths 返回 server 证书相关路径。
//
//	dir/cert/key/pass/csr
func (l *PathLayout) ServerPaths(dirNo, pubSM3 string) (
	dir, cert, key, pass, csr string, err error,
) {
	dir, err = l.ServerDir(dirNo)
	if err != nil {
		return
	}
	cert = filepath.Join(dir, pubSM3+".cert.pem")
	key = filepath.Join(dir, pubSM3+".key.pem")
	pass = filepath.Join(dir, pubSM3+".key.pem.pass")
	csr = filepath.Join(dir, pubSM3+".req.csr")
	return
}

// AllocServerDir 分配下一个可用的 server 目录编号（0000 起递增）。
func (l *PathLayout) AllocServerDir() (string, error) {
	if err := os.MkdirAll(l.ServerRoot, 0750); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(l.ServerRoot)
	if err != nil {
		return "", err
	}
	maxNo := -1
	for _, e := range entries {
		if !e.IsDir() || !dirNoRe.MatchString(e.Name()) {
			continue
		}
		n, _ := strconv.Atoi(e.Name())
		if n > maxNo {
			maxNo = n
		}
	}
	dirNo := fmt.Sprintf("%04d", maxNo+1)
	dir := filepath.Join(l.ServerRoot, dirNo)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}
	return dirNo, nil
}

// FindCSRDirByPubkeySM3 通过 CSR 公钥 SM3 定位其所在目录。
func (l *PathLayout) FindCSRDirByPubkeySM3(pubSM3 string) (string, error) {
	entries, err := os.ReadDir(l.ServerRoot)
	if err != nil {
		return "", err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if !e.IsDir() || !dirNoRe.MatchString(e.Name()) {
			continue
		}
		csr := filepath.Join(l.ServerRoot, e.Name(), pubSM3+".req.csr")
		if st, err := os.Stat(csr); err == nil && !st.IsDir() {
			return filepath.Join(l.ServerRoot, e.Name()), nil
		}
	}
	return "", exception.New(exception.CodeNotFound, "未找到对应 CSR 目录", 404, nil)
}

// EnsureUnderCertRoot 保证路径位于证书根目录下。
func (l *PathLayout) EnsureUnderCertRoot(p string) error {
	abs := filepath.Clean(p)
	root := filepath.Clean(l.CertRoot)
	if abs == root || strings.HasPrefix(abs, root+string(os.PathSeparator)) {
		return nil
	}
	return exception.New(exception.CodeParamInvalid, "路径不在证书根目录下", 400, nil)
}

// =============================================================================
// ★ 域管理
// =============================================================================

// ListDomains 列出 CARoot 下的所有合法域目录（含 safe）。
//
// 规则：
//   - 只统计 CARoot 的直接子目录；
//   - 目录名必须匹配 ^[A-Za-z0-9_-]+$；
//   - 目录名以 "." 开头的一律忽略；
//   - 结果按字典序排序，safe 优先排在首位；
//   - 若无任何合法域，至少返回 ["safe"]。
func (l *PathLayout) ListDomains() ([]string, error) {
	if err := os.MkdirAll(l.CARoot, 0750); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", l.CARoot, err)
	}
	entries, err := os.ReadDir(l.CARoot)
	if err != nil {
		return nil, err
	}
	var domains []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") || !domainRe.MatchString(name) {
			continue
		}
		domains = append(domains, name)
	}
	sort.Strings(domains)

	// safe 优先
	for i, d := range domains {
		if d == "safe" && i != 0 {
			domains = append([]string{"safe"}, append(domains[:i], domains[i+1:]...)...)
			break
		}
	}
	if len(domains) == 0 {
		domains = []string{"safe"}
	}
	return domains, nil
}

// CreateDomain 创建一个新的域目录。
//
// 校验规则：
//   - 名称必须匹配 ^[A-Za-z0-9_-]+$；
//   - 长度 1~64；
//   - 不能与已有域重名；
//   - 目录权限 0750。
func (l *PathLayout) CreateDomain(name string) error {
	n := strings.TrimSpace(name)
	if n == "" || len(n) > 64 {
		return exception.New(exception.CodeParamInvalid,
			"域名称长度必须在 1~64 之间", 400, nil)
	}
	if !domainRe.MatchString(n) {
		return exception.New(exception.CodeParamInvalid,
			"域名称只能包含字母、数字、下划线、短横线", 400, nil)
	}
	dir := filepath.Join(l.CARoot, n)
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return exception.New(exception.CodeParamInvalid,
			"域已存在："+n, 409, nil)
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建域目录失败：%v", err), 500, nil)
	}
	return nil
}

// ★ DeleteDomain 删除一个域目录。
//
// 约束：
//   - safe 为默认域，禁止删除；
//   - 名称必须匹配 ^[A-Za-z0-9_-]+$；
//   - 目录必须为空（不含任何文件 / 子目录），否则拒绝删除；
//   - 目录不存在 → 404。
func (l *PathLayout) DeleteDomain(name string) error {
	n := strings.TrimSpace(name)
	if n == "" {
		return exception.New(exception.CodeParamInvalid, "域名称不能为空", 400, nil)
	}
	if n == "safe" {
		return exception.New(exception.CodeParamInvalid,
			"默认域 safe 不允许删除", 400, nil)
	}
	if !domainRe.MatchString(n) {
		return exception.New(exception.CodeParamInvalid,
			"域名称只能包含字母、数字、下划线、短横线", 400, nil)
	}
	dir := filepath.Join(l.CARoot, n)
	st, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return exception.New(exception.CodeNotFound, "域不存在："+n, 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("查询域目录失败：%v", err), 500, nil)
	}
	if !st.IsDir() {
		return exception.New(exception.CodeParamInvalid,
			"域路径不是目录："+n, 400, nil)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取域目录失败：%v", err), 500, nil)
	}
	if len(entries) > 0 {
		return exception.New(exception.CodeParamInvalid,
			"域目录非空，无法删除："+n+"（请先删除该域下的所有 CA）", 400, nil)
	}
	if err := os.Remove(dir); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("删除域目录失败：%v", err), 500, nil)
	}
	return nil
}
