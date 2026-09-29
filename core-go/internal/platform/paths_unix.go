//go:build !windows
package platform
import "os"
// SecureChmod 设置安全文件权限（Unix）。
func SecureChmod(path string, mode os.FileMode) error {
    return os.Chmod(path, mode)
}
// SecureMkdirAll 创建目录并设置安全权限（Unix）。
func SecureMkdirAll(path string, mode os.FileMode) error {
    if err := os.MkdirAll(path, mode); err != nil {
        return err
    }
    // 显式 chmod，绕过 umask
    return os.Chmod(path, mode)
}
// HasSecurePermission 检查文件权限是否符合要求（Unix 强校验）。
func HasSecurePermission(path string, expect os.FileMode) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return info.Mode().Perm() == expect
}
