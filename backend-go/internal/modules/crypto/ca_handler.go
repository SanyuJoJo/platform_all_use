package crypto
import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// CAHandler 只注册 /cas/* 路由。
type CAHandler struct {
	svc *CAService
}
func NewCAHandler(svc *CAService) *CAHandler { return &CAHandler{svc: svc} }
func (h *CAHandler) RegisterRoutes(r *gin.RouterGroup) {
	cas := r.Group("/cas")
	cas.GET("", middleware.RequirePermission("crypto_console:ca:view"), h.List)
	cas.POST("/import", middleware.RequirePermission("crypto_console:ca:import"), h.Import)
	cas.GET("/:ca_id", middleware.RequirePermission("crypto_console:ca:view"), h.Get)
	cas.GET("/:ca_id/detail", middleware.RequirePermission("crypto_console:ca:view"), h.GetDetail)
	cas.POST("/:ca_id/export", middleware.RequirePermission("crypto_console:ca:export"), h.Export)
	cas.GET("/:ca_id/download", middleware.RequirePermission("crypto_console:ca:view"), h.Download) // 预留
	cas.DELETE("/:ca_id", middleware.RequirePermission("crypto_console:ca:delete"), h.Delete)
}
func (h *CAHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size)
	if err != nil { handleError(c, err); return }
	response.Success(c, data, "success")
}
func (h *CAHandler) Get(c *gin.Context) {
	data, err := h.svc.Get(c.Param("ca_id"))
	if err != nil { handleError(c, err); return }
	response.Success(c, data, "success")
}
func (h *CAHandler) GetDetail(c *gin.Context) {
	data, err := h.svc.GetDetail(c.Param("ca_id"))
	if err != nil { handleError(c, err); return }
	response.Success(c, data, "success")
}
func (h *CAHandler) Import(c *gin.Context) {
	var req ImportCARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	data, err := h.svc.Import(c.Request.Context(), &req)
	if err != nil { handleError(c, err); return }
	response.SuccessWithStatus(c, http.StatusCreated, data, "导入成功")
}
func (h *CAHandler) Export(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	result, err := h.svc.Export(c.Request.Context(),
		c.Param("ca_id"), ExportFormat(req.Type), req.Password)
	if err != nil { handleError(c, err); return }
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", `attachment; filename="`+result.Filename+`"`)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}
func (h *CAHandler) Download(c *gin.Context) {
	// 【预留】从 core 读取证书文件流式返回，与 Export(cert) 类似但走独立权限点
	result, err := h.svc.Export(c.Request.Context(),
		c.Param("ca_id"), ExportFormatCert, "")
	if err != nil { handleError(c, err); return }
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition", `attachment; filename="`+result.Filename+`"`)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}
func (h *CAHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("ca_id")); err != nil {
		handleError(c, err); return
	}
	response.Success(c, nil, "删除成功")
}
