package crypto

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"backend-go/internal/exception"
)

// coreRootOf 从 dispatch.sh 的绝对路径推导 core 根目录。
//
// 例：/opt/core/sbin/dispatch.sh → /opt/core
func coreRootOf(dispatchPath string) string {
	return filepath.Dir(filepath.Dir(dispatchPath))
}

// newShortID 生成 16 位十六进制短 ID。
func newShortID() string {
	return strings.ReplaceAll(time.Now().Format("20060102150405.000000"), ".", "")[:16]
}

// timeNowUTC 返回当前 UTC 时间。
func timeNowUTC() time.Time {
	return time.Now().UTC()
}

// extractCNFromSubject 从 "CN=xxx,O=yyy" 中提取 CN 值。
func extractCNFromSubject(subject string) string {
	s := strings.TrimSpace(subject)
	if s == "" {
		return "imported-ca"
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToUpper(part), "CN=") {
			return part[3:]
		}
	}
	return s
}

// exceptionCodeFromHTTP 把 HTTP 状态码映射为业务异常码。
func exceptionCodeFromHTTP(status int) int {
	switch status {
	case http.StatusBadRequest, http.StatusForbidden:
		return exception.CodeParamInvalid
	case http.StatusNotFound:
		return exception.CodeNotFound
	case http.StatusUnprocessableEntity:
		return exception.CodeValidationFail
	default:
		return exception.CodeInternalError
	}
}
