package openssl
import (
    "context"
    "time"
)
func contextWithTimeout() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 15*time.Second)
}
