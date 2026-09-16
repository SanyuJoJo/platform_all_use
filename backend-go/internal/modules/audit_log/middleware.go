package audit_log
import (
	"fmt"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"backend-go/internal/middleware"
	"backend-go/internal/models"
	"backend-go/internal/security"
)
// AuditLogMiddleware 全局审计日志中间件。
//
// v1.1（P0-1 / P1-1）：
//   - 必须注册在 Recovery 之外层，确保路由 panic 时本中间件的 defer 逻辑
//     能读取最终响应状态码与 X-Error-Code；
//   - 记录逻辑用 defer 封装，保证任何路径（正常返回、panic、Abort）都能记录；
//   - 异步写入，不阻塞响应。
func AuditLogMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// 1. 跳过白名单与非 /api/v1/ 路径
		if shouldSkip(path) {
			c.Next()
			return
		}
		// 2. 跳过 OPTIONS 预检
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		// 3. 解析路径
		moduleID, resource, resourceID := parsePath(path)
		if moduleID == "" {
			c.Next()
			return
		}
		// 4. 提取用户
		userID, username := extractUserFromContext(c)
		if userID == nil {
			userID, username = extractUserFromToken(c, svc.cfg.SecretKey)
		}
		// 5. 客户端信息
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID := c.GetString(middleware.RequestIDKey)
		// 6. 解析 action
		action := resolveAction(c.Request.Method, path)
		// 7. 构造详情
		detail := fmt.Sprintf("%s %s", c.Request.Method, path)
		// v1.1（P1-1）：defer 记录，保证异常路径不丢失
		defer func() {
			statusCode := c.Writer.Status()
			if statusCode == 0 {
				statusCode = 500
			}
			status := "success"
			var errorCode *int
			if statusCode >= 400 {
				status = "fail"
				if v := c.Writer.Header().Get("X-Error-Code"); v != "" {
					if code, err := strconv.Atoi(v); err == nil {
						errorCode = &code
					}
				}
				if errorCode == nil {
					if code, ok := STATUS_ERROR_MAP[statusCode]; ok {
						errorCode = &code
					} else {
						code := 90000
						errorCode = &code
					}
				}
			}
			entry := &models.AuditLogOperation{
				UserID:     userID,
				Username:   username,
				ModuleID:   moduleID,
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				Detail:     strPtrOrNil(detail),
				IP:         strPtrOrNil(ip),
				UserAgent:  strPtrOrNil(userAgent),
				Status:     status,
				ErrorCode:  errorCode,
				RequestID:  strPtrOrNil(requestID),
			}
			svc.WriteOperationLogAsync(entry)
		}()
		c.Next()
	}
}
// strPtrOrNil 空字符串转 nil，非空返回指针。
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
// shouldSkip 判断路径是否跳过审计。
//
// v1.1（P2-5）：精确匹配（path == prefix 或 path 以 prefix + "/" 开头），
// 避免 `/health` 误匹配 `/health-check`。
//
// v1.2（P1-NEW-1）：测试断言修正后，本函数语义不变：
//   - 白名单前缀使用精确匹配；
//   - 非 `/api/v1/*` 路径一律跳过。
func shouldSkip(path string) bool {
	for _, prefix := range SKIP_PATH_PREFIXES {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	if path != "/api/v1" && !strings.HasPrefix(path, "/api/v1/") {
		return true
	}
	return false
}
// parsePath 从路径中提取 (module_id, resource, resource_id)。
func parsePath(path string) (string, *string, *string) {
	prefix := "/api/v1/"
	if !strings.HasPrefix(path, prefix) {
		return "", nil, nil
	}
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", nil, nil
	}
	pathModule := parts[0]
	moduleID, ok := PATH_MODULE_MAP[pathModule]
	if !ok {
		moduleID = pathModule
	}
	var resource *string
	var resourceID *string
	if len(parts) == 1 {
		resource = &pathModule
	} else {
		r := parts[1]
		resource = &r
		if len(parts) > 2 {
			rid := parts[2]
			resourceID = &rid
		}
	}
	return moduleID, resource, resourceID
}
// resolveAction 解析操作的 action。
func resolveAction(method, path string) string {
	if strings.HasSuffix(path, "/export") {
		return "export"
	}
	if action, ok := METHOD_ACTION_MAP[method]; ok {
		return action
	}
	return strings.ToLower(method)
}
// extractUserFromContext 从 gin.Context 提取已认证用户信息。
func extractUserFromContext(c *gin.Context) (*uint, *string) {
	if v, ok := c.Get(middleware.ContextKeyUserID); ok {
		if uid, ok := v.(uint); ok {
			var username *string
			if u, ok := c.Get(middleware.ContextKeyUserInfo); ok {
				if uc, ok := u.(middleware.UserContext); ok {
					if uc.Username != "" {
						username = &uc.Username
					}
				}
			}
			return &uid, username
		}
	}
	return nil, nil
}
// extractUserFromToken 从 Authorization 头解析 JWT 获取用户信息。
//
// v1.1（P2-8）：仅接受 `""`（兼容旧 token）或 `"access"` 类型；
// 拒绝 refresh token，避免在审计日志中记录非预期的 user_id。
func extractUserFromToken(c *gin.Context, secretKey string) (*uint, *string) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return nil, nil
	}
	token := strings.TrimSpace(authHeader[7:])
	if token == "" {
		return nil, nil
	}
	claims, err := security.ParseToken(secretKey, token)
	if err != nil {
		return nil, nil
	}
	// v1.1（P2-8）：仅接受 access token
	if claims.Type != "" && claims.Type != "access" {
		return nil, nil
	}
	var uid uint
	if _, err := fmt.Sscanf(claims.Subject, "%d", &uid); err != nil {
		return nil, nil
	}
	var username *string
	if claims.Username != "" {
		u := claims.Username
		username = &u
	}
	return &uid, username
}
