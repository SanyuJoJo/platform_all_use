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
	TaskID       string                 `json:"task_id"`
	OperationID  string                 `json:"operation_id"`
	RequestID    string                 `json:"request_id"`
	Status       string                 `json:"status"`
	Progress     *int                   `json:"progress,omitempty"`
	ResultRef    *string                `json:"result_ref,omitempty"`
	ErrorCode    *string                `json:"error_code,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	StartedAt    *string                `json:"started_at,omitempty"`
	FinishedAt   *string                `json:"finished_at,omitempty"`
	TimeoutMs    *int                   `json:"timeout_ms,omitempty"`
}
