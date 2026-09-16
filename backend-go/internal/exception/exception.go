package exception
import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"backend-go/internal/response"
)
type PlatformError struct {
	Code       int
	Message    string
	HTTPStatus int
	Data       any
}
func (e *PlatformError) Error() string {
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}
func New(code int, message string, status int, data any) *PlatformError {
	if status == 0 {
		status = http.StatusBadRequest
	}
	return &PlatformError{Code: code, Message: message, HTTPStatus: status, Data: data}
}
func Abort(c *gin.Context, err *PlatformError) {
	response.Error(c, err.HTTPStatus, err.Code, err.Message, err.Data)
	c.Abort()
}
// Recovery 捕获 panic，返回 code=99999 的统一错误响应。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Interface("panic", r).
					Str("request_id", response.RequestID(c)).
					Msg("panic recovered")
				response.Error(
					c,
					http.StatusInternalServerError,
					CodeInternalError,
					"服务器内部错误",
					nil,
				)
				c.Abort()
			}
		}()
		c.Next()
	}
}
// NoRoute 未注册路径：404 + 90002。
func NoRoute(c *gin.Context) {
	response.Error(c, http.StatusNotFound, CodeNotFound, "资源不存在", nil)
}
// NoMethod 方法不允许。
//
// 与 Python 版 StarletteHTTPException 处理器分支保持一致：
//   - code = 90001
//   - message = "请求错误"（Starlette 分支的兜底文案）
func NoMethod(c *gin.Context) {
	response.Error(c, http.StatusMethodNotAllowed, CodeParamInvalid, "请求错误", nil)
}
