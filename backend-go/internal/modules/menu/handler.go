package menu

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)

// Handler 菜单 HTTP 处理器。
type Handler struct {
	svc *Service
}

// NewHandler 构造菜单处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes 注册菜单管理与动态路由接口。
//
// 路由规划（注意 Gin 静态段与参数段不能同层级）：
//   - GET    /menus/routers          当前用户路由（前端 addRoute 的数据源，免权限）
//   - GET    /menus/tree             菜单树（管理端）
//   - GET    /menus/micro-apps       已知微应用列表（菜单管理下拉，v1.7 新增）
//   - GET    /menus                  菜单列表
//   - GET    /menus/detail/:id       菜单详情
//   - POST   /menus                  新增菜单
//   - PUT    /menus/detail/:id       修改菜单
//   - DELETE /menus/detail/:id       删除菜单
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	g := group.Group("/menus")
	g.GET("/routers", h.GetRouters)
	g.GET("/tree", h.Tree)
	g.GET("/micro-apps", h.ListMicroApps) // ★ 新增
	g.GET("", h.List)
	g.GET("/detail/:id", h.Get)
	g.POST("", h.Create)
	g.PUT("/detail/:id", h.Update)
	g.DELETE("/detail/:id", h.Delete)
}

// GetRouters 获取当前登录用户可见的前端路由树。
func (h *Handler) GetRouters(c *gin.Context) {
	uc, ok := middleware.GetCurrentUser(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	routers, err := h.svc.BuildRouters(uc.Username, uc.Roles, uc.Permissions)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"routers": routers}, "success")
}

// List 菜单列表。
func (h *Handler) List(c *gin.Context) {
	var q QueryMenuRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeParamInvalid, "参数错误："+err.Error(), nil)
		return
	}
	list, err := h.svc.List(&q)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list, "success")
}

// Tree 菜单树（管理端）。
func (h *Handler) Tree(c *gin.Context) {
	tree, err := h.svc.Tree()
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, tree, "success")
}

// ListMicroApps 返回已知微应用列表，供菜单管理下拉使用。
//
// 权限：auth:menu:list（在 route_perm 种子里配置）。
func (h *Handler) ListMicroApps(c *gin.Context) {
	items, err := h.svc.ListKnownMicroApps()
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, gin.H{"items": items}, "success")
}

// Get 菜单详情。
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	m, err := h.svc.GetByID(id)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, m, "success")
}

// Create 新增菜单。
func (h *Handler) Create(c *gin.Context) {
	var req CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeValidationFail, "请求体校验失败："+err.Error(), nil)
		return
	}
	m, err := h.svc.Create(&req, operatorOf(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, m, "创建成功")
}

// Update 修改菜单。
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeValidationFail, "请求体校验失败："+err.Error(), nil)
		return
	}
	m, err := h.svc.Update(id, &req, operatorOf(c))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, m, "更新成功")
}

// Delete 删除菜单。
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}

func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, exception.CodeParamInvalid, "id 无效", nil)
		return 0, false
	}
	return uint(id), true
}

func operatorOf(c *gin.Context) string {
	if uc, ok := middleware.GetCurrentUser(c); ok {
		return uc.Username
	}
	return ""
}

func writeError(c *gin.Context, err error) {
	if pe, ok := err.(*exception.PlatformError); ok {
		exception.Abort(c, pe)
		return
	}
	exception.Abort(c, exception.New(exception.CodeInternalError, err.Error(), 500, nil))
}
