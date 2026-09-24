package crypto

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// CertHandler 证书领域 Handler，只注册 /certs/* 路由。
type CertHandler struct {
	svc *CertService
}

// NewCertHandler 创建证书 Handler。
func NewCertHandler(svc *CertService) *CertHandler {
	return &CertHandler{svc: svc}
}

// RegisterRoutes 注册证书路由。
//
// 注意：静态段 /sign、/import 必须早于 /:cert_id。
//
// 路由清单：
//   GET    /certs                  列表
//   POST   /certs/sign             申请/签发证书
//   POST   /certs/import           导入证书
//   GET    /certs/:cert_id         元数据
//   GET    /certs/:cert_id/detail  详情解析
//   POST   /certs/:cert_id/export  导出（cert / key / pkcs12）
//   DELETE /certs/:cert_id         删除（软删除）
func (h *CertHandler) RegisterRoutes(r *gin.RouterGroup) {
	certs := r.Group("/certs")

	certs.GET("",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.List,
	)
	certs.POST("/sign",
		middleware.RequirePermission("crypto_console:cert:sign"),
		h.Sign,
	)
	certs.POST("/import",
		middleware.RequirePermission("crypto_console:cert:import"),
		h.Import,
	)

	certs.GET("/:cert_id",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.Get,
	)
	certs.GET("/:cert_id/detail",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.GetDetail,
	)
	certs.POST("/:cert_id/export",
		middleware.RequirePermission("crypto_console:cert:export"),
		h.Export,
	)
	certs.DELETE("/:cert_id",
		middleware.RequirePermission("crypto_console:cert:delete"),
		h.Delete,
	)
}

// List 证书列表。
func (h *CertHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size, c.Query("cert_type"), c.Query("ca_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// Get 证书元数据。
func (h *CertHandler) Get(c *gin.Context) {
	data, err := h.svc.Get(c.Param("cert_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// GetDetail 证书详情（解析）。
func (h *CertHandler) GetDetail(c *gin.Context) {
	data, err := h.svc.GetDetail(c.Param("cert_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// Sign 申请/签发证书。
func (h *CertHandler) Sign(c *gin.Context) {
	var req SignCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": err.Error()},
			}})
		return
	}
	data, err := h.svc.Sign(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, data, "签发成功")
}

// Import 导入证书。
func (h *CertHandler) Import(c *gin.Context) {
	var req ImportCertRequest
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

// Export 导出证书。
func (h *CertHandler) Export(c *gin.Context) {
	var req ExportCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	result, err := h.svc.Export(
		c.Request.Context(), c.Param("cert_id"), req.Type, req.Password,
	)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Type", result.ContentType)
	c.Header("Content-Disposition",
		`attachment; filename="`+result.Filename+`"`)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

// Delete 删除证书。
func (h *CertHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("cert_id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
