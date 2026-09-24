package crypto

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// MetaHandler 密码元数据查询与 CA 管理处理器。
type MetaHandler struct {
	svc *MetaService
}

// NewMetaHandler 创建元数据查询处理器。
func NewMetaHandler(svc *MetaService) *MetaHandler {
	return &MetaHandler{svc: svc}
}

// RegisterRoutes 注册元数据查询路由（受保护）。
//
// 注意：静态段 /import 必须早于 /:ca_id。
func (h *MetaHandler) RegisterRoutes(r *gin.RouterGroup) {
	// ---- CA ----
	cas := r.Group("/cas")
	cas.GET("",
		middleware.RequirePermission("crypto_console:ca:view"),
		h.ListCAs,
	)
	cas.POST("/import",
		middleware.RequirePermission("crypto_console:ca:import"),
		h.ImportCA,
	)
	cas.GET("/:ca_id",
		middleware.RequirePermission("crypto_console:ca:view"),
		h.GetCA,
	)
	cas.GET("/:ca_id/detail",
		middleware.RequirePermission("crypto_console:ca:view"),
		h.GetCADetail,
	)
	cas.POST("/:ca_id/export",
		middleware.RequirePermission("crypto_console:ca:export"),
		h.ExportCA,
	)
	cas.DELETE("/:ca_id",
		middleware.RequirePermission("crypto_console:ca:delete"),
		h.DeleteCA,
	)

	// ---- 证书 ----
	certs := r.Group("/certs")
	certs.GET("",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.ListCerts,
	)
	certs.GET("/:cert_id",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.GetCert,
	)

	// ---- CSR ----
	csrs := r.Group("/csrs")
	csrs.GET("",
		middleware.RequirePermission("crypto_console:csr:view"),
		h.ListCSRs,
	)

	// ---- CRL ----
	crls := r.Group("/crls")
	crls.GET("",
		middleware.RequirePermission("crypto_console:crl:view"),
		h.ListCRLs,
	)

	// ---- 密钥 ----
	keys := r.Group("/keys")
	keys.GET("",
		middleware.RequirePermission("crypto_console:key:view"),
		h.ListKeys,
	)
}

// ListCAs 查询 CA 列表。
func (h *MetaHandler) ListCAs(c *gin.Context) {
	page, pageSize := parsePageQuery(c)
	data, err := h.svc.ListCAs(page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// GetCA 查询 CA 详情。
func (h *MetaHandler) GetCA(c *gin.Context) {
	caID := c.Param("ca_id")
	data, err := h.svc.GetCA(caID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// GetCADetail 查询 CA 证书解析详情。
func (h *MetaHandler) GetCADetail(c *gin.Context) {
	caID := c.Param("ca_id")
	data, err := h.svc.GetCADetail(caID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// ImportCA 导入 CA（支持带私钥与仅证书）。
func (h *MetaHandler) ImportCA(c *gin.Context) {
	var req ImportCARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": err.Error()},
			}})
		return
	}
	data, err := h.svc.ImportCA(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, data, "导入成功")
}

// ExportCA 导出 CA（cert / key / pkcs12）。
func (h *MetaHandler) ExportCA(c *gin.Context) {
	caID := c.Param("ca_id")
	var req struct {
		Type     string `json:"type" binding:"required"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	result, err := h.svc.ExportCA(
		c.Request.Context(), caID, ExportFormat(req.Type), req.Password,
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

// DeleteCA 删除 CA。
func (h *MetaHandler) DeleteCA(c *gin.Context) {
	caID := c.Param("ca_id")
	if err := h.svc.DeleteCA(caID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}

// ListCerts 查询证书列表。
func (h *MetaHandler) ListCerts(c *gin.Context) {
	page, pageSize := parsePageQuery(c)
	data, err := h.svc.ListCerts(page, pageSize, c.Query("cert_type"), c.Query("ca_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// GetCert 查询证书详情。
func (h *MetaHandler) GetCert(c *gin.Context) {
	data, err := h.svc.GetCert(c.Param("cert_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// ListCSRs 查询 CSR 列表。
func (h *MetaHandler) ListCSRs(c *gin.Context) {
	page, pageSize := parsePageQuery(c)
	data, err := h.svc.ListCSRs(page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// ListCRLs 查询 CRL 列表。
func (h *MetaHandler) ListCRLs(c *gin.Context) {
	page, pageSize := parsePageQuery(c)
	data, err := h.svc.ListCRLs(page, pageSize, c.Query("ca_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// ListKeys 查询密钥元数据列表。
func (h *MetaHandler) ListKeys(c *gin.Context) {
	page, pageSize := parsePageQuery(c)
	data, err := h.svc.ListKeys(page, pageSize, c.Query("algorithm"), c.Query("state"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// parsePageQuery 解析分页参数。
func parsePageQuery(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}
	return page, pageSize
}
