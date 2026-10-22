package route_perm
import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"backend-go/internal/exception"
	"backend-go/internal/response"
)
// Handler 路由权限 HTTP 处理器。
type Handler struct {
	svc *Service
}
// NewHandler 构造路由权限处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
// RegisterRoutes 注册路由权限管理接口。
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	g := group.Group("/route-permissions")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
// List 查询路由权限列表。
func (h *Handler) List(c *gin.Context) {
	list, err := h.svc.List(c.Query("moduleId"))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, list, "success")
}
// Create 新增路由权限。
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeValidationFail, "请求体校验失败："+err.Error(), nil)
		return
	}
	rp, err := h.svc.Create(&req)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, rp, "创建成功")
}
// Update 修改路由权限。
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, exception.CodeParamInvalid, "id 无效", nil)
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, exception.CodeValidationFail, "请求体校验失败："+err.Error(), nil)
		return
	}
	var status *int8
	if v := c.Query("status"); v != "" {
		n, _ := strconv.Atoi(v)
		s := int8(n)
		status = &s
	}
	rp, err := h.svc.Update(uint(id), &req, status)
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, rp, "更新成功")
}
// Delete 删除路由权限。
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, exception.CodeParamInvalid, "id 无效", nil)
		return
	}
	if err := h.svc.Delete(uint(id)); err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, nil, "删除成功")
}
func writeError(c *gin.Context, err error) {
	if pe, ok := err.(*exception.PlatformError); ok {
		exception.Abort(c, pe)
		return
	}
	exception.Abort(c, exception.New(exception.CodeInternalError, err.Error(), 500, nil))
}
