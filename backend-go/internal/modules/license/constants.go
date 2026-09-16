package license
// LicenseTypes License 类型白名单。
var LicenseTypes = []string{"trial", "standard", "enterprise"}
// LicenseFileExtensions License 文件允许的扩展名。
var LicenseFileExtensions = []string{".lic", ".json"}
// LicenseFileMaxSize License 文件最大大小（1MB，防止内存爆炸）。
const LicenseFileMaxSize = 1 * 1024 * 1024
// ExpiringSoonDays 过期提醒阈值（天）。
const ExpiringSoonDays = 7
// CoreModulesBypass 受 License 限制的核心模块白名单（始终允许访问）。
//
// 避免平台自身被锁死。与 Python 版 constants.CORE_MODULES_BYPASS 一致。
var CoreModulesBypass = map[string]struct{}{
	"auth":           {},
	"platform":       {},
	"module_manager": {},
	"audit_log":      {},
	"license":        {},
}
// LicenseKeyMinLen license_key 最小长度。
const LicenseKeyMinLen = 1
// LicenseKeyMaxLen license_key 最大长度。
const LicenseKeyMaxLen = 255
// SignatureAlgorithm 签名算法。
//
// 当前仅使用 HMAC-SHA256（对称密钥，部署简单）。
// 本常量作为**未来 RSA 非对称签名升级的预留点**：
//   - 若未来切换为 RSA-SHA256，需同步修改 validator.go 中的
//     ComputeSignature / VerifySignature 实现；
//   - 切换前需保证老版本 License 文件仍可验证（保留 HMAC 兼容路径）。
//
// 当前版本中此常量不参与运行时逻辑，仅作语义标记。
const SignatureAlgorithm = "HMAC-SHA256"
// ActiveLicenseUniqueIndex 部分唯一索引名。
const ActiveLicenseUniqueIndex = "uq_license_license_active"
// RequiredPayloadFields License 文件必需字段。
var RequiredPayloadFields = []string{
	"license_key",
	"license_type",
	"issued_at",
	"expires_at",
}
