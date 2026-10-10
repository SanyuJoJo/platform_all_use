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
	TmpDir     string // ★ 白盒临时解密目录
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
//
// 规则：取当前 ServerRoot 下所有 4 位数字目录的最大值 + 1。
// 不复用已删除编号，保证历史审计可追溯。
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
//
// 遍历 server/<dir_no>/ 下是否存在 <pubSM3>.req.csr。
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
