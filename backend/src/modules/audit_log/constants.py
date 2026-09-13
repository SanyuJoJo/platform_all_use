"""日志审计模块 - 常量定义。"""
# 中间件跳过的路径前缀（健康检查、文档、静态资源等）
SKIP_PATH_PREFIXES = (
    "/health",
    "/docs",
    "/redoc",
    "/openapi.json",
    "/favicon.ico",
    "/static",
)
# 路径中的模块段 → 平台 module_id 映射
PATH_MODULE_MAP = {
    "modules": "module_manager",
    "audit-logs": "audit_log",
}
# HTTP 方法 → 操作类型映射
METHOD_ACTION_MAP = {
    "GET": "view",
    "POST": "create",
    "PUT": "update",
    "PATCH": "update",
    "DELETE": "delete",
}
# HTTP 状态码 → 粗略错误码映射（无法从 X-Error-Code 获取精确错误码时使用）
STATUS_ERROR_MAP = {
    400: 90001,
    401: 10001,
    403: 20051,
    404: 90002,
    409: 90003,
    422: 90004,
    429: 90008,
    500: 90000,
}
# 导出格式白名单
EXPORT_FORMATS = ("csv", "json")
# 单次导出最大行数（防止内存爆炸）
EXPORT_MAX_ROWS = 10000
# CSV 导出 BOM（兼容 Excel 中文显示）
CSV_BOM = "\ufeff"
