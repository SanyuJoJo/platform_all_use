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
// Stop（★ 所有返回点都回写 p.Status / p.PID）
// =============================================================================

// Stop 停止实例。
//
// ★ 关键约束：**每次 return 前都必须把 p 的内存状态和 DB 状态对齐**，
//   否则上层（如 Restart）复用 p 指针时会读到过期状态。
func (m *ProcessManager) Stop(p *PlatformSSLProxy) error {
	// 幂等：已停止返回
	if p.Status == StatusStopped {
		if p.PID > 0 {
			m.killProcessGroup(p.PID)
		}
		// ★ 回写：即使 DB 是 STOPPED，也保证内存干净
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
		// ★ 回写
		p.Status = StatusStopped
		p.PID = 0
		return nil
	}

	// -s stop：通知 master 快速停止
	confPath := p.ConfigNginxPath
	if confPath == "" {
		confPath = filepath.Join(confDir, "nginx.conf")
	}
	cmd := exec.Command(m.binPath, "-p", CoreRoot+"/", "-c", confPath, "-s", "stop")
	cmd.Env = m.childEnv()
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Warn().Err(err).Str("output", string(out)).
			Str("service_id", p.ServiceID).
			Msg("ssl_proxy: nginx -s stop 失败（继续等待 + SIGKILL 兜底）")
	}

	// 等待 master 退出：10ms 粒度，200ms 上限（实测 nginx 13ms 退出）
	if !waitExit(p.PID, 200*time.Millisecond) {
		log.Warn().
			Int("pid", p.PID).
			Str("service_id", p.ServiceID).
			Msg("ssl_proxy: nginx -s stop 超时，SIGKILL 兜底")
		m.killProcessGroup(p.PID)
	}

	m.cleanupAfterStop(p, confDir, pidDir)
	m.updateStatus(p.ServiceID, StatusStopped, 0)
	// ★ 正常退出流程的最终回写（关键修复点）
	p.Status = StatusStopped
	p.PID = 0
	return nil
}

// =============================================================================
// ★ Restart（核心修复：Stop 后重新读 DB，确保 Start 拿到最新状态）
// =============================================================================

// Restart 重启实例（停 + 启）。
//
// ★ 修复要点：
//   虽然 Stop 已经回写 p.Status = StatusStopped，但为了让 Restart
//   对"Stop 内部未来可能新增的 early return 分支"保持健壮，这里
//   仍从 DB 重新读取一份 fresh 记录，再传给 Start。
//   这样即使 Stop 有任何漏写，Restart 都不会误判为"已在运行中"。
func (m *ProcessManager) Restart(p *PlatformSSLProxy) error {
	if err := m.Stop(p); err != nil {
		return err
	}

	// ★ 关键：从 DB 重新读取，避免 p 内存状态与 DB 不一致
	var fresh PlatformSSLProxy
	if err := m.db.Where("service_id = ?", p.ServiceID).First(&fresh).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return newErr(ErrNotFound, "重启时实例已不存在", p.ServiceID)
		}
		return newErr(ErrStartFailed, "重启时读取实例状态失败", err.Error())
	}

	// 二次保险：如果 fresh 仍是 RUNNING（Stop 没成功写 DB），主动置为 STOPPED
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

	// 端口释放极快（nginx 13ms 退出），50ms 足够
	time.Sleep(50 * time.Millisecond)

	if err := m.Start(&fresh); err != nil {
		return err
	}

	// ★ 把 fresh 的成功状态回写给调用方的 p（Service 层可能还会用 p）
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
// 进程组杀 / 端口预检 / 内部辅助
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
