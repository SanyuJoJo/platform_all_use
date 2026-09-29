package envelope
// Request core 请求（C-04）。
type Request struct {
    SchemaVersion string                 `json:"schema_version"`
    OperationID   string                 `json:"operation_id"`
    RequestID     string                 `json:"request_id"`
    TaskID        *string                `json:"task_id,omitempty"`
    Actor         Actor                  `json:"actor"`
    Params        map[string]interface{} `json:"params"`
    Options       *Options               `json:"options,omitempty"`
}
type Actor struct {
    Type string `json:"type"`
    ID   string `json:"id"`
}
type Options struct {
    TimeoutMs *int  `json:"timeout_ms,omitempty"`
    DryRun    *bool `json:"dry_run,omitempty"`
}
