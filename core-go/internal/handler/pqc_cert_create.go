package handler

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/yourorg/core-go/internal/audit"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)

// PQCCertCreateHandler pqc.cert.create 后量子证书。
type PQCCertCreateHandler struct{}

func (h *PQCCertCreateHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    alg := strParam(req.Params, "algorithm")
    switch alg {
    case "ML-KEM", "ML-DSA", "SLH-DSA":
    default:
        return genErr("ALGORITHM_NOT_ALLOWED", "PQC requires ML-KEM/ML-DSA/SLH-DSA")
    }

    subject, ok := subjectFromParams(req.Params)
    if !ok {
        return genErr("INVALID_PARAM", "missing subject")
    }
    days := intParam(req.Params, "validity_days", 365)

    tmpDir, err := mkTmpDir("pqc-cert")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    keyPath := filepath.Join(tmpDir, "key.pem")
    if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
        return genErr("VERSION_UNSUPPORTED", "PQC keygen not supported: "+err.Error())
    }

    plainKey, _ := os.ReadFile(keyPath)
    keyRef, err := saveEncryptedKey(plainKey)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    certID := newID("pqc")
    certRel := "data/pqc/" + certID + ".pem"
    certAbs, _ := pathguard.Resolve(certRel)
    _ = pathguard.EnsureDir(filepath.Dir(certAbs), 0750)

    subjArg := subjectToOpenSSL(subject)
    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    extPath := filepath.Join(tmpDir, "pqc.ext")
    ext := "basicConstraints=CA:FALSE\n"
    if alg == "ML-KEM" {
        ext += "keyUsage=critical,keyEncipherment,keyAgreement\n"
    } else {
        ext += "keyUsage=critical,digitalSignature\n"
    }
    _ = os.WriteFile(extPath, []byte(ext), 0600)

    args := []string{
        "req", "-new", "-x509",
        "-key", keyPath,
        "-out", certAbs,
        "-days", fmt.Sprintf("%d", days),
        "-subj", subjArg,
        "-extfile", extPath,
    }

    caCertRel := strParam(req.Params, "ca_cert_path")
    caKeyRef := strParam(req.Params, "ca_key_ref")
    if caCertRel != "" && caKeyRef != "" {
        caCertAbs, _ := pathguard.Resolve(caCertRel)
        caKeyAbs, _ := loadDecryptedKey(caKeyRef, tmpDir)
        if caKeyAbs != "" {
            // 用 CA 签发
            csrPath := filepath.Join(tmpDir, "req.csr")
            if _, err := client.Run(ctx, "req", "-new", "-key", keyPath,
                "-out", csrPath, "-subj", subjArg); err != nil {
                return genErr("CORE_EXEC_FAILED", err.Error())
            }
            digest := digestForCertAlg(ctx, client, caCertAbs)
            args2 := []string{"x509", "-req", "-in", csrPath,
                "-CA", caCertAbs, "-CAkey", caKeyAbs, "-CAcreateserial",
                "-out", certAbs, "-days", fmt.Sprintf("%d", days),
                "-extfile", extPath}
            if digest != "" {
                args2 = append(args2, digest)
            }
            if _, err := client.Run(ctx, args2...); err != nil {
                return genErr("VERSION_UNSUPPORTED", "PQC sign failed: "+err.Error())
            }
            _ = os.Chmod(certAbs, 0640)
            audit.Log(req, "pqc.cert.create", start, "SUCCESS", "")
            return envelope.NewSuccess(map[string]interface{}{
                "cert_path": certRel,
                "key_ref":   keyRef,
                "algorithm": alg,
            })
        }
    }

    if _, err := client.Run(ctx, args...); err != nil {
        return genErr("VERSION_UNSUPPORTED", "PQC self-sign failed: "+err.Error())
    }
    _ = os.Chmod(certAbs, 0640)

    audit.Log(req, "pqc.cert.create", start, "SUCCESS", "")
    return envelope.NewSuccess(map[string]interface{}{
        "cert_path": certRel,
        "key_ref":   keyRef,
        "algorithm": alg,
    })
}
