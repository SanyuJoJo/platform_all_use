package handler

import (
    "context"
    "os"
    "path/filepath"

    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)

// KeyManageHandler key.manage 密钥管理。
type KeyManageHandler struct{}

func (h *KeyManageHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    action := strParam(req.Params, "action")
    switch action {
    case "generate":
        return h.generate(ctx, req)
    case "import":
        return h.importKey(ctx, req)
    case "export":
        return h.exportKey(ctx, req)
    case "delete":
        return h.deleteKey(req)
    default:
        return genErr("INVALID_PARAM", "invalid action: "+action)
    }
}

func (h *KeyManageHandler) generate(ctx context.Context, req *envelope.Request) *envelope.Response {
    alg := strParamDefault(req.Params, "algorithm", "SM2")
    tmpDir, err := mkTmpDir("keygen")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    keyPath := filepath.Join(tmpDir, "key.pem")
    if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
        return genErr("INVALID_PARAM", err.Error())
    }
    plain, _ := os.ReadFile(keyPath)
    keyRef, err := saveEncryptedKey(plain)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    return envelope.NewSuccess(map[string]interface{}{
        "key_ref":   keyRef,
        "algorithm": alg,
        "state":     "ACTIVE",
    })
}

func (h *KeyManageHandler) importKey(ctx context.Context, req *envelope.Request) *envelope.Response {
    keyRel := strParam(req.Params, "key_path")
    if keyRel == "" {
        return genErr("INVALID_PARAM", "missing key_path")
    }
    abs, err := pathguard.Resolve(keyRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if _, err := os.Stat(abs); err != nil {
        return genErr("KEY_NOT_FOUND", "key file not found")
    }
    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    if _, err := client.Run(ctx, "pkey", "-in", abs, "-noout"); err != nil {
        return genErr("CERT_PARSE_FAILED", "invalid private key")
    }
    plain, _ := os.ReadFile(abs)
    keyRef, err := saveEncryptedKey(plain)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    return envelope.NewSuccess(map[string]interface{}{
        "key_ref": keyRef,
    })
}

func (h *KeyManageHandler) exportKey(ctx context.Context, req *envelope.Request) *envelope.Response {
    keyRef := strParam(req.Params, "key_ref")
    exportRel := strParamDefault(req.Params, "export_path",
        "data/export/"+keyRef+".pem")
    allow := false
    if v, ok := req.Params["allow_plain_export"].(bool); ok {
        allow = v
    }
    if !allow {
        return genErr("PERMISSION_DENIED", "plain export disabled")
    }
    tmpDir, err := mkTmpDir("keyexport")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)
    plainPath, err := loadDecryptedKey(keyRef, tmpDir)
    if err != nil {
        return genErr("KEY_NOT_FOUND", err.Error())
    }
    exportAbs, err := pathguard.Resolve(exportRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if err := pathguard.EnsureDir(filepath.Dir(exportAbs), 0750); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    data, _ := os.ReadFile(plainPath)
    if err := os.WriteFile(exportAbs, data, 0600); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    return envelope.NewSuccess(map[string]interface{}{
        "key_ref":     keyRef,
        "export_path": exportRel,
    })
}

func (h *KeyManageHandler) deleteKey(req *envelope.Request) *envelope.Response {
    keyRef := strParam(req.Params, "key_ref")
    if keyRef == "" {
        return genErr("INVALID_PARAM", "missing key_ref")
    }
    keysDir := pathguard.MustResolve("data/keys")
    src := filepath.Join(keysDir, keyRef+".key.enc")
    dst := src + ".deleted"
    if err := os.Rename(src, dst); err != nil {
        return genErr("KEY_NOT_FOUND", err.Error())
    }
    return envelope.NewSuccess(map[string]interface{}{
        "key_ref": keyRef,
        "state":   "DELETED",
    })
}
