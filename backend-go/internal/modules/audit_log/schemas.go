package audit_log
import "time"
// AuditLogOut 操作日志响应对象。
type AuditLogOut struct {
	ID         uint    `json:"id"`
	UserID     *uint   `json:"user_id"`
	Username   *string `json:"username"`
	ModuleID   string  `json:"module_id"`
	Action     string  `json:"action"`
	Resource   *string `json:"resource"`
	ResourceID *string `json:"resource_id"`
	Detail     *string `json:"detail"`
	IP         *string `json:"ip"`
	UserAgent  *string `json:"user_agent"`
	Status     string  `json:"status"`
	ErrorCode  *int    `json:"error_code"`
	RequestID  *string `json:"request_id"`
	CreatedAt  string  `json:"created_at"`
}
// AuditLogListQuery 日志列表查询参数。
//
// v1.2（P2-NEW-1）：Page / PageSize 改为 *int，与 Python 版契约完全对齐：
//   - 未传（nil）→ 使用默认值 1 / 20；
//   - 显式传 0 → 触发 binding:"min=1" 校验 → 422/90004；
//   - 显式传 -1 → 触发 min=1 校验 → 422/90004；
//   - 显式传 >100 → 触发 max=100 校验 → 422/90004；
//   - 传非数字 → strconv.ParseInt 失败 → 422/90004，type=type_error。
//
// v1.1（P1-2）：增加分页参数 binding 约束。
type AuditLogListQuery struct {
	Page      *int   `form:"page" binding:"omitempty,min=1"`
	PageSize  *int   `form:"page_size" binding:"omitempty,min=1,max=100"`
	ModuleID  string `form:"module_id"`
	UserID    *uint  `form:"user_id"`
	Action    string `form:"action"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Keyword   string `form:"keyword"`
}
// AuditLogExportQuery 日志导出查询参数。
type AuditLogExportQuery struct {
	Format    string `form:"format"`
	ModuleID  string `form:"module_id"`
	UserID    *uint  `form:"user_id"`
	Action    string `form:"action"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Keyword   string `form:"keyword"`
}
// formatDateTime 将时间格式化为 Python datetime.isoformat() 兼容的字符串。
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
