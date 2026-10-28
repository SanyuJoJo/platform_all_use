// backend-go/internal/modules/ssl_proxy/process_manager.go
package ssl_proxy

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type ProcessManager struct {
	db       *gorm.DB
	binPath  string
	gen      *NginxGenerator
	kv       *ConfigGenerator
	resolver CertResolver
}

func NewProcessManager(db *gorm.DB, resolver CertResolver) *ProcessManager {
	return &ProcessManager{
		db:       db,
		binPath:  TengineBin,
		gen:      NewNginxGenerator(resolver),
		kv:       NewConfigGenerator(UserConfRoot),
		resolver: resolver,
	}
}

// =============================================================================
// Start
// =============================================================================

func (m *ProcessManager) Start(p *PlatformSSLProxy) error {
	if p.IsRunning() {
		return newErr(ErrRunningConflict, "实例已在运行中", p.ServiceID)
	}
	if err := m.checkTengine(); err != nil {
		return err
	}
	if err := m.precheckPort(p); err != nil {
		return err
	}

	m.updateStatus(p.ServiceID, StatusStarting, 0)
	p.Status = StatusStarting
	p.PID = 0

	confPath, pidPath, err := m.gen.Render(p)
	if err != nil {
		m.updateStatus(p.ServiceID, StatusFailed, 0)
		p.Status = StatusFailed
		return err
	}

	_ = os.Remove(pidPath)

	if err := m.testConfig(p, confPath); err != nil {
		m.updateStatus(p.ServiceID, StatusFailed, 0)
		p.Status = StatusFailed
		return err
	}

	kvPath, _ := m.kv.WriteKV(p)

	cmd := exec.Command(m.binPath, "-p", CoreRoot+"/", "-c", confPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = m.childEnv()

	if err := cmd.Start(); err != nil {
		m.updateStatus(p.ServiceID, StatusFailed, 0)
		p.Status = StatusFailed
		return newErr(ErrStartFailed, "启动 tengine 失败", err.Error())
	}

	// ★ 修复：异步回收 master 进程，避免 SIGTERM/SIGKILL 后变 zombie
	//
	// 背景：cmd.Start() 后 master 是 backend-go 的子进程。master 退出后
	// （SIGTERM / SIGKILL）会变 zombie，直到 parent 调 cmd.Wait() 回收。
	// 如果不回收，waitExit() 里的 signal(0) 探测会一直返回"活着"，
	// 导致 Stop 时误判超时并 SIGKILL，zombie 越积越多。
	//
	// 这里用一个 goroutine 长期等待，master 退出后自然完成回收。
	go func(serviceID string, c *exec.Cmd) {
		err := c.Wait()
		if err != nil {
			// nginx 被信号杀死时 ExitError 是正常的，用 Debug 级别
			log.Debug().
				Err(err).
				Str("service_id", serviceID).
				Msg("ssl_proxy: nginx master 进程已退出")
		} else {
			log.Debug().
				Str("service_id", serviceID).
				Msg("ssl_proxy: nginx master 进程正常退出")
		}
	}(p.ServiceID, cmd)

	pid, err := waitPID(pidPath, 5*time.Second)
	if err != nil {
		var pidFromFile int
		if data, rerr := os.ReadFile(pidPath); rerr == nil {
			if v, perr := strconv.Atoi(strings.TrimSpace(string(data))); perr == nil {
				pidFromFile = v
			}
		}
		if pidFromFile > 0 {
			m.killProcessGroup(pidFromFile)
		}
		logTail := m.readErrorLogTail(p.ServiceID, 4096)
		m.updateStatus(p.ServiceID, StatusFailed, 0)
		p.Status = StatusFailed
		detail := "pid not found"
		if logTail != "" {
			detail = "nginx error.log:\n" + logTail
		}
		log.Error().
			Str("service_id", p.ServiceID).
			Str("error_log_tail", logTail).
			Msg("ssl_proxy: nginx 启动失败，详见 error.log")
		return newErr(ErrStartFailed, "等待 PID 超时（nginx 启动失败）", detail)
	}

	confDir := filepath.Dir(confPath)
	if err := m.resolver.Seal(p, confDir); err != nil {
		log.Warn().Err(err).Str("service_id", p.ServiceID).
			Msg("ssl_proxy: 启动后私钥加密失败（nginx 已加载，不阻断）")
	}

	p.Status = StatusRunning
	p.PID = pid
	p.ConfigKVPath = kvPath
	p.ConfigNginxPath = confPath
	p.LogDir = filepath.Join(m.gen.LogRoot, p.ServiceID)
	m.db.Save(p)
	return nil
}

// =============================================================================
// Stop（★ 本次核心修复）
// =============================================================================

// Stop 停止实例。
//
// ★ 关键修复（本版）：
//
// 原实现用 `nginx -s stop` 通知 master，但 `nginx -s stop` 会 fork 一个
// 临时客户端进程并「重新解析 nginx.conf」→ 加载证书私钥。而私钥在启动
// 成功后被 Seal 白盒 SM4 加密（不是 PEM 格式），临时进程加载失败：
//
//	nginx: [emerg] cannot load certificate key "...client_enc.key":
//	PEM_read_bio_PrivateKey() failed (DECODER routines::unsupported)
//	exit status 1
//
// 结果：master 收不到信号 → 200ms 超时 → SIGKILL 兜底 → worker 不能优雅退出。
//
// 修复：直接用 syscall.Kill(master_pid, SIGTERM)，跳过 nginx 客户端命令：
//   - 不解析配置
//   - 不加载私钥
//   - 直接把 SIGTERM 送到 master，master 内部通知 worker 优雅退出
//
// 优雅停止语义：
//   - SIGTERM 给 master → master 停止 accept 新连接
//   - master 通知 worker 处理完当前请求后退出
//   - 如果 200ms 内没退完（长连接场景），SIGKILL 整个进程组兜底
func (m *ProcessManager) Stop(p *PlatformSSLProxy) error {
	// 幂等：已停止返回
	if p.Status == StatusStopped {
		if p.PID > 0 {
			m.killProcessGroup(p.PID)
		}
		p.Status = StatusStopped
		p.PID = 0
		return nil
	}

	m.updateStatus(p.ServiceID, StatusStopping, p.PID)
	p.Status = StatusStopping

	confDir := filepath.Join(m.gen.ConfRoot, p.ServiceID)
	pidDir := filepath.Join(m.gen.PidRoot, p.ServiceID)

	// master 已死 → 直接清理
	if p.PID <= 0 || !processAlive(p.PID) {
		if p.PID > 0 {
			m.killProcessGroup(p.PID)
		}
		m.cleanupAfterStop(p, confDir, pidDir)
		m.updateStatus(p.ServiceID, StatusStopped, 0)
		p.Status = StatusStopped
		p.PID = 0
		return nil
	}

	// ★ 直接用 SIGTERM 通知 master
	//
	// 与 `nginx -s stop` 的等价性：
	//   `nginx -s stop` = 读 pid 文件 + kill(pid, SIGTERM)
	// 我们已经有 p.PID，直接发信号即可，跳过读配置/加载私钥。
	if err := syscall.Kill(p.PID, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			// 进程已不存在（可能刚好死了），直接清理
			log.Debug().
				Int("pid", p.PID).
				Str("service_id", p.ServiceID).
				Msg("ssl_proxy: master 已不存在，跳过信号发送")
		} else {
			// 信号发送失败（权限等），仅在此情况回退 nginx -s stop
			log.Warn().Err(err).
				Int("pid", p.PID).
				Str("service_id", p.ServiceID).
				Msg("ssl_proxy: 发送 SIGTERM 失败，回退 nginx -s stop")

			confPath := p.ConfigNginxPath
			if confPath == "" {
				confPath = filepath.Join(confDir, "nginx.conf")
			}
			cmd := exec.Command(m.binPath, "-p", CoreRoot+"/", "-c", confPath, "-s", "stop")
			cmd.Env = m.childEnv()
			if out, e := cmd.CombinedOutput(); e != nil {
				log.Warn().Err(e).
					Str("output", string(out)).
					Str("service_id", p.ServiceID).
					Msg("ssl_proxy: nginx -s stop 也失败，走 SIGKILL 兜底")
			}
		}
	}

	// 等待 master 退出：200ms 上限
	// master 优雅退出通常 10-50ms（实测）；超过说明有长连接未断
	if !waitExit(p.PID, 200*time.Millisecond) {
		log.Warn().
			Int("pid", p.PID).
			Str("service_id", p.ServiceID).
			Msg("ssl_proxy: nginx master 200ms 未退出，SIGKILL 兜底")
		m.killProcessGroup(p.PID)
		// SIGKILL 后再等等，确保进程组真的清理干净
		waitExit(p.PID, 500*time.Millisecond)
	}

	m.cleanupAfterStop(p, confDir, pidDir)
	m.updateStatus(p.ServiceID, StatusStopped, 0)
	p.Status = StatusStopped
	p.PID = 0
	return nil
}

// =============================================================================
// Restart（不变）
// =============================================================================

func (m *ProcessManager) Restart(p *PlatformSSLProxy) error {
	if err := m.Stop(p); err != nil {
		return err
	}

	var fresh PlatformSSLProxy
	if err := m.db.Where("service_id = ?", p.ServiceID).First(&fresh).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return newErr(ErrNotFound, "重启时实例已不存在", p.ServiceID)
		}
		return newErr(ErrStartFailed, "重启时读取实例状态失败", err.Error())
	}

	if fresh.Status == StatusRunning || fresh.Status == StatusStarting {
		log.Warn().
			Str("service_id", p.ServiceID).
			Str("stale_status", fresh.Status).
			Msg("ssl_proxy: Restart 前发现 DB 状态非 STOPPED，强制重置")
		_ = m.db.Model(&PlatformSSLProxy{}).
			Where("service_id = ?", p.ServiceID).
			Updates(map[string]interface{}{"status": StatusStopped, "pid": 0}).Error
		fresh.Status = StatusStopped
		fresh.PID = 0
	}

	time.Sleep(50 * time.Millisecond)

	if err := m.Start(&fresh); err != nil {
		return err
	}
	*p = fresh
	return nil
}

func (m *ProcessManager) Status(p *PlatformSSLProxy) string {
	if p.PID <= 0 {
		return StatusStopped
	}
	if processAlive(p.PID) {
		return StatusRunning
	}
	return StatusStopped
}

// =============================================================================
// 内部辅助（与旧版一致）
// =============================================================================

func (m *ProcessManager) killProcessGroup(masterPid int) {
	if masterPid <= 0 {
		return
	}
	if err := syscall.Kill(-masterPid, syscall.SIGKILL); err != nil {
		if !errors.Is(err, syscall.ESRCH) {
			log.Debug().Int("pgid", masterPid).Err(err).
				Msg("ssl_proxy: kill -KILL 进程组失败")
		}
	}
}

func (m *ProcessManager) precheckPort(p *PlatformSSLProxy) error {
	if err := tryBind(p.ListenIP, p.ListenPort); err == nil {
		return nil
	}
	log.Warn().
		Str("service_id", p.ServiceID).
		Str("listen", p.ListenIP+":"+strconv.Itoa(p.ListenPort)).
		Msg("ssl_proxy: 端口被占，尝试清理 nginx 残留")

	ownerPid := m.findPortOwner(p.ListenPort)
	if ownerPid > 0 && m.isNginxProcess(ownerPid) {
		pgid := getPgid(ownerPid)
		if pgid > 0 {
			log.Info().
				Int("owner_pid", ownerPid).
				Int("pgid", pgid).
				Str("service_id", p.ServiceID).
				Msg("ssl_proxy: 清理占端口的 nginx 进程组")
			m.killProcessGroup(pgid)
			time.Sleep(50 * time.Millisecond)
			if err := tryBind(p.ListenIP, p.ListenPort); err == nil {
				log.Info().
					Str("service_id", p.ServiceID).
					Msg("ssl_proxy: 端口已清理，继续启动")
				return nil
			}
		}
	}

	return newErr(ErrStartFailed,
		"监听端口已被占用，无法启动",
		fmt.Sprintf("listen=%s:%d，请手动排查: ss -tlnp | grep %d",
			p.ListenIP, p.ListenPort, p.ListenPort))
}

func (m *ProcessManager) findPortOwner(port int) int {
	out, err := exec.Command("ss", "-tlnp", fmt.Sprintf("sport = :%d", port)).Output()
	if err != nil {
		out, err = exec.Command("ss", "-tlnp").Output()
		if err != nil {
			return 0
		}
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, fmt.Sprintf(":%d ", port)) {
			continue
		}
		idx := strings.Index(line, "pid=")
		if idx < 0 {
			continue
		}
		rest := line[idx+4:]
		end := strings.IndexAny(rest, ",)")
		if end < 0 {
			continue
		}
		pid, err := strconv.Atoi(rest[:end])
		if err == nil && pid > 0 {
			return pid
		}
	}
	return 0
}

func (m *ProcessManager) isNginxProcess(pid int) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "nginx")
}

func getPgid(pid int) int {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return -1
	}
	fields := strings.Fields(string(data))
	if len(fields) < 5 {
		return -1
	}
	pgid, err := strconv.Atoi(fields[4])
	if err != nil {
		return -1
	}
	return pgid
}

func (m *ProcessManager) cleanupAfterStop(p *PlatformSSLProxy, confDir, pidDir string) {
	_ = m.resolver.Cleanup(p, confDir)
	_ = os.Remove(filepath.Join(pidDir, "nginx.pid"))
}

func (m *ProcessManager) childEnv() []string {
	existing := os.Getenv("LD_LIBRARY_PATH")
	tongsuoLib := CoreRoot + "/libs/tongsuo/libs"
	var ld string
	if existing != "" {
		ld = tongsuoLib + ":" + existing
	} else {
		ld = tongsuoLib
	}
	return append(os.Environ(), "LD_LIBRARY_PATH="+ld)
}

func (m *ProcessManager) checkTengine() error {
	if _, err := os.Stat(m.binPath); err != nil {
		return newErr(ErrTengineMissing, "Tengine 二进制不可用", m.binPath)
	}
	return nil
}

func (m *ProcessManager) testConfig(p *PlatformSSLProxy, confPath string) error {
	cmd := exec.Command(m.binPath, "-p", CoreRoot+"/", "-t", "-c", confPath)
	cmd.Env = m.childEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return newErr(ErrConfigGenerate, "nginx -t 校验失败", string(out))
	}
	return nil
}

func (m *ProcessManager) readErrorLogTail(serviceID string, maxBytes int64) string {
	logPath := filepath.Join(m.gen.LogRoot, serviceID, "error.log")
	f, err := os.Open(logPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return ""
	}
	size := st.Size()
	offset := int64(0)
	if size > maxBytes {
		offset = size - maxBytes
	}
	if _, err := f.Seek(offset, 0); err != nil {
		return ""
	}
	buf := make([]byte, size-offset)
	n, _ := f.Read(buf)
	return strings.TrimSpace(string(buf[:n]))
}

func (m *ProcessManager) updateStatus(sid, status string, pid int) {
	m.db.Model(&PlatformSSLProxy{}).
		Where("service_id = ?", sid).
		Updates(map[string]interface{}{"status": status, "pid": pid})
}

// ---------- 进程工具 ----------

func waitPID(pidPath string, d time.Duration) (int, error) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(pidPath)
		if err == nil {
			pid, perr := strconv.Atoi(strings.TrimSpace(string(b)))
			if perr == nil && pid > 0 && processAlive(pid) {
				return pid, nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, errors.New("pid not found")
}

func waitExit(pid int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}
