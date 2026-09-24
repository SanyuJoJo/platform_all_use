package crypto

// OperationRequest 密码操作请求（前端传入）。
type OperationRequest struct {
	Params  map[string]interface{} `json:"params" binding:"required"`
	Options *OperationOptions      `json:"options,omitempty"`
}

// OperationOptions 运行选项。
type OperationOptions struct {
	TimeoutMs *int  `json:"timeout_ms,omitempty"`
	DryRun    *bool `json:"dry_run,omitempty"`
}

// CoreRequest core/sbin 请求 JSON 结构（C-04）。
type CoreRequest struct {
	SchemaVersion string                 `json:"schema_version"`
	OperationID   string                 `json:"operation_id"`
	RequestID     string                 `json:"request_id"`
	TaskID        *string                `json:"task_id,omitempty"`
	Actor         CoreActor              `json:"actor"`
	Params        map[string]interface{} `json:"params"`
	Options       *CoreOptions           `json:"options,omitempty"`
}

// CoreActor 执行者标识。
type CoreActor struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// CoreOptions core 运行选项。
type CoreOptions struct {
	TimeoutMs *int  `json:"timeout_ms,omitempty"`
	DryRun    *bool `json:"dry_run,omitempty"`
}

// CoreResponse core/sbin 响应 JSON 结构（C-04）。
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

// CoreError core 错误详情。
type CoreError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Detail    map[string]interface{} `json:"detail,omitempty"`
	Retryable bool                   `json:"retryable"`
}

// CoreAudit core 审计摘要。
type CoreAudit struct {
	AuditID      string `json:"audit_id"`
	ParamsDigest string `json:"params_digest"`
	Result       string `json:"result"`
	DurationMs   int    `json:"duration_ms"`
}

// OperationResponse 平台响应（同步）。
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

// PlatformErrorDetail 平台错误详情。
type PlatformErrorDetail struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Detail    map[string]interface{} `json:"detail,omitempty"`
	Retryable bool                   `json:"retryable"`
}

// PlatformAudit 平台审计摘要。
type PlatformAudit struct {
	AuditID      string `json:"audit_id"`
	ParamsDigest string `json:"params_digest"`
	Result       string `json:"result"`
	DurationMs   int    `json:"duration_ms"`
}

// TaskResponse 任务状态响应。
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

// -----------------------------------------------------------------------------
// 证书相关 DTO
// -----------------------------------------------------------------------------

// SignCertRequest 申请/签发证书请求（无 CSR）。
//
// CA 来源：
//   - local：使用平台已有本地 CA，仅需 ca_id；后端查表自动补全 cert_path + key_ref
//   - manual：手动上传 CA 证书 PEM + 私钥 PEM + 可选私钥密码
//
// 证书类型支持多选：cert_types 允许组合，例如 ["server","client"]。
type SignCertRequest struct {
	// ---- CA 来源 ----
	CASource      string `json:"ca_source" binding:"required"` // local | manual
	CAID          string `json:"ca_id,omitempty"`              // local：只需 ca_id
	CACertPEM     string `json:"ca_cert_pem,omitempty"`        // manual
	CAKeyPEM      string `json:"ca_key_pem,omitempty"`         // manual
	CAKeyPassword string `json:"ca_key_password,omitempty"`    // manual，可选

	// ---- 证书参数 ----
	Algorithm    string                 `json:"algorithm,omitempty"`
	KeyParams    map[string]interface{} `json:"key_params,omitempty"`
	Subject      map[string]string      `json:"subject"`
	SAN          []string               `json:"san,omitempty"`
	CertTypes    []string               `json:"cert_types" binding:"required"`
	ValidityDays int                    `json:"validity_days"`
}

// ImportCertRequest 导入证书请求。
type ImportCertRequest struct {
	CertPEM     string `json:"cert_pem" binding:"required"`
	KeyPEM      string `json:"key_pem,omitempty"`
	KeyPassword string `json:"key_password,omitempty"`
	KeyRef      string `json:"key_ref,omitempty"`
}

// ExportCertRequest 导出证书请求。
type ExportCertRequest struct {
	Type     string `json:"type" binding:"required"` // cert | key | pkcs12
	Password string `json:"password,omitempty"`
}
