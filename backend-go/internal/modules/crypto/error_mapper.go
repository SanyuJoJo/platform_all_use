package crypto
// 平台错误码常量。
const (
	PlatformSuccess                    = "SUCCESS"
	PlatformCryptoInvalidParam         = "CRYPTO_INVALID_PARAM"
	PlatformCryptoAlgorithmNotAllowed  = "CRYPTO_ALGORITHM_NOT_ALLOWED"
	PlatformCryptoPathNotAllowed       = "CRYPTO_PATH_NOT_ALLOWED"
	PlatformAuthForbidden              = "AUTH_FORBIDDEN"
	PlatformCryptoKeyNotFound          = "CRYPTO_KEY_NOT_FOUND"
	PlatformCryptoCertNotFound         = "CRYPTO_CERT_NOT_FOUND"
	PlatformCryptoCertParseFailed      = "CRYPTO_CERT_PARSE_FAILED"
	PlatformCryptoCoreFailed           = "CRYPTO_CORE_FAILED"
	PlatformCryptoCoreTimeout          = "CRYPTO_CORE_TIMEOUT"
	PlatformCryptoCoreResponseInvalid  = "CRYPTO_CORE_RESPONSE_INVALID"
	PlatformCryptoVersionUnsupported   = "CRYPTO_VERSION_UNSUPPORTED"
	PlatformTaskNotFound               = "TASK_NOT_FOUND"
	PlatformTaskStateInvalid           = "TASK_STATE_INVALID"
	PlatformTaskCancelled              = "TASK_CANCELLED"
	PlatformTaskTimeout                = "TASK_TIMEOUT"
	PlatformInternalError              = "INTERNAL_ERROR"
)
// CoreErrorMapping core 错误码到平台错误码的映射。
type CoreErrorMapping struct {
	PlatformCode string
	HTTPStatus   int
	Retryable    bool
	FrontendHint string
}
// CoreErrorMap core 错误码 → 平台映射表（C-05）。
var CoreErrorMap = map[string]CoreErrorMapping{
	"OK": {
		PlatformCode: PlatformSuccess,
		HTTPStatus:   200,
		Retryable:    false,
		FrontendHint: "成功",
	},
	"INVALID_PARAM": {
		PlatformCode: PlatformCryptoInvalidParam,
		HTTPStatus:   400,
		Retryable:    false,
		FrontendHint: "参数不合法",
	},
	"ALGORITHM_NOT_ALLOWED": {
		PlatformCode: PlatformCryptoAlgorithmNotAllowed,
		HTTPStatus:   400,
		Retryable:    false,
		FrontendHint: "算法不在白名单",
	},
	"PATH_NOT_ALLOWED": {
		PlatformCode: PlatformCryptoPathNotAllowed,
		HTTPStatus:   400,
		Retryable:    false,
		FrontendHint: "路径不允许",
	},
	"PERMISSION_DENIED": {
		PlatformCode: PlatformAuthForbidden,
		HTTPStatus:   403,
		Retryable:    false,
		FrontendHint: "无权限",
	},
	"KEY_NOT_FOUND": {
		PlatformCode: PlatformCryptoKeyNotFound,
		HTTPStatus:   404,
		Retryable:    false,
		FrontendHint: "密钥不存在",
	},
	"CERT_NOT_FOUND": {
		PlatformCode: PlatformCryptoCertNotFound,
		HTTPStatus:   404,
		Retryable:    false,
		FrontendHint: "证书不存在",
	},
	"CERT_PARSE_FAILED": {
		PlatformCode: PlatformCryptoCertParseFailed,
		HTTPStatus:   422,
		Retryable:    false,
		FrontendHint: "证书解析失败",
	},
	"CORE_EXEC_FAILED": {
		PlatformCode: PlatformCryptoCoreFailed,
		HTTPStatus:   500,
		Retryable:    false,
		FrontendHint: "密码服务执行失败",
	},
	"CORE_TIMEOUT": {
		PlatformCode: PlatformCryptoCoreTimeout,
		HTTPStatus:   504,
		Retryable:    false,
		FrontendHint: "密码服务超时",
	},
	"CORE_JSON_INVALID": {
		PlatformCode: PlatformCryptoCoreResponseInvalid,
		HTTPStatus:   500,
		Retryable:    false,
		FrontendHint: "密码服务响应异常",
	},
	"VERSION_UNSUPPORTED": {
		PlatformCode: PlatformCryptoVersionUnsupported,
		HTTPStatus:   500,
		Retryable:    false,
		FrontendHint: "铜锁版本不满足",
	},
	"INTERNAL_ERROR": {
		PlatformCode: PlatformInternalError,
		HTTPStatus:   500,
		Retryable:    false,
		FrontendHint: "系统内部错误",
	},
}
// MapCoreError 将 core 错误码映射为平台错误码与 HTTP 状态码。
// 未知 core code 兜底为 INTERNAL_ERROR，并记录原始 code。
func MapCoreError(coreCode string) CoreErrorMapping {
	if m, ok := CoreErrorMap[coreCode]; ok {
		return m
	}
	// 未知 core code 兜底
	return CoreErrorMapping{
		PlatformCode: PlatformInternalError,
		HTTPStatus:   500,
		Retryable:    false,
		FrontendHint: "系统内部错误",
	}
}
