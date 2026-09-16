package middleware
import (
	"fmt"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/response"
	"backend-go/internal/security"
)
// ContextKeyUserID gin.Context 中用户 ID 的键。
const ContextKeyUserID = "user_id"
// ContextKeyUserInfo gin.Context 中用户信息的键。
const ContextKeyUserInfo = "user_info"
// UserContext 当前用户上下文。
//
// 字段名与 `auth.UserInfo` 保持一致，JSON 序列化结果相同。
type UserContext struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Email       *string  `json:"email"`
	Avatar      *string  `json:"avatar"`
	Status      int8     `json:"status"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
// AuthMiddleware 认证中间件：解析 JWT，加载用户与权限。
//
// 用法（在 main.go 中）：
//   protected := api.Group("")
//   protected.Use(middleware.AuthMiddleware(db, cfg.SecretKey))
//   authHandler.RegisterProtectedRoutes(protected)
func AuthMiddleware(db *gorm.DB, secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "未提供认证令牌", nil)
			c.Abort()
			return
		}
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "无效的认证格式", nil)
			c.Abort()
			return
		}
		token := strings.TrimSpace(authHeader[7:])
		if token == "" {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "未提供认证令牌", nil)
			c.Abort()
			return
		}
		claims, err := security.ParseToken(secretKey, token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "Token无效或已过期", nil)
			c.Abort()
			return
		}
		if claims.Type != "access" && claims.Type != "" {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "无效的访问令牌", nil)
			c.Abort()
			return
		}
		var uid uint
		if _, err := fmt.Sscanf(claims.Subject, "%d", &uid); err != nil {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "无效的Token", nil)
			c.Abort()
			return
		}
		// 加载用户与 RBAC
		var user models.User
		if err := db.Preload("Roles.Permissions").First(&user, uid).Error; err != nil {
			response.Error(c, http.StatusNotFound,
				exception.CodeAuthUserNotFound, "用户不存在", nil)
			c.Abort()
			return
		}
		if user.Status != 1 {
			response.Error(c, http.StatusForbidden,
				exception.CodeAuthForbidden, "用户已被禁用", nil)
			c.Abort()
			return
		}
		// 构建权限集合
		roles := make([]string, 0, len(user.Roles))
		permSet := make(map[string]struct{})
		for _, role := range user.Roles {
			roles = append(roles, role.Code)
			for _, perm := range role.Permissions {
				permSet[perm.Code] = struct{}{}
			}
		}
		permissions := make([]string, 0, len(permSet))
		for code := range permSet {
			permissions = append(permissions, code)
		}
		uc := UserContext{
			ID:          user.ID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Email:       user.Email,
			Avatar:      user.Avatar,
			Status:      user.Status,
			Roles:       roles,
			Permissions: permissions,
		}
		c.Set(ContextKeyUserID, user.ID)
		c.Set(ContextKeyUserInfo, uc)
		c.Next()
	}
}
// RequirePermission 权限校验中间件工厂。
func RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(ContextKeyUserInfo)
		if !exists {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "未认证", nil)
			c.Abort()
			return
		}
		uc, ok := val.(UserContext)
		if !ok {
			response.Error(c, http.StatusUnauthorized,
				exception.CodeAuthUnauthorized, "未认证", nil)
			c.Abort()
			return
		}
		for _, p := range uc.Permissions {
			if p == permissionCode {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden,
			exception.CodePermissionDenied, "无权限访问", nil)
		c.Abort()
	}
}
// GetCurrentUser 从 gin.Context 获取当前用户。
func GetCurrentUser(c *gin.Context) (*UserContext, bool) {
	val, exists := c.Get(ContextKeyUserInfo)
	if !exists {
		return nil, false
	}
	uc, ok := val.(UserContext)
	if !ok {
		return nil, false
	}
	return &uc, true
}
