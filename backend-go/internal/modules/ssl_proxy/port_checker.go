// backend-go/internal/modules/ssl_proxy/port_checker.go
package ssl_proxy

import (
	"fmt"
	"net"
	"time"

	"gorm.io/gorm"
)

// PortChecker 检测监听端口可用性
type PortChecker struct {
	db *gorm.DB
}

func NewPortChecker(db *gorm.DB) *PortChecker {
	return &PortChecker{db: db}
}

// Check 检测 listen_ip:listen_port
//
// excludeServiceID：编辑/重启时排除自身
//
// ★ 检测顺序：
//  1. 数据库配置冲突（与 status 无关，只看端口重叠 + IP 覆盖）；
//  2. 系统绑定（内核层试探）；
//  3. 若内核层拒绝但"占用者就是本实例自身"，视为可用（编辑运行中实例端口未变的场景）。
func (c *PortChecker) Check(req *PortCheckReq) *PortCheckResult {
	res := &PortCheckResult{}

	// 1. 配置冲突（数据库）
	if conflict, ok := c.checkDBConflict(req.ListenIP, req.ListenPort, req.ServiceID); ok {
		res.Available = false
		res.ConflictType = "config"
		res.OccupiedBy = conflict
		res.SuggestedPort = c.suggestPort(req.ListenIP, req.ListenPort)
		res.Message = fmt.Sprintf("端口已被配置 %s 占用", conflict)
		return res
	}

	// 2. 系统绑定
	if err := tryBind(req.ListenIP, req.ListenPort); err != nil {
		// ★ 编辑自身：运行中实例的端口没变 → 视为"自己占自己"，可用
		if c.isSelfOccupied(req.ServiceID, req.ListenPort) {
			res.Available = true
			res.Message = "端口可用（当前由本实例占用）"
			return res
		}
		res.Available = false
		res.ConflictType = "system"
		res.OccupiedBy = "OS process"
		res.SuggestedPort = c.suggestPort(req.ListenIP, req.ListenPort)
		res.Message = fmt.Sprintf("端口被系统进程占用：%v", err)
		return res
	}

	res.Available = true
	res.Message = "端口可用"
	return res
}

// checkDBConflict 返回冲突的 service_id（若有）。
//
// ★ 判定规则：
//   - 只排除自身（excludeSID）；
//   - 不区分实例是否在运行：只要配置里占用同一端口，即视为冲突，
//     避免两个配置端口重叠，后续同时启动时撞端口；
//   - IP 覆盖关系：0.0.0.0 与任意 IP 视为重叠（0.0.0.0 抢占所有网卡）；
//     其余需 IP 完全相等。
func (c *PortChecker) checkDBConflict(ip string, port int, excludeSID string) (string, bool) {
	q := c.db.Model(&PlatformSSLProxy{}).
		Where("listen_port = ?", port)
	if excludeSID != "" {
		q = q.Where("service_id <> ?", excludeSID)
	}

	var rows []PlatformSSLProxy
	if err := q.Find(&rows).Error; err != nil {
		return "", false
	}
	for _, r := range rows {
		if ipOverlap(ip, r.ListenIP) {
			return r.ServiceID, true
		}
	}
	return "", false
}

// ipOverlap 判断两个监听 IP 是否可能产生内核层绑定冲突。
//
//   - 完全相同：冲突；
//   - 任一方为 0.0.0.0：冲突（0.0.0.0 会绑定所有网卡）；
//   - 其他情况（如 127.0.0.1 vs 192.168.1.5）：视为不冲突。
func ipOverlap(a, b string) bool {
	if a == b {
		return true
	}
	if a == "0.0.0.0" || b == "0.0.0.0" {
		return true
	}
	return false
}

// isSelfOccupied 判断"待检测端口"是否正被指定实例自身占用。
//
// 判定条件：
//   - serviceID 非空；
//   - 实例存在且状态为运行中（STARTING / RUNNING）；
//   - 实例当前监听的端口与待检测端口相同。
//
// 满足即视为"端口未变、自己占用自己"，避免编辑运行中实例时误报冲突。
// 不改端口只改其它字段，或只改 IP 不改端口，都走这一分支。
func (c *PortChecker) isSelfOccupied(serviceID string, port int) bool {
	if serviceID == "" {
		return false
	}
	var self PlatformSSLProxy
	if err := c.db.Where("service_id = ?", serviceID).First(&self).Error; err != nil {
		return false
	}
	if !self.IsRunning() {
		return false
	}
	return self.ListenPort == port
}

// tryBind 尝试绑定端口并立即释放
func tryBind(ip string, port int) error {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	_ = l.Close()
	// 等一小段时间确保 TIME_WAIT 不阻塞
	time.Sleep(10 * time.Millisecond)
	return nil
}

// suggestPort 从 port+1 开始找可用端口
func (c *PortChecker) suggestPort(ip string, port int) int {
	for i := 1; i <= 1000; i++ {
		cand := port + i
		if cand > 65535 {
			break
		}
		if _, ok := c.checkDBConflict(ip, cand, ""); ok {
			continue
		}
		if err := tryBind(ip, cand); err == nil {
			return cand
		}
	}
	return 0
}
