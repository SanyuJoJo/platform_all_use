package audit
import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
    "github.com/google/uuid"
    "github.com/yourorg/core-go/internal/envelope"
)
// Entry 审计条目（C-06）。
type Entry struct {
    AuditID          string `json:"audit_id"`
    RequestID        string `json:"request_id"`
    TaskID           string `json:"task_id,omitempty"`
    OperationID      string `json:"operation_id"`
    ActorType        string `json:"actor_type"`
    ActorID          string `json:"actor_id"`
    Algorithm        string `json:"algorithm,omitempty"`
    ParamsDigest     string `json:"params_digest"`
    WhitelistVersion string `json:"whitelist_version,omitempty"`
    Result           string `json:"result"`
    DurationMs       int64  `json:"duration_ms"`
    TS               string `json:"ts"`
    ErrorCode        string `json:"error_code,omitempty"`
}
var auditFilePath string
// Init 初始化审计文件路径。
func Init(coreRoot, relPath string) {
    if filepath.IsAbs(relPath) {
        auditFilePath = relPath
    } else {
        auditFilePath = filepath.Join(coreRoot, filepath.FromSlash(relPath))
    }
    _ = os.MkdirAll(filepath.Dir(auditFilePath), 0750)
}
// Log 写审计（JSON Lines，追加）。
func Log(req *envelope.Request, op string, start time.Time, result, errCode string) {
    if auditFilePath == "" {
        return
    }
    entry := Entry{
        AuditID:      "audit-" + uuid.NewString(),
        OperationID:  op,
        ActorType:    req.Actor.Type,
        ActorID:      req.Actor.ID,
        ParamsDigest: ParamsDigest(req.Params),
        Result:       result,
        DurationMs:   time.Since(start).Milliseconds(),
        TS:           time.Now().UTC().Format(time.RFC3339),
        ErrorCode:    errCode,
    }
    if req.RequestID != "" {
        entry.RequestID = req.RequestID
    }
    if req.TaskID != nil {
        entry.TaskID = *req.TaskID
    }
    if alg, ok := req.Params["algorithm"].(string); ok {
        entry.Algorithm = alg
    }
    data, err := json.Marshal(entry)
    if err != nil {
        return
    }
    f, err := os.OpenFile(auditFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
    if err != nil {
        return
    }
    defer f.Close()
    _, _ = f.Write(append(data, '\n'))
}
