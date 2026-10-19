// backend-go/internal/modules/ssl_proxy/errors.go
package ssl_proxy

import "fmt"

// ErrCode 统一错误码
type ErrCode int

const (
	ErrOK              ErrCode = 0
	ErrInvalidParam    ErrCode = 40001
	ErrNotFound        ErrCode = 40401
	ErrPortConflict    ErrCode = 40901
	ErrRunningConflict ErrCode = 40902
	ErrNameConflict    ErrCode = 40903 // ★ 新增：服务名称冲突
	ErrConfigGenerate  ErrCode = 50001
	ErrCertResolve     ErrCode = 50002
	ErrStartFailed     ErrCode = 50003
	ErrStopFailed      ErrCode = 50004
	ErrTengineMissing  ErrCode = 50005
	ErrProtocolUnsup   ErrCode = 50006
	ErrPermission      ErrCode = 50007
	ErrBatchPartial    ErrCode = 50008
)

// APIError 业务错误封装
type APIError struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
	Detail  string  `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("[%d] %s", e.Code, e.Message)
	}
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
}

func newErr(code ErrCode, msg, detail string) *APIError {
	return &APIError{Code: code, Message: msg, Detail: detail}
}
