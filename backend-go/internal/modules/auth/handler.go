package auth
import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// Handler 认证路由处理器。
type Handler struct {
	svc *Service
}
// NewHandler 创建认证路由处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
// RegisterPublicRoutes 注册不需要认证的路由。
func (h *Handler) RegisterPublicRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
}
// RegisterProtectedRoutes 注册需要认证的认证路由。
func (h *Handler) RegisterProtectedRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/logout", h.Logout)
	auth.GET("/me", h.Me)
	auth.PUT("/me/password", h.ChangePassword)
}
// bindJSONOr422 绑定 JSON 请求体；失败时统一返回 422 / 90004。
func bindJSONOr422(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(
			c,
			http.StatusUnprocessableEntity,
			exception.CodeValidationFail,
			"请求参数校验失败",
			gin.H{"errors": buildValidationErrors(req, err, "body")},
		)
		return false
	}
	return true
}
// buildValidationErrors 将绑定错误转换为结构化数组。
//
// v1.3 修复 P2-C：
//   - 类型转换错误分支现在**只匹配 strconv.Parse* 错误**（来自 query 参数）；
//   - body 场景的 "cannot unmarshal" 错误归入 invalid_json 分支，
//     语义上属于请求体格式错误，与 query 类型错误区分。
func buildValidationErrors(req interface{}, err error, location string) []map[string]interface{} {
	var result []map[string]interface{}
	// 1. validator.ValidationErrors（字段校验失败）
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
	// 2. 类型转换错误（如 ?page=abc 时 strconv.ParseInt 失败）
	if isTypeConversionError(err) {
		fieldName := extractFieldFromTypeError(err)
		result = append(result, map[string]interface{}{
			"type": "type_error",
			"loc":  []string{location, fieldName},
			"msg":  err.Error(),
		})
		return result
	}
	// 3. 其他（JSON body 解析失败、类型不匹配等）
	result = append(result, map[string]interface{}{
		"type": "invalid_json",
		"msg":  err.Error(),
	})
	return result
}
// isTypeConversionError 判断错误是否为 query 参数的类型转换失败。
//
// v1.3 修复 P2-C：
//   - 仅匹配 strconv.ParseInt / ParseUint / ParseFloat / ParseBool 错误；
//   - **不再**匹配 "cannot unmarshal"（该错误来自 JSON body 解析，
//     语义上属于 invalid_json，而非 query type_error）。
func isTypeConversionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "strconv.ParseInt") ||
		strings.Contains(msg, "strconv.ParseUint") ||
		strings.Contains(msg, "strconv.ParseFloat") ||
		strings.Contains(msg, "strconv.ParseBool")
}
// extractFieldFromTypeError 从类型转换错误中提取字段名。
//
// P2-D 已知限制：gin 在类型转换错误时返回的消息为
// `strconv.ParseInt: parsing "abc": invalid syntax`，
// 不含 `Key: 'Xxx.Field'` 前缀，因此本函数通常返回 "unknown"。
// 前端应以 errors[0].msg 为主进行错误提示。
func extractFieldFromTypeError(err error) string {
	msg := err.Error()
	const prefix = "Key: '"
	if idx := strings.Index(msg, prefix); idx >= 0 {
		rest := msg[idx+len(prefix):]
		if end := strings.Index(rest, "'"); end >= 0 {
			full := rest[:end]
			if dot := strings.LastIndex(full, "."); dot >= 0 {
				structField := full[dot+1:]
				return strings.ToLower(structField)
			}
		}
	}
	return "unknown"
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
// Login 用户登录。
func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if !bindJSONOr422(c, &req) {
		return
	}
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	userInfo, err := h.svc.AuthenticateUser(req.Username, req.Password, ip, userAgent)
	if err != nil {
		handleError(c, err)
		return
	}
	tokens, err := h.svc.CreateTokensForUser(userInfo, ip, userAgent)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, tokens, "登录成功")
}
// Refresh 刷新 Access Token。
func (h *Handler) Refresh(c *gin.Context) {
	var token string
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = strings.TrimSpace(authHeader[7:])
	}
	if token == "" && c.Request.ContentLength > 0 {
		var req RefreshReq
		if !bindJSONOr422(c, &req) {
			return
		}
		token = req.RefreshToken
	}
	if token == "" {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "缺少refresh token", nil)
		return
	}
	resp, err := h.svc.RefreshAccessToken(token)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, resp, "刷新成功")
}
// Logout 用户登出。
func (h *Handler) Logout(c *gin.Context) {
	var req RefreshReq
	if c.Request.ContentLength > 0 {
		if !bindJSONOr422(c, &req) {
			return
		}
	}
	if req.RefreshToken != "" {
		_ = h.svc.RevokeRefreshToken(req.RefreshToken)
	}
	if userID, exists := c.Get(middleware.ContextKeyUserID); exists {
		if uid, ok := userID.(uint); ok {
			logAuthEvent("logout", &uid, "", "success", nil,
				c.ClientIP(), c.GetHeader("User-Agent"), "")
		}
	}
	response.Success(c, nil, "登出成功")
}
// Me 获取当前用户信息。
func (h *Handler) Me(c *gin.Context) {
	val, exists := c.Get(middleware.ContextKeyUserInfo)
	if !exists {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	response.Success(c, val, "success")
}
// ChangePassword 修改当前用户密码。
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if !bindJSONOr422(c, &req) {
		return
	}
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		response.Error(c, http.StatusUnauthorized,
			exception.CodeAuthUnauthorized, "未认证", nil)
		return
	}
	if err := h.svc.ChangePassword(
		uid,
		*req.OldPassword,
		*req.NewPassword,
		*req.ConfirmPassword,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil, "密码修改成功")
}
// handleError 统一处理 PlatformError。
//
// 使用 errors.As 替代类型断言，支持包装错误。
func handleError(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}
