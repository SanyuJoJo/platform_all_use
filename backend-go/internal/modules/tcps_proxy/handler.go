// backend-go/internal/modules/tcps_proxy/handler.go
package tcps_proxy

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go/internal/response"
)

// Handler TCPS 代理 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造 TCPS 代理处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册 TCPS 代理路由（挂在受保护的 group 下）。
func (h *Handler) Register(group *gin.RouterGroup) {
	g := group.Group("/tcps-proxies")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	g.POST("/:id/start", h.Start)
	g.POST("/:id/stop", h.Stop)
	g.POST("/:id/restart", h.Restart)

	g.POST("/check-port", h.CheckPort)
	g.GET("/:id/config", h.GetConfig)
	g.GET("/:id/logs", h.GetLogs)

	g.POST("/batch/start", h.BatchStart)
	g.POST("/batch/stop", h.BatchStop)

	g.GET("/available-certs", h.ListAvailableCerts)
}

// =============================================================================
// 统一响应
// =============================================================================

func respond(c *gin.Context, apiErr *APIError, data interface{}) {
	if apiErr != nil {
		response.Error(c, httpStatus(apiErr.Code), int(apiErr.Code), apiErr.Message, nil)
		return
	}
	response.Success(c, data, "success")
}

func httpStatus(code ErrCode) int {
	switch code {
	case ErrInvalidParam:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrPortConflict, ErrRunningConflict, ErrNameConflict:
		return http.StatusConflict
	case ErrPermission:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func toAPIError(err error) *APIError {
	if ae, ok := err.(*APIError); ok {
		return ae
	}
	return newErr(50000, err.Error(), "")
}

// =============================================================================
// CRUD Handlers
// =============================================================================

// List 列表。
func (h *Handler) List(c *gin.Context) {
	var q ProxyListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 200 {
		q.Size = 20
	}
	items, total, err := h.svc.List(&q)
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, gin.H{
		"items": items,
		"total": total,
		"page":  q.Page,
		"size":  q.Size,
	})
}

// Create 新建。
func (h *Handler) Create(c *gin.Context) {
	var req ProxyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "请求参数错误", err.Error()), nil)
		return
	}
	p, err := h.svc.Create(&req)
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, p)
}

// Get 详情。
func (h *Handler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Param("id"))
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, p)
}

// Update 编辑。
func (h *Handler) Update(c *gin.Context) {
	var req ProxyUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "请求参数错误", err.Error()), nil)
		return
	}
	p, err := h.svc.Update(c.Param("id"), &req)
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, p)
}

// Delete 删除。
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, gin.H{"service_id": c.Param("id")})
}

// =============================================================================
// 生命周期 Handlers
// =============================================================================

// Start 启动。
func (h *Handler) Start(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Start(id); err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	p, _ := h.svc.Get(id)
	respond(c, nil, gin.H{"service_id": id, "status": statusOf(p)})
}

// Stop 停止。
func (h *Handler) Stop(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Stop(id); err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	p, _ := h.svc.Get(id)
	respond(c, nil, gin.H{"service_id": id, "status": statusOf(p)})
}

// Restart 重启。
func (h *Handler) Restart(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Restart(id); err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	p, _ := h.svc.Get(id)
	respond(c, nil, gin.H{"service_id": id, "status": statusOf(p)})
}

// =============================================================================
// 端口 / 配置 / 日志 Handlers
// =============================================================================

// CheckPort 端口检测。
func (h *Handler) CheckPort(c *gin.Context) {
	var req PortCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	res := h.svc.CheckPort(&req)
	respond(c, nil, res)
}

// GetConfig 查看 KV / Nginx 配置。
func (h *Handler) GetConfig(c *gin.Context) {
	kv, nginx, err := h.svc.GetConfig(c.Param("id"))
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, gin.H{"kv": kv, "nginx": nginx})
}

// GetLogs 查看日志（tail N 行）。
func (h *Handler) GetLogs(c *gin.Context) {
	id := c.Param("id")
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "200"))
	if tail <= 0 || tail > 2000 {
		tail = 200
	}
	p, err := h.svc.Get(id)
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	logPath := ""
	if p.LogDir != "" {
		logPath = filepath.Join(p.LogDir, "error.log")
	}
	content, msg := readLastLines(logPath, tail)
	respond(c, nil, gin.H{
		"path":    logPath,
		"content": content,
		"message": msg,
	})
}

// =============================================================================
// 批量 Handlers
// =============================================================================

// BatchStart 批量启动。
func (h *Handler) BatchStart(c *gin.Context) {
	var req BatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	resp := h.svc.BatchStart(&req)
	respond(c, nil, resp)
}

// BatchStop 批量停止。
func (h *Handler) BatchStop(c *gin.Context) {
	var req BatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	resp := h.svc.BatchStop(&req)
	respond(c, nil, resp)
}

// =============================================================================
// 证书 Handlers
// =============================================================================

// ListAvailableCerts 可绑定证书列表。
func (h *Handler) ListAvailableCerts(c *gin.Context) {
	var q AvailableCertsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	items, err := h.svc.ListAvailableCerts(q.Purpose, q.Protocol)
	if err != nil {
		respond(c, toAPIError(err), nil)
		return
	}
	respond(c, nil, gin.H{"items": items})
}

// =============================================================================
// 内部工具
// =============================================================================

func statusOf(p *PlatformTCPSProxy) string {
	if p == nil {
		return StatusStopped
	}
	return p.Status
}

// readLastLines 读取文件末尾 N 行。
func readLastLines(path string, n int) (string, string) {
	if path == "" {
		return "", "日志路径为空（实例可能未启动）"
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "日志文件尚未生成（实例可能未启动）"
		}
		return "", "读取日志失败：" + err.Error()
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return "", "读取日志失败：" + err.Error()
	}
	if st.Size() == 0 {
		return "", "日志文件为空"
	}

	// 简单读最后 n*512 字节（避免大文件全量读）
	maxBytes := int64(n * 512)
	offset := int64(0)
	if st.Size() > maxBytes {
		offset = st.Size() - maxBytes
	}
	if _, err := f.Seek(offset, 0); err != nil {
		return "", "读取日志失败：" + err.Error()
	}
	buf := make([]byte, st.Size()-offset)
	read, _ := f.Read(buf)
	raw := string(buf[:read])
	lines := splitLines(raw)
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return joinLines(lines), ""
}

func splitLines(s string) []string {
	out := make([]string, 0, 16)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	n := 0
	for _, l := range lines {
		n += len(l) + 1
	}
	buf := make([]byte, 0, n)
	for i, l := range lines {
		if i > 0 {
			buf = append(buf, '\n')
		}
		buf = append(buf, l...)
	}
	return string(buf)
}
