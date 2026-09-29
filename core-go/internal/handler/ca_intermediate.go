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

// CAIntermediateHandler ca.intermediate.create 中间 CA 生成。
type CAIntermediateHandler struct{}

func (h *CAIntermediateHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    wl, err := config.GetWhitelist(cfg.WhitelistPath)
    if err != nil {
        return genErr("INTERNAL_ERROR", "whitelist load failed")
    }
    alg := strParam(req.Params, "algorithm")
    if !wl.IsAlgorithmAllowed("ca.intermediate.create", alg) {
        return genErr("ALGORITHM_NOT_ALLOWED", "algorithm not allowed")
    }

    parentCertRel := strParam(req.Params, "parent_ca_cert_path")
    parentKeyRef := strParam(req.Params, "parent_ca_key_ref")
    if parentCertRel == "" || parentKeyRef == "" {
        return genErr("INVALID_PARAM", "missing parent_ca_cert_path or parent_ca_key_ref")
    }

    parentCertAbs, err := pathguard.Resolve(parentCertRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if _, err := os.Stat(parentCertAbs); err != nil {
        return genErr("CERT_NOT_FOUND", "parent ca cert not found")
    }

    subject, ok := subjectFromParams(req.Params)
    if !ok {
        return genErr("INVALID_PARAM", "missing subject")
    }
    days := intParam(req.Params, "validity_days", 1825)
    pathLen := intParam(req.Params, "path_len", 0)

    tmpDir, err := mkTmpDir("ca-inter")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    parentKeyAbs, err := loadDecryptedKey(parentKeyRef, tmpDir)
    if err != nil {
        return genErr("KEY_NOT_FOUND", err.Error())
    }

    keyPath := filepath.Join(tmpDir, "key.pem")
    if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
        return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
    }

    plainKey, _ := os.ReadFile(keyPath)
    keyRef, err := saveEncryptedKey(plainKey)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    caID := newID("ca-int")
    certRel := "data/ca/" + caID + ".pem"
    certAbs, _ := pathguard.Resolve(certRel)
    _ = pathguard.EnsureDir(filepath.Dir(certAbs), 0750)

    csrPath := filepath.Join(tmpDir, "req.csr")
    subjArg := subjectToOpenSSL(subject)

    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    if _, err := client.Run(ctx, "req", "-new", "-key", keyPath,
        "-out", csrPath, "-subj", subjArg); err != nil {
        return genErr("CORE_EXEC_FAILED", "csr: "+err.Error())
    }

    extPath := filepath.Join(tmpDir, "ca.ext")
    extContent := fmt.Sprintf(
        "basicConstraints=critical,CA:TRUE,pathlen:%d\nkeyUsage=critical,keyCertSign,cRLSign\n",
        pathLen)
    _ = os.WriteFile(extPath, []byte(extContent), 0600)

    digest := digestForCertAlg(ctx, client, parentCertAbs)

    if _, err := client.Run(ctx, "x509", "-req", "-in", csrPath,
        "-CA", parentCertAbs, "-CAkey", parentKeyAbs, "-CAcreateserial",
        "-out", certAbs, "-days", fmt.Sprintf("%d", days),
        digest, "-extfile", extPath); err != nil {
        return genErr("CORE_EXEC_FAILED", "sign: "+err.Error())
    }
    _ = os.Chmod(certAbs, 0640)

    chainRel := "data/ca/" + caID + "-chain.pem"
    chainAbs, _ := pathguard.Resolve(chainRel)
    leaf, _ := os.ReadFile(certAbs)
    parent, _ := os.ReadFile(parentCertAbs)
    _ = os.WriteFile(chainAbs, append(leaf, parent...), 0640)

    serial := extractSerial(ctx, client, certAbs)
    audit.Log(req, "ca.intermediate.create", start, "SUCCESS", "")

    return envelope.NewSuccess(map[string]interface{}{
        "ca_id":      caID,
        "cert_path":  certRel,
        "chain_path": chainRel,
        "key_ref":    keyRef,
        "serial":     serial,
    })
}
