package upgrade

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// Handler 升级模块路由处理器。
//
// ★ 安全要点：
//   - 升级相关写接口一律要求「拥有显式权限码」且「admin/super_admin 角色」；
//   - 不进入 DynamicPermission 的超管直通分支，避免降权后旧 token 仍可升级；
//   - 只读取本地 updateDir 预置包，不接受外部来源串。
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes 注册升级路由（受保护）。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	upg := r.Group("/upgrade")
	// 读接口：需 upgrade:package:view
	upg.GET("/status", middleware.RequirePermission("upgrade:package:view"), h.Status)
	upg.GET("/packages", middleware.RequirePermission("upgrade:package:view"), h.ListPackages)
	// 写接口：需 upgrade:package:run / upgrade:package:rollback，且必须是管理员
	upg.POST("/run", middleware.RequirePermission("upgrade:package:run"), h.adminOnly(h.Run))
	upg.POST("/rollback", middleware.RequirePermission("upgrade:package:rollback"), h.adminOnly(h.Rollback))
}

// Status 当前版本 + 可用升级提醒（推送提醒的数据源）。
func (h *Handler) Status(c *gin.Context) {
	status, err := h.svc.Status()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, status, "success")
}

// ListPackages 升级包列表。
func (h *Handler) ListPackages(c *gin.Context) {
	packages, err := h.svc.ListPackages()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, gin.H{"items": packages}, "success")
}

// Run 执行升级。
func (h *Handler) Run(c *gin.Context) {
	var req RunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	status, err := h.svc.Run(req.PackageID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, status, "升级已提交")
}

// Rollback 回退。
func (h *Handler) Rollback(c *gin.Context) {
	var req RollbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败", nil)
		return
	}
	status, err := h.svc.Rollback(req.ID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, status, "回退已提交")
}

// adminOnly 包装：仅允许 admin / super_admin 角色（写接口强制管理员）。
func (h *Handler) adminOnly(next func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAdminUser(c) {
			response.Error(c, http.StatusForbidden,
				exception.CodePermissionDenied, "仅管理员可执行该操作", nil)
			c.Abort()
			return
		}
		next(c)
	}
}

// isAdminUser 判断当前用户是否 admin / super_admin。
func isAdminUser(c *gin.Context) bool {
	if val, ok := c.Get(middleware.ContextKeyUserInfo); ok {
		if uc, ok := val.(middleware.UserContext); ok {
			for _, r := range uc.Roles {
				if r == "admin" || r == "super_admin" {
					return true
				}
			}
		}
	}
	return false
}

func handleError(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if ok := asPlatformError(err, &pe); ok {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}

func asPlatformError(err error, out **exception.PlatformError) bool {
	if e, ok := err.(*exception.PlatformError); ok {
		*out = e
		return true
	}
	return false
}
