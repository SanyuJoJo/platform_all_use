package crypto

import (
	"github.com/gin-gonic/gin"

	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

type CertHandler struct {
	svc *CertService
}

func NewCertHandler(svc *CertService) *CertHandler { return &CertHandler{svc: svc} }

func (h *CertHandler) RegisterRoutes(r *gin.RouterGroup) {
	certs := r.Group("/certs")
	certs.GET("", middleware.RequirePermission("crypto_console:cert:view"), h.List)
	certs.GET("/:cert_id", middleware.RequirePermission("crypto_console:cert:view"), h.Get)

	// 【预留】证书导入 / 导出 / 下载
	// certs.POST("/import", middleware.RequirePermission("crypto_console:cert:import"), h.Import)
	// certs.POST("/:cert_id/export", middleware.RequirePermission("crypto_console:cert:export"), h.Export)
	// certs.GET("/:cert_id/download", middleware.RequirePermission("crypto_console:cert:view"), h.Download)
}

func (h *CertHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size, c.Query("cert_type"), c.Query("ca_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

func (h *CertHandler) Get(c *gin.Context) {
	data, err := h.svc.Get(c.Param("cert_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
