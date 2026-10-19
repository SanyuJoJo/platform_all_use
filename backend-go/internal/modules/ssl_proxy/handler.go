// backend-go/internal/modules/ssl_proxy/handler.go
package ssl_proxy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 暴露 REST API
type Handler struct {
	svc *Service
}

// NewHandler 构造 Handler
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册路由（挂载在 protected 分组下，分组前缀应为 /api/v1）
func (h *Handler) Register(r gin.IRouter) {
	g := r.Group("/ssl-proxies")

	// 基础
	g.GET("", h.List)
	g.POST("", h.Create)

	// 静态段必须在 /:id 之前
	g.POST("/check-port", h.CheckPort)
	g.GET("/available-certs", h.AvailableCerts)
	g.POST("/batch/start", h.BatchStart)
	g.POST("/batch/stop", h.BatchStop)

	// 动态段
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/start", h.Start)
	g.POST("/:id/stop", h.Stop)
	g.POST("/:id/restart", h.Restart)
	g.GET("/:id/config", h.GetConfig)
	g.GET("/:id/logs", h.Logs)
}

// ---------- 基础 CRUD ----------

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
		respond(c, err, nil)
		return
	}
	respond(c, nil, gin.H{
		"items": items,
		"total": total,
		"page":  q.Page,
		"size":  q.Size,
	})
}

func (h *Handler) Create(c *gin.Context) {
	var req ProxyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	p, err := h.svc.Create(&req)
	if err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, p)
}

func (h *Handler) Get(c *gin.Context) {
	p, err := h.svc.Get(c.Param("id"))
	if err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, p)
}

func (h *Handler) Update(c *gin.Context) {
	var req ProxyUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	p, err := h.svc.Update(c.Param("id"), &req)
	if err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, p)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, gin.H{"service_id": c.Param("id")})
}

// ---------- 生命周期 ----------

func (h *Handler) Start(c *gin.Context) {
	if err := h.svc.Start(c.Param("id")); err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, gin.H{"service_id": c.Param("id"), "status": StatusRunning})
}

func (h *Handler) Stop(c *gin.Context) {
	if err := h.svc.Stop(c.Param("id")); err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, gin.H{"service_id": c.Param("id"), "status": StatusStopped})
}

func (h *Handler) Restart(c *gin.Context) {
	if err := h.svc.Restart(c.Param("id")); err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, gin.H{"service_id": c.Param("id"), "status": StatusRunning})
}

// ---------- 批量启停 ----------

func (h *Handler) BatchStart(c *gin.Context) {
	var req BatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	resp := h.svc.BatchStart(&req)
	respond(c, nil, resp)
}

func (h *Handler) BatchStop(c *gin.Context) {
	var req BatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	resp := h.svc.BatchStop(&req)
	respond(c, nil, resp)
}

// ---------- 端口检测 ----------

func (h *Handler) CheckPort(c *gin.Context) {
	var req PortCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	res := NewPortChecker(h.svc.db).Check(&req)
	respond(c, nil, res)
}

// ---------- 证书目录查询 ----------

func (h *Handler) AvailableCerts(c *gin.Context) {
	var q AvailableCertsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respond(c, newErr(ErrInvalidParam, "参数错误", err.Error()), nil)
		return
	}
	items, err := h.svc.ListAvailableCerts(q.Purpose, q.Protocol)
	if err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, AvailableCertsResp{Items: items})
}

// ---------- 配置 / 日志 ----------

func (h *Handler) GetConfig(c *gin.Context) {
	kv, nginx, err := h.svc.GetConfig(c.Param("id"))
	if err != nil {
		respond(c, err, nil)
		return
	}
	respond(c, nil, gin.H{"kv": kv, "nginx": nginx})
}

func (h *Handler) Logs(c *gin.Context) {
	respond(c, nil, gin.H{
		"message": "请通过日志聚合组件或文件接口读取",
		"path":    LogRoot + "/" + c.Param("id"),
	})
}

// ---------- 响应封装 ----------

func respond(c *gin.Context, err error, data interface{}) {
	if err == nil {
		c.JSON(http.StatusOK, Resp{Code: 0, Message: "success", Data: data})
		return
	}
	if ae, ok := err.(*APIError); ok {
		c.JSON(httpStatusOf(ae.Code), Resp{
			Code:    int(ae.Code),
			Message: ae.Message,
			Data:    gin.H{"detail": ae.Detail},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Resp{Code: 50000, Message: err.Error()})
}

func httpStatusOf(code ErrCode) int {
	switch code {
	case ErrInvalidParam:
		return http.StatusBadRequest
	case ErrNotFound:
		return http.StatusNotFound
	case ErrPortConflict, ErrRunningConflict, ErrNameConflict: // ★ 加 NameConflict
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
