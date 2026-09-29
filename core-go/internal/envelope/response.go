package envelope
// Response core 响应（C-04）。
type Response struct {
    SchemaVersion string                 `json:"schema_version"`
    Code          string                 `json:"code"`
    Message       string                 `json:"message"`
    RequestID     string                 `json:"request_id,omitempty"`
    OperationID   string                 `json:"operation_id,omitempty"`
    TaskID        *string                `json:"task_id,omitempty"`
    Data          map[string]interface{} `json:"data,omitempty"`
    Error         *ErrorDetail           `json:"error,omitempty"`
    Audit         *AuditSummary          `json:"audit,omitempty"`
}
type ErrorDetail struct {
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Detail    map[string]interface{} `json:"detail,omitempty"`
    Retryable bool                   `json:"retryable"`
}
type AuditSummary struct {
    AuditID      string `json:"audit_id"`
    ParamsDigest string `json:"params_digest"`
    Result       string `json:"result"`
    DurationMs   int    `json:"duration_ms"`
}
// NewSuccess 成功响应。
func NewSuccess(data map[string]interface{}) *Response {
    return &Response{
        SchemaVersion: "1.0",
        Code:          "OK",
        Message:       "success",
        Data:          data,
    }
}
// NewError 错误响应。
func NewError(code, msg string) *Response {
    return &Response{
        SchemaVersion: "1.0",
        Code:          code,
        Message:       msg,
        Error: &ErrorDetail{
            Code:      code,
            Message:   msg,
            Retryable: false,
        },
    }
}
// NewErrorWithDetail 带 detail 的错误响应。
func NewErrorWithDetail(code, msg string, detail map[string]interface{}) *Response {
    r := NewError(code, msg)
    r.Error.Detail = detail
    return r
}
