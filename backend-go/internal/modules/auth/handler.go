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
// RegisterPublicRoutes 注册**不需要认证**的路由。
//
// v1.1（P0-01）：拆分出公开路由分组。调用方不得为这些路由挂载
//   AuthMiddleware，否则会破坏登录/刷新流程。
func (h *Handler) RegisterPublicRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
}
// RegisterProtectedRoutes 注册**需要认证**的路由。
//
// v1.1（P0-01）：调用方**必须**在调用本方法前挂载
//   `middleware.AuthMiddleware(db, cfg.SecretKey)`，否则
//   `/me`、`/me/password` 会因为缺少 `user_info` / `user_id` 而返回 401。
func (h *Handler) RegisterProtectedRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/logout", h.Logout)
	auth.GET("/me", h.Me)
	auth.PUT("/me/password", h.ChangePassword)
}
// bindJSONOr422 绑定 JSON 请求体；失败时统一返回 422 / 90004。
//
// v1.1（P1-01）：与 Python Pydantic 的 422 / 90004 语义对齐。
// v1.2（P2-NEW-01）：错误体 `data.errors` 改为结构化数组，与 Python
//   Pydantic 的 `loc / msg / type / input` 字段对齐。
func bindJSONOr422(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(
			c,
			http.StatusUnprocessableEntity,
			exception.CodeValidationFail,
			"请求参数校验失败",
			gin.H{"errors": buildValidationErrors(req, err)},
		)
		return false
	}
	return true
}
// buildValidationErrors 将绑定错误转换为结构化数组。
//
// v1.2（P2-NEW-01）：
//   - 若为 `validator.ValidationErrors`：遍历每个字段错误，
//     通过反射从结构体的 `json` tag 提取 JSON 字段名；
//   - 否则（JSON 解析失败）：返回包含 `invalid_json` 类型的单元素数组。
func buildValidationErrors(req interface{}, err error) []map[string]interface{} {
	var result []map[string]interface{}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			jsonName := toJSONFieldName(req, e.StructField())
			result = append(result, map[string]interface{}{
				"type":  e.Tag(),
				"loc":   []string{"body", jsonName},
				"msg":   e.Error(),
				"input": e.Value(),
			})
		}
		return result
	}
	// 非 validator 错误：JSON 解析失败等
	result = append(result, map[string]interface{}{
		"type": "invalid_json",
		"msg":  err.Error(),
	})
	return result
}
// toJSONFieldName 通过反射从结构体的 `json` tag 提取 JSON 字段名。
//
// 若 tag 为空或为 "-"，回退到 `strings.ToLower(structField)`。
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
//
// v1.1（P2-02）：区分"无 body"与"body 无效"：
//   - 优先从 Authorization 头读取 token；
//   - 否则尝试解析 body；body 存在但无效 → 422/90004；
//   - 都没有 → 401/10001。
func (h *Handler) Refresh(c *gin.Context) {
	var token string
	// 1. 优先从 Authorization 读取
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = strings.TrimSpace(authHeader[7:])
	}
	// 2. 若 header 无 token，尝试从 body 读取
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
//
// v1.1（P0-01 / P2-10）：本方法通过 RegisterProtectedRoutes 注册，
//   已挂载 AuthMiddleware；未认证访问将返回 401/10001。
// v1.2（P2-NEW-04）：body 存在但无效时返回 422 / 90004，
//   与 Python 契约保持一致；无 body 或有效空 body 仍返回 200。
func (h *Handler) Logout(c *gin.Context) {
	var req RefreshReq
	// 区分"无 body"与"body 无效"：
	//   - ContentLength == 0：无 body，跳过绑定；
	//   - ContentLength > 0：尝试绑定，失败返回 422。
	if c.Request.ContentLength > 0 {
		if !bindJSONOr422(c, &req) {
			return
		}
	}
	if req.RefreshToken != "" {
		// 幂等：忽略错误
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
//
// v1.1（P0-01）：通过 RegisterProtectedRoutes 注册；AuthMiddleware
//   已在 context 中写入 `user_info`，此处直接透传。
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
//
// v1.1（P0-01 / P2-01）：
//   - 通过 RegisterProtectedRoutes 注册；AuthMiddleware 保证用户已认证；
//   - 请求体字段为指针类型，binding:"required" 保证字段存在；
//   - 缺失字段返回 422/90004；空字符串进入 Service 返回 10003。
// v1.2（P2-NEW-02）：传入 c.ClientIP() 与 User-Agent 供审计日志使用。
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
func handleError(c *gin.Context, err error) {
	if pe, ok := err.(*exception.PlatformError); ok {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}
