package exception
// 统一错误码常量。
//
// P0 阶段仅包含框架底座所需；
// P1 认证模块补充 10003/10004/10005/10007。
const (
	// ---- 通用成功 ----
	CodeSuccess = 0
	// ---- 通用错误 ----
	CodeParamInvalid   = 90001 // 请求参数错误
	CodeNotFound       = 90002 // 资源不存在
	CodeDataConflict   = 90003 // 数据冲突
	CodeValidationFail = 90004 // 请求体验证失败
	CodeInternalError  = 99999 // 服务器内部错误
	// ---- 认证相关（P1） ----
	CodeAuthUnauthorized   = 10001 // 未认证 / Token 无效 / 用户名或密码错误
	CodeAuthForbidden      = 10002 // 用户被禁用
	CodeAuthOldPasswordErr = 10003 // 原密码错误
	CodeAuthPasswordDiff   = 10004 // 两次密码不一致
	CodeAuthUserNotFound   = 10005 // 用户不存在
	CodeAuthPasswordFormat = 10007 // 密码格式无效
	// ---- 权限相关（P1） ----
	CodePermissionDenied = 20051 // 无权限
)
