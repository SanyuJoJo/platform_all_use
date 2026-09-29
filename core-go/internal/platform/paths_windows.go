//go:build windows
package platform
import "os"
// SecureChmod Windows 无 POSIX 权限位，仅保证文件存在（可选：后续可加 ACL）。
func SecureChmod(path string, mode os.FileMode) error {
    // Windows 下 chmod 语义有限（只区分只读/可写），此处不报错
    _ = os.Chmod(path, mode)
    return nil
}
// SecureMkdirAll 创建目录（Windows）。
func SecureMkdirAll(path string, mode os.FileMode) error {
    return os.MkdirAll(path, mode)
}
// HasSecurePermission Windows 不做严格权限校验，仅保证文件存在。
func HasSecurePermission(path string, expect os.FileMode) bool {
    _, err := os.Stat(path)
    return err == nil
}
