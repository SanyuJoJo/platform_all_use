package pathguard
import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "github.com/yourorg/core-go/internal/platform"
)
var coreRoot string
// Init 初始化路径守卫。
func Init(root string) {
    coreRoot = root
}
// CoreRoot 返回 core 根目录。
func CoreRoot() string { return coreRoot }
// Resolve 解析受控路径。
//
//   - 相对路径 → 相对于 coreRoot
//   - 返回绝对路径
//   - 若路径不在 coreRoot/data 或 coreRoot/tmp 下，返回 error
func Resolve(p string) (string, error) {
    if p == "" {
        return "", fmt.Errorf("path is empty")
    }
    abs := p
    if !filepath.IsAbs(abs) {
        abs = filepath.Join(coreRoot, filepath.FromSlash(p))
    }
    abs = filepath.Clean(abs)
    dataDir := filepath.Join(coreRoot, "data")
    tmpDir := filepath.Join(coreRoot, "tmp")
    if !isUnder(abs, dataDir) && !isUnder(abs, tmpDir) {
        return "", fmt.Errorf("path not allowed: %s", p)
    }
    return abs, nil
}
// MustResolve 解析失败 panic（仅用于内部约定路径）。
func MustResolve(p string) string {
    abs, err := Resolve(p)
    if err != nil {
        panic(err)
    }
    return abs
}
// isUnder 判断 child 是否在 parent 下（含自身）。
func isUnder(child, parent string) bool {
    child = filepath.Clean(child)
    parent = filepath.Clean(parent)
    if child == parent {
        return true
    }
    sep := string(os.PathSeparator)
    return strings.HasPrefix(child, parent+sep)
}
// RequirePasswordFile 校验口令文件（受控目录 + Unix 权限 0600）。
func RequirePasswordFile(p string) (string, error) {
    abs, err := Resolve(p)
    if err != nil {
        return "", err
    }
    info, err := os.Stat(abs)
    if err != nil || info.IsDir() {
        return "", fmt.Errorf("password file not found: %s", p)
    }
    // Unix 强校验；Windows 跳过
    if !platform.IsWindows() {
        if info.Mode().Perm() != 0600 {
            return "", fmt.Errorf("password file permission must be 0600, got %o",
                info.Mode().Perm())
        }
    }
    return abs, nil
}
// EnsureDir 创建目录并设置安全权限。
func EnsureDir(p string, mode os.FileMode) error {
    return platform.SecureMkdirAll(p, mode)
}
