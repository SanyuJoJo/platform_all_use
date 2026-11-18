package live

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// Handler 直播路由。
type Handler struct {
	svc    *Service
	engine *EngineManager
}

// NewHandler 创建 Handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, engine: svc.Engine()}
}

// RegisterRoutes 注册路由。
//
// 新增直播间管理：
//   POST   /live/rooms       创建（权限：media:live:push）
//   GET    /live/rooms       列表（权限：media:live:play）
//   DELETE /live/rooms/:id   删除（权限：media:live:stop）
//
// 老接口保留兼容：
//   /live/push/*、/live/proxy/*、/live/streams、/live/play/:streamId、/live/engine/*
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/live")
	{
		// ★ 直播间
		g.POST("/rooms",
			middleware.RequirePermission("media:live:push"), h.CreateRoom)
		g.GET("/rooms",
			middleware.RequirePermission("media:live:play"), h.ListRooms)
		g.DELETE("/rooms/:id",
			middleware.RequirePermission("media:live:stop"), h.DeleteRoom)

		// 老接口（保留）
		g.POST("/push/start",
			middleware.RequirePermission("media:live:push"), h.StartPush)
		g.POST("/push/stop",
			middleware.RequirePermission("media:live:stop"), h.StopPush)
		g.POST("/push/webrtc",
			middleware.RequirePermission("media:live:push"), h.PushWebRTC)

		// 拉流代理：保留旧接口（add/remove），同时兼容 RESTful 风格（start/DELETE）
		// 前端统一使用 POST /proxy/start 与 DELETE /proxy/:streamId。
		g.POST("/proxy/add",
			middleware.RequirePermission("media:live:push"), h.AddStreamProxy)
		g.POST("/proxy/remove",
			middleware.RequirePermission("media:live:stop"), h.RemoveStreamProxy)
		g.POST("/proxy/start",
			middleware.RequirePermission("media:live:push"), h.AddStreamProxy)
		g.DELETE("/proxy/:streamId",
			middleware.RequirePermission("media:live:stop"), h.RemoveStreamProxy)

		g.GET("/streams",
			middleware.RequirePermission("media:live:play"), h.ListStreams)
		g.GET("/play/:streamId",
			middleware.RequirePermission("media:live:play"), h.Play)

		g.GET("/engine/status",
			middleware.RequirePermission("media:live:play"), h.EngineStatus)
		g.POST("/engine/start",
			middleware.RequirePermission("media:live:push"), h.EngineStart)
		g.POST("/engine/stop",
			middleware.RequirePermission("media:live:stop"), h.EngineStop)
	}
}

// ---------------------------------------------------------------------------
// 直播间 handlers
// ---------------------------------------------------------------------------

// CreateRoom POST /live/rooms
func (h *Handler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败", gin.H{"error": err.Error()})
		return
	}
	uid := currentUserID(c)
	room, err := h.svc.CreateRoom(uid, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, room, "直播间已创建")
}

// ListRooms GET /live/rooms
func (h *Handler) ListRooms(c *gin.Context) {
	resp, err := h.svc.ListRooms()
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

// DeleteRoom DELETE /live/rooms/:id
func (h *Handler) DeleteRoom(c *gin.Context) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "id 无效", nil)
		return
	}
	if err := h.svc.DeleteRoom(currentUserID(c), uint(id)); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, gin.H{}, "已删除")
}

// ---------------------------------------------------------------------------
// 老接口 handlers（保留）
// ---------------------------------------------------------------------------

func (h *Handler) StartPush(c *gin.Context) {
	var req PushStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败", gin.H{"error": err.Error()})
		return
	}
	sess, err := h.svc.StartPush(&req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, sess, "推流会话已创建")
}

func (h *Handler) StopPush(c *gin.Context) {
	var req PushStopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败", gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.StopPush(req.StreamID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, gin.H{}, "已停止")
}

func (h *Handler) PushWebRTC(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	var req PushWebRTCReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败", gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.PushWebRTC(req.StreamID, req.SDP)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

func (h *Handler) AddStreamProxy(c *gin.Context) {
	var req AddStreamProxyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败", gin.H{"error": err.Error()})
		return
	}
	resp, err := h.svc.AddStreamProxy(&req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "拉流代理已创建")
}

func (h *Handler) RemoveStreamProxy(c *gin.Context) {
	// 优先读取 RESTful 路径参数 DELETE /live/proxy/:streamId
	streamID := c.Param("streamId")
	if streamID == "" {
		if c.Request.ContentLength > 0 {
			var req RemoveStreamProxyReq
			if err := c.ShouldBindJSON(&req); err == nil {
				streamID = req.StreamID
			}
		}
		if streamID == "" {
			streamID = c.Query("stream_id")
		}
	}
	if streamID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "stream_id 必填", nil)
		return
	}
	if err := h.svc.RemoveStreamProxy(streamID); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, gin.H{}, "已删除")
}

func (h *Handler) ListStreams(c *gin.Context) {
	resp, err := h.svc.ListStreams()
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

func (h *Handler) Play(c *gin.Context) {
	format := c.DefaultQuery("format", "hls")
	resp, err := h.svc.GetPlayURL(c.Param("streamId"), format)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

// ---------------------------------------------------------------------------
// 引擎管理 handlers（保留）
// ---------------------------------------------------------------------------

func (h *Handler) EngineStatus(c *gin.Context) {
	if h.engine == nil {
		response.Success(c, &EngineStatus{Running: false}, "success")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	response.Success(c, h.engine.Status(ctx), "success")
}

func (h *Handler) EngineStart(c *gin.Context) {
	if h.engine == nil {
		response.Error(c, http.StatusServiceUnavailable,
			exception.CodeInternalError, "引擎管理器未初始化", nil)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	if err := h.engine.Start(ctx); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, h.engine.Status(ctx), "引擎启动中")
}

func (h *Handler) EngineStop(c *gin.Context) {
	if h.engine == nil {
		response.Error(c, http.StatusServiceUnavailable,
			exception.CodeInternalError, "引擎管理器未初始化", nil)
		return
	}
	if err := h.engine.Stop(); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, gin.H{"running": false}, "引擎已停止")
}

// ---------------------------------------------------------------------------
// 工具
// ---------------------------------------------------------------------------

func currentUserID(c *gin.Context) uint {
	if v, ok := c.Get(middleware.ContextKeyUserID); ok {
		if uid, ok := v.(uint); ok {
			return uid
		}
	}
	return 0
}

func handleErr(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}
