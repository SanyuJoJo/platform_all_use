package license
import "time"
// LicenseStatusOut License 状态响应。
//
// 严格对应 Python 版 LicenseStatusOut：不含 machine_code / current_machine_code
// 字段（v1.1 起已移除），机器码信息仅用于服务端日志排障。
type LicenseStatusOut struct {
	IsValid           bool     `json:"is_valid"`
	LicenseType       *string  `json:"license_type"`
	ExpiresAt         *string  `json:"expires_at"`
	DaysRemaining     *int     `json:"days_remaining"`
	MaxUsers          *int     `json:"max_users"`
	CurrentUsers      int      `json:"current_users"`
	AuthorizedModules []string `json:"authorized_modules"`
	IsExpired         bool     `json:"is_expired"`
	IsExpiringSoon    bool     `json:"is_expiring_soon"`
}
// LicenseImportResp License 导入/激活响应。
type LicenseImportResp struct {
	ID                uint     `json:"id"`
	LicenseKey        string   `json:"license_key"`
	LicenseType       string   `json:"license_type"`
	ExpiresAt         string   `json:"expires_at"`
	AuthorizedModules []string `json:"authorized_modules"`
	MaxUsers          *int     `json:"max_users"`
	MachineCode       *string  `json:"machine_code"`
}
// LicenseActivateReq 在线激活请求。
type LicenseActivateReq struct {
	ActivationCode string `json:"activation_code" binding:"required,min=1,max=255"`
	MachineCode    string `json:"machine_code" binding:"required,min=1,max=255"`
}
// ModuleAuthorizationOut 模块授权状态。
type ModuleAuthorizationOut struct {
	ModuleID     string  `json:"module_id"`
	ModuleName   string  `json:"module_name"`
	IsAuthorized bool    `json:"is_authorized"`
	ExpiresAt    *string `json:"expires_at"`
}
// formatDateTime 将时间格式化为 Python datetime.isoformat() 兼容的字符串。
//
// 规则：
//   - 零值返回空串；
//   - 微秒为 0 时省略小数部分（与 Python 默认行为一致）；
//   - 微秒非 0 时保留 6 位（与 Python 的微秒精度一致）。
func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	t = t.UTC()
	if t.Nanosecond()/1000 == 0 {
		return t.Format("2006-01-02T15:04:05")
	}
	return t.Format("2006-01-02T15:04:05.000000")
}
