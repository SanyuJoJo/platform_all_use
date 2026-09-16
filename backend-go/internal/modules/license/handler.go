package license
import (
	"errors"
	"io"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// Handler License 管理路由处理器。
type Handler struct {
	svc *Service
}
// NewHandler 创建 License 管理路由处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
// RegisterRoutes 注册 License 管理路由（受保护）。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	licenses := r.Group("/license")
	licenses.GET("/status", middleware.RequirePermission("license:license:view"), h.GetStatus)
	licenses.POST("/import", middleware.RequirePermission("license:license:create"), h.Import)
	licenses.POST("/activate", middleware.RequirePermission("license:license:create"), h.Activate)
	licenses.GET("/modules", middleware.RequirePermission("license:license:view"), h.GetModules)
}
// GetStatus 获取 License 状态。
func (h *Handler) GetStatus(c *gin.Context) {
	data, err := h.svc.GetLicenseStatus()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// GetModules 获取模块授权状态。
func (h *Handler) GetModules(c *gin.Context) {
	data, err := h.svc.GetModuleAuthorization()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// Import 导入 License（multipart/form-data）。
//
// 支持：
//   - license_file：.lic / .json 文件（≤ 1MB）；
//   - activation_code：在线激活码（≤ 255 位）。
//
// 二者二选一。
func (h *Handler) Import(c *gin.Context) {
	operatorID, operatorName, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	var fileContent []byte
	file, err := c.FormFile("license_file")
	if err == nil && file != nil {
		if file.Size > LicenseFileMaxSize {
			response.Error(c, http.StatusBadRequest,
				exception.CodeLicenseInvalidFile, "License 文件超过大小上限", nil)
			return
		}
		// 扩展名校验
		lower := strings.ToLower(file.Filename)
		extOK := false
		for _, ext := range LicenseFileExtensions {
			if strings.HasSuffix(lower, ext) {
				extOK = true
				break
			}
		}
		if !extOK {
			response.Error(c, http.StatusBadRequest,
				exception.CodeLicenseInvalidFile,
				"License 文件扩展名无效，仅支持 .lic / .json", nil)
			return
		}
		f, err := file.Open()
		if err != nil {
			response.Error(c, http.StatusInternalServerError,
				exception.CodeInternalError, "读取上传文件失败", nil)
			return
		}
		defer f.Close()
		fileContent, err = io.ReadAll(io.LimitReader(f, LicenseFileMaxSize+1))
		if err != nil {
			response.Error(c, http.StatusInternalServerError,
				exception.CodeInternalError, "读取上传文件失败", nil)
			return
		}
		if int64(len(fileContent)) > LicenseFileMaxSize {
			response.Error(c, http.StatusBadRequest,
				exception.CodeLicenseInvalidFile, "License 文件超过大小上限", nil)
			return
		}
	}
	activationCode := c.PostForm("activation_code")
	licenseObj, err := h.svc.ImportLicense(fileContent, activationCode, operatorID, operatorName)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, SerializeLicenseForImport(licenseObj), "License 导入成功")
}
// Activate 在线激活。
func (h *Handler) Activate(c *gin.Context) {
	operatorID, operatorName, ok := currentOperator(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	var req LicenseActivateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{"type": "validation", "msg": err.Error()},
			}})
		return
	}
	licenseObj, err := h.svc.ActivateLicense(
		req.ActivationCode, req.MachineCode, operatorID, operatorName,
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, SerializeLicenseForImport(licenseObj), "License 激活成功")
}
// currentOperator 从 gin.Context 获取操作者 ID 与用户名。
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
