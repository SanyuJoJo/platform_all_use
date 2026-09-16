package middleware
import (
	"time"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)
// AccessLog 结构化访问日志，替代 gin.Logger()，统一由 zerolog 输出。
// 按状态码分级：5xx=Error，4xx=Warn，其他=Info。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		evt := log.Info()
		if status >= 500 {
			evt = log.Error()
		} else if status >= 400 {
			evt = log.Warn()
		}
		rid := ""
		if v, ok := c.Get(RequestIDKey); ok {
			if s, ok := v.(string); ok {
				rid = s
			}
		}
		evt.
			Str("request_id", rid).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Int("size", c.Writer.Size()).
			Msg("http_request")
	}
}
