package handler

import (
    "bytes"
    "context"
    "os/exec"
    "time"

    "github.com/yourorg/core-go/internal/audit"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)

// ChainVerifyHandler chain.verify 证书链验证。
type ChainVerifyHandler struct{}

func (h *ChainVerifyHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    leafRel := strParam(req.Params, "leaf_path")
    if leafRel == "" {
        return genErr("INVALID_PARAM", "missing leaf_path")
    }
    leafAbs, err := pathguard.Resolve(leafRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }

    caRel := strParamDefault(req.Params, "ca_path", leafRel)
    caAbs, err := pathguard.Resolve(caRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }

    bin, err := openssl.Resolve()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    args := []string{"verify", "-CAfile", caAbs, leafAbs}
    if chainRel := strParam(req.Params, "chain_path"); chainRel != "" {
        chainAbs, _ := pathguard.Resolve(chainRel)
        args = []string{"verify", "-CAfile", caAbs, "-untrusted", chainAbs, leafAbs}
    }

    cmd := exec.CommandContext(ctx, bin, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err = cmd.Run()

    valid := err == nil
    reason := ""
    if !valid {
        reason = stderr.String()
    }

    audit.Log(req, "chain.verify", start, "SUCCESS", "")
    return envelope.NewSuccess(map[string]interface{}{
        "valid":  valid,
        "reason": reason,
    })
}
