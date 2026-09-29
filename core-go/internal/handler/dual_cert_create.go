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

// DualCertCreateHandler dual_cert.create。
type DualCertCreateHandler struct{}

func (h *DualCertCreateHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    signAlg := strParamDefault(req.Params, "sign_algorithm", "SM2")
    encAlg := strParamDefault(req.Params, "enc_algorithm", "SM2")
    tlcp := strParamDefault(req.Params, "tlcp_profile", "GB/T 38636-2020")

    if signAlg != "SM2" || encAlg != "SM2" || tlcp != "GB/T 38636-2020" {
        return envelope.NewErrorWithDetail("ALGORITHM_NOT_ALLOWED",
            "dual cert requires SM2+SM2+GB/T 38636-2020",
            map[string]interface{}{
                "sign_algorithm": signAlg,
                "enc_algorithm":  encAlg,
                "tlcp_profile":   tlcp,
            })
    }

    caCertRel := strParam(req.Params, "ca_cert_path")
    caKeyRef := strParam(req.Params, "ca_key_ref")
    if caCertRel == "" || caKeyRef == "" {
        return genErr("INVALID_PARAM", "missing ca_cert_path or ca_key_ref")
    }

    caCertAbs, err := pathguard.Resolve(caCertRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }

    subject, ok := subjectFromParams(req.Params)
    if !ok {
        return genErr("INVALID_PARAM", "missing subject")
    }
    days := intParam(req.Params, "validity_days", 365)

    tmpDir, err := mkTmpDir("dual-cert")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    caKeyAbs, err := loadDecryptedKey(caKeyRef, tmpDir)
    if err != nil {
        return genErr("KEY_NOT_FOUND", err.Error())
    }

    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    signKeyPath := filepath.Join(tmpDir, "sign.key")
    encKeyPath := filepath.Join(tmpDir, "enc.key")
    if err := genPrivateKey(ctx, "SM2", req.Params, signKeyPath); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    if err := genPrivateKey(ctx, "SM2", req.Params, encKeyPath); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    signPlain, _ := os.ReadFile(signKeyPath)
    encPlain, _ := os.ReadFile(encKeyPath)
    signKeyRef, _ := saveEncryptedKey(signPlain)
    encKeyRef, _ := saveEncryptedKey(encPlain)

    subjArg := subjectToOpenSSL(subject)
    baseID := newID("dual")

    signCSR := filepath.Join(tmpDir, "sign.csr")
    encCSR := filepath.Join(tmpDir, "enc.csr")
    if _, err := client.Run(ctx, "req", "-new", "-key", signKeyPath,
        "-out", signCSR, "-subj", subjArg); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    if _, err := client.Run(ctx, "req", "-new", "-key", encKeyPath,
        "-out", encCSR, "-subj", subjArg); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    signExt := filepath.Join(tmpDir, "sign.ext")
    encExt := filepath.Join(tmpDir, "enc.ext")
    _ = os.WriteFile(signExt,
        []byte("basicConstraints=CA:FALSE\nkeyUsage=critical,digitalSignature\nextendedKeyUsage=clientAuth\n"),
        0600)
    _ = os.WriteFile(encExt,
        []byte("basicConstraints=CA:FALSE\nkeyUsage=critical,keyEncipherment,keyAgreement\nextendedKeyUsage=clientAuth\n"),
        0600)

    signRel := "data/dual/" + baseID + "-sign.pem"
    encRel := "data/dual/" + baseID + "-enc.pem"
    chainRel := "data/dual/" + baseID + "-chain.pem"
    signAbs, _ := pathguard.Resolve(signRel)
    encAbs, _ := pathguard.Resolve(encRel)
    chainAbs, _ := pathguard.Resolve(chainRel)
    _ = pathguard.EnsureDir(filepath.Dir(signAbs), 0750)

    if _, err := client.Run(ctx, "x509", "-req", "-in", signCSR,
        "-CA", caCertAbs, "-CAkey", caKeyAbs, "-CAcreateserial",
        "-out", signAbs, "-days", fmt.Sprintf("%d", days),
        "-sm3", "-extfile", signExt); err != nil {
        return genErr("CORE_EXEC_FAILED", "sign cert: "+err.Error())
    }

    if _, err := client.Run(ctx, "x509", "-req", "-in", encCSR,
        "-CA", caCertAbs, "-CAkey", caKeyAbs, "-CAcreateserial",
        "-out", encAbs, "-days", fmt.Sprintf("%d", days),
        "-sm3", "-extfile", encExt); err != nil {
        return genErr("CORE_EXEC_FAILED", "enc cert: "+err.Error())
    }

    signData, _ := os.ReadFile(signAbs)
    caData, _ := os.ReadFile(caCertAbs)
    _ = os.WriteFile(chainAbs, append(signData, caData...), 0640)
    _ = os.Chmod(signAbs, 0640)
    _ = os.Chmod(encAbs, 0640)

    audit.Log(req, "dual_cert.create", start, "SUCCESS", "")

    return envelope.NewSuccess(map[string]interface{}{
        "sign_cert_path": signRel,
        "enc_cert_path":  encRel,
        "sign_key_ref":   signKeyRef,
        "enc_key_ref":    encKeyRef,
        "chain_path":     chainRel,
    })
}
