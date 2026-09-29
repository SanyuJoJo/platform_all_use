package handler

import (
    "context"
    "time"

    "github.com/yourorg/core-go/internal/audit"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
)

// BatchExecuteHandler batch.execute 批量操作。
//
// 说明：为避免 dispatch 包与 handler 包循环依赖，
// 本 handler 只做参数校验，实际子任务由平台后端循环调用 dispatch。
type BatchExecuteHandler struct{}

func (h *BatchExecuteHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    items, ok := req.Params["items"].([]interface{})
    if !ok || len(items) == 0 {
        return genErr("INVALID_PARAM", "items is required")
    }

    // 校验每个子任务的 operation_id
    for _, it := range items {
        m, ok := it.(map[string]interface{})
        if !ok {
            return genErr("INVALID_PARAM", "items must be objects")
        }
        op, _ := m["operation_id"].(string)
        if op == "" {
            return genErr("INVALID_PARAM", "sub-item missing operation_id")
        }
        if op == "batch.execute" {
            return genErr("INVALID_PARAM", "recursive batch.execute not allowed")
        }
    }

    audit.Log(req, "batch.execute", start, "SUCCESS", "")
    return envelope.NewSuccess(map[string]interface{}{
        "total":   len(items),
        "message": "batch.execute 由平台后端拆分执行，core 不递归",
    })
}
