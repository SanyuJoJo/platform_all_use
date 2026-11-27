package handler

import (
	"fmt"
	"strings"
)

// -----------------------------------------------------------------------------
// 摘要选择策略（按公钥算法约束可选摘要）
//
// 规则：
//   - RSA（1024/2048/3072/4096）：SHA1 / SHA224 / SHA256 / SHA384 / SHA512
//   - ECC（P-256 等）：            SHA1 / SHA224 / SHA256 / SHA384 / SHA512
//   - SM2（密钥位数 256）：        仅 SM3
//   - PQC（ML-DSA / SLH-DSA / Falcon）：
//     摘要由子算法自带，且“摘要 == 子算法标识本身”，
//     例如 slh-dsa-shake-256s 的摘要只能是 slh-dsa-shake-256s。
//     因此不传 digest_algorithm 时由 openssl 使用算法自身散列。
// -----------------------------------------------------------------------------

// digestFamilyOf 判断密钥算法所属族。
//
// 兼容平台算法 ID（SM2 / RSA2048 / ECC / MLDSA65 / falcon512 / slh-dsa-shake-256s）
// 与已归一化的族名（SM2 / RSA / ECC / ML-DSA / SLH-DSA / Falcon）。
func digestFamilyOf(alg string) string {
	id := strings.ToUpper(strings.TrimSpace(alg))
	switch {
	case id == "SM2":
		return "SM2"
	case id == "RSA" || strings.HasPrefix(id, "RSA"):
		return "RSA"
	case id == "ECC" || id == "EC":
		return "ECC"
	case id == "ML-DSA" || id == "MLDSA" || strings.HasPrefix(id, "MLDSA") || strings.HasPrefix(id, "ML-DSA"),
		id == "SLH-DSA" || id == "SLHDSA" || strings.HasPrefix(id, "SLH-DSA") || strings.HasPrefix(id, "SLHDSA"),
		id == "FALCON" || strings.HasPrefix(id, "FALCON"):
		return "PQC"
	}
	return ""
}

// allowedDigestsForAlgorithm 返回某密钥算法允许的摘要规范 ID 列表。
//
// 返回 nil 表示“不可另行指定摘要”（PQC：摘要由子算法自带）。
func allowedDigestsForAlgorithm(alg string) []string {
	switch digestFamilyOf(alg) {
	case "SM2":
		return []string{"SM3"}
	case "RSA", "ECC":
		return []string{"SHA1", "SHA224", "SHA256", "SHA384", "SHA512"}
	}
	return nil
}

// ValidateDigestForAlgorithm 校验 digest_algorithm 与密钥算法是否匹配。
//
// 规则：
//   - 未指定摘要（空 / AUTO）：通过，由调用方按默认值选择；
//   - PQC 算法：不允许指定摘要（摘要即子算法本身），指定即报错；
//   - SM2：只允许 SM3；
//   - RSA / ECC：只允许 SHA1 / SHA224 / SHA256 / SHA384 / SHA512。
//
// 返回归一化后的摘要 ID（未指定时返回空串）。
func ValidateDigestForAlgorithm(alg, raw string) (string, error) {
	id := NormalizeDigestID(raw)
	if id == "" {
		return "", nil
	}
	family := digestFamilyOf(alg)
	if family == "PQC" {
		return "", fmt.Errorf(
			"invalid digest_algorithm: %s（后量子算法 %s 的摘要由子算法自带，不可另行选择）",
			raw, alg,
		)
	}
	allowed := allowedDigestsForAlgorithm(alg)
	if len(allowed) == 0 {
		return "", fmt.Errorf("invalid digest_algorithm: %s (unknown algorithm: %s)", raw, alg)
	}
	for _, a := range allowed {
		if a == id {
			return id, nil
		}
	}
	return "", fmt.Errorf(
		"invalid digest_algorithm: %s（算法 %s 仅支持：%s）",
		raw, alg, strings.Join(allowed, ", "),
	)
}
