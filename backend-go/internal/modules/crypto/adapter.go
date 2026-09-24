package crypto
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"github.com/rs/zerolog/log"
)
// CoreAdapter 负责以参数列表方式调用 core/sbin/dispatch.sh。
//
// 安全要求（C-08）：
//   - 禁止 shell=True、eval、sh -c、字符串拼接
//   - 必须使用 argv 参数列表调用
//   - 用户输入不得进入 shell 解析
type CoreAdapter struct {
	DispatchPath string // core/sbin/dispatch.sh 绝对路径
	TimeoutMs    int    // 默认超时
	MaxConcurr   int    // 最大并发（由 Service 层控制信号量）
}
// NewCoreAdapter 创建 Core Adapter。
func NewCoreAdapter(dispatchPath string, timeoutMs, maxConcurr int) *CoreAdapter {
	if timeoutMs <= 0 {
		timeoutMs = DefaultTimeoutMs
	}
	if maxConcurr <= 0 {
		maxConcurr = DefaultMaxConcurrency
	}
	return &CoreAdapter{
		DispatchPath: dispatchPath,
		TimeoutMs:    timeoutMs,
		MaxConcurr:   maxConcurr,
	}
}
// Call 以参数列表方式调用 dispatch.sh。
//
// 参数：
//   - ctx：上下文，用于超时控制
//   - op：operation_id
//   - req：core 请求 JSON 结构
//
// 返回：
//   - CoreResponse：core 响应
//   - exitCode：core 退出码
//   - err：仅当无法启动进程、JSON 解析失败等平台级错误时返回
func (a *CoreAdapter) Call(ctx context.Context, op string, req *CoreRequest) (*CoreResponse, int, error) {
	// 1. 创建临时文件
	tmpDir, err := os.MkdirTemp("", "crypto-call-")
	if err != nil {
		return nil, -1, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	inPath := filepath.Join(tmpDir, "req.json")
	outPath := filepath.Join(tmpDir, "resp.json")
	// 2. 序列化请求 JSON 并写入 --in 文件
	reqBytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, -1, fmt.Errorf("序列化请求失败: %w", err)
	}
	if err := os.WriteFile(inPath, reqBytes, 0600); err != nil {
		return nil, -1, fmt.Errorf("写入 --in 文件失败: %w", err)
	}
	// 3. 超时控制
	timeoutMs := a.TimeoutMs
	if req.Options != nil && req.Options.TimeoutMs != nil && *req.Options.TimeoutMs > 0 {
		timeoutMs = *req.Options.TimeoutMs
	}
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()
	// 4. 参数列表调用 dispatch.sh（禁止 shell 字符串拼接）
	cmd := exec.CommandContext(
		callCtx,
		a.DispatchPath,
		"--op", op,
		"--in", inPath,
		"--out", outPath,
	)
	// 捕获 stderr 用于日志
	var stderrBuf []byte
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, -1, fmt.Errorf("创建 stderr 管道失败: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, -1, fmt.Errorf("启动 dispatch.sh 失败: %w", err)
	}
	// 读取 stderr（JSON Lines 日志）
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				stderrBuf = append(stderrBuf, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
	}()
	// 5. 等待完成
	waitErr := cmd.Wait()
	exitCode := 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// 超时或启动失败
			if callCtx.Err() == context.DeadlineExceeded {
				log.Warn().
					Str("operation_id", op).
					Str("request_id", req.RequestID).
					Msg("core 调用超时")
				return nil, -1, fmt.Errorf("core 调用超时（%dms）", timeoutMs)
			}
			return nil, -1, fmt.Errorf("等待 dispatch.sh 失败: %w", waitErr)
		}
	}
	// 记录 stderr 日志（JSON Lines）
	if len(stderrBuf) > 0 {
		log.Debug().
			Str("operation_id", op).
			Str("request_id", req.RequestID).
			Str("stderr", string(stderrBuf)).
			Msg("core stderr 日志")
	}
	// 6. 读取 --out 文件
	outBytes, err := os.ReadFile(outPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error().
				Str("operation_id", op).
				Str("request_id", req.RequestID).
				Int("exit_code", exitCode).
				Msg("core --out 文件不存在")
			return nil, exitCode, fmt.Errorf("core --out 文件不存在")
		}
		return nil, exitCode, fmt.Errorf("读取 --out 文件失败: %w", err)
	}
	// 7. 解析 JSON
	var resp CoreResponse
	if err := json.Unmarshal(outBytes, &resp); err != nil {
		log.Error().
			Str("operation_id", op).
			Str("request_id", req.RequestID).
			Int("exit_code", exitCode).
			Str("raw", string(outBytes)).
			Msg("core 响应 JSON 非法")
		return nil, exitCode, fmt.Errorf("core 响应 JSON 非法: %w", err)
	}
	return &resp, exitCode, nil
}
// HealthCheck 检查 dispatch.sh 是否存在且可执行。
func (a *CoreAdapter) HealthCheck() error {
	info, err := os.Stat(a.DispatchPath)
	if err != nil {
		return fmt.Errorf("dispatch.sh 不存在: %w", err)
	}
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("dispatch.sh 不可执行: %s", a.DispatchPath)
	}
	return nil
}
// VersionCheck 调用 version_check.sh 校验铜锁版本。
func (a *CoreAdapter) VersionCheck(ctx context.Context) (string, error) {
	versionScript := filepath.Join(filepath.Dir(a.DispatchPath), "lib", "version_check.sh")
	if _, err := os.Stat(versionScript); err != nil {
		return "", fmt.Errorf("version_check.sh 不存在: %w", err)
	}
	cmd := exec.CommandContext(ctx, versionScript, "--json")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("版本校验失败: %w", err)
	}
	var result struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", fmt.Errorf("版本校验响应解析失败: %w", err)
	}
	if result.Code != "OK" {
		return "", fmt.Errorf("版本校验未通过: %s", result.Message)
	}
	return result.Version, nil
}
// 辅助函数：从环境变量读取整数
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
