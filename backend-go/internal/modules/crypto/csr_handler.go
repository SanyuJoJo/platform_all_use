package crypto

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// CSRHandler P10 领域 Handler，只注册 /csrs/* 路由。
//
// 路由清单（静态段必须早于 /:csr_id）：
//   GET    /csrs                  列表
//   POST   /csrs/import           导入 P10
//   GET    /csrs/:csr_id          元数据
//   GET    /csrs/:csr_id/download 下载 P10
//   GET    /csrs/:csr_id/key      下载私钥（支持 ?encrypt=true&password=xxx）
//   DELETE /csrs/:csr_id          删除
type CSRHandler struct {
	svc *CSRService
}

// NewCSRHandler 创建 P10 Handler。
func NewCSRHandler(svc *CSRService) *CSRHandler {
	return &CSRHandler{svc: svc}
}

// RegisterRoutes 注册 P10 路由。
//
// 权限点说明：
//   - 查看  ：crypto_console:csr:view
//   - 创建  ：crypto_console:csr:create
//   - 删除  ：crypto_console:csr:delete
//
// 注意：删除权限用独立的 crypto_console:csr:delete，
// 与 auth/seed.go 中的权限定义保持一致，避免与 create 混用。
func (h *CSRHandler) RegisterRoutes(r *gin.RouterGroup) {
	csrs := r.Group("/csrs")

	csrs.GET("",
		middleware.RequirePermission("crypto_console:csr:view"),
		h.List,
	)
	csrs.POST("/import",
		middleware.RequirePermission("crypto_console:csr:create"),
		h.Import,
	)
	csrs.GET("/:csr_id",
		middleware.RequirePermission("crypto_console:csr:view"),
		h.Get,
	)
	csrs.GET("/:csr_id/download",
		middleware.RequirePermission("crypto_console:csr:view"),
		h.Download,
	)
	csrs.GET("/:csr_id/key",
		middleware.RequirePermission("crypto_console:csr:view"),
		h.DownloadKey,
	)
	csrs.DELETE("/:csr_id",
		middleware.RequirePermission("crypto_console:csr:delete"),
		h.Delete,
	)
}

// List P10 列表。
func (h *CSRHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// Get P10 元数据。
func (h *CSRHandler) Get(c *gin.Context) {
	data, err := h.svc.Get(c.Param("csr_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// Import 导入 P10。
func (h *CSRHandler) Import(c *gin.Context) {
	var req ImportCsrRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": err.Error()},
			}})
		return
	}
	data, err := h.svc.Import(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, data, "导入成功")
}

// Download 下载 P10 文件。
func (h *CSRHandler) Download(c *gin.Context) {
	csrID := c.Param("csr_id")
	data, err := h.svc.Download(csrID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Type", "application/pkcs10")
	c.Header("Content-Disposition",
		`attachment; filename="`+csrID+`.csr"`)
	c.Data(http.StatusOK, "application/pkcs10", data)
}

// DownloadKey 下载 P10 私钥，支持口令加密。
func (h *CSRHandler) DownloadKey(c *gin.Context) {
	csrID := c.Param("csr_id")
	encrypt := c.Query("encrypt") == "true"
	password := c.Query("password")

	data, err := h.svc.DownloadKey(csrID, encrypt, password)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Type", "application/x-pem-file")
	c.Header("Content-Disposition",
		`attachment; filename="`+csrID+`.key.pem"`)
	c.Data(http.StatusOK, "application/x-pem-file", data)
}

// Delete 删除 P10。
func (h *CSRHandler) Delete(c *gin.Context) {
	csrID := c.Param("csr_id")
	if csrID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "缺少 csr_id", nil)
		return
	}
	if err := h.svc.Delete(csrID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
