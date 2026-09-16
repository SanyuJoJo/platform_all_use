package exception
// 统一错误码常量。
//
// P1 阶段补充用户/角色/权限管理相关错误码。
const (
	// ---- 通用成功 ----
	CodeSuccess = 0
	// ---- 通用错误 ----
	CodeParamInvalid   = 90001 // 请求参数错误
	CodeNotFound       = 90002 // 资源不存在
	CodeDataConflict   = 90003 // 数据冲突
	CodeValidationFail = 90004 // 请求体验证失败
	CodeInternalError  = 99999 // 服务器内部错误
	// ---- 认证相关 ----
	CodeAuthUnauthorized   = 10001 // 未认证 / Token 无效 / 用户名或密码错误
	CodeAuthForbidden      = 10002 // 用户被禁用
	CodeAuthOldPasswordErr = 10003 // 原密码错误
	CodeAuthPasswordDiff   = 10004 // 两次密码不一致
	CodeAuthUserNotFound   = 10005 // 用户不存在
	CodeAuthPasswordFormat = 10007 // 密码格式无效
	// ---- 用户管理相关 ----
	CodeAuthUsernameExists    = 10000 // 用户名已存在
	CodeAuthUsernameFormat    = 10006 // 用户名格式无效
	CodeAuthEmailExists       = 10009 // 邮箱已被使用
	CodeAuthCannotDeleteSelf  = 10010 // 不能删除自己
	CodeAuthCannotDisableSelf = 10011 // 不能禁用自己
	// ---- 角色管理相关 ----
	CodeAuthRoleCodeExists      = 20001 // 角色编码已存在
	CodeAuthRoleNotFound        = 20002 // 角色不存在
	CodeAuthSystemRoleProtected = 20003 // 不能删除系统内置角色
	CodeAuthRoleNameExists      = 20004 // 角色名称已存在
	CodeAuthRoleInUse           = 20005 // 角色已被用户使用
	// ---- 权限管理相关 ----
	CodeAuthPermissionNotFound       = 20052 // 权限不存在
	CodeAuthPermissionModuleConflict = 20053 // 权限编码已被其他模块占用
	CodeAuthPermissionCodeFormat     = 20054 // 权限编码格式无效
	// ---- 权限相关 ----
	CodePermissionDenied = 20051 // 无权限
)
