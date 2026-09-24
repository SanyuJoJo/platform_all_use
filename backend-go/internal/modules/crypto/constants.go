package crypto
// OperationID 密码操作标识。
type OperationID string
const (
	// P0 操作
	OpCACreate             OperationID = "ca.create"
	OpCAIntermediateCreate OperationID = "ca.intermediate.create"
	OpCSRCreate            OperationID = "csr.create"
	OpCertSign             OperationID = "cert.sign"
	OpDualCertCreate       OperationID = "dual_cert.create"
	OpCRLCreate            OperationID = "crl.create"
	// P1 操作
	OpCertConvert    OperationID = "cert.convert"
	OpCertParse      OperationID = "cert.parse"
	OpKeyManage      OperationID = "key.manage"
	OpPQCCertCreate  OperationID = "pqc.cert.create"
	// P2 操作
	OpChainVerify     OperationID = "chain.verify"
	OpBatchExecute    OperationID = "batch.execute"
	OpSSLConfigGenerate OperationID = "ssl.config.generate"
	OpCryptoService   OperationID = "crypto.service"
)
// AllOperations 所有已注册的 operation_id。
var AllOperations = []OperationID{
	OpCACreate, OpCAIntermediateCreate, OpCSRCreate, OpCertSign,
	OpDualCertCreate, OpCRLCreate,
	OpCertConvert, OpCertParse, OpKeyManage, OpPQCCertCreate,
	OpChainVerify, OpBatchExecute, OpSSLConfigGenerate, OpCryptoService,
}
// IsValidOperation 判断 operation_id 是否已注册。
func IsValidOperation(op string) bool {
	for _, o := range AllOperations {
		if string(o) == op {
			return true
		}
	}
	return false
}
// 默认超时（毫秒）
const DefaultTimeoutMs = 5000
// 默认最大并发
const DefaultMaxConcurrency = 8
// 任务状态
const (
	TaskStatusPending   = "PENDING"
	TaskStatusRunning   = "RUNNING"
	TaskStatusSuccess   = "SUCCESS"
	TaskStatusFailed    = "FAILED"
	TaskStatusTimeout   = "TIMEOUT"
	TaskStatusCancelled = "CANCELLED"
)
