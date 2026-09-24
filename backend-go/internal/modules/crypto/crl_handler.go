package crypto

import (
	"github.com/gin-gonic/gin"

	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

type CRLHandler struct {
	svc *CRLService
}

func NewCRLHandler(svc *CRLService) *CRLHandler { return &CRLHandler{svc: svc} }

func (h *CRLHandler) RegisterRoutes(r *gin.RouterGroup) {
	crls := r.Group("/crls")
	crls.GET("", middleware.RequirePermission("crypto_console:crl:view"), h.List)
}

func (h *CRLHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size, c.Query("ca_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
