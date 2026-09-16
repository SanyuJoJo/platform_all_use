package module_manager
import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// Handler 模块管理路由处理器。
type Handler struct {
	svc *Service
}
// NewHandler 创建模块管理路由处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
// RegisterRoutes 注册模块管理路由（受保护）。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	modules := r.Group("/modules")
	modules.GET("", middleware.RequirePermission("module_manager:module:view"), h.ListModules)
	modules.POST("", middleware.RequirePermission("module_manager:module:create"), h.InstallModule)
	modules.POST("/upload", middleware.RequirePermission("module_manager:module:create"), h.UploadAndInstall)
	modules.DELETE("/:module_id", middleware.RequirePermission("module_manager:module:delete"), h.UninstallModule)
	modules.POST("/:module_id/enable", middleware.RequirePermission("module_manager:module:edit"), h.EnableModule)
	modules.POST("/:module_id/disable", middleware.RequirePermission("module_manager:module:edit"), h.DisableModule)
	modules.POST("/:module_id/upgrade", middleware.RequirePermission("module_manager:module:create"), h.UpgradeModule)
	modules.GET("/:module_id/config", middleware.RequirePermission("module_manager:module:view"), h.GetModuleConfig)
	modules.PUT("/:module_id/config", middleware.RequirePermission("module_manager:module:edit"), h.UpdateModuleConfig)
}
// ListModules 查询模块列表。
func (h *Handler) ListModules(c *gin.Context) {
	var q ModuleListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(&q, err, "query")})
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	var userPerms []string
	if val, ok := c.Get(middleware.ContextKeyUserInfo); ok {
		if uc, ok := val.(middleware.UserContext); ok {
			userPerms = uc.Permissions
		}
	}
	data, err := h.svc.ListModules(q, userPerms)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// InstallModule 从受控目录安装模块。
func (h *Handler) InstallModule(c *gin.Context) {
	var req ModuleInstallReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(&req, err, "body")})
		return
	}
	if req.InstallType == "" {
		req.InstallType = "path"
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	data, err := h.svc.InstallModule(req.InstallType, "", req.SourcePath, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "模块安装成功")
}
// UploadAndInstall 上传 ZIP 并安装。
func (h *Handler) UploadAndInstall(c *gin.Context) {
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "缺少 file 字段", nil)
		return
	}
	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError,
			exception.CodeInternalError, "读取上传文件失败", nil)
		return
	}
	defer f.Close()
	maxSize := int64(h.svc.cfg.ModuleZipMaxSize)
	if maxSize <= 0 {
		maxSize = DefaultZipMaxSize
	}
	content, err := io.ReadAll(io.LimitReader(f, maxSize+1))
	if err != nil {
		response.Error(c, http.StatusInternalServerError,
			exception.CodeInternalError, "读取上传文件失败", nil)
		return
	}
	if int64(len(content)) > maxSize {
		response.Error(c, http.StatusBadRequest,
			exception.CodeModuleZipInvalid, "ZIP 文件超过大小上限", nil)
		return
	}
	savedPath, err := h.svc.SaveUploadedZip(file.Filename, content)
	if err != nil {
		handleError(c, err)
		return
	}
	data, err := h.svc.InstallModule("zip", savedPath, "", uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "模块上传并安装成功")
}
// UninstallModule 卸载模块。
func (h *Handler) UninstallModule(c *gin.Context) {
	moduleID := c.Param("module_id")
	force := c.Query("force") == "true"
	dropTables := c.Query("drop_tables") == "true"
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	if err := h.svc.UninstallModule(moduleID, force, dropTables, uid, name); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "卸载成功")
}
// EnableModule 启用模块。
func (h *Handler) EnableModule(c *gin.Context) {
	moduleID := c.Param("module_id")
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	data, err := h.svc.EnableModule(moduleID, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "模块已启用")
}
// DisableModule 停用模块。
func (h *Handler) DisableModule(c *gin.Context) {
	moduleID := c.Param("module_id")
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	data, err := h.svc.DisableModule(moduleID, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "模块已停用")
}
// UpgradeModule 升级模块。
func (h *Handler) UpgradeModule(c *gin.Context) {
	moduleID := c.Param("module_id")
	var req ModuleUpgradeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(&req, err, "body")})
		return
	}
	filePath := ""
	if req.FilePath != nil {
		filePath = *req.FilePath
	}
	sourcePath := ""
	if req.SourcePath != nil {
		sourcePath = *req.SourcePath
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	data, err := h.svc.UpgradeModule(moduleID, req.InstallType, filePath, sourcePath, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "模块升级成功")
}
// GetModuleConfig 获取模块配置。
func (h *Handler) GetModuleConfig(c *gin.Context) {
	moduleID := c.Param("module_id")
	data, err := h.svc.GetModuleConfig(moduleID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// UpdateModuleConfig 更新模块配置。
func (h *Handler) UpdateModuleConfig(c *gin.Context) {
	moduleID := c.Param("module_id")
	var req ModuleConfigUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(&req, err, "body")})
		return
	}
	uid, name, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	data, err := h.svc.UpdateModuleConfig(moduleID, req.Config, uid, name)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "配置更新成功")
}
// currentOperator 从 gin.Context 获取操作者 ID 与用户名。
//
// v1.2（P2-NEW-04）：不写响应，仅返回三元组。
func currentOperator(c *gin.Context) (uint, string, bool) {
	uid := uint(0)
	name := ""
	if val, ok := c.Get(middleware.ContextKeyUserID); ok {
		if v, ok := val.(uint); ok {
			uid = v
		}
	}
	if val, ok := c.Get(middleware.ContextKeyUserInfo); ok {
		if uc, ok := val.(middleware.UserContext); ok {
			name = uc.Username
		}
	}
	return uid, name, uid != 0
}
// handleError 统一处理 PlatformError。
func handleError(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}
// buildValidationErrors 将绑定错误转换为结构化数组。
func buildValidationErrors(req interface{}, err error, location string) []map[string]interface{} {
	var result []map[string]interface{}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			jsonName := toJSONFieldName(req, e.StructField())
			result = append(result, map[string]interface{}{
				"type":  e.Tag(),
				"loc":   []string{location, jsonName},
				"msg":   e.Error(),
				"input": e.Value(),
			})
		}
		return result
	}
	result = append(result, map[string]interface{}{
		"type": "invalid_json",
		"msg":  err.Error(),
	})
	return result
}
// toJSONFieldName 通过反射从结构体的 json tag 提取 JSON 字段名。
func toJSONFieldName(req interface{}, structField string) string {
	t := reflect.TypeOf(req)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return strings.ToLower(structField)
	}
	f, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}
	jsonTag := f.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return strings.ToLower(structField)
	}
	name := strings.Split(jsonTag, ",")[0]
	if name == "" {
		return strings.ToLower(structField)
	}
	return name
}
