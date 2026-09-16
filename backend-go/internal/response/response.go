package response
import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)
type HealthData struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
type SuccessResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
}
type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
}
// formatPythonISO 复刻 Python datetime.isoformat() 的输出：
//   - 微秒为 0：2006-01-02T15:04:05+00:00
//   - 微秒非 0：2006-01-02T15:04:05.123456+00:00
func formatPythonISO(t time.Time) string {
	t = t.UTC()
	if t.Nanosecond()/1000 == 0 {
		return t.Format("2006-01-02T15:04:05+00:00")
	}
	return t.Format("2006-01-02T15:04:05.000000+00:00")
}
func RequestID(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if h := c.GetHeader("X-Request-ID"); h != "" {
		return h
	}
	return "unknown"
}
// Success 保持 200 便捷方法。
func Success(c *gin.Context, data any, message string) {
	SuccessWithStatus(c, http.StatusOK, data, message)
}
// SuccessWithStatus 允许调用方指定 HTTP 状态码（如创建接口 201）。
func SuccessWithStatus(c *gin.Context, status int, data any, message string) {
	if message == "" {
		message = "success"
	}
	c.PureJSON(status, SuccessResponse{
		Code:      0,
		Message:   message,
		Data:      data,
		Timestamp: formatPythonISO(time.Now()),
		RequestID: RequestID(c),
	})
}
func Error(c *gin.Context, httpStatus int, code int, message string, data any) {
	c.PureJSON(httpStatus, ErrorResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: formatPythonISO(time.Now()),
		RequestID: RequestID(c),
	})
}
