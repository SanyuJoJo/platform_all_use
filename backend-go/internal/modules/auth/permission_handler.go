package auth
import (
	"github.com/gin-gonic/gin"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// RegisterPermissionRoutes 注册权限管理路由（受保护）。
func (h *Handler) RegisterPermissionRoutes(r *gin.RouterGroup) {
	perms := r.Group("/auth/permissions")
	perms.GET("", middleware.RequirePermission("auth:permission:view"), h.ListPermissions)
	perms.GET("/modules/:module_id", middleware.RequirePermission("auth:permission:view"), h.GetPermissionsByModule)
}
// ListPermissions 查询权限列表。
func (h *Handler) ListPermissions(c *gin.Context) {
	moduleID := c.Query("module_id")
	resource := c.Query("resource")
	data, err := h.svc.ListPermissions(moduleID, resource)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// GetPermissionsByModule 获取指定模块的所有权限。
func (h *Handler) GetPermissionsByModule(c *gin.Context) {
	moduleID := c.Param("module_id")
	data, err := h.svc.GetPermissionsByModule(moduleID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
