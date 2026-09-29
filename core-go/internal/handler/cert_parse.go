package handler

import (
    "context"
    "strings"

    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)

type CertParseHandler struct{}

func (h *CertParseHandler) Execute(
    ctx context.Context, req *envelope.Request, cfg *config.Config,
) *envelope.Response {
    certRel := strParam(req.Params, "cert_path")
    if certRel == "" {
        return genErr("INVALID_PARAM", "missing cert_path")
    }
    certAbs, err := pathguard.Resolve(certRel)
    if err != nil {
        return genErr("PATH_NOT_ALLOWED", err.Error())
    }
    client, err := openssl.NewClient()
    if err != nil {
        return genErr("CORE_EXEC_FAILED", err.Error())
    }

    out, err := client.Run(ctx, "x509", "-in", certAbs, "-noout",
        "-text", "-nameopt", "RFC2253")
    if err != nil {
        return genErr("CERT_PARSE_FAILED", err.Error())
    }
    text := string(out)

    subject := extractField(text, "Subject:")
    issuer := extractField(text, "Issuer:")
    serial := extractField(text, "Serial Number:")
    notBefore := extractField(text, "Not Before:")
    notAfter := extractField(text, "Not After :")
    pubAlg := extractField(text, "Public Key Algorithm:")
    sigAlg := extractField(text, "Signature Algorithm:")

    fpOut, _ := client.Run(ctx, "x509", "-in", certAbs, "-noout",
        "-fingerprint", "-sha256")
    fp := parseFingerprint(string(fpOut))

    return envelope.NewSuccess(map[string]interface{}{
        "subject":              strings.TrimSpace(subject),
        "issuer":               strings.TrimSpace(issuer),
        "serial":               strings.TrimSpace(serial),
        "not_before":           strings.TrimSpace(notBefore),
        "not_after":            strings.TrimSpace(notAfter),
        "public_key_algorithm": strings.TrimSpace(pubAlg),
        "signature_algorithm":  strings.TrimSpace(sigAlg),
        "fingerprint_sha256":   fp,
    })
}

func extractField(text, prefix string) string {
    lines := strings.Split(text, "\n")
    for i, line := range lines {
        t := strings.TrimSpace(line)
        if strings.HasPrefix(t, prefix) {
            v := strings.TrimSpace(strings.TrimPrefix(t, prefix))
            if v != "" {
                return v
            }
            if i+1 < len(lines) {
                return strings.TrimSpace(lines[i+1])
            }
        }
    }
    return ""
}

func parseFingerprint(s string) string {
    s = strings.TrimSpace(s)
    if i := strings.Index(s, "="); i >= 0 {
        s = s[i+1:]
    }
    s = strings.ReplaceAll(s, ":", "")
    return strings.TrimSpace(s)
}
