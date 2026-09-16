package auth
import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// RegisterUserRoutes 注册用户管理路由（受保护）。
//
// 逐路由挂载权限中间件，保证每个 HTTP 方法对应正确的权限点。
func (h *Handler) RegisterUserRoutes(r *gin.RouterGroup) {
	users := r.Group("/auth/users")
	users.GET("", middleware.RequirePermission("auth:user:view"), h.ListUsers)
	users.POST("", middleware.RequirePermission("auth:user:create"), h.CreateUser)
	users.GET("/:user_id", middleware.RequirePermission("auth:user:view"), h.GetUserDetail)
	users.PUT("/:user_id", middleware.RequirePermission("auth:user:edit"), h.UpdateUser)
	users.DELETE("/:user_id", middleware.RequirePermission("auth:user:delete"), h.DeleteUser)
	users.PATCH("/:user_id/status", middleware.RequirePermission("auth:user:edit"), h.UpdateUserStatus)
	users.PATCH("/:user_id/password", middleware.RequirePermission("auth:user:edit"), h.ResetUserPassword)
}
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest,
			exception.CodeParamInvalid, "路径参数无效", nil)
		return 0, false
	}
	return uint(id), true
}
func currentOperator(c *gin.Context) (uint, string, bool) {
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return 0, "", false
	}
	uid, ok := userID.(uint)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return 0, "", false
	}
	name := ""
	if val, ok := c.Get(middleware.ContextKeyUserInfo); ok {
		if uc, ok := val.(middleware.UserContext); ok {
			name = uc.Username
		}
	}
	return uid, name, true
}
// bindQueryOr422 绑定 query 参数，失败时返回 422 / 90004。
func bindQueryOr422(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindQuery(req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(req, err, "query")})
		return false
	}
	return true
}
// ListUsers 查询用户列表。
func (h *Handler) ListUsers(c *gin.Context) {
	var q UserListQuery
	if !bindQueryOr422(c, &q) {
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	data, err := h.svc.ListUsers(q)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// CreateUser 创建用户。
func (h *Handler) CreateUser(c *gin.Context) {
	var req UserCreate
	if !bindJSONOr422(c, &req) {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	data, err := h.svc.CreateUser(req, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "用户创建成功")
}
// GetUserDetail 获取用户详情。
func (h *Handler) GetUserDetail(c *gin.Context) {
	userID, ok := parseUintParam(c, "user_id")
	if !ok {
		return
	}
	data, err := h.svc.GetUserDetail(userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// UpdateUser 更新用户。
func (h *Handler) UpdateUser(c *gin.Context) {
	userID, ok := parseUintParam(c, "user_id")
	if !ok {
		return
	}
	var req UserUpdate
	if !bindJSONOr422(c, &req) {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	data, err := h.svc.UpdateUser(userID, req, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "用户更新成功")
}
// DeleteUser 删除用户。
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, ok := parseUintParam(c, "user_id")
	if !ok {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteUser(userID, uid, name); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
// UpdateUserStatus 启用/禁用用户。
//
// UserStatusUpdate.Status 已声明 `binding:"required"`，
// gin binding 层保证非 nil；若缺失，bindJSONOr422 已返回 422 / 90004。
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	userID, ok := parseUintParam(c, "user_id")
	if !ok {
		return
	}
	var req UserStatusUpdate
	if !bindJSONOr422(c, &req) {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	data, err := h.svc.UpdateUserStatus(userID, *req.Status, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "状态更新成功")
}
// ResetUserPassword 管理员重置用户密码。
func (h *Handler) ResetUserPassword(c *gin.Context) {
	userID, ok := parseUintParam(c, "user_id")
	if !ok {
		return
	}
	var req UserPasswordReset
	if !bindJSONOr422(c, &req) {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	if err := h.svc.ResetUserPassword(userID, req.NewPassword, uid, name); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "密码重置成功")
}
