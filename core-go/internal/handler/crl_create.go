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

// CRLCreateHandler crl.create。
type CRLCreateHandler struct{}

func (h *CRLCreateHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    start := time.Now()

    caCertRel := strParam(req.Params, "ca_cert_path")
    caKeyRef := strParam(req.Params, "ca_key_ref")
    if caCertRel == "" || caKeyRef == "" {
        return genErr("INVALID_PARAM", "missing ca_cert_path or ca_key_ref")
    }

    digest := strParamDefault(req.Params, "digest_algorithm", "SM3")
    switch digest {
    case "SM3", "SHA256", "SHA384", "SHA512":
    default:
        return genErr("INVALID_PARAM", "invalid digest_algorithm")
    }

    caCertAbs, err := pathguard.Resolve(caCertRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    if _, err := os.Stat(caCertAbs); err != nil {
        return genErr("CERT_NOT_FOUND", "ca cert not found")
    }

    tmpDir, err := mkTmpDir("crl-create")
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }
    defer cleanupTmpDir(tmpDir)

    caKeyAbs, err := loadDecryptedKey(caKeyRef, tmpDir)
    if err != nil {
        return genErr("KEY_NOT_FOUND", err.Error())
    }

    indexFile := filepath.Join(tmpDir, "index.txt")
    _ = os.WriteFile(indexFile, nil, 0600)

    crlNumberFile := filepath.Join(tmpDir, "crlnumber")
    _ = os.WriteFile(crlNumberFile, []byte("1000\n"), 0600)

    count := 0
    if arr, ok := req.Params["revoked_serials"].([]interface{}); ok {
        var sb strings.Builder
        for _, v := range arr {
            if s, ok := v.(string); ok && s != "" {
                sb.WriteString(fmt.Sprintf(
                    "R\t99991231235959Z\t260101000000Z\t%s\tunknown\t/CN=revoked\n", s))
                count++
            }
        }
        _ = os.WriteFile(indexFile, []byte(sb.String()), 0600)
    }

    confFile := filepath.Join(tmpDir, "openssl.cnf")
    conf := fmt.Sprintf(`[ca]
default_ca = CA_default
[CA_default]
database = %s
crlnumber = %s
default_md = %s
default_crl_days = 30
crl_extensions = crl_ext
unique_subject = no
[crl_ext]
authorityKeyIdentifier = keyid:always
`, indexFile, crlNumberFile, digest)
    _ = os.WriteFile(confFile, []byte(conf), 0600)

    crlID := newID("crl")
    crlRel := "data/crl/" + crlID + ".crl.pem"
    crlAbs, _ := pathguard.Resolve(crlRel)
    _ = pathguard.EnsureDir(filepath.Dir(crlAbs), 0750)

    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    if _, err := client.Run(ctx, "ca", "-gencrl", "-config", confFile,
        "-cert", caCertAbs, "-keyfile", caKeyAbs,
        "-out", crlAbs); err != nil {
        return genErr("CORE_EXEC_FAILED", "openssl ca -gencrl: "+err.Error())
    }
    _ = os.Chmod(crlAbs, 0640)

    audit.Log(req, "crl.create", start, "SUCCESS", "")

    return envelope.NewSuccess(map[string]interface{}{
        "crl_id":        crlID,
        "crl_path":      crlRel,
        "revoked_count": count,
    })
}
