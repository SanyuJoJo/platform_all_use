package crypto

import "testing"

// TestGetNestedString 锁定 subject_cn 落库路径的取值语义。
//
// 回归背景：getNestedString 曾只断言 map[string]interface{}，
// 平台侧手工构造的 map[string]string 会被静默吞成空串，
// 表现为 platform_ca.subject_cn 为空但不报任何错。
func TestGetNestedString(t *testing.T) {
	cases := []struct {
		name   string
		params map[string]interface{}
		outer  string
		inner  string
		want   string
	}{
		{
			name: "JSON 反序列化 map[string]interface{}",
			params: map[string]interface{}{
				"subject": map[string]interface{}{"C": "CNbeijing", "CN": "test.example.com"},
			},
			outer: "subject", inner: "CN", want: "test.example.com",
		},
		{
			name: "平台侧手工构造 map[string]string",
			params: map[string]interface{}{
				"subject": map[string]string{"C": "CNbeijing", "CN": "test.example.com"},
			},
			outer: "subject", inner: "CN", want: "test.example.com",
		},
		{
			name: "小写 inner 兼容",
			params: map[string]interface{}{
				"subject": map[string]string{"cn": "lower.example.com"},
			},
			outer: "subject", inner: "CN", want: "lower.example.com",
		},
		{
			name: "非字符串值 fmt.Sprint 兜底",
			params: map[string]interface{}{
				"subject": map[string]interface{}{"CN": 12345},
			},
			outer: "subject", inner: "CN", want: "12345",
		},
		{
			name:   "外层缺失",
			params: map[string]interface{}{},
			outer:  "subject", inner: "CN", want: "",
		},
		{
			name: "内层缺失",
			params: map[string]interface{}{
				"subject": map[string]interface{}{"O": "Example"},
			},
			outer: "subject", inner: "CN", want: "",
		},
		{
			name:   "外层类型错误",
			params: map[string]interface{}{"subject": "CN=not-a-map"},
			outer:  "subject", inner: "CN", want: "",
		},
		{
			name: "nil 值",
			params: map[string]interface{}{
				"subject": map[string]interface{}{"CN": nil},
			},
			outer: "subject", inner: "CN", want: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// json/gin 下发后的真实形态：inner 恒为大写 CN。
			if got := getNestedString(c.params, c.outer, c.inner); got != c.want {
				t.Fatalf("getNestedString=%q, want %q", got, c.want)
			}
		})
	}
}

// TestGetNestedStringFromLowercaseKey 小写 inner 也应命中大写 DN 属性。
func TestGetNestedStringFromLowercaseKey(t *testing.T) {
	params := map[string]interface{}{
		"subject": map[string]string{"CN": "upper.example.com"},
	}
	if got := getNestedString(params, "subject", "cn"); got != "upper.example.com" {
		t.Fatalf("lowercase inner lookup=%q, want upper.example.com", got)
	}
}
