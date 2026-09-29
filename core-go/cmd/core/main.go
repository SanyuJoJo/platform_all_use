package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "time"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"

    "github.com/yourorg/core-go/internal/audit"
    "github.com/yourorg/core-go/internal/config"
    "github.com/yourorg/core-go/internal/dispatch"
    "github.com/yourorg/core-go/internal/envelope"
    "github.com/yourorg/core-go/internal/keystore"
    "github.com/yourorg/core-go/internal/openssl"
    "github.com/yourorg/core-go/internal/pathguard"
)

const (
    exitOK               = 0
    exitInvalidParam     = 2
    exitAlgoNotAllowed   = 3
    exitPathNotAllowed   = 4
    exitPermissionDenied = 5
    exitKeyNotFound      = 6
    exitCertNotFound     = 7
    exitCertParseFailed  = 8
    exitCoreExecFailed   = 9
    exitCoreTimeout      = 10
    exitCoreJSONInvalid  = 11
    exitVersionUnsupport = 12
    exitInternalError    = 99
)

func main() {
    var (
        op      = flag.String("op", "", "operation_id (e.g. ca.create)")
        inPath  = flag.String("in", "", "request JSON path")
        outPath = flag.String("out", "", "response JSON path")
    )
    flag.Parse()

    // 日志走 stderr（JSON Lines），stdout 保持纯净
    log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
    zerolog.SetGlobalLevel(zerolog.InfoLevel)

    // 参数必填校验
    if *op == "" || *inPath == "" || *outPath == "" {
        emitFatal(*outPath, "INVALID_PARAM",
            "missing --op/--in/--out", exitInvalidParam)
    }

    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        emitFatal(*outPath, "INTERNAL_ERROR",
            "config load failed: "+err.Error(), exitInternalError)
    }

    // 初始化全局基础设施
    pathguard.Init(cfg.CoreRoot)

    // ★ 修复：将 conf/core.yaml 的 tongsuo_openssl_bin 传给 openssl.Init
    openssl.Init(cfg.CoreRoot, cfg.OpenSSLBin, cfg.TongsuoMinVersion)

    if err := keystore.Init(cfg.MasterKeyPath); err != nil {
        emitFatal(*outPath, "INTERNAL_ERROR",
            "keystore init failed: "+err.Error(), exitInternalError)
    }

    // 审计日志初始化
    audit.Init(cfg.CoreRoot, cfg.AuditLogPath)

    // 版本校验
    if err := openssl.CheckVersion(); err != nil {
        emitFatal(*outPath, "VERSION_UNSUPPORTED", err.Error(), exitVersionUnsupport)
    }

    // 读取请求 JSON
    reqBytes, err := os.ReadFile(*inPath)
    if err != nil {
        emitFatal(*outPath, "INVALID_PARAM",
            "read --in failed: "+err.Error(), exitInvalidParam)
    }

    var req envelope.Request
    if err := json.Unmarshal(reqBytes, &req); err != nil {
        emitFatal(*outPath, "CORE_JSON_INVALID",
            "invalid request json: "+err.Error(), exitCoreJSONInvalid)
    }

    // 执行
    ctx, cancel := context.WithTimeout(
        context.Background(),
        time.Duration(cfg.TimeoutMs)*time.Millisecond,
    )
    defer cancel()

    resp := dispatch.Dispatch(ctx, *op, &req, cfg)
    resp.RequestID = req.RequestID
    resp.OperationID = *op

    // 写响应
    outBytes, err := json.MarshalIndent(resp, "", "  ")
    if err != nil {
        emitFatal(*outPath, "INTERNAL_ERROR",
            "marshal response failed", exitInternalError)
    }
    if err := os.WriteFile(*outPath, outBytes, 0600); err != nil {
        fmt.Fprintf(os.Stderr,
            `{"level":"error","msg":"write --out failed: %s"}`+"\n", err)
        os.Exit(exitInternalError)
    }

    os.Exit(exitCodeOf(resp.Code))
}

// emitFatal 写错误响应并退出。
func emitFatal(outPath, code, msg string, exitCode int) {
    resp := envelope.NewError(code, msg)
    if outPath != "" {
        if b, err := json.MarshalIndent(resp, "", "  "); err == nil {
            _ = os.WriteFile(outPath, b, 0600)
        }
    }
    fmt.Fprintf(os.Stderr,
        `{"level":"error","code":"%s","msg":"%s"}`+"\n", code, msg)
    os.Exit(exitCode)
}

// exitCodeOf 平台错误码 → 退出码。
func exitCodeOf(code string) int {
    switch code {
    case "OK":
        return exitOK
    case "INVALID_PARAM":
        return exitInvalidParam
    case "ALGORITHM_NOT_ALLOWED":
        return exitAlgoNotAllowed
    case "PATH_NOT_ALLOWED":
        return exitPathNotAllowed
    case "PERMISSION_DENIED":
        return exitPermissionDenied
    case "KEY_NOT_FOUND":
        return exitKeyNotFound
    case "CERT_NOT_FOUND":
        return exitCertNotFound
    case "CERT_PARSE_FAILED":
        return exitCertParseFailed
    case "CORE_EXEC_FAILED":
        return exitCoreExecFailed
    case "CORE_TIMEOUT":
        return exitCoreTimeout
    case "CORE_JSON_INVALID":
        return exitCoreJSONInvalid
    case "VERSION_UNSUPPORTED":
        return exitVersionUnsupport
    default:
        return exitInternalError
    }
}
