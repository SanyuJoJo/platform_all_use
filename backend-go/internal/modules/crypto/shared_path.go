package crypto
import (
	"os"
	"path/filepath"
	"strings"
	"backend-go/internal/exception"
)
// PathGuard 受控路径校验。所有涉及文件路径的入口都应经过它。
type PathGuard struct {
	coreRoot string
}
func NewPathGuard(coreRoot string) *PathGuard {
	return &PathGuard{coreRoot: coreRoot}
}
// UnderCore 判断给定路径是否落在 coreRoot/data 或 coreRoot/tmp 下。
func (g *PathGuard) UnderCore(relOrAbs string) bool {
	if relOrAbs == "" {
		return false
	}
	abs := relOrAbs
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(g.coreRoot, relOrAbs)
	}
	abs = filepath.Clean(abs)
	for _, sub := range []string{"data", "tmp"} {
		base := filepath.Join(g.coreRoot, sub)
		if abs == base || strings.HasPrefix(abs, base+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}
// Resolve 返回绝对路径。若不在受控目录则抛出 PATH_NOT_ALLOWED。
func (g *PathGuard) Resolve(path, field string) (string, error) {
	if !g.UnderCore(path) {
		return "", exception.New(
			exception.CodeParamInvalid, "路径不在受控目录", 400, nil,
		)
	}
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(g.coreRoot, path)
	}
	return filepath.Clean(abs), nil
}
// RequirePasswordFile 校验口令文件：受控目录 + 权限 0600，返回绝对路径。
func (g *PathGuard) RequirePasswordFile(path string) (string, error) {
	abs, err := g.Resolve(path, "password_file")
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return "", exception.New(
			exception.CodeParamInvalid, "口令文件不存在", 400, nil,
		)
	}
	if st.Mode().Perm() != 0600 {
		return "", exception.New(
			exception.CodeParamInvalid, "口令文件权限必须为 0600", 400, nil,
		)
	}
	return abs, nil
}
