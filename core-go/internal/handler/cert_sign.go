package handler

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/yourorg/core-go/internal/audit"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)

// CertSignHandler cert.sign 终端证书签发。
type CertSignHandler struct{}

func (h *CertSignHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    // 1. 白名单
    wl, err := config.GetWhitelist(cfg.WhitelistPath)
    if err != nil {
        return genErr("INTERNAL_ERROR", "whitelist load failed")
    }
    alg := strParamDefault(req.Params, "algorithm", "SM2")
    if !wl.IsAlgorithmAllowed("cert.sign", alg) {
        audit.Log(req, "cert.sign", start, "FAILED", "ALGORITHM_NOT_ALLOWED")
        return genErr("ALGORITHM_NOT_ALLOWED", "algorithm not allowed")
    }

    caSource := strParamDefault(req.Params, "ca_source", "local")
    if !wl.IsCAKindAllowed("cert.sign", caSource) {
        return genErr("INVALID_PARAM", "invalid ca_source: "+caSource)
    }

    certType := strParamDefault(req.Params, "cert_type", "server")
    if !wl.IsCertTypeAllowed("cert.sign", certType) {
        return genErr("INVALID_PARAM", "invalid cert_type: "+certType)
    }

    // 2. 临时目录
    tmpDir, err := mkTmpDir("cert-sign")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    // 3. 解析 CA 来源
    var caCertAbs, caKeyAbs string
    switch caSource {
    case "local":
        caCertRel := strParam(req.Params, "ca_cert_path")
        caKeyRef := strParam(req.Params, "ca_key_ref")
        if caCertRel == "" || caKeyRef == "" {
            return genErr("INVALID_PARAM", "missing ca_cert_path or ca_key_ref")
        }
        caCertAbs, err = pathguard.Resolve(caCertRel)
        if err != nil {
            return genErr("PATH_NOT_ALLOWED", err.Error())
        }
        if _, err := os.Stat(caCertAbs); err != nil {
            return genErr("CERT_NOT_FOUND", "ca cert not found")
        }
        caKeyAbs, err = loadDecryptedKey(caKeyRef, tmpDir)
        if err != nil {
            return genErr("KEY_NOT_FOUND", err.Error())
        }
    case "manual":
        caCertRel := strParam(req.Params, "ca_cert_path")
        caKeyRel := strParam(req.Params, "ca_key_path")
        if caCertRel == "" || caKeyRel == "" {
            return genErr("INVALID_PARAM", "missing manual ca cert/key")
        }
        caCertAbs, err = pathguard.Resolve(caCertRel)
        if err != nil {
            return genErr("PATH_NOT_ALLOWED", err.Error())
        }
        caKeyAbs, err = pathguard.Resolve(caKeyRel)
        if err != nil {
            return genErr("PATH_NOT_ALLOWED", err.Error())
        }
    default:
        return genErr("INVALID_PARAM", "invalid ca_source")
    }

    // 4. 生成终端私钥
    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    keyPath := filepath.Join(tmpDir, "leaf.key")
    if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
        return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
    }

    // 5. 生成 CSR
    subject, _ := subjectFromParams(req.Params)
    subjArg := subjectToOpenSSL(subject)
    csrPath := filepath.Join(tmpDir, "leaf.csr")
    if _, err := client.Run(ctx, "req", "-new", "-key", keyPath,
        "-out", csrPath, "-subj", subjArg); err != nil {
        return genErr("CORE_EXEC_FAILED", "csr: "+err.Error())
    }

    // 6. 扩展配置（KeyUsage / EKU）
    extPath := filepath.Join(tmpDir, "ext.cnf")
    if err := writeCertExt(extPath, certType); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    // 7. 签发
    certID := newID("cert")
    certRel := "data/certs/" + certID + ".pem"
    certAbs, err := pathguard.Resolve(certRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if err := pathguard.EnsureDir(filepath.Dir(certAbs), 0750); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    digest := digestForCertAlg(ctx, client, caCertAbs)
    days := intParam(req.Params, "validity_days", 365)
    if _, err := client.Run(ctx, "x509", "-req", "-in", csrPath,
        "-CA", caCertAbs, "-CAkey", caKeyAbs, "-CAcreateserial",
        "-out", certAbs, "-days", fmt.Sprintf("%d", days),
        digest, "-extfile", extPath); err != nil {
        return genErr("CORE_EXEC_FAILED", "sign: "+err.Error())
    }
    _ = os.Chmod(certAbs, 0640)

    // 8. 链文件
    chainRel := "data/certs/" + certID + "-chain.pem"
    chainAbs, _ := pathguard.Resolve(chainRel)
    if chainAbs != "" {
        leaf, _ := os.ReadFile(certAbs)
        ca, _ := os.ReadFile(caCertAbs)
        _ = os.WriteFile(chainAbs, append(leaf, ca...), 0640)
    }

    // 9. 加密终端私钥（注意：saveEncryptedKey 只接 1 个参数）
    plainKey, _ := os.ReadFile(keyPath)
    keyRef, err := saveEncryptedKey(plainKey)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", "save leaf key: "+err.Error())
    }

    // 10. 审计
    audit.Log(req, "cert.sign", start, "SUCCESS", "")

    return envelope.NewSuccess(map[string]interface{}{
        "cert_id":    certID,
        "cert_path":  certRel,
        "chain_path": chainRel,
        "serial":     extractSerial(ctx, client, certAbs),
        "key_ref":    keyRef,
    })
}

// writeCertExt 写证书扩展配置文件。
func writeCertExt(path, certType string) error {
    var ku, eku string
    switch certType {
    case "server":
        ku = "digitalSignature,keyEncipherment"
        eku = "serverAuth"
    case "client":
        ku = "digitalSignature"
        eku = "clientAuth"
    case "signature":
        ku = "digitalSignature,nonRepudiation"
    case "encryption":
        ku = "keyEncipherment,keyAgreement"
    default:
        ku = "digitalSignature"
    }
    var sb strings.Builder
    sb.WriteString("basicConstraints=CA:FALSE\n")
    sb.WriteString("keyUsage=critical," + ku + "\n")
    if eku != "" {
        sb.WriteString("extendedKeyUsage=" + eku + "\n")
    }
    return os.WriteFile(path, []byte(sb.String()), 0600)
}
