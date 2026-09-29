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
func genPrivateKey(ctx context.Context, alg string, params map[string]interface{}, out string) error {
    client, err := openssl.NewClient()
    if err != nil {
        return err
    }
    switch alg {
    case "SM2":
        _, err = client.Run(ctx, "genpkey", "-algorithm", "SM2", "-out", out)
    case "RSA":
        bits := 2048
        if kp, ok := params["key_params"].(map[string]interface{}); ok {
            bits = intParam(kp, "key_size", 2048)
        }
        _, err = client.Run(ctx, "genpkey", "-algorithm", "RSA",
            "-pkeyopt", fmt.Sprintf("rsa_keygen_bits:%d", bits), "-out", out)
    case "ECC":
        curve := "prime256v1"
        if kp, ok := params["key_params"].(map[string]interface{}); ok {
            curve = strParamDefault(kp, "curve", "prime256v1")
        }
        _, err = client.Run(ctx, "ecparam", "-name", curve, "-genkey",
            "-noout", "-out", out)
    case "ML-DSA":
        param := "ML-DSA-65"
        if kp, ok := params["key_params"].(map[string]interface{}); ok {
            param = strParamDefault(kp, "parameter", "ML-DSA-65")
        }
        _, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)
    case "ML-KEM":
        param := "ML-KEM-768"
        if kp, ok := params["key_params"].(map[string]interface{}); ok {
            param = strParamDefault(kp, "parameter", "ML-KEM-768")
        }
        _, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)
    case "SLH-DSA":
        param := "SLH-DSA-SHA2-128s"
        if kp, ok := params["key_params"].(map[string]interface{}); ok {
            param = strParamDefault(kp, "parameter", "SLH-DSA-SHA2-128s")
        }
        _, err = client.Run(ctx, "genpkey", "-algorithm", param, "-out", out)
    default:
        return fmt.Errorf("unsupported algorithm: %s", alg)
    }
    if err == nil {
        _ = os.Chmod(out, 0600)
    }
    return err
}

// -----------------------------------------------------------------------------
// 摘要选择
// -----------------------------------------------------------------------------

// digestForAlgorithm 返回 openssl 摘要参数（按算法）。
func digestForAlgorithm(alg string) string {
    switch alg {
    case "SM2":
        return "-sm3"
    case "RSA", "ECC":
        return "-sha256"
    case "ML-DSA", "SLH-DSA":
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
