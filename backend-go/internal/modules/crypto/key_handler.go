package crypto

import (
	"github.com/gin-gonic/gin"

	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

type KeyHandler struct {
	svc *KeyService
}

func NewKeyHandler(svc *KeyService) *KeyHandler { return &KeyHandler{svc: svc} }

func (h *KeyHandler) RegisterRoutes(r *gin.RouterGroup) {
	keys := r.Group("/keys")
	keys.GET("", middleware.RequirePermission("crypto_console:key:view"), h.List)
}

func (h *KeyHandler) List(c *gin.Context) {
	page, size := parsePageQuery(c)
	data, err := h.svc.List(page, size, c.Query("algorithm"), c.Query("state"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
