package audit_log
// 中间件跳过的路径前缀（健康检查、文档、静态资源等）。
var SKIP_PATH_PREFIXES = []string{
	"/health",
	"/docs",
	"/redoc",
	"/openapi.json",
	"/favicon.ico",
	"/static",
}
// 路径中的模块段 → 平台 module_id 映射
var PATH_MODULE_MAP = map[string]string{
	"modules":    "module_manager",
	"audit-logs": "audit_log",
}
// HTTP 方法 → 操作类型映射
var METHOD_ACTION_MAP = map[string]string{
	"GET":    "view",
	"POST":   "create",
	"PUT":    "update",
	"PATCH":  "update",
	"DELETE": "delete",
}
// HTTP 状态码 → 粗略错误码映射（无法从 X-Error-Code 获取精确错误码时使用）
var STATUS_ERROR_MAP = map[int]int{
	400: 90001,
	401: 10001,
	403: 20051,
	404: 90002,
	409: 90003,
	422: 90004,
	429: 90008,
	500: 90000,
}
const (
	// ExportFormatCSV CSV 导出格式
	ExportFormatCSV = "csv"
	// ExportFormatJSON JSON 导出格式
	ExportFormatJSON = "json"
	// ExportMaxRows 单次导出最大行数
	ExportMaxRows = 10000
	// CSV_BOM UTF-8 BOM，兼容 Excel 中文显示
	CSV_BOM = "\ufeff"
)
