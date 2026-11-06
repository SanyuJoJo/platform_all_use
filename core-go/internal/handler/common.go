package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourorg/core-go/internal/envelope"
	"github.com/yourorg/core-go/internal/keystore"
	"github.com/yourorg/core-go/internal/openssl"
	"github.com/yourorg/core-go/internal/pathguard"
	"github.com/yourorg/core-go/internal/platform"
)

// -----------------------------------------------------------------------------
// ID 生成
// -----------------------------------------------------------------------------

// newID 生成短 ID。
func newID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s-%s", prefix,
		time.Now().UTC().Format("20060102150405"),
		hex.EncodeToString(b))
}

// newKeyRef 生成 key_ref。
func newKeyRef() string { return newID("key") }

// -----------------------------------------------------------------------------
// 参数提取
// -----------------------------------------------------------------------------

// strParam 取字符串参数。
func strParam(params map[string]interface{}, key string) string {
	v, _ := params[key].(string)
	return v
}

// strParamDefault 取字符串参数（带默认值）。
func strParamDefault(params map[string]interface{}, key, def string) string {
	if v := strParam(params, key); v != "" {
		return v
	}
	return def
}

// intParam 取 int 参数。
func intParam(params map[string]interface{}, key string, def int) int {
	switch v := params[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return def
}

// subjectFromParams 提取 subject。
func subjectFromParams(params map[string]interface{}) (map[string]string, bool) {
	raw, ok := params["subject"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if s, ok := v.(string); ok && s != "" {
			out[strings.ToUpper(k)] = s
		}
	}
	return out, len(out) > 0
}

// subjectToOpenSSL 拼接 openssl -subj 参数（/C=CN/CN=Test/O=Example）。
func subjectToOpenSSL(subject map[string]string) string {
	order := []string{"C", "ST", "L", "O", "OU", "CN", "E",
		"SERIALNUMBER", "SURNAME", "GIVENNAME"}
	var sb strings.Builder
	used := map[string]bool{}
	for _, k := range order {
		if v, ok := subject[k]; ok && v != "" {
			sb.WriteString("/")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			used[k] = true
		}
	}
	for k, v := range subject {
		if used[k] || v == "" {
			continue
		}
		sb.WriteString("/")
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	return sb.String()
}

// -----------------------------------------------------------------------------
// 密钥加密存储
// -----------------------------------------------------------------------------

// saveEncryptedKey 用 AES-256-GCM 加密私钥并落盘，返回 key_ref。
func saveEncryptedKey(plainKey []byte) (string, error) {
	keyRef := newKeyRef()
	keysDir := pathguard.MustResolve("data/keys")
	if err := pathguard.EnsureDir(keysDir, 0700); err != nil {
		return "", err
	}
	enc, err := keystore.Default().Encrypt(plainKey)
	if err != nil {
		return "", err
	}
	encPath := filepath.Join(keysDir, keyRef+".key.enc")
	if err := os.WriteFile(encPath, enc, 0600); err != nil {
		return "", err
	}
	_ = platform.SecureChmod(encPath, 0600)
	return keyRef, nil
}

// loadDecryptedKey 解密私钥到临时文件，返回临时文件绝对路径。
func loadDecryptedKey(keyRef, tmpDir string) (string, error) {
	keysDir := pathguard.MustResolve("data/keys")
	encPath := filepath.Join(keysDir, keyRef+".key.enc")
	enc, err := os.ReadFile(encPath)
	if err != nil {
		return "", fmt.Errorf("read encrypted key: %w", err)
	}
	plain, err := keystore.Default().Decrypt(enc)
	if err != nil {
		return "", fmt.Errorf("decrypt key: %w", err)
	}
	tmpFile := filepath.Join(tmpDir, "decrypted-"+keyRef+".pem")
	if err := os.WriteFile(tmpFile, plain, 0600); err != nil {
		return "", err
	}
	_ = platform.SecureChmod(tmpFile, 0600)
	return tmpFile, nil
}

// -----------------------------------------------------------------------------
// 临时目录
// -----------------------------------------------------------------------------

// mkTmpDir 创建 core/tmp 下的临时目录。
func mkTmpDir(prefix string) (string, error) {
	base := pathguard.MustResolve("tmp")
	if err := pathguard.EnsureDir(base, 0700); err != nil {
		return "", err
	}
	return os.MkdirTemp(base, prefix+"-")
}

// cleanupTmpDir 清理临时目录。
func cleanupTmpDir(dir string) {
	if dir == "" {
		return
	}
	_ = os.RemoveAll(dir)
}

// -----------------------------------------------------------------------------
// 错误响应
// -----------------------------------------------------------------------------

// genErr 生成错误响应。
func genErr(code, msg string) *envelope.Response {
	return envelope.NewError(code, msg)
}

// -----------------------------------------------------------------------------
// 私钥生成
// -----------------------------------------------------------------------------

// genPrivateKey 生成私钥。
//
// ★ 字段兼容：
//   - RSA 位数：优先 key_params.key_size，回退 key_params.bits，再回退 2048；
//     平台侧 DefaultParams 用 "bits"，core 原代码读 "key_size"，
//     不兼容会导致选 RSA4096 实际生成 RSA2048。
//   - PQC 参数集：优先 key_params.parameter，回退 key_params.param_set；
//     平台侧 DefaultParams 用 "param_set"，core 原代码读 "parameter"，
//     不兼容会导致选 ML-DSA-44 实际生成 ML-DSA-65（默认值）。
//   - Falcon 算法名：自动归一为 oqs-provider 注册名（全小写无连字符）。
//
// ★ 关于 Falcon 的命名兼容：
//
//	oqs-provider 注册名  →  falcon512 / falcon1024
//	NIST 规范名          →  Falcon-512 / Falcon-1024
//	环境变量覆盖          →  任意用户自定义名
//
//	openssl -algorithm 是大小写敏感的**精确匹配**，传 "Falcon-512" 会报：
//	  "unsupported ... Algorithm (Falcon-512 : 0), Properties (<null>)"
//	本函数统一归一为全小写无连字符形式，对以下输入都能正确工作：
//	  "falcon512" / "Falcon-512" / "Falcon512" / "FALCON-512" / "falcon-512"
//
// ★ 关于 oqs-provider：
//   Falcon 由 oqs-provider 提供，不在 default provider 中。
//   首次调用可能失败，openssl.Client.Run 会自动注入
//   `-provider oqsprovider -provider default` 重试（见 internal/openssl/client.go）。
func genPrivateKey(ctx context.Context, alg string, params map[string]interface{}, out string) error {
	client, err := openssl.NewClient()
	if err != nil {
		return err
	}

	// 统一提取 key_params（便于各 case 复用）
	kp := map[string]interface{}{}
	if raw, ok := params["key_params"].(map[string]interface{}); ok {
		kp = raw
	}

	switch alg {
	case "SM2":
		_, err = client.Run(ctx, "genpkey", "-algorithm", "SM2", "-out", out)

	case "RSA":
		bits := intParam(kp, "key_size", 0)
		if bits == 0 {
			bits = intParam(kp, "bits", 0)
		}
		if bits == 0 {
			bits = 2048
		}
		_, err = client.Run(ctx, "genpkey", "-algorithm", "RSA",
			"-pkeyopt", fmt.Sprintf("rsa_keygen_bits:%d", bits), "-out", out)

	case "ECC":
		curve := strParamDefault(kp, "curve", "prime256v1")
		_, err = client.Run(ctx, "ecparam", "-name", curve, "-genkey",
			"-noout", "-out", out)

	case "ML-DSA":
		param := pqcParamSet(kp, "ML-DSA-65")
		_, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)

	case "ML-KEM":
		param := pqcParamSet(kp, "ML-KEM-768")
		_, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)

	case "SLH-DSA":
		param := pqcParamSet(kp, "SLH-DSA-SHA2-128s")
		_, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)

	case "Falcon":
		// ★ 归一为 oqs-provider 注册名（全小写无连字符）
		param := normalizeFalconName(pqcParamSet(kp, "falcon512"))
		_, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)

	default:
		return fmt.Errorf("unsupported algorithm: %s", alg)
	}
	if err == nil {
		_ = os.Chmod(out, 0600)
	}
	return err
}

// pqcParamSet 读取 PQC 参数集，兼容 parameter / param_set 两种字段名。
//
// 平台侧 backend-go 使用 "param_set"，core 原代码使用 "parameter"；
// 本次同时写入两字段（见 algorithms.go 的 applyCoreAlgorithmMapping），
// 此函数再双读，保证任何一侧命名变化都不会静默降级到默认参数集。
func pqcParamSet(kp map[string]interface{}, def string) string {
	if v := strParam(kp, "parameter"); v != "" {
		return v
	}
	if v := strParam(kp, "param_set"); v != "" {
		return v
	}
	return def
}

// normalizeFalconName 把 Falcon 算法名的各种写法归一为 oqs-provider 注册名。
//
// 输入变体 → 输出（oqs-provider 注册名）：
//
//	"Falcon-512"  → "falcon512"
//	"Falcon512"   → "falcon512"
//	"FALCON-512"  → "falcon512"
//	"falcon-512"  → "falcon512"
//	"Falcon-1024" → "falcon1024"
//	"falconpadded512" → "falconpadded512"（前缀仍是 falcon，仅去分隔符，结果不变）
//
// 非 falcon 前缀的名字原样返回（保留未来扩展空间）。
//
// 为什么需要这个函数：
//
//	oqs-provider 的注册名是 OQS 项目规范（全小写无连字符），
//	与 NIST FIPS 206 规范名（Falcon-512）不同。openssl -algorithm
//	按注册名做**大小写敏感精确匹配**，任何偏差都会报 unsupported。
//	平台侧 DisplayName 面向用户（Falcon-512），ParamSet 面向 openssl
//	（falcon512），此函数作为 core 侧的**最后一道防线**兜底。
func normalizeFalconName(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	if !strings.HasPrefix(lower, "falcon") {
		return s
	}
	// 去掉所有分隔符（- / _ / 空格）
	lower = strings.NewReplacer("-", "", "_", "", " ", "").Replace(lower)
	return lower
}

// -----------------------------------------------------------------------------
// 摘要选择
// -----------------------------------------------------------------------------

// digestForAlgorithm 返回 openssl 摘要参数（按算法）。
//
// PQC 算法（ML-DSA / SLH-DSA / Falcon）的签名是"自带哈希"的，
// 不需要 openssl 再指定摘要，因此返回空串。
func digestForAlgorithm(alg string) string {
	switch alg {
	case "SM2":
		return "-sm3"
	case "RSA", "ECC":
		return "-sha256"
	case "ML-DSA", "SLH-DSA", "Falcon":
		return ""
	}
	return "-sha256"
}

// digestForCertAlg 根据 CA 证书公钥算法选摘要。
//
// 说明：这是全局唯一实现。cert_sign.go 等文件不应重复定义。
func digestForCertAlg(ctx context.Context, client *openssl.Client, caCert string) string {
	out, err := client.Run(ctx, "x509", "-in", caCert, "-noout", "-text")
	if err != nil {
		return "-sha256"
	}
	s := string(out)
	if strings.Contains(s, "SM2") || strings.Contains(s, "sm2") ||
		strings.Contains(s, "1.2.156.10197.1.301") {
		return "-sm3"
	}
	return "-sha256"
}

// -----------------------------------------------------------------------------
// 序列号提取
// -----------------------------------------------------------------------------

// extractSerial 提取证书序列号。
func extractSerial(ctx context.Context, client *openssl.Client, certPath string) string {
	out, err := client.Run(ctx, "x509", "-in", certPath, "-noout", "-serial")
	if err != nil {
		return ""
	}
	s := string(out)
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return trimSpace(s[i+1:])
		}
	}
	return trimSpace(s)
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\r' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			s = s[:len(s)-1]
		} else {
			break
		}
	}
	return s
}
