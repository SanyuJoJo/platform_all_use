package crypto

import "strings"

// normalizeCertTypes 归一化证书类型。
//
// 规则：
//   - 'server' / 'client' → 'signature'（兼容旧前端与历史调用）
//   - 去重、去空
//   - 空输入返回 ['signature']
//
// 背景：早期代码把 X.509 EKU 的 server / client 当作业务分类写进 cert_type，
// 现已收敛为 signature / encryption 两类；本函数作为后端防御，避免脏数据落库。
func normalizeCertTypes(types []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range types {
		t = strings.TrimSpace(strings.ToLower(t))
		switch t {
		case "server", "client":
			t = "signature"
		case "":
			continue
		}
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		out = []string{"signature"}
	}
	return out
}
