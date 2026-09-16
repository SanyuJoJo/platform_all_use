package auth
import (
	"encoding/json"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// RegisterRoleRoutes 注册角色管理路由（受保护）。
func (h *Handler) RegisterRoleRoutes(r *gin.RouterGroup) {
	roles := r.Group("/auth/roles")
	roles.GET("", middleware.RequirePermission("auth:role:view"), h.ListRoles)
	roles.POST("", middleware.RequirePermission("auth:role:create"), h.CreateRole)
	roles.GET("/:role_id", middleware.RequirePermission("auth:role:view"), h.GetRoleDetail)
	roles.PUT("/:role_id", middleware.RequirePermission("auth:role:edit"), h.UpdateRole)
	roles.DELETE("/:role_id", middleware.RequirePermission("auth:role:delete"), h.DeleteRole)
}
// decodeStrictJSON 严格解析 JSON，拒绝未知字段。
//
// 用于实现 Python 版 RoleUpdate 的 extra="forbid" 语义。
func decodeStrictJSON(c *gin.Context, target interface{}) bool {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		errs := []map[string]interface{}{}
		if isUnknownFieldError(err) {
			field := extractUnknownField(err)
			errs = append(errs, map[string]interface{}{
				"type":  "extra_forbidden",
				"loc":   []string{"body", field},
				"msg":   "Extra inputs are not permitted",
				"input": nil,
			})
		} else {
			errs = append(errs, map[string]interface{}{
				"type": "invalid_json",
				"msg":  err.Error(),
			})
		}
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": errs})
		return false
	}
	if decoder.More() {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "invalid_json", "msg": "unexpected trailing data"},
			}})
		return false
	}
	return true
}
func isUnknownFieldError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unknown field")
}
func extractUnknownField(err error) string {
	msg := err.Error()
	const prefix = `json: unknown field "`
	idx := strings.Index(msg, prefix)
	if idx < 0 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(msg[start:], `"`)
	if end < 0 {
		return ""
	}
	return msg[start : start+end]
}
// ListRoles 查询角色列表。
func (h *Handler) ListRoles(c *gin.Context) {
	var q RoleListQuery
	if !bindQueryOr422(c, &q) {
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	data, err := h.svc.ListRoles(q)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// CreateRole 创建角色。
func (h *Handler) CreateRole(c *gin.Context) {
	var req RoleCreate
	if !bindJSONOr422(c, &req) {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	data, err := h.svc.CreateRole(req, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "角色创建成功")
}
// GetRoleDetail 获取角色详情。
func (h *Handler) GetRoleDetail(c *gin.Context) {
	roleID, ok := parseUintParam(c, "role_id")
	if !ok {
		return
	}
	data, err := h.svc.GetRoleDetail(roleID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// UpdateRole 更新角色。
func (h *Handler) UpdateRole(c *gin.Context) {
	roleID, ok := parseUintParam(c, "role_id")
	if !ok {
		return
	}
	var req RoleUpdate
	if !decodeStrictJSON(c, &req) {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	data, err := h.svc.UpdateRole(roleID, req, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "角色更新成功")
}
// DeleteRole 删除角色。
func (h *Handler) DeleteRole(c *gin.Context) {
	roleID, ok := parseUintParam(c, "role_id")
	if !ok {
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteRole(roleID, uid, name); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
