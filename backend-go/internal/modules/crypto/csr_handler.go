package crypto

import (
	"github.com/gin-gonic/gin"

	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

type CSRHandler struct {
	svc *CSRService
}

func NewCSRHandler(svc *CSRService) *CSRHandler { return &CSRHandler{svc: svc} }

func (h *CSRHandler) RegisterRoutes(r *gin.RouterGroup) {
	csrs := r.Group("/csrs")
	csrs.GET("", middleware.RequirePermission("crypto_console:csr:view"), h.List)
}

func (h *CSRHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
