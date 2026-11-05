package crypto

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

type CFCAHandler struct {
	svc *CFCAService
}

func NewCFCAHandler(svc *CFCAService) *CFCAHandler { return &CFCAHandler{svc: svc} }

func (h *CFCAHandler) RegisterRoutes(r *gin.RouterGroup) {
	cfca := r.Group("/cfca")
	cfca.POST("/sign",
		middleware.RequirePermission("crypto_console:cert:sign"),
		h.Sign,
	)
	cfca.POST("/decode-envelope",
		middleware.RequirePermission("crypto_console:cert:view"),
		h.DecodeEnvelope,
	)
}

func (h *CFCAHandler) Sign(c *gin.Context) {
	var req CFCASignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	data, err := h.svc.SignDual(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithStatus(c, http.StatusCreated, data, "CFCA 双证签发成功")
}

func (h *CFCAHandler) DecodeEnvelope(c *gin.Context) {
	var req CFCAQueryEnvelopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	data, err := h.svc.DecodeEnvelope(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
