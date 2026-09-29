package handler
import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)
// CertConvertHandler cert.convert 格式转换。
type CertConvertHandler struct{}
func (h *CertConvertHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    srcFmt := strings.ToUpper(strParamDefault(req.Params, "source_format", "PEM"))
    dstFmt := strings.ToUpper(strParamDefault(req.Params, "target_format", "PEM"))
    srcRel := strParam(req.Params, "source_path")
    dstRel := strParamDefault(req.Params, "target_path", "")
    if srcRel == "" {
        return genErr("INVALID_PARAM", "missing source_path")
    }
    srcAbs, err := pathguard.Resolve(srcRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if _, err := os.Stat(srcAbs); err != nil {
        return genErr("INVALID_PARAM", "source not found")
    }
    if dstRel == "" {
        id := newID("convert")
        dstRel = "data/convert/" + id
        switch dstFmt {
        case "PEM":
            dstRel += ".pem"
        case "DER":
            dstRel += ".der"
        case "PKCS12":
            dstRel += ".p12"
        }
    }
    dstAbs, err := pathguard.Resolve(dstRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if err := pathguard.EnsureDir(filepath.Dir(dstAbs), 0750); err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    switch {
    case srcFmt == "PEM" && dstFmt == "DER":
        if _, err := client.Run(ctx, "x509", "-in", srcAbs, "-outform", "DER",
            "-out", dstAbs); err != nil {
            return genErr("CERT_PARSE_FAILED", err.Error())
        }
    case srcFmt == "DER" && dstFmt == "PEM":
        if _, err := client.Run(ctx, "x509", "-in", srcAbs, "-inform", "DER",
            "-out", dstAbs); err != nil {
            return genErr("CERT_PARSE_FAILED", err.Error())
        }
    case srcFmt == "PEM" && dstFmt == "PEM":
        data, _ := os.ReadFile(srcAbs)
        if err := os.WriteFile(dstAbs, data, 0640); err != nil {
            return genErr("CORE_EXEC_FAILED", err.Error())
        }
    default:
        return genErr("INVALID_PARAM",
            fmt.Sprintf("unsupported conversion: %s → %s", srcFmt, dstFmt))
    }
    _ = os.Chmod(dstAbs, 0640)
    return envelope.NewSuccess(map[string]interface{}{
        "converted_path": dstRel,
        "format":         dstFmt,
    })
}
