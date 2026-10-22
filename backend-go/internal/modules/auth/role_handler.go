package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/response"
)

// RegisterRoleRoutes 注册角色管理路由。
func (h *Handler) RegisterRoleRoutes(group *gin.RouterGroup) {
	g := group.Group("/auth/roles")
	g.GET("", h.ListRoles)
	g.GET("/:id", h.GetRole)
	g.POST("", h.CreateRole)
	g.PUT("/:id", h.UpdateRole)
	g.DELETE("/:id", h.DeleteRole)
}

// ListRoles 角色列表。
func (h *Handler) ListRoles(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	pageSize := atoiDefault(c.Query("pageSize"), 20)
	if v := c.Query("page_size"); v != "" {
		pageSize = atoiDefault(v, pageSize)
	}
	keyword := c.Query("keyword")
	if keyword == "" {
		keyword = c.Query("name")
	}
	data, err := h.svc.ListRoles(page, pageSize, keyword)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// GetRole 角色详情。
func (h *Handler) GetRole(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	data, err := h.svc.GetRoleDetail(id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}

// CreateRole 新增角色。
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeValidationFail, "请求体校验失败："+err.Error(), nil)
		return
	}
	data, err := h.svc.CreateRole(&req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "创建成功")
}

// UpdateRole 修改角色。
func (h *Handler) UpdateRole(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeValidationFail, "请求体校验失败："+err.Error(), nil)
		return
	}
	data, err := h.svc.UpdateRole(id, &req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "更新成功")
}

// DeleteRole 删除角色。
func (h *Handler) DeleteRole(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteRole(id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}

// atoiDefault 解析 query 整数，失败返回默认值。
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
