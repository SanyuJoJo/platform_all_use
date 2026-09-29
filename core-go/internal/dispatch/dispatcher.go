package dispatch
import (
    "context"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
)
// Dispatch 路由到 handler。
func Dispatch(
    ctx context.Context, op string, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    h, ok := registry[op]
    if !ok {
        return envelope.NewError("INVALID_PARAM", "unknown operation_id: "+op)
    }
    return h.Execute(ctx, req, cfg)
}
