// backend-go/internal/modules/tcps_proxy/paths.go
package tcps_proxy
import (
	"os"
	"path/filepath"
)
// DeployRoot 部署根目录（与 ssl_proxy 共享）。
func DeployRoot() string {
	if v := os.Getenv("PLATS_DEPLOY_ROOT"); v != "" {
		return v
	}
	return "/opt/plats_tool"
}
func CoreRoot() string      { return filepath.Join(DeployRoot(), "core") }
func TengineBin() string    { return filepath.Join(CoreRoot(), "sbin", "nginx") }
func TemplateDir() string   { return filepath.Join(CoreRoot(), "conf", "template") }
// ConfRoot TCPS 配置根目录：core/conf/tcps。
func ConfRoot() string { return filepath.Join(CoreRoot(), "conf", TcpsSubDir) }
// PidRoot TCPS PID 根目录：core/pid/tcps。
func PidRoot() string { return filepath.Join(CoreRoot(), "pid", TcpsSubDir) }
// TmpRoot TCPS 临时目录：core/tmp/tcps。
func TmpRoot() string { return filepath.Join(CoreRoot(), "tmp", TcpsSubDir) }
// LogRoot TCPS 日志根目录：user_data/logs/tcps_proxy。
func LogRoot() string { return filepath.Join(DeployRoot(), "user_data", "logs", "tcps_proxy") }
// CertRoot 证书根目录：user_data/cert（与 ssl_proxy 共用）。
func CertRoot() string { return filepath.Join(DeployRoot(), "user_data", "cert") }
