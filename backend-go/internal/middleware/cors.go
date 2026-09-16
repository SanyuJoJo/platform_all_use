package middleware
import (
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)
// SetupCORS 启用跨域中间件。
//   - origins 包含 "*" 时，自动关闭 AllowCredentials（浏览器不允许）。
//   - AllowHeaders：请求头白名单，仅列入客户端会主动发送的头。
//   - ExposeHeaders：响应头白名单，用于让浏览器 JS 能读取自定义响应头。
//
// V11-P1-03 修复：AllowHeaders 不再列入 X-Request-ID
// （X-Request-ID 是请求头但通常由服务端生成，语义不应出现在 AllowHeaders 中）。
func SetupCORS(origins []string) gin.HandlerFunc {
	allowAll := false
	for _, o := range origins {
		if o == "*" {
			allowAll = true
			break
		}
	}
	cfg := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"X-Request-ID", "X-Error-Code", "Content-Disposition"},
		AllowCredentials: !allowAll,
		MaxAge:           12 * time.Hour,
	}
	if allowAll {
		cfg.AllowOrigins = []string{"*"}
		cfg.AllowCredentials = false
	}
	return cors.New(cfg)
}
