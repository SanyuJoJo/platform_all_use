package response
import (
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
)
// HealthData 健康检查数据。
type HealthData struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
// SuccessResponse 成功响应。
type SuccessResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
}
// ErrorResponse 错误响应。
type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
}
// formatPythonISO 复刻 Python datetime.isoformat() 的输出。
func formatPythonISO(t time.Time) string {
	t = t.UTC()
	if t.Nanosecond()/1000 == 0 {
		return t.Format("2006-01-02T15:04:05+00:00")
	}
	return t.Format("2006-01-02T15:04:05.000000+00:00")
}
// RequestID 从 gin.Context 获取 request_id。
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
// SuccessWithStatus 允许调用方指定 HTTP 状态码。
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
// Error 统一错误响应，并设置 X-Error-Code 响应头。
func Error(c *gin.Context, httpStatus int, code int, message string, data any) {
	c.Header("X-Error-Code", strconv.Itoa(code))
	c.PureJSON(httpStatus, ErrorResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: formatPythonISO(time.Now()),
		RequestID: RequestID(c),
	})
}
