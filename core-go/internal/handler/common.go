package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
//   - RSA 位数：优先 key_params.key_size，回退 key_params.bits，
//     再回退顶层 bits / key_size / key_bits，最后回退 2048；
//     平台侧 DefaultParams 用 "bits"，core 原代码读 "key_size"，
//     不兼容会导致选 RSA4096 实际生成 RSA2048。详见 rsaBitsFromParams。
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
//
//	Falcon 由 oqs-provider 提供，不在 default provider 中。
//	首次调用可能失败，openssl.Client.Run 会自动注入
//	`-provider oqsprovider -provider default` 重试（见 internal/openssl/client.go）。
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
		bits := rsaBitsFromParams(params)
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

// rsaBitsFromParams 解析 RSA 密钥位数，兼容多种字段名与位置。
//
// 解析顺序（先命中先返回）：
//  1. key_params.key_size  —— core 原始字段
//  2. key_params.bits      —— 平台 DefaultParams 字段
//  3. 顶层 key_size / bits —— 部分调用方（如 ca.create）直接把位数摊平在 params 上
//  4. 平台算法 ID 推导      —— algorithm=RSA4096 / RSA2048 / RSA1024
//  5. 兜底 2048
//
// 为什么要第 4 步：
//
//	平台侧 ca.create 只透传 algorithm="RSA4096"，**不写 key_params**，
//	core 只读 key_params.key_size 会静默降级为 2048 —— 即"选 4096 实得 2048"。
//	必须从算法 ID 反推位数，才能与用户选择保持一致。
func rsaBitsFromParams(params map[string]interface{}) int {
	if params == nil {
		return 2048
	}

	// 1/2：key_params 内
	kp := map[string]interface{}{}
	if raw, ok := params["key_params"].(map[string]interface{}); ok {
		kp = raw
	}
	if bits := intParam(kp, "key_size", 0); bits > 0 {
		return bits
	}
	if bits := intParam(kp, "bits", 0); bits > 0 {
		return bits
	}

	// 3：顶层
	if bits := intParam(params, "key_size", 0); bits > 0 {
		return bits
	}
	if bits := intParam(params, "bits", 0); bits > 0 {
		return bits
	}
	if bits := intParam(params, "key_bits", 0); bits > 0 {
		return bits
	}

	// 4：算法 ID 推导（RSA4096 → 4096）
	alg := strParam(params, "algorithm")
	if len(alg) > 3 && strings.EqualFold(alg[:3], "RSA") {
		if bits := atoiSafe(alg[3:]); bits > 0 {
			return bits
		}
	}

	// 5：兜底
	return 2048
}

// atoiSafe 宽松解析十进制数字串，失败返回 0。
func atoiSafe(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
		if n > 1_000_000 {
			return 0
		}
	}
	return n
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

// Digest 摘要算法描述（平台 → openssl 参数）。
type Digest struct {
	// ID 平台规范 ID，如 SM3 / SHA256 / SHA1。
	ID string
	// Arg openssl 摘要参数，如 -sm3 / -sha256 / -sha1。
	Arg string
}

// digestRegistry 支持的摘要算法（除 SM3 外均为国际摘要，按算法族约束可用性）。
//
// ★ 默认行为：未显式指定 digest_algorithm 时，沿用"按算法自动选择"的逻辑
// （SM2 → SM3，RSA/ECC → SHA256）。显式指定时严格按用户选择，不做自动切换，
// 但会按公钥算法校验合法性（见 digest_policy.go：SM2 只能 SM3；PQC 不可指定）。
var digestRegistry = map[string]Digest{
	"SM3":    {ID: "SM3", Arg: "-sm3"},
	"SHA1":   {ID: "SHA1", Arg: "-sha1"},
	"SHA224": {ID: "SHA224", Arg: "-sha224"},
	"SHA256": {ID: "SHA256", Arg: "-sha256"},
	"SHA384": {ID: "SHA384", Arg: "-sha384"},
	"SHA512": {ID: "SHA512", Arg: "-sha512"},
}

// defaultDigestID 显式指定摘要时使用的默认值。
const defaultDigestID = "SM3"

// AllowedDigestAlgorithms 返回平台允许使用的全部摘要算法 ID。
// 具体某算法可用哪些由 allowedDigestsForAlgorithm 决定。
func AllowedDigestAlgorithms() []string {
	return []string{"SM3", "SHA1", "SHA224", "SHA256", "SHA384", "SHA512"}
}

// DefaultDigestForAlgorithm 返回某密钥算法对应的默认摘要 ID。
func DefaultDigestForAlgorithm(alg string) string {
	return defaultDigestForAlgorithm(alg)
}

// defaultDigestForAlgorithm 返回某密钥算法对应的默认摘要 ID。
func defaultDigestForAlgorithm(alg string) string {
	switch alg {
	case "RSA", "ECC":
		return "SHA256"
	}
	// SM2 / PQC（ML-DSA、SLH-DSA、Falcon）：SM2 默认 SM3；
	// PQC 实际摘要由 openssl 子算法决定，这里仅用于展示兜底。
	return defaultDigestID
}

// digestForAlgorithm 返回 openssl 摘要参数（按密钥算法的默认值）。
//
// ★ 必须始终显式返回一个 -xxx 参数：
// openssl req -x509 / openssl req -new 在不指定 -md 时默认使用 SHA1，
// 这正是历史上"界面申请到的证书摘要全是 SHA1"的根因。
func digestForAlgorithm(alg string) string {
	switch alg {
	case "SM2":
		return "-sm3"
	case "RSA", "ECC":
		return "-sha256"
	case "ML-DSA", "SLH-DSA", "Falcon":
		// PQC 算法签名自带哈希，不需要 openssl 再指定摘要
		return ""
	}
	return "-sha256"
}

// NormalizeDigestID 把前端/调用方传入的摘要名归一化为规范 ID。
//
// 兼容：sm3 / SM3 / sha256 / SHA-256 / sha2-256 / sha1 / SHA-1 / md5 等写法。
// 返回空串表示未指定或未识别。
func NormalizeDigestID(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.NewReplacer("-", "", "_", "", ".", "", " ", "").Replace(s)
	switch s {
	case "", "AUTO":
		return ""
	case "SM3":
		return "SM3"
	case "SHA1":
		return "SHA1"
	case "SHA224":
		return "SHA224"
	case "SHA256":
		return "SHA256"
	case "SHA384":
		return "SHA384"
	case "SHA512":
		return "SHA512"
	case "MD5":
		return "MD5"
	}
	// 兼容 ECDSA-WITH-SHA256 / SHA256WITHRSAENCRYPTION / SM2-WITH-SM3 等写法
	switch {
	case strings.Contains(s, "SM3"):
		return "SM3"
	case strings.Contains(s, "SHA1"):
		return "SHA1"
	case strings.Contains(s, "SHA224"):
		return "SHA224"
	case strings.Contains(s, "SHA256"):
		return "SHA256"
	case strings.Contains(s, "SHA384"):
		return "SHA384"
	case strings.Contains(s, "SHA512"):
		return "SHA512"
	}
	return ""
}

// ErrInvalidDigest 摘要参数错误（由调用方转换为 INVALID_PARAM）。
var ErrInvalidDigest = errors.New("invalid digest_algorithm")

// ResolveDigest 解析 digest_algorithm 参数，返回摘要描述符。
//
// 返回 ok=false 表示调用方未显式指定摘要，应沿用"按算法自动选择"的逻辑。
// 返回 error 表示显式指定的摘要不可识别。
func ResolveDigest(raw string) (Digest, bool, error) {
	id := NormalizeDigestID(raw)
	if id == "" {
		return Digest{}, false, nil
	}
	if d, ok := digestRegistry[id]; ok {
		return d, true, nil
	}
	return Digest{}, true, ErrInvalidDigest
}

// explicitDigestFromParams 从 params 中解析调用方显式指定的摘要 ID（未指定返回空串）。
//
// 兼容两个键：digest_algorithm（推荐）与 md（openssl 原生写法）。
// 两个键：digest_algorithm（推荐）与 md（openssl 原生写法）。
// 调用方若传了不可识别的值，会返回 error。
func explicitDigestFromParams(params map[string]interface{}) (string, error) {
	raw := strParam(params, "digest_algorithm")
	if raw == "" {
		raw = strParam(params, "md")
	}
	if d, ok, err := ResolveDigest(raw); ok || err != nil {
		return d.ID, err
	}
	return "", nil
}

// digestArgFor 返回 digest ID 对应的 openssl 参数；空 ID 返回空串（PQC）。
func digestArgFor(id string) string {
	if id == "" {
		return ""
	}
	if d, ok := digestRegistry[id]; ok {
		return d.Arg
	}
	return ""
}

// pickDigest 摘要选择总入口（CA / CSR 自签）：
//   - params 显式指定 digest_algorithm → 按公钥算法校验后严格使用；
//   - 未指定 → 按密钥算法自动选择（SM2→SM3，RSA/ECC→SHA256，PQC 无需摘要）。
//
// 校验确保：SM2 只能 SM3；RSA/ECC 可选 SHA1/SHA224/SHA256/SHA384/SHA512；
// PQC 不可指定（摘要即子算法本身）。
func pickDigest(params map[string]interface{}, alg string) (string, error) {
	explicit, err := explicitDigestFromParams(params)
	if err != nil {
		return "", err
	}
	if explicit != "" {
		id, verr := ValidateDigestForAlgorithm(alg, explicit)
		if verr != nil {
			return "", verr
		}
		return digestArgFor(id), nil
	}
	// ★ 关键修复：RSA / ECC 的自签名证书过去不带 -md，openssl req -x509
	//   会退化为 SHA1 默认摘要；这里显式给出 SHA256，保证"未指定也安全"。
	//   PQC 则返回空串，由 openssl 使用子算法自带散列。
	return digestForAlgorithm(alg), nil
}

// resolvedDigestID 返回本次实际使用的摘要规范 ID（写入审计 / 响应）。
func resolvedDigestID(params map[string]interface{}, arg string) string {
	if d, ok, err := ResolveDigest(strParam(params, "digest_algorithm")); ok && err == nil {
		return d.ID
	}
	if d, ok, err := ResolveDigest(strParam(params, "md")); ok && err == nil {
		return d.ID
	}
	switch arg {
	case "-sm3":
		return "SM3"
	case "-sha256":
		return "SHA256"
	case "-sha384":
		return "SHA384"
	case "-sha512":
		return "SHA512"
	case "-sha1":
		return "SHA1"
	case "-sha224":
		return "SHA224"
	case "-md5":
		return "MD5"
	}
	return ""
}

// pickDigestForCA 按 CA 证书选择摘要（签发终端证书 / 中间 CA 使用）。
//
// caFamily 为 CA 公钥算法族（SM2 / RSA / ECC / PQC），空则从证书探测。
// params 未显式指定时，按 CA 公钥算法自动选择（SM2→SM3，RSA/ECC→SHA256）。
// 显式指定时按 caFamily 校验合法性。
func pickDigestForCA(
	ctx context.Context, client *openssl.Client, params map[string]interface{},
	caCert, caFamily string,
) (string, error) {
	explicit, err := explicitDigestFromParams(params)
	if err != nil {
		return "", err
	}
	if explicit != "" {
		if caFamily == "" {
			caFamily = digestFamilyOfCert(ctx, client, caCert)
		}
		id, verr := ValidateDigestForAlgorithm(caFamily, explicit)
		if verr != nil {
			return "", verr
		}
		return digestArgFor(id), nil
	}
	return digestForCertAlg(ctx, client, caCert), nil
}

// digestFamilyOfCert 探测 CA 证书公钥算法族。
func digestFamilyOfCert(ctx context.Context, client *openssl.Client, caCert string) string {
	out, err := client.Run(ctx, "x509", "-in", caCert, "-noout", "-text")
	if err != nil {
		return "RSA"
	}
	s := string(out)
	if strings.Contains(s, "SM2") || strings.Contains(s, "sm2") ||
		strings.Contains(s, "1.2.156.10197.1.301") {
		return "SM2"
	}
	lower := strings.ToLower(s)
	switch {
	case strings.Contains(lower, "ml-dsa"), strings.Contains(lower, "mldsa"):
		return "ML-DSA"
	case strings.Contains(lower, "slh-dsa"), strings.Contains(lower, "slhdsa"):
		return "SLH-DSA"
	case strings.Contains(lower, "falcon"):
		return "Falcon"
	case strings.Contains(lower, "ecpublickey"), strings.Contains(lower, "prime256v1"),
		strings.Contains(lower, "secp256r1"), strings.Contains(lower, "ecdsa"):
		return "ECC"
	}
	return "RSA"
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
