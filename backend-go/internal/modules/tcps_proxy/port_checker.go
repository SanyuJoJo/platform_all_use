// backend-go/internal/modules/tcps_proxy/port_checker.go
package tcps_proxy
import (
	"fmt"
	"net"
	"time"
	"gorm.io/gorm"
)
type PortChecker struct{ db *gorm.DB }
func NewPortChecker(db *gorm.DB) *PortChecker { return &PortChecker{db: db} }
func (c *PortChecker) Check(req *PortCheckReq) *PortCheckResult {
	res := &PortCheckResult{}
	if conflict, ok := c.checkDBConflict(req.ListenIP, req.ListenPort, req.ServiceID); ok {
		res.Available = false
		res.ConflictType = "config"
		res.OccupiedBy = conflict
		res.SuggestedPort = c.suggestPort(req.ListenIP, req.ListenPort)
		res.Message = fmt.Sprintf("端口已被配置 %s 占用", conflict)
		return res
	}
	if err := tryBind(req.ListenIP, req.ListenPort); err != nil {
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
func (c *PortChecker) checkDBConflict(ip string, port int, excludeSID string) (string, bool) {
	q := c.db.Model(&PlatformTCPSProxy{}).Where("listen_port = ?", port)
	if excludeSID != "" {
		q = q.Where("service_id <> ?", excludeSID)
	}
	var rows []PlatformTCPSProxy
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
func ipOverlap(a, b string) bool {
	if a == b {
		return true
	}
	if a == "0.0.0.0" || b == "0.0.0.0" {
		return true
	}
	return false
}
func (c *PortChecker) isSelfOccupied(serviceID string, port int) bool {
	if serviceID == "" {
		return false
	}
	var self PlatformTCPSProxy
	if err := c.db.Where("service_id = ?", serviceID).First(&self).Error; err != nil {
		return false
	}
	if !self.IsRunning() {
		return false
	}
	return self.ListenPort == port
}
func tryBind(ip string, port int) error {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	_ = l.Close()
	time.Sleep(10 * time.Millisecond)
	return nil
}
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
