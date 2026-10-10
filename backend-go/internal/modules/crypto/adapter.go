package crypto

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// CoreAdapter 负责以参数列表方式调用 core 二进制。
//
// 历史命名说明：
//   - 字段名 DispatchPath 沿用旧版（core/sbin/dispatch.sh 时代），
//     实际值是 core 二进制的绝对路径（如 /opt/plats_tool/core/bin/core）。
//   - 不再使用任何 shell 脚本。
//
// 安全要求（C-08）：
//   - 禁止 shell=True、eval、sh -c、字符串拼接
//   - 必须使用 argv 参数列表调用
//   - 用户输入不得进入 shell 解析
type CoreAdapter struct {
	DispatchPath string // core 二进制绝对路径（命名遗留）
	TimeoutMs    int    // 默认超时
	MaxConcurr   int    // 最大并发（由 Service 层控制信号量）
}

// NewCoreAdapter 创建 Core Adapter。
//
// 构造时执行 preflight 自检，输出诊断日志：
//   - bin 为空 / 不存在 / 是目录 → Error
//   - CORE_ROOT 未设置 → Warn
func NewCoreAdapter(dispatchPath string, timeoutMs, maxConcurr int) *CoreAdapter {
	if timeoutMs <= 0 {
		timeoutMs = DefaultTimeoutMs
	}
	if maxConcurr <= 0 {
		maxConcurr = DefaultMaxConcurrency
	}
	a := &CoreAdapter{
		DispatchPath: dispatchPath,
		TimeoutMs:    timeoutMs,
		MaxConcurr:   maxConcurr,
	}
	a.preflight()
	return a
}

// preflight 启动期自检，输出诊断日志（不阻断）。
func (a *CoreAdapter) preflight() {
	if strings.TrimSpace(a.DispatchPath) == "" {
		log.Error().
			Msg("CoreAdapter: core 二进制路径为空（检查 PLATS_CORE_GO_BIN / CORE_DISPATCH_PATH）")
	} else if st, err := os.Stat(a.DispatchPath); err != nil {
		log.Error().
			Str("bin", a.DispatchPath).
			Err(err).
			Msg("CoreAdapter: core 二进制不存在或不可访问")
	} else if st.IsDir() {
		log.Error().
			Str("bin", a.DispatchPath).
			Msg("CoreAdapter: core 路径是目录，不是可执行文件")
	} else if st.Mode()&0o111 == 0 {
		log.Error().
			Str("bin", a.DispatchPath).
			Msg("CoreAdapter: core 二进制没有可执行权限")
	} else {
		log.Info().
			Str("bin", a.DispatchPath).
			Int("timeout_ms", a.TimeoutMs).
			Int("max_concurrency", a.MaxConcurr).
			Msg("CoreAdapter: 初始化完成")
	}

	// CORE_ROOT 告警（决定 core 是否回退到系统 openssl）
	if strings.TrimSpace(os.Getenv("CORE_ROOT")) == "" {
		log.Warn().
			Msg("CoreAdapter: CORE_ROOT 未设置，core 可能回退到系统 openssl " +
				"（典型症状：tongsuo version too low: OpenSSL 1.1.1f < 8.5.0）")
	} else {
		root := os.Getenv("CORE_ROOT")
		ossl := filepath.Join(root, "libs", "tongsuo", "bin", "openssl")
		if _, err := os.Stat(ossl); err != nil {
			log.Warn().
				Str("core_root", root).
				Str("expected", ossl).
				Msg("CoreAdapter: CORE_ROOT 下未找到铜锁 openssl")
		} else {
			log.Info().
				Str("openssl", ossl).
				Msg("CoreAdapter: 铜锁 openssl 已就绪")
		}
	}
}

// Call 以参数列表方式调用 core 二进制。
//
// 日志分级：
//   - Debug: 调用开始（op / request_id / params 摘要）、调用结束（耗时）
//   - Warn : 业务失败（非零退出码）、调用超时
//   - Error: 启动失败、响应文件缺失、响应 JSON 非法
//
// 返回：
//   - CoreResponse: core 响应
//   - exitCode    : core 退出码
//   - err         : 平台级错误（启动失败 / 超时 / JSON 解析失败）
func (a *CoreAdapter) Call(
	ctx context.Context, op string, req *CoreRequest,
) (*CoreResponse, int, error) {
	start := time.Now()
	reqID := ""
	if req != nil {
		reqID = req.RequestID
	}

	// ---- 入口 Debug：op / request_id / bin / params 摘要 ----
	if log.Debug().Enabled() {
		paramsKeys := ""
		paramsCount := 0
		if req != nil && req.Params != nil {
			paramsCount = len(req.Params)
			paramsKeys = strings.Join(adapterSortedKeys(req.Params), ",")
		}
		log.Debug().
			Str("op", op).
			Str("request_id", reqID).
			Str("bin", a.DispatchPath).
			Int("params_count", paramsCount).
			Str("params_keys", paramsKeys).
			Msg("core call: start")
	}

	// 1. 创建临时文件
	tmpDir, err := os.MkdirTemp("", "crypto-call-")
	if err != nil {
		log.Error().Str("op", op).Err(err).Msg("CoreAdapter: 创建临时目录失败")
		return nil, -1, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "req.json")
	outPath := filepath.Join(tmpDir, "resp.json")

	// 2. 序列化请求 JSON
	reqBytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Error().Str("op", op).Err(err).Msg("CoreAdapter: 序列化请求失败")
		return nil, -1, fmt.Errorf("序列化请求失败: %w", err)
	}
	if err := os.WriteFile(inPath, reqBytes, 0600); err != nil {
		log.Error().Str("op", op).Err(err).Msg("CoreAdapter: 写入 --in 文件失败")
		return nil, -1, fmt.Errorf("写入 --in 文件失败: %w", err)
	}

	// 3. 超时控制
	timeoutMs := a.TimeoutMs
	if req != nil && req.Options != nil &&
		req.Options.TimeoutMs != nil && *req.Options.TimeoutMs > 0 {
		timeoutMs = *req.Options.TimeoutMs
	}
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// 4. 参数列表调用（禁止 shell 字符串拼接）
	cmd := exec.CommandContext(
		callCtx,
		a.DispatchPath,
		"--op", op,
		"--in", inPath,
		"--out", outPath,
	)

	// stderr 采集：用 bytes.Buffer，避免 goroutine 与 Wait 之间的数据竞争
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	cmd.Stdout = nil
	cmd.Stdin = nil

	// 5. Start（显式区分启动失败）
	if err := cmd.Start(); err != nil {
		elapsed := time.Since(start)

		var execErr *exec.Error
		var pathErr *os.PathError
		switch {
		case errors.As(err, &execErr) || errors.As(err, &pathErr):
			log.Error().
				Str("op", op).
				Str("request_id", reqID).
				Str("bin", a.DispatchPath).
				Dur("elapsed", elapsed).
				Err(err).
				Msg("core call: 启动失败（exec.Error/PathError）")
		default:
			log.Error().
				Str("op", op).
				Str("request_id", reqID).
				Str("bin", a.DispatchPath).
				Dur("elapsed", elapsed).
				Err(err).
				Msg("core call: 启动失败")
		}
		return nil, -1, fmt.Errorf("启动 core 失败（bin=%s）: %w", a.DispatchPath, err)
	}

	// 6. Wait
	waitErr := cmd.Wait()
	elapsed := time.Since(start)
	exitCode := 0

	if waitErr != nil {
		// 6a. 业务失败：进程正常起来但退出码非零
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
			log.Warn().
				Str("op", op).
				Str("request_id", reqID).
				Int("exit_code", exitCode).
				Dur("elapsed", elapsed).
				Str("stderr", strings.TrimSpace(stderrBuf.String())).
				Msg("core call: 非零退出码（业务失败）")
		} else if errors.Is(callCtx.Err(), context.DeadlineExceeded) {
			// 6b. 超时
			log.Warn().
				Str("op", op).
				Str("request_id", reqID).
				Dur("elapsed", elapsed).
				Dur("timeout", time.Duration(timeoutMs)*time.Millisecond).
				Str("stderr", strings.TrimSpace(stderrBuf.String())).
				Msg("core call: 调用超时")
			return nil, -1, fmt.Errorf("core 调用超时（%dms）", timeoutMs)
		} else {
			// 6c. 其它 Wait 错误
			log.Error().
				Str("op", op).
				Str("request_id", reqID).
				Dur("elapsed", elapsed).
				Str("stderr", strings.TrimSpace(stderrBuf.String())).
				Err(waitErr).
				Msg("core call: wait 失败")
			return nil, -1, fmt.Errorf("等待 core 失败: %w", waitErr)
		}
	}

	// Debug：即使成功也输出 stderr（JSON Lines 日志）
	if log.Debug().Enabled() && stderrBuf.Len() > 0 {
		log.Debug().
			Str("op", op).
			Str("request_id", reqID).
			Str("stderr", strings.TrimSpace(stderrBuf.String())).
			Msg("core call: stderr")
	}

	// 7. 读取 --out 文件
	outBytes, err := os.ReadFile(outPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error().
				Str("op", op).
				Str("request_id", reqID).
				Int("exit_code", exitCode).
				Dur("elapsed", elapsed).
				Str("stderr", strings.TrimSpace(stderrBuf.String())).
				Msg("core call: --out 文件不存在")
			return nil, exitCode, fmt.Errorf(
				"core 响应文件不存在（op=%s, exit=%d, stderr=%s）",
				op, exitCode, strings.TrimSpace(stderrBuf.String()),
			)
		}
		log.Error().
			Str("op", op).
			Str("request_id", reqID).
			Int("exit_code", exitCode).
			Err(err).
			Msg("core call: 读取 --out 文件失败")
		return nil, exitCode, fmt.Errorf("读取 --out 文件失败: %w", err)
	}

	// 8. 解析 JSON
	var resp CoreResponse
	if err := json.Unmarshal(outBytes, &resp); err != nil {
		log.Error().
			Str("op", op).
			Str("request_id", reqID).
			Int("exit_code", exitCode).
			Dur("elapsed", elapsed).
			Str("raw", adapterTruncate(string(outBytes), 512)).
			Msg("core call: 响应 JSON 非法")
		return nil, exitCode, fmt.Errorf("core 响应 JSON 非法: %w", err)
	}

	// 成功日志（Debug 级，避免噪音）
	log.Debug().
		Str("op", op).
		Str("request_id", reqID).
		Int("exit_code", exitCode).
		Dur("elapsed", elapsed).
		Str("code", resp.Code).
		Msg("core call: done")

	return &resp, exitCode, nil
}

// HealthCheck 检查 core 二进制是否存在且可执行。
func (a *CoreAdapter) HealthCheck() error {
	if strings.TrimSpace(a.DispatchPath) == "" {
		return fmt.Errorf("core 二进制路径为空")
	}
	info, err := os.Stat(a.DispatchPath)
	if err != nil {
		return fmt.Errorf("core 二进制不存在（%s）: %w", a.DispatchPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("core 路径是目录：%s", a.DispatchPath)
	}
	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("core 二进制不可执行：%s", a.DispatchPath)
	}
	return nil
}

// VersionCheck 校验铜锁版本（已废弃）。
//
// 历史实现依赖 core/sbin/lib/version_check.sh；Go 化后脚本已移除。
// 保留此方法仅为兼容旧调用；新代码不应使用。
func (a *CoreAdapter) VersionCheck(ctx context.Context) (string, error) {
	return "", fmt.Errorf("VersionCheck 已废弃（Go 化后由 core 内部自行校验铜锁版本）")
}

// -----------------------------------------------------------------------------
// 辅助
// -----------------------------------------------------------------------------

// adapterSortedKeys 返回 map 的键（排序，日志稳定）。
func adapterSortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// adapterTruncate 截断字符串（日志用）。
func adapterTruncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}

// getEnvInt 从环境变量读取整数（保留，供其它模块使用）。
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
