package exception
// 统一错误码常量。
//
// P5 阶段（License 管理）补充 50001~50010。
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
	CodeAuthUnauthorized   = 10001
	CodeAuthForbidden      = 10002
	CodeAuthOldPasswordErr = 10003
	CodeAuthPasswordDiff   = 10004
	CodeAuthUserNotFound   = 10005
	CodeAuthPasswordFormat = 10007
	// ---- 用户管理相关 ----
	CodeAuthUsernameExists    = 10000
	CodeAuthUsernameFormat    = 10006
	CodeAuthEmailExists       = 10009
	CodeAuthCannotDeleteSelf  = 10010
	CodeAuthCannotDisableSelf = 10011
	// ---- 角色管理相关 ----
	CodeAuthRoleCodeExists      = 20001
	CodeAuthRoleNotFound        = 20002
	CodeAuthSystemRoleProtected = 20003
	CodeAuthRoleNameExists      = 20004
	CodeAuthRoleInUse           = 20005
	// ---- 权限管理相关 ----
	CodeAuthPermissionNotFound       = 20052
	CodeAuthPermissionModuleConflict = 20053
	CodeAuthPermissionCodeFormat     = 20054
	// ---- 权限相关 ----
	CodePermissionDenied = 20051
	// ---- 模块管理相关 ----
	CodeModuleDependencyMissing  = 30001
	CodeModuleExists             = 30002
	CodeModuleNotFound           = 30003
	CodeModuleManifestInvalid    = 30004
	CodeModuleDependencyConflict = 30006
	CodeModuleConfigInvalid      = 30007
	CodeModuleEntryNotFound      = 30008
	CodeModuleLoadFailed         = 30009
	CodeModuleIDInvalid          = 30010
	CodeModuleInUse              = 30011
	CodeModuleZipInvalid         = 30012
	CodeModuleCoreProtected      = 30013
	CodeModuleVersionInvalid     = 30015
	// ---- 日志审计相关 ----
	CodeAuditLogInvalidTimeRange = 40001
	CodeAuditLogNotFound         = 40002
	CodeAuditLogExportFormat     = 40004
	CodeAuditLogExportFailed     = 40005
	// ---- License 管理相关（v1.0 新增）----
	CodeLicenseInvalidFile     = 50001 // License 文件无效
	CodeLicenseExpired         = 50002 // License 已过期
	CodeLicenseSignatureFailed = 50003 // License 签名验证失败
	CodeLicenseAlreadyActive   = 50004 // License 已激活
	CodeLicenseActivationCode  = 50005 // 激活码无效
	CodeLicenseMachineMismatch = 50006 // 机器码不匹配
	CodeLicenseModuleNotAuth   = 50007 // 模块未授权
	CodeLicenseUserQuotaExceed = 50008 // 用户数超限
	CodeLicenseNotFound        = 50009 // License 不存在
	CodeLicenseTypeUnsupported = 50010 // License 类型不支持
)
