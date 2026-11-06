package crypto

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"

	"backend-go/internal/exception"
)

// AlgorithmSpec 算法元信息。
//
// ★ DisplayName 是**展示名**（面向用户），必须与前端 ALGORITHM_DISPLAY 一致：
//   - SM2 / RSA-1024 / RSA-2048 / RSA-4096
//   - ML-DSA-44 / ML-DSA-65 / ML-DSA-87
//   - Falcon(512) / Falcon(1024)         ← 括号形态，与前端对齐
//   - SLH-DSA-SHA2-128s / ...
type AlgorithmSpec struct {
	ID            string                 `json:"id"`
	DisplayName   string                 `json:"display_name"`
	Family        string                 `json:"family"`
	Category      string                 `json:"category"`
	KeyBits       int                    `json:"key_bits"`
	DefaultParams map[string]interface{} `json:"default_params"`
	SupportsDual  bool                   `json:"supports_dual"`
	Notes         string                 `json:"notes"`
	Available     bool                   `json:"available"`

	// CoreName 传给 core 的算法家族名（SM2/RSA/ECC/ML-DSA/SLH-DSA/Falcon）。
	CoreName string `json:"-"`

	// ParamSet 传给 core 的参数集（ML-DSA-65 / falcon1024 等）。
	ParamSet string `json:"-"`
}

// CoreAlgorithm 返回传给 core 的算法家族名。
func (s AlgorithmSpec) CoreAlgorithm() string {
	if v := strings.TrimSpace(os.Getenv("CRYPTO_CORE_ALG_" + s.ID)); v != "" {
		return v
	}
	if strings.TrimSpace(s.CoreName) != "" {
		return s.CoreName
	}
	return s.ID
}

var AlgorithmRegistry = map[string]AlgorithmSpec{
	"SM2": {
		ID: "SM2", DisplayName: "SM2", Family: "SM2", Category: "国密",
		KeyBits: 256, SupportsDual: true, Available: true,
		Notes: "国密标准；支持数字信封、双证签发、PKCS#12 导出",
	},
	"RSA1024": {
		ID: "RSA1024", DisplayName: "RSA-1024", Family: "RSA", Category: "RSA",
		KeyBits: 1024, CoreName: "RSA", Available: rsa1024Allowed(),
		DefaultParams: map[string]interface{}{"bits": 1024, "key_size": 1024},
		Notes:         "安全强度约 80 bit，仅在兼容旧系统时启用",
	},
	"RSA2048": {
		ID: "RSA2048", DisplayName: "RSA-2048", Family: "RSA", Category: "RSA",
		KeyBits: 2048, CoreName: "RSA", Available: true,
		DefaultParams: map[string]interface{}{"bits": 2048, "key_size": 2048},
		Notes:         "主流安全级别",
	},
	"RSA4096": {
		ID: "RSA4096", DisplayName: "RSA-4096", Family: "RSA", Category: "RSA",
		KeyBits: 4096, CoreName: "RSA", Available: true,
		DefaultParams: map[string]interface{}{"bits": 4096, "key_size": 4096},
		Notes:         "高强度；签发/握手较慢",
	},
	"MLDSA44": {
		ID: "MLDSA44", DisplayName: "ML-DSA-44", Family: "ML-DSA", Category: "后量子",
		CoreName: "ML-DSA", ParamSet: "ML-DSA-44", Available: true,
		DefaultParams: map[string]interface{}{"param_set": "ML-DSA-44"},
		Notes:         "FIPS 204（原 Dilithium2），NIST Level 2",
	},
	"MLDSA65": {
		ID: "MLDSA65", DisplayName: "ML-DSA-65", Family: "ML-DSA", Category: "后量子",
		CoreName: "ML-DSA", ParamSet: "ML-DSA-65", Available: true,
		DefaultParams: map[string]interface{}{"param_set": "ML-DSA-65"},
		Notes:         "FIPS 204（原 Dilithium3），NIST Level 3",
	},
	"MLDSA87": {
		ID: "MLDSA87", DisplayName: "ML-DSA-87", Family: "ML-DSA", Category: "后量子",
		CoreName: "ML-DSA", ParamSet: "ML-DSA-87", Available: true,
		DefaultParams: map[string]interface{}{"param_set": "ML-DSA-87"},
		Notes:         "FIPS 204（原 Dilithium5），NIST Level 5",
	},
	// ★ Falcon 展示名用括号形态（与前端对齐）
	"falcon512": {
		ID: "falcon512", DisplayName: "Falcon(512)", Family: "Falcon", Category: "后量子",
		CoreName: "Falcon", ParamSet: "falcon512", Available: falconAllowed(),
		DefaultParams: map[string]interface{}{"param_set": "falcon512"},
		Notes:         "NIST Level 1；签名/公钥短，密钥生成较慢",
	},
	"falcon1024": {
		ID: "falcon1024", DisplayName: "Falcon(1024)", Family: "Falcon", Category: "后量子",
		CoreName: "Falcon", ParamSet: "falcon1024", Available: falconAllowed(),
		DefaultParams: map[string]interface{}{"param_set": "falcon1024"},
		Notes:         "NIST Level 5",
	},
}

func init() {
	for _, h := range []string{"sha2", "shake"} {
		for _, lvl := range []string{"128", "192", "256"} {
			for _, spd := range []string{"s", "f"} {
				id := fmt.Sprintf("slh-dsa-%s-%s%s", h, lvl, spd)
				disp := fmt.Sprintf("SLH-DSA-%s-%s%s", strings.ToUpper(h), lvl, spd)

				var available bool
				if h == "sha2" {
					available = true
				} else {
					available = slhShakeAllowed()
				}

				note := "FIPS 205（原 SPHINCS+）"
				if h == "shake" && !available {
					note = "SHAKE 系列已通过 CRYPTO_ALLOW_SLH_SHAKE=false 关闭"
				}

				AlgorithmRegistry[id] = AlgorithmSpec{
					ID:            id,
					DisplayName:   disp,
					Family:        "SLH-DSA",
					Category:      "后量子",
					CoreName:      "SLH-DSA",
					ParamSet:      disp,
					Available:     available,
					DefaultParams: map[string]interface{}{"param_set": disp},
					Notes:         note,
				}
			}
		}
	}
}

// ValidateAlgorithm 校验算法 ID，返回规范 Spec。
func ValidateAlgorithm(raw string) (AlgorithmSpec, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return AlgorithmRegistry["SM2"], nil
	}
	if spec, ok := AlgorithmRegistry[id]; ok {
		return spec, nil
	}
	for k, spec := range AlgorithmRegistry {
		if strings.EqualFold(k, id) {
			return spec, nil
		}
	}
	switch strings.ToUpper(id) {
	case "RSA":
		return AlgorithmRegistry["RSA2048"], nil
	case "ECC", "EC":
		return AlgorithmRegistry["SM2"], nil
	case "ML-DSA", "MLDSA":
		return AlgorithmRegistry["MLDSA65"], nil
	case "FALCON":
		return AlgorithmRegistry["falcon512"], nil
	case "SLH-DSA", "SLHDSA":
		return AlgorithmRegistry["slh-dsa-sha2-128s"], nil
	}
	return AlgorithmSpec{}, exception.New(
		exception.CodeParamInvalid,
		fmt.Sprintf("不支持的算法：%s（可选：%s）", raw, AlgorithmIDList()),
		400, nil,
	)
}

// FillAlgorithmParams 补齐 params.algorithm / params.key_params。
func FillAlgorithmParams(params map[string]interface{}) (AlgorithmSpec, error) {
	if params == nil {
		return AlgorithmSpec{}, fmt.Errorf("params 为空")
	}
	raw, _ := params["algorithm"].(string)
	spec, err := ValidateAlgorithm(raw)
	if err != nil {
		return spec, err
	}
	params["algorithm"] = spec.ID
	if _, ok := params["key_params"]; !ok && len(spec.DefaultParams) > 0 {
		params["key_params"] = cloneKeyParams(spec.DefaultParams)
	}
	return spec, nil
}

// applyCoreAlgorithmMapping 把平台 ID 映射为 core 白名单可识别的形式。
func applyCoreAlgorithmMapping(params map[string]interface{}) {
	if params == nil {
		return
	}
	raw, ok := params["algorithm"].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return
	}
	spec, err := ValidateAlgorithm(raw)
	if err != nil {
		log.Debug().
			Str("raw", raw).
			Err(err).
			Msg("crypto: applyCoreAlgorithmMapping 跳过未知算法")
		return
	}

	coreAlg := spec.CoreAlgorithm()
	if coreAlg != "" && coreAlg != raw {
		params["algorithm"] = coreAlg
	}

	if spec.ParamSet != "" {
		kp := cloneKeyParams(params["key_params"])
		kp["param_set"] = spec.ParamSet
		kp["parameter"] = spec.ParamSet
		params["key_params"] = kp
	}

	log.Debug().
		Str("platform_id", spec.ID).
		Str("core_name", coreAlg).
		Str("param_set", spec.ParamSet).
		Interface("key_params", params["key_params"]).
		Bool("available", spec.Available).
		Msg("crypto: 算法名已映射为 core 形式")
}

func cloneKeyParams(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		out := make(map[string]interface{}, len(m)+2)
		for k, vv := range m {
			out[k] = vv
		}
		return out
	}
	if m, ok := v.(map[string]string); ok {
		out := make(map[string]interface{}, len(m)+2)
		for k, vv := range m {
			out[k] = vv
		}
		return out
	}
	return map[string]interface{}{}
}

// AlgorithmIDList 返回全部合法算法 ID（逗号分隔，供错误提示）。
func AlgorithmIDList() string {
	ids := make([]string, 0, len(AlgorithmRegistry))
	for k := range AlgorithmRegistry {
		ids = append(ids, k)
	}
	sort.Strings(ids)
	return strings.Join(ids, ", ")
}

// ListAlgorithms 返回注册表快照（只返回 Available=true，按 ID 排序）。
func ListAlgorithms() []AlgorithmSpec {
	ids := make([]string, 0, len(AlgorithmRegistry))
	for k, spec := range AlgorithmRegistry {
		if spec.Available {
			ids = append(ids, k)
		}
	}
	sort.Strings(ids)
	out := make([]AlgorithmSpec, 0, len(ids))
	for _, id := range ids {
		out = append(out, AlgorithmRegistry[id])
	}
	return out
}

// =============================================================================
// 环境变量开关（默认开启）
// =============================================================================

func envBoolDefaultTrue(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch v {
	case "", "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	}
	return true
}

func rsa1024Allowed() bool     { return envBoolDefaultTrue("CRYPTO_ALLOW_RSA1024") }
func falconAllowed() bool      { return envBoolDefaultTrue("CRYPTO_ALLOW_FALCON") }
func slhShakeAllowed() bool    { return envBoolDefaultTrue("CRYPTO_ALLOW_SLH_SHAKE") }
