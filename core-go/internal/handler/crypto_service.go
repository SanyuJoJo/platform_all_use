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

// CryptoServiceHandler crypto.service。
type CryptoServiceHandler struct{}

func (h *CryptoServiceHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    action := strParam(req.Params, "action")
    switch action {
    case "sign", "verify", "encrypt", "decrypt", "hmac", "kdf":
    default:
        return genErr("INVALID_PARAM", "invalid action")
    }

    keyRef := strParam(req.Params, "key_ref")
    inputRel := strParam(req.Params, "input_path")
    if keyRef == "" || inputRel == "" {
        return genErr("INVALID_PARAM", "missing key_ref or input_path")
    }

    inputAbs, err := pathguard.Resolve(inputRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }

    outputRel := strParamDefault(req.Params, "output_path",
        "data/crypto/"+newID("crypto")+".out")
    outputAbs, _ := pathguard.Resolve(outputRel)
    _ = pathguard.EnsureDir(filepath.Dir(outputAbs), 0750)

    tmpDir, err := mkTmpDir("crypto-svc")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    keyAbs, err := loadDecryptedKey(keyRef, tmpDir)
    if err != nil {
        return genErr("KEY_NOT_FOUND", err.Error())
    }

    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    switch action {
    case "sign":
        if _, err := client.Run(ctx, "dgst", "-sign", keyAbs,
            "-out", outputAbs, inputAbs); err != nil {
            return genErr("CORE_EXEC_FAILED", err.Error())
        }
    case "encrypt":
        if _, err := client.Run(ctx, "pkeyutl", "-encrypt",
            "-inkey", keyAbs, "-in", inputAbs, "-out", outputAbs); err != nil {
            return genErr("CORE_EXEC_FAILED", err.Error())
        }
    case "decrypt":
        if _, err := client.Run(ctx, "pkeyutl", "-decrypt",
            "-inkey", keyAbs, "-in", inputAbs, "-out", outputAbs); err != nil {
            return genErr("CORE_EXEC_FAILED", err.Error())
        }
    case "hmac":
        secretRel := strParam(req.Params, "secret_file")
        secretAbs, err := pathguard.Resolve(secretRel)
        if err != nil {
            return genErr("INVALID_PARAM", "secret_file required")
        }
        secret, _ := os.ReadFile(secretAbs)
        if _, err := client.Run(ctx, "dgst", "-hmac", string(secret),
            "-out", outputAbs, inputAbs); err != nil {
            return genErr("CORE_EXEC_FAILED", err.Error())
        }
    case "verify", "kdf":
        return genErr("INTERNAL_ERROR", fmt.Sprintf("%s not implemented", action))
    }

    _ = os.Chmod(outputAbs, 0600)
    audit.Log(req, "crypto.service", start, "SUCCESS", "")

    return envelope.NewSuccess(map[string]interface{}{
        "output_path": outputRel,
        "action":      action,
    })
}
