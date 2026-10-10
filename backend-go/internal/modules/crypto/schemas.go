package crypto

import "backend-go/internal/models"

// =============================================================================
// 通用密码操作请求 / 响应（C-04 契约）
// =============================================================================

type OperationRequest struct {
	Params  map[string]interface{} `json:"params" binding:"required"`
	Options *OperationOptions      `json:"options,omitempty"`
}

type OperationOptions struct {
	TimeoutMs *int  `json:"timeout_ms,omitempty"`
	DryRun    *bool `json:"dry_run,omitempty"`
}

type CoreRequest struct {
	SchemaVersion string                 `json:"schema_version"`
	OperationID   string                 `json:"operation_id"`
	RequestID     string                 `json:"request_id"`
	TaskID        *string                `json:"task_id,omitempty"`
	Actor         CoreActor              `json:"actor"`
	Params        map[string]interface{} `json:"params"`
	Options       *CoreOptions           `json:"options,omitempty"`
}

type CoreActor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CoreOptions struct {
	TimeoutMs *int  `json:"timeout_ms,omitempty"`
	DryRun    *bool `json:"dry_run,omitempty"`
}

type CoreResponse struct {
	SchemaVersion string                 `json:"schema_version"`
	Code          string                 `json:"code"`
	Message       string                 `json:"message"`
	RequestID     string                 `json:"request_id"`
	OperationID   string                 `json:"operation_id"`
	TaskID        *string                `json:"task_id,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Error         *CoreError             `json:"error,omitempty"`
	Audit         *CoreAudit             `json:"audit,omitempty"`
}

type CoreError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Detail    map[string]interface{} `json:"detail,omitempty"`
	Retryable bool                   `json:"retryable"`
}

type CoreAudit struct {
	AuditID      string `json:"audit_id"`
	ParamsDigest string `json:"params_digest"`
	Result       string `json:"result"`
	DurationMs   int    `json:"duration_ms"`
}

type OperationResponse struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id"`
	OperationID string                 `json:"operation_id"`
	TaskID      *string                `json:"task_id,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       *PlatformErrorDetail   `json:"error,omitempty"`
	Audit       *PlatformAudit         `json:"audit,omitempty"`
}

type PlatformErrorDetail struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Detail    map[string]interface{} `json:"detail,omitempty"`
	Retryable bool                   `json:"retryable"`
}

type PlatformAudit struct {
	AuditID      string `json:"audit_id"`
	ParamsDigest string `json:"params_digest"`
	Result       string `json:"result"`
	DurationMs   int    `json:"duration_ms"`
}

type TaskResponse struct {
	TaskID       string  `json:"task_id"`
	OperationID  string  `json:"operation_id"`
	RequestID    string  `json:"request_id"`
	Status       string  `json:"status"`
	Progress     *int    `json:"progress,omitempty"`
	ResultRef    *string `json:"result_ref,omitempty"`
	ErrorCode    *string `json:"error_code,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	CreatedAt    string  `json:"created_at"`
	StartedAt    *string `json:"started_at,omitempty"`
	FinishedAt   *string `json:"finished_at,omitempty"`
	TimeoutMs    *int    `json:"timeout_ms,omitempty"`
}

// =============================================================================
// 证书相关 DTO
// =============================================================================

// SignCertRequest 签发证书请求。
//
// 覆盖两类场景：
//   - normal：普通单证书签发；
//   - dual  ：国密双证签发（签名证书 + 加密证书）。
//
// CA 来源二选一：ca_source = "local"（平台内 CA）或 "manual"（上传 CA 证书+私钥）。
// CSR 来源三选一：csr_source = ""（不使用）、"existing"（引用已导入 P10）、"upload"（随请求上传）。
type SignCertRequest struct {
	// ---------- CA 来源 ----------
	CASource      string `json:"ca_source" binding:"required"`
	CAID          string `json:"ca_id,omitempty"`
	CACertPEM     string `json:"ca_cert_pem,omitempty"`
	CAKeyPEM      string `json:"ca_key_pem,omitempty"`
	CAKeyPassword string `json:"ca_key_password,omitempty"`

	// ---------- CSR 来源 ----------
	CSRSource      string `json:"csr_source,omitempty"`
	CSRID          string `json:"csr_id,omitempty"`
	CSRPEM         string `json:"csr_pem,omitempty"`
	CSRKeyPEM      string `json:"csr_key_pem,omitempty"`
	CSRKeyPassword string `json:"csr_key_password,omitempty"`

	// ---------- 证书参数 ----------
	Algorithm    string                 `json:"algorithm,omitempty"`
	KeyParams    map[string]interface{} `json:"key_params,omitempty"`
	Subject      map[string]string      `json:"subject,omitempty"`
	SAN          []string               `json:"san,omitempty"`
	CertTypes    []string               `json:"cert_types,omitempty"`
	ValidityDays int                    `json:"validity_days"`
	CertMode     string                 `json:"cert_mode,omitempty"` // normal（默认） | dual

	// ---------- 私钥回传 ----------
	ReturnKey         bool   `json:"return_key,omitempty"`
	KeyExportPassword string `json:"key_export_password,omitempty"`

	// ---------- ★ 证书路径改造新增 ----------
	// Domain 域名目录（可选）。
	//   - 用于 CA 场景 / 目录标识；
	//   - 为空时由 core 或 layout 默认处理。
	Domain string `json:"domain,omitempty"`
}

type ImportCertRequest struct {
	CertPEM     string `json:"cert_pem" binding:"required"`
	KeyPEM      string `json:"key_pem,omitempty"`
	KeyPassword string `json:"key_password,omitempty"`
	KeyRef      string `json:"key_ref,omitempty"`
}

// ImportDualCertRequest 国密双证导入请求。
//
// 落盘规则（证书路径改造）：
//   - 签名证书：server/<csr_dir_no>/<sign_pub_sm3>.cert.pem；
//   - 加密证书：server/<new_dir_no>/<enc_pub_sm3>.cert.pem；
//   - 加密私钥：server/<new_dir_no>/<enc_pub_sm3>.key.pem（白盒密文）
//     与 <enc_pub_sm3>.key.pem.pass（0600）。
type ImportDualCertRequest struct {
	CSRPubSM3      string `json:"csr_pub_sm3"   binding:"required"`
	SignCertPEM    string `json:"sign_cert_pem" binding:"required"`
	EncCertPEM     string `json:"enc_cert_pem"  binding:"required"`
	EncKeyPEM      string `json:"enc_key_pem"   binding:"required"`
	EncKeyPassword string `json:"enc_key_password,omitempty"`
}

type ExportCertRequest struct {
	Type     string `json:"type" binding:"required"`
	Password string `json:"password,omitempty"`
}

type ImportCsrRequest struct {
	CSRPEM      string `json:"csr_pem" binding:"required"`
	KeyPEM      string `json:"key_pem,omitempty"`
	KeyPassword string `json:"key_password,omitempty"`
}

// =============================================================================
// 内部数字信封结构
// =============================================================================

type envelopedKey struct {
	Version             string `json:"version"`
	Algorithm           string `json:"algorithm"`
	SignCertID          string `json:"sign_cert_id,omitempty"`
	EncCertID           string `json:"enc_cert_id,omitempty"`
	SignAlg             string `json:"sign_alg"`
	EncAlg              string `json:"enc_alg"`
	SymmetricKeyCipher  string `json:"symmetric_key_cipher"`
	IV                  string `json:"iv"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

// =============================================================================
// 签发证书返回结构
// =============================================================================

// SignCertResult 签发证书返回给前端的结果。
//
// 普通证书：
//   - cert_pem ：证书 PEM
//   - key_pem  ：私钥 PEM（可选）
//
// 国密双证：
//   - sign_cert_pem      ：签名证书 PEM
//   - enc_cert_pem       ：加密证书 PEM
//   - encrypted_envelope ：加密的数字信封（单一 base64 字符串，用户需保存）
type SignCertResult struct {
	Certificate *models.Certificate `json:"certificate"`

	// 普通证书
	CertPEM string `json:"cert_pem,omitempty"`
	KeyPEM  string `json:"key_pem,omitempty"`

	// 国密双证
	SignCertPEM       string `json:"sign_cert_pem,omitempty"`
	EncCertPEM        string `json:"enc_cert_pem,omitempty"`
	EncryptedEnvelope string `json:"encrypted_envelope,omitempty"`
}

// =============================================================================
// 信封信息查询
// =============================================================================

// QueryEnvelopeRequest 查询信封信息请求。
//
// 两种模式：
//   - mode=cert   ：从签名证书反查私钥。
//                   若签名证书有 KeyRef（客户导入 P10 时上传过私钥），
//                   从 core 导出私钥解密。
//                   若签名证书没有 KeyRef，返回明确错误引导用 manual 模式。
//   - mode=manual ：客户上传自己的 CSR 私钥 s_pri，Server B 临时解密。
//
// ★ 去掉 binding:"required"，允许不同模式下选择性传参；
//   实际校验在 QueryEnvelope 方法内完成，错误提示更友好。
type QueryEnvelopeRequest struct {
	Format            string `json:"format"`             // pkcs10 | cfca
	Mode              string `json:"mode,omitempty"`     // cert | manual，默认 cert
	EncryptedEnvelope string `json:"encrypted_envelope"` // 两种模式都必填

	// mode=cert
	CertID string `json:"cert_id,omitempty"`

	// mode=manual
	SignKeyPEM      string `json:"sign_key_pem,omitempty"`
	SignKeyPassword string `json:"sign_key_password,omitempty"`
}

type QueryEnvelopeResponse struct {
	Format              string `json:"format"`
	Mode                string `json:"mode"`
	SymmetricKeyCipher  string `json:"symmetric_key_cipher"`
	IV                  string `json:"iv"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
	DecryptedKeyPEM     string `json:"decrypted_key_pem,omitempty"`
}
