package openssl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Client 铜锁 openssl 子进程调用封装。
//
// 安全要点：
//   - 严格 argv 参数列表，禁止 shell 拼接
//   - 显式 cmd.Stdin = nil，防止交互式阻塞
//   - 超时由 context 控制
//
// ★ oqs-provider 自动 fallback：
//   - 每条命令先按**默认方式**执行（走 openssl.cnf 或 default provider 自动加载）；
//   - 如果 CORE_OQS_ENABLED=1 且默认执行失败，**自动重试**并注入：
//       -provider <oqs-provider-name> -provider default
//   - 两次都失败则返回**第一次**的原始错误（更贴近用户意图）；
//   - 非 PQC 命令（SM2/RSA/ECC/ML-DSA/SLH-DSA/SM3）一次成功，不触发重试。
//
// ⚠️ Provider 名（重要）：
//   - 官方 oqs-provider 的模块名是 `oqsprovider`（全小写，无连字符），
//     对应 oqsprovider.so / oqsprovider.dylib / oqsprovider.dll。
//   - 部分自编译版本可能命名为 `oqs`，可通过环境变量覆盖：
//       CORE_OQS_PROVIDER_NAME=oqs
//   - 若注入的 -provider 名与 .so 注册名不符，openssl 会报
//     "unable to load provider ..." 或静默忽略（导致算法找不到）。
type Client struct {
	bin string
}

// NewClient 创建 Client。
func NewClient() (*Client, error) {
	bin, err := Resolve()
	if err != nil {
		return nil, err
	}
	return &Client{bin: bin}, nil
}

// BinPath 返回 openssl 路径。
func (c *Client) BinPath() string { return c.bin }

// Run 执行 openssl 命令。
func (c *Client) Run(ctx context.Context, args ...string) ([]byte, error) {
	return c.run(ctx, nil, args...)
}

// RunWithStdin 带 stdin 执行（用于 -passin stdin 场景）。
func (c *Client) RunWithStdin(ctx context.Context, stdin []byte, args ...string) ([]byte, error) {
	return c.run(ctx, stdin, args...)
}

// =============================================================================
// oqs-provider fallback
// =============================================================================

// oqsEnabled 判断是否启用 oqs-provider fallback。
//
// 由 config.Load 在启动期根据 core.yaml 的 oqs_provider.enabled 写入环境变量，
// 保持 openssl 包对 config 包零依赖（避免循环依赖）。
func oqsEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("CORE_OQS_ENABLED")), "1")
}

// oqsProviderName 返回 oqs-provider 的注册名。
//
// 默认 `oqsprovider`（官方发行版命名）；
// 可通过环境变量 CORE_OQS_PROVIDER_NAME 覆盖（兼容自编译版本）。
func oqsProviderName() string {
	if v := strings.TrimSpace(os.Getenv("CORE_OQS_PROVIDER_NAME")); v != "" {
		return v
	}
	return "oqsprovider"
}

// injectProviders 在子命令之后插入 `-provider <name>` 序列。
//
// 前提：args 非空且 args[0] 为 openssl 子命令（genpkey / x509 / req / ...）。
//
//	args     = ["genpkey", "-algorithm", "falcon1024", "-out", "k.pem"]
//	providers= ["oqsprovider", "default"]
//	→ ["genpkey", "-provider", "oqsprovider", "-provider", "default",
//	   "-algorithm", "falcon1024", "-out", "k.pem"]
func injectProviders(args, providers []string) []string {
	if len(args) == 0 || len(providers) == 0 {
		return args
	}
	out := make([]string, 0, len(args)+len(providers)*2)
	out = append(out, args[0]) // 子命令
	for _, p := range providers {
		out = append(out, "-provider", p)
	}
	out = append(out, args[1:]...)
	return out
}

// =============================================================================
// 命令执行
// =============================================================================

// run 执行 openssl 命令。
//
// 流程：
//  1. 按 args 原样执行一次（不注入 -provider）；
//  2. 若失败且 CORE_OQS_ENABLED=1，注入 `-provider <oqsprovider> -provider default` 重试；
//  3. 两次都失败时返回**第一次**的错误（保留原始失败原因）。
//
// stdin 在两次尝试中都会重置，因为 bytes.Buffer 一旦被 cmd 消费就不可复用。
func (c *Client) run(ctx context.Context, stdin []byte, args ...string) ([]byte, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// 1. 第一次：默认执行
	out1, err1 := c.execOnce(ctx, stdin, args)
	if err1 == nil {
		return out1, nil
	}

	// 2. oqs fallback（仅启用时）
	if !oqsEnabled() || len(args) == 0 {
		return nil, err1
	}

	fallbackArgs := injectProviders(args, []string{oqsProviderName(), "default"})
	out2, err2 := c.execOnce(ctx, stdin, fallbackArgs)
	if err2 == nil {
		if log.Debug().Enabled() {
			log.Debug().
				Str("openssl_args", strings.Join(args, " ")).
				Str("fallback_args", strings.Join(fallbackArgs, " ")).
				Msg("openssl: oqs fallback succeeded")
		}
		return out2, nil
	}

	// 3. 两次都失败：返回第一次的原始错误
	//    （第二次错误通常只是 "oqs 也找不到"，对用户无额外信息价值）
	if log.Debug().Enabled() {
		log.Debug().
			Str("openssl_args", strings.Join(args, " ")).
			Str("first_err", err1.Error()).
			Str("fallback_err", err2.Error()).
			Msg("openssl: oqs fallback also failed")
	}
	return nil, err1
}

// execOnce 执行一次 openssl 命令（不含 fallback 逻辑）。
func (c *Client) execOnce(ctx context.Context, stdin []byte, args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("openssl: empty args")
	}

	cmd := exec.CommandContext(ctx, c.bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if stdin != nil {
		// ★ 每次调用都用新的 Reader：bytes.Buffer 被消费后不可复用
		cmd.Stdin = bytes.NewReader(stdin)
	} else {
		// 显式置空，防止 openssl 从 tty 读取阻塞
		cmd.Stdin = bytes.NewReader(nil)
	}

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("openssl timeout")
		}
		return nil, fmt.Errorf("openssl %v failed: %w (stderr: %s)",
			args, err, stderr.String())
	}
	return stdout.Bytes(), nil
}
