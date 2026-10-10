package crypto

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"backend-go/internal/exception"
)

// coreRootOf 从 dispatch.sh 的绝对路径推导 core 根目录。
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

// cnFromSubject 从 "CN=xxx,O=yyy" 中提取 CN 值。
//
// ★ 无 fallback：subject 为空或无 CN 时返回空字符串。
//   用于**普通证书**（CN 必须来自用户输入，不允许伪造）。
//
// 输入格式（RFC2253 或 OpenSSL 原生输出）：
//
//	"CN=test-server"                       → "test-server"
//	"CN=test-server,O=Example,C=CN"        → "test-server"
//	"O=Example,C=CN"                       → ""
//	""                                     → ""
func cnFromSubject(subject string) string {
	s := strings.TrimSpace(subject)
	if s == "" {
		return ""
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToUpper(part), "CN=") {
			return strings.TrimSpace(part[3:])
		}
	}
	return ""
}

// extractCNFromSubject 从 subject 提取 CN 值，**空时返回 "imported-ca"**。
//
// ★ 带 fallback：仅供 **CA 导入** 场景使用。
//   普通证书请用 cnFromSubject，避免把 imported-ca 写进 cert.subject_cn。
//
// 历史遗留：早期 CA 导入逻辑用它兜底无 CN 的第三方 CA。
func extractCNFromSubject(subject string) string {
	cn := cnFromSubject(subject)
	if cn == "" {
		return "imported-ca"
	}
	return cn
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
