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
// 注意：静态段 /sign、/import、/dual/import、/envelope/query 必须早于 /:cert_id。
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
	certs.POST("/dual/import",
		middleware.RequirePermission("crypto_console:cert:import"),
		h.ImportDualCert,
	)
	certs.POST("/envelope/query",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.QueryEnvelope,
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

// Sign 签发证书（普通 / 双证）。
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
	result, err := h.svc.Sign(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, result, "签发成功")
}

// Import 导入单证书。
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

// ImportDualCert 国密双证导入。
func (h *CertHandler) ImportDualCert(c *gin.Context) {
	var req ImportDualCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": err.Error()},
			}})
		return
	}
	data, err := h.svc.ImportDualCert(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, data, "双证导入成功")
}

// Export 导出证书 / 私钥 / PKCS#12。
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

// QueryEnvelope 查询信封信息。
func (h *CertHandler) QueryEnvelope(c *gin.Context) {
	var req QueryEnvelopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": err.Error()},
			}})
		return
	}
	data, err := h.svc.QueryEnvelope(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// Delete 硬删除证书（不可恢复）。
//
// 行为：
//   - 删除 DB 记录（Unscoped，物理删除）
//   - 删除磁盘文件（.cert.pem / .key.pem，路径必须在 CertRoot 下）
//
// 审计与恢复能力由上层自行设计。
func (h *CertHandler) Delete(c *gin.Context) {
	certID := c.Param("cert_id")
	if certID == "" {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "缺少 cert_id", nil)
		return
	}
	if err := h.svc.Delete(certID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
