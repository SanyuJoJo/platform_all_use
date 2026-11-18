package media

import (
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
	"backend-go/internal/security"
)

// Handler 媒体路由。
type Handler struct {
	svc *Service
	cfg *config.Config
}

// NewHandler 创建 Handler。
func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// RegisterRoutes 注册：
//   - protected 下的正常业务路由；
//   - api（未受保护组）下的 /media/:id/download 与 /media/:id/raw
//     （handler 内部手工 token 校验）。
//
// ★ 下载 / 播放直链使用 ?token= 是因为前端 <a href> / <video src>
//   无法附加 Authorization 头。
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup, api *gin.RouterGroup) {
	g := protected.Group("/media")
	{
		g.POST("/upload/init",
			middleware.RequirePermission("media:media:upload"), h.InitUpload)
		g.POST("/upload/chunk",
			middleware.RequirePermission("media:media:upload"), h.UploadChunk)
		g.POST("/upload/complete",
			middleware.RequirePermission("media:media:upload"), h.CompleteUpload)
		g.GET("/list",
			middleware.RequirePermission("media:media:view"), h.List)
		g.GET("/:id",
			middleware.RequirePermission("media:media:view"), h.Detail)
		g.DELETE("/:id",
			middleware.RequirePermission("media:media:delete"), h.Remove)
		g.GET("/:id/play",
			middleware.RequirePermission("media:media:view"), h.Play)
	}

	// 直链接口（download / raw / thumbnail）：既支持 Authorization，也支持 ?token=
	pub := api.Group("/media")
	{
		pub.GET("/:id/download", h.authByHeaderOrQuery(), h.Download)
		pub.GET("/:id/raw", h.authByHeaderOrQuery(), h.Raw)
		pub.GET("/:id/thumbnail", h.authByHeaderOrQuery(), h.Thumbnail)
	}
}

// authByHeaderOrQuery 手工鉴权中间件：
// 优先 Authorization: Bearer，回退 ?token=。
// 校验通过后把用户上下文写入 gin.Context（保持与 AuthMiddleware 一致）。
func (h *Handler) authByHeaderOrQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		if hd := c.GetHeader("Authorization"); strings.HasPrefix(strings.ToLower(hd), "bearer ") {
			token = strings.TrimSpace(hd[7:])
		}
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "未提供认证令牌", nil)
			c.Abort()
			return
		}

		claims, err := security.ParseToken(h.cfg.SecretKey, token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "Token无效或已过期", nil)
			c.Abort()
			return
		}

		uid64, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "无效的Token", nil)
			c.Abort()
			return
		}

		c.Set(middleware.ContextKeyUserID, uint(uid64))
		c.Next()
	}
}

// InitUpload POST /media/upload/init
func (h *Handler) InitUpload(c *gin.Context) {
	var req InitUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败",
			gin.H{"error": err.Error()})
		return
	}
	uid := currentUserID(c)
	resp, err := h.svc.InitUpload(&req, uid)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

// UploadChunk POST /media/upload/chunk (multipart)
func (h *Handler) UploadChunk(c *gin.Context) {
	uploadID := c.PostForm("upload_id")
	if uploadID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "upload_id 必填", nil)
		return
	}

	chunkIndexStr := c.PostForm("chunk_index")
	if chunkIndexStr == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "chunk_index 必填", nil)
		return
	}
	idx, err := strconv.Atoi(chunkIndexStr)
	if err != nil || idx < 0 {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "chunk_index 非法", nil)
		return
	}

	fh, err := c.FormFile("chunk_file")
	if err != nil {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "chunk_file 必填", nil)
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError,
			exception.CodeInternalError, "读取分片失败", nil)
		return
	}
	defer f.Close()

	received, err := h.svc.SaveChunk(uploadID, idx, f, currentUserID(c))
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, gin.H{"received": received}, "success")
}

// CompleteUpload POST /media/upload/complete
func (h *Handler) CompleteUpload(c *gin.Context) {
	var req CompleteUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "参数校验失败",
			gin.H{"error": err.Error()})
		return
	}
	info, err := h.svc.CompleteUpload(req.UploadID, currentUserID(c))
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, &CompleteUploadResp{Media: *info}, "上传成功")
}

// List GET /media/list
func (h *Handler) List(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	pageSize := atoiDefault(c.Query("page_size"), 20)
	resp, err := h.svc.List(&ListQuery{
		Page:     page,
		PageSize: pageSize,
		Type:     c.Query("type"),
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

// Detail GET /media/:id
//
// 注意：Gin 路由树不允许同一位置同时注册静态段（/media/list）与参数段（/media/:id），
// 因此 Detail 兼具 "list 转发" 职责——当 :id == "list" 时转发给 List。
// media.id 为 36 位 UUID，不可能与字面量 "list" 冲突。
func (h *Handler) Detail(c *gin.Context) {
	if c.Param("id") == "list" {
		h.List(c)
		return
	}
	info, err := h.svc.Detail(c.Param("id"))
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, info, "success")
}

// Remove DELETE /media/:id
func (h *Handler) Remove(c *gin.Context) {
	if err := h.svc.Remove(c.Param("id")); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, gin.H{}, "已删除")
}

// Play GET /media/:id/play
func (h *Handler) Play(c *gin.Context) {
	resp, err := h.svc.Play(c.Param("id"))
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, resp, "success")
}

// Download GET /media/:id/download
func (h *Handler) Download(c *gin.Context) {
	m, err := h.svc.GetRaw(c.Param("id"))
	if err != nil {
		handleErr(c, err)
		return
	}
	abs := h.svc.AbsPath(m)
	c.Header("Content-Disposition",
		`attachment; filename="`+safeFilename(m.Name)+`"`)
	c.Header("Content-Type", "application/octet-stream")
	c.File(abs)
}

// Raw GET /media/:id/raw  内联播放
func (h *Handler) Raw(c *gin.Context) {
	m, err := h.svc.GetRaw(c.Param("id"))
	if err != nil {
		handleErr(c, err)
		return
	}
	abs := h.svc.AbsPath(m)
	c.Header("Content-Disposition", "inline")
	ct := "video/" + m.Format
	if m.Type == "audio" {
		ct = "audio/" + m.Format
	}
	c.Header("Content-Type", ct)
	// 支持 Range 请求（拖动进度条 / 断点续传）
	http.ServeFile(c.Writer, c.Request, abs)
}

// Thumbnail GET /media/:id/thumbnail  媒体缩略图
//
// 前端 <img> 无法附加 Authorization 头，因此该直链走 ?token= 手工鉴权。
// 无缩略图或文件缺失时返回 404。
func (h *Handler) Thumbnail(c *gin.Context) {
	m, err := h.svc.GetRaw(c.Param("id"))
	if err != nil {
		handleErr(c, err)
		return
	}
	if m.Thumbnail == "" {
		response.Error(c, http.StatusNotFound,
			exception.CodeNotFound, "该媒体无缩略图", nil)
		return
	}
	abs := filepath.Join(h.cfg.MediaRoot, m.Thumbnail)
	c.Header("Content-Type", "image/jpeg")
	c.File(abs)
}

// ---------------------------------------------------------------------------
// 工具
// ---------------------------------------------------------------------------

func handleErr(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}

func currentUserID(c *gin.Context) uint {
	if v, ok := c.Get(middleware.ContextKeyUserID); ok {
		if uid, ok := v.(uint); ok {
			return uid
		}
	}
	return 0
}

func safeFilename(name string) string {
	replacer := strings.NewReplacer(`"`, `\"`, "\r", "", "\n", "")
	return replacer.Replace(name)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}
