package crypto

import (
	"fmt"
	"strings"

	"backend-go/internal/exception"
)

// dnAllowedKeys 允许的 DN 项，与前端 dn.ts 的 DN_ALLOWED_KEYS 对齐。
var dnAllowedKeys = map[string]struct{}{
	"C":            {},
	"CN":           {},
	"O":            {},
	"OU":           {},
	"ST":           {},
	"L":            {},
	"E":            {},
	"SERIALNUMBER": {},
	"SURNAME":      {},
	"GIVENNAME":    {},
}

// ValidateSubject 校验 subject 字段，提前拦截非法输入。
//
// 规则（与前端 dn.ts 对齐）：
//   - key 必须是允许的 DN 项，大小写不敏感
//   - 值不能为空，长度 ≤ 256
//   - C 必须是 2 位国家码（ISO 3166-1 alpha-2）
//
// 用于防止绕过前端直接调 REST API 时把非法 subject 传给 core，
// 触发 OpenSSL ASN1_mbstring_ncopy:string too long 之类的底层错误。
func ValidateSubject(subject map[string]interface{}) error {
	if len(subject) == 0 {
		return exception.New(exception.CodeParamInvalid, "subject 不能为空", 400, nil)
	}
	for rawKey, rawVal := range subject {
		key := strings.ToUpper(strings.TrimSpace(rawKey))
		if _, ok := dnAllowedKeys[key]; !ok {
			return exception.New(
				exception.CodeParamInvalid,
				fmt.Sprintf("不支持的 DN 项：%s", rawKey), 400, nil,
			)
		}
		val, ok := rawVal.(string)
		if !ok {
			return exception.New(
				exception.CodeParamInvalid,
				fmt.Sprintf("DN 项 %s 的值类型必须是字符串", key), 400, nil,
			)
		}
		val = strings.TrimSpace(val)
		if val == "" {
			return exception.New(
				exception.CodeParamInvalid,
				fmt.Sprintf("DN 项 %s 的值不能为空", key), 400, nil,
			)
		}
		if key == "C" && len(val) != 2 {
			return exception.New(
				exception.CodeParamInvalid,
				fmt.Sprintf("DN 项 C 必须是 2 位国家码，当前为「%s」", val), 400, nil,
			)
		}
		if len(val) > 256 {
			return exception.New(
				exception.CodeParamInvalid,
				fmt.Sprintf("DN 项 %s 的值过长（>256）", key), 400, nil,
			)
		}
	}
	return nil
}

// subjectOperations 需要校验 subject 的 operation_id 白名单。
//
// 这些操作直接接收用户传入的 subject；其他操作（如 key.manage）不涉及。
var subjectOperations = map[string]struct{}{
	"ca.create":               {},
	"ca.intermediate.create":  {},
	"csr.create":              {},
	"cert.sign":               {},
	"dual_cert.create":        {},
	"pqc.cert.create":         {},
}

// ValidateSubjectForOperation 按 operation_id 判断是否需要校验 subject。
//
// 若 operation 在白名单内且 params 中存在 subject，则执行 ValidateSubject；
// 否则直接返回 nil（不强制要求 subject 存在，交给 core 兜底）。
func ValidateSubjectForOperation(op string, params map[string]interface{}) error {
	if _, need := subjectOperations[op]; !need {
		return nil
	}
	raw, ok := params["subject"]
	if !ok {
		return nil
	}
	subj, ok := raw.(map[string]interface{})
	if !ok {
		return exception.New(
			exception.CodeParamInvalid,
			"subject 必须是对象", 400, nil,
		)
	}
	return ValidateSubject(subj)
}
