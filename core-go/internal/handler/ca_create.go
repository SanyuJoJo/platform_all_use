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

// CACreateHandler ca.create 根 CA 生成。
type CACreateHandler struct{}

func (h *CACreateHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    wl, err := config.GetWhitelist(cfg.WhitelistPath)
    if err != nil {
        return genErr("INTERNAL_ERROR", "whitelist load failed: "+err.Error())
    }
    alg := strParam(req.Params, "algorithm")
    if !wl.IsAlgorithmAllowed("ca.create", alg) {
        audit.Log(req, "ca.create", start, "FAILED", "ALGORITHM_NOT_ALLOWED")
        return envelope.NewErrorWithDetail("ALGORITHM_NOT_ALLOWED",
            "algorithm not allowed",
            map[string]interface{}{"algorithm": alg})
    }

    subject, ok := subjectFromParams(req.Params)
    if !ok {
        audit.Log(req, "ca.create", start, "FAILED", "INVALID_PARAM")
        return genErr("INVALID_PARAM", "missing subject")
    }
    validityDays := intParam(req.Params, "validity_days", 3650)
    if validityDays <= 0 {
        validityDays = 3650
    }

    tmpDir, err := mkTmpDir("ca-create")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    keyPath := filepath.Join(tmpDir, "key.pem")
    if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
        audit.Log(req, "ca.create", start, "FAILED", "CORE_EXEC_FAILED")
        return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
    }

    plainKey, err := os.ReadFile(keyPath)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    keyRef, err := saveEncryptedKey(plainKey)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", "save key: "+err.Error())
    }

    caID := newID("ca")
    certRel := "data/ca/" + caID + ".pem"
    certAbs, err := pathguard.Resolve(certRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if err := pathguard.EnsureDir(filepath.Dir(certAbs), 0750); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    subjArg := subjectToOpenSSL(subject)
    digest := digestForAlgorithm(alg)

    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    args := []string{
        "req", "-new", "-x509",
        "-key", keyPath,
        "-out", certAbs,
        "-days", fmt.Sprintf("%d", validityDays),
        "-subj", subjArg,
        "-addext", "basicConstraints=critical,CA:TRUE",
        "-addext", "keyUsage=critical,keyCertSign,cRLSign",
    }
    if digest != "" {
        args = append(args, digest)
    }
    if _, err := client.Run(ctx, args...); err != nil {
        audit.Log(req, "ca.create", start, "FAILED", "CORE_EXEC_FAILED")
        return genErr("CORE_EXEC_FAILED", "openssl req: "+err.Error())
    }
    _ = os.Chmod(certAbs, 0640)

    serial := extractSerial(ctx, client, certAbs)
    audit.Log(req, "ca.create", start, "SUCCESS", "")

    return envelope.NewSuccess(map[string]interface{}{
        "ca_id":     caID,
        "cert_path": certRel,
        "key_ref":   keyRef,
        "serial":    serial,
    })
}
