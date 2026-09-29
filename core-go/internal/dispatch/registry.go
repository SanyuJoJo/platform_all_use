package dispatch

import (
    "context"

    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/handler"
)

// Handler 每个 operation 实现此接口。
type Handler interface {
    Execute(ctx context.Context, req *envelope.Request, cfg *config.Config) *envelope.Response
}

// registry operation_id → Handler。
var registry = map[string]Handler{
    "ca.create":              &handler.CACreateHandler{},
    "ca.intermediate.create": &handler.CAIntermediateHandler{},
    "csr.create":             &handler.CSRCreateHandler{},
    "cert.sign":              &handler.CertSignHandler{},
    "dual_cert.create":       &handler.DualCertCreateHandler{},
    "crl.create":             &handler.CRLCreateHandler{},
    "cert.convert":           &handler.CertConvertHandler{},
    "cert.parse":             &handler.CertParseHandler{},
    "key.manage":             &handler.KeyManageHandler{},
    "pqc.cert.create":        &handler.PQCCertCreateHandler{},
    "chain.verify":           &handler.ChainVerifyHandler{},
    "batch.execute":          &handler.BatchExecuteHandler{},
    "ssl.config.generate":    &handler.SSLConfigGenerateHandler{},
    "crypto.service":         &handler.CryptoServiceHandler{},
}
