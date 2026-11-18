package live

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"backend-go/internal/config"
	"backend-go/internal/exception"
)

// EngineManager ZLMediaKit 进程管理器。
//
// 职责：
//   - Probe   ：HTTP API 探活
//   - Start   ：托管启动子进程（同步检测秒退 + 注入 LD_LIBRARY_PATH）
//   - Stop    ：SIGTERM → 5s 超时 → SIGKILL
//   - Status  ：running/managed/pid/auto_start + 日志尾
//   - EnsureRunning：启动时探活失败 + ZLM_AUTO_START=true 则托管拉起
//   - Shutdown：后端退出时按 ZLM_ENGINE_SHUTDOWN_ON_EXIT 决定是否 kill
//
// 锁设计：
//   - m.mu      ：保护 client / cmd / managed / startupErr
//   - m.reloadMu：串行化 reloadFromINI，避免并发读 config.ini
//
// Start 里绝不在持有 m.mu 时调用 Probe / reloadFromINI，
// 否则会因 Probe 内部再次拿锁而死锁。
type EngineManager struct {
	cfg *config.Config

	mu         sync.RWMutex
	client     *zlmClient
	cmd        *exec.Cmd
	managed    bool
	startupErr string

	reloadMu sync.Mutex
}

// EngineStatus 引擎状态（对外契约）。
type EngineStatus struct {
	Running        bool   `json:"running"`
	Managed        bool   `json:"managed"`
	PID            int    `json:"pid,omitempty"`
	BaseURL        string `json:"base_url"`
	LastError      string `json:"last_error,omitempty"`
	AutoStart      bool   `json:"auto_start"`
	ShutdownOnExit bool   `json:"shutdown_on_exit"`
	HTTPPort       int    `json:"http_port"`
	LogFile        string `json:"log_file,omitempty"`
	LogTail        string `json:"log_tail,omitempty"`
}

// NewEngineManager 创建引擎管理器。
func NewEngineManager(cfg *config.Config) *EngineManager {
	return &EngineManager{
		cfg:    cfg,
		client: newZLMClient(cfg.ZLMBaseURL, cfg.ZLMSecret),
	}
}

// Client 返回当前 client（★ 每次调用都 reloadFromINI，确保 baseURL/secret 最新）。
//
// 为什么每次都 reload：
//   - ZLM 启动时如果改写了 config.ini 的 secret，前端能拿到新值；
//   - 运维手工改 config.ini 后无需重启后端；
//   - reloadFromINI 内部有 m.reloadMu 保护，开销极低（读几 KB 文本）。
//
// ★ 这是修复 ListRooms 拿不到流的关键：
//   之前 Client() 直接读缓存，如果 ZLM 启动后 secret 变了，
//   ListRooms 里调 GetMediaList 会因 secret 错误返回空。
func (m *EngineManager) Client() *zlmClient {
	m.reloadFromINI() // ★ 关键修复
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// reloadFromINI 从 config.ini 同步端口与 secret 到 cfg 与 client。
//
// 关键约束：
//   - 用 m.reloadMu 串行化，避免两次并发 reload 读到不同 ini 状态
//   - 内层 m.mu.Lock 只保护 client 与 cfg 的原子更新，不可嵌套调用其他持锁函数
//   - 无变化直接返回，不拿锁（性能优化）
func (m *EngineManager) reloadFromINI() {
	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	if m.cfg.ZLMEngineConfig == "" {
		return
	}
	iniCfg, err := config.ReadZLMConfigINI(m.cfg.ZLMEngineConfig)
	if err != nil {
		return
	}

	// 先在无锁状态下计算新值
	m.mu.RLock()
	newPort := m.cfg.ZLMHTTPPort
	newSecret := m.cfg.ZLMSecret
	curBaseURL := m.cfg.ZLMBaseURL
	m.mu.RUnlock()

	changed := false
	if iniCfg.HTTPPort > 0 && iniCfg.HTTPPort != newPort {
		newPort = iniCfg.HTTPPort
		changed = true
	}
	if iniCfg.APISecret != "" && iniCfg.APISecret != newSecret {
		newSecret = iniCfg.APISecret
		changed = true
	}
	if !changed {
		return
	}

	// 派生 baseURL：显式配置尊重显式值，否则跟随端口
	newBaseURL := curBaseURL
	if !config.EnvExplicit("ZLM_BASE_URL") {
		newBaseURL = fmt.Sprintf("http://127.0.0.1:%d", newPort)
	}

	// 原子更新
	m.mu.Lock()
	m.cfg.ZLMHTTPPort = newPort
	m.cfg.ZLMSecret = newSecret
	m.cfg.ZLMBaseURL = newBaseURL
	m.client = newZLMClient(newBaseURL, newSecret)
	m.mu.Unlock()

	log.Info().
		Int("http_port", newPort).
		Str("base_url", newBaseURL).
		Msg("ZLM 配置已从 config.ini 重新加载")
}

// Probe 探活。
func (m *EngineManager) Probe(ctx context.Context) error {
	m.reloadFromINI()
	client := m.Client()

	m.mu.RLock()
	baseURL := m.cfg.ZLMBaseURL
	secret := m.cfg.ZLMSecret
	m.mu.RUnlock()

	u := fmt.Sprintf("%s/index/api/getServerConfig?secret=%s",
		strings.TrimRight(baseURL, "/"), secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := client.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ZLM HTTP 返回 %d", resp.StatusCode)
	}
	return nil
}

// Status 查询当前状态（含日志尾）。
func (m *EngineManager) Status(ctx context.Context) *EngineStatus {
	m.reloadFromINI()

	m.mu.RLock()
	baseURL := m.cfg.ZLMBaseURL
	httpPort := m.cfg.ZLMHTTPPort
	cmd := m.cmd
	startupErr := m.startupErr
	m.mu.RUnlock()

	st := &EngineStatus{
		BaseURL:        baseURL,
		AutoStart:      m.cfg.ZLMAutoStart,
		ShutdownOnExit: m.cfg.ZLMEngineShutdownOnExit,
		HTTPPort:       httpPort,
		LogFile:        m.cfg.ZLMEngineLogFile,
	}

	if err := m.Probe(ctx); err == nil {
		st.Running = true
	} else {
		st.LastError = err.Error()
	}

	if cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Signal(syscall.Signal(0)); err == nil {
			st.Managed = true
			st.PID = cmd.Process.Pid
		}
	}

	if !st.Running && startupErr != "" {
		st.LastError = startupErr
	}
	if !st.Running && st.LogTail == "" {
		if tail := readFileTail(m.cfg.ZLMEngineLogFile, 30); tail != "" {
			st.LogTail = tail
		}
	}
	return st
}

// Start 托管启动 ZLM。
//
// 关键：不在持 m.mu 时调用 Probe / reloadFromINI，否则死锁。
func (m *EngineManager) Start(ctx context.Context) error {
	// 1. 先 reload + 探活（不持锁）
	m.reloadFromINI()
	if err := m.Probe(ctx); err == nil {
		return exception.New(exception.CodeParamInvalid,
			"ZLM 已在运行（外部进程），无需托管启动", 409, nil)
	}

	// 2. 幂等检查（读锁）
	m.mu.RLock()
	existingCmd := m.cmd
	m.mu.RUnlock()

	if existingCmd != nil && existingCmd.Process != nil {
		if err := existingCmd.Process.Signal(syscall.Signal(0)); err == nil {
			return nil
		}
		m.mu.Lock()
		if m.cmd == existingCmd {
			m.cmd = nil
			m.managed = false
		}
		m.mu.Unlock()
	}

	// 3. 前置校验
	bin := strings.TrimSpace(m.cfg.ZLMEngineBin)
	if bin == "" {
		return m.failStart("未配置 ZLM_ENGINE_BIN，无法托管启动")
	}
	if _, err := os.Stat(bin); err != nil {
		return m.failStart(fmt.Sprintf("ZLM 可执行文件不存在：%s（%v）", bin, err))
	}

	logFile := strings.TrimSpace(m.cfg.ZLMEngineLogFile)
	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
			log.Warn().Err(err).Str("dir", filepath.Dir(logFile)).Msg("创建日志目录失败")
			logFile = ""
		}
	}

	args := []string{}
	if cfgPath := strings.TrimSpace(m.cfg.ZLMEngineConfig); cfgPath != "" {
		args = append(args, "-c", cfgPath)
	}

	// 4. 构造 cmd（不涉及锁）
	cmd := exec.CommandContext(ctx, bin, args...)
	env := os.Environ()
	if dir := strings.TrimSpace(m.cfg.ZLMEngineLibDir); dir != "" {
		env = append(env, "LD_LIBRARY_PATH="+dir)
	}
	cmd.Env = env
	if wd := strings.TrimSpace(m.cfg.ZLMEngineWorkDir); wd != "" {
		cmd.Dir = wd
	}

	var logHandle *os.File
	var stderrBuf strings.Builder
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			logHandle = f
			cmd.Stdout = f
			cmd.Stderr = f
		}
	}
	if logHandle == nil {
		cmd.Stdout = &stderrBuf
		cmd.Stderr = &stderrBuf
	}

	// 5. 启动
	startAt := time.Now()
	if err := cmd.Start(); err != nil {
		if logHandle != nil {
			_ = logHandle.Close()
		}
		return m.failStart(fmt.Sprintf("exec 失败：%v", err))
	}
	pid := cmd.Process.Pid
	log.Info().Int("pid", pid).Str("bin", bin).Strs("args", args).Msg("ZLM 子进程已 fork")

	// 6. 同步检测秒退（500ms）
	time.Sleep(500 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		var detail string
		if logFile != "" {
			detail = readFileTail(logFile, 30)
		}
		if detail == "" && stderrBuf.Len() > 0 {
			detail = stderrBuf.String()
		}
		if logHandle != nil {
			_ = logHandle.Close()
		}
		return m.failStart(fmt.Sprintf(
			"ZLM 启动后立即退出（耗时 %v）：\n%s", time.Since(startAt), detail,
		))
	}

	// 7. 成功登记
	m.mu.Lock()
	m.cmd = cmd
	m.managed = true
	m.startupErr = ""
	m.mu.Unlock()

	// 8. 后台回收 + 异步探活
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		if m.cmd == cmd {
			m.cmd = nil
			m.managed = false
		}
		m.mu.Unlock()
		if logHandle != nil {
			_ = logHandle.Close()
		}
		if err != nil {
			log.Warn().Err(err).Int("pid", pid).Msg("ZLM 托管进程退出")
		} else {
			log.Info().Int("pid", pid).Msg("ZLM 托管进程正常退出")
		}
	}()

	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(500 * time.Millisecond)
			if err := m.Probe(context.Background()); err == nil {
				log.Info().Int("pid", pid).Dur("elapsed", time.Since(startAt)).
					Msg("ZLM HTTP API 已就绪")
				return
			}
		}
		tail := readFileTail(logFile, 30)
		log.Error().
			Int("pid", pid).
			Str("log_file", logFile).
			Str("log_tail", tail).
			Msg("ZLM 启动后 10s 内 HTTP API 仍未就绪")
	}()

	return nil
}

// failStart 记录失败原因。
func (m *EngineManager) failStart(msg string) error {
	m.mu.Lock()
	m.startupErr = msg
	m.mu.Unlock()
	log.Error().Str("reason", msg).Msg("ZLM 启动失败")
	return exception.New(exception.CodeInternalError, msg, 500, nil)
}

// Stop 停止本进程托管的 ZLM。
func (m *EngineManager) Stop() error {
	m.mu.RLock()
	cmd := m.cmd
	m.mu.RUnlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	pid := cmd.Process.Pid
	_ = cmd.Process.Signal(syscall.SIGTERM)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			log.Info().Int("pid", pid).Msg("ZLM 托管进程已停止")
			return nil
		}
	}
	log.Warn().Int("pid", pid).Msg("ZLM 未响应 SIGTERM，强制 SIGKILL")
	_ = cmd.Process.Kill()
	return nil
}

// EnsureRunning 启动时调用。
func (m *EngineManager) EnsureRunning(ctx context.Context) {
	if err := m.Probe(ctx); err == nil {
		log.Info().Str("base_url", m.cfg.ZLMBaseURL).Msg("ZLM 已运行（外部）")
		return
	}
	if !m.cfg.ZLMAutoStart {
		log.Warn().Str("base_url", m.cfg.ZLMBaseURL).
			Msg("ZLM 未运行且未开启托管（ZLM_AUTO_START=false）")
		return
	}
	log.Info().Msg("ZLM 未运行，尝试托管启动")
	if err := m.Start(ctx); err != nil {
		log.Error().Err(err).Msg("托管启动 ZLM 失败")
	}
}

// Shutdown 后端退出时调用。
func (m *EngineManager) Shutdown() {
	if !m.cfg.ZLMEngineShutdownOnExit {
		return
	}
	if err := m.Stop(); err != nil {
		log.Warn().Err(err).Msg("关闭 ZLM 失败")
	}
}

// readFileTail 读取文件最后 n 行。
func readFileTail(path string, n int) string {
	if path == "" || n <= 0 {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n*2 {
			lines = lines[len(lines)-n:]
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
