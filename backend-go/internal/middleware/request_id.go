package middleware
import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
const RequestIDKey = "request_id"
// RequestID 生成或透传 X-Request-ID，写入 gin.Context 与响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(RequestIDKey, rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
