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

// CSRCreateHandler csr.create。
type CSRCreateHandler struct{}

func (h *CSRCreateHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    wl, _ := config.GetWhitelist(cfg.WhitelistPath)
    alg := strParamDefault(req.Params, "algorithm", "SM2")
    if wl != nil && !wl.IsAlgorithmAllowed("csr.create", alg) {
        return genErr("ALGORITHM_NOT_ALLOWED", "algorithm not allowed")
    }

    subject, ok := subjectFromParams(req.Params)
    if !ok {
        return genErr("INVALID_PARAM", "missing subject")
    }

    tmpDir, err := mkTmpDir("csr-create")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    keyPath := filepath.Join(tmpDir, "key.pem")
    if err := genPrivateKey(ctx, alg, req.Params, keyPath); err != nil {
        return genErr("CORE_EXEC_FAILED", "gen key: "+err.Error())
    }

    plainKey, _ := os.ReadFile(keyPath)
    keyRef, err := saveEncryptedKey(plainKey)
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    csrID := newID("csr")
    csrRel := "data/csr/" + csrID + ".csr"
    csrAbs, _ := pathguard.Resolve(csrRel)
    _ = pathguard.EnsureDir(filepath.Dir(csrAbs), 0750)

    // SAN
    var sanLine string
    if raw, ok := req.Params["san"].([]interface{}); ok && len(raw) > 0 {
        parts := make([]string, 0, len(raw))
        for _, v := range raw {
            if s, ok := v.(string); ok && s != "" {
                parts = append(parts, "DNS:"+s)
            }
        }
        if len(parts) > 0 {
            sanLine = "subjectAltName=" + strings.Join(parts, ",") + "\n"
        }
    }

    // 生成扩展配置
    extPath := filepath.Join(tmpDir, "csr.ext")
    var eb strings.Builder
    eb.WriteString("[req]\ndistinguished_name=dn\nreq_extensions=v3_req\nprompt=no\n")
    eb.WriteString("[dn]\n")
    for k, v := range subject {
        eb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
    }
    eb.WriteString("[v3_req]\nbasicConstraints=CA:FALSE\n")
    if sanLine != "" {
        eb.WriteString(sanLine)
    }
    _ = os.WriteFile(extPath, []byte(eb.String()), 0600)

    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    // 注意：subjArg 由 -config extPath 隐式提供，不需要显式 -subj
    if _, err := client.Run(ctx, "req", "-new", "-key", keyPath,
        "-out", csrAbs, "-config", extPath); err != nil {
        return genErr("CORE_EXEC_FAILED", "csr: "+err.Error())
    }
    _ = os.Chmod(csrAbs, 0640)

    audit.Log(req, "csr.create", start, "SUCCESS", "")
    return envelope.NewSuccess(map[string]interface{}{
        "csr_id":   csrID,
        "csr_path": csrRel,
        "key_ref":  keyRef,
    })
}
