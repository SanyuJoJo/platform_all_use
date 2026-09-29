package platform
import (
    "os"
    "runtime"
)
// OS 返回归一化的操作系统标识。
func OS() string {
    switch runtime.GOOS {
    case "linux":
        return "linux"
    case "darwin":
        return "darwin"
    case "windows":
        return "windows"
    default:
        return runtime.GOOS
    }
}
// Arch 返回归一化的 CPU 架构。
func Arch() string {
    switch runtime.GOARCH {
    case "amd64":
        return "amd64"
    case "arm64":
        return "arm64"
    case "386":
        return "386"
    default:
        return runtime.GOARCH
    }
}
// IsWindows 是否为 Windows 平台。
func IsWindows() bool { return runtime.GOOS == "windows" }
// ExeSuffix 返回平台可执行文件后缀。
func ExeSuffix() string {
    if IsWindows() {
        return ".exe"
    }
    return ""
}
// PathSeparator 平台路径分隔符（用于配置解析）。
func PathSeparator() string { return string(os.PathSeparator) }
