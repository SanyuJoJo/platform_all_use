package middleware
import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"backend-go/internal/exception"
	"backend-go/internal/response"
)
// RoutePermissionLookup 由 route_perm.Service 实现，避免 middleware 直接依赖具体包。
type RoutePermissionLookup interface {
	LookupCodes(method, pattern string) ([]string, error)
}
// DynamicPermission 动态权限校验中间件。
//
// 行为：
//   - 超管（admin / super_admin）直接放行；
//   - c.FullPath() 为空（未匹配到路由）→ 放行，由 NoRoute 处理；
//   - 查表未配置（codes 为空）→ 放行（兼容模式）；
//   - 查表出错 → 放行（fail-open），并交由 AccessLog 记录；
//   - 已配置 → 用户需持有其中任一权限码，否则 403。
//
// 使用位置：AuthMiddleware 之后，业务路由之前。
func DynamicPermission(lookup RoutePermissionLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		uc, ok := GetCurrentUser(c)
		if ok && isAdminUserContext(uc) {
			c.Next()
			return
		}
		pattern := c.FullPath()
		if pattern == "" {
			c.Next()
			return
		}
		codes, err := lookup.LookupCodes(c.Request.Method, pattern)
		if err != nil {
			// fail-open：查表失败不影响主流程
			c.Next()
			return
		}
		if len(codes) == 0 {
			c.Next()
			return
		}
		if !ok {
			response.Error(c, http.StatusUnauthorized, exception.CodeAuthUnauthorized, "未认证", nil)
			c.Abort()
			return
		}
		permSet := make(map[string]struct{}, len(uc.Permissions))
		for _, p := range uc.Permissions {
			permSet[p] = struct{}{}
		}
		for _, code := range codes {
			if _, hit := permSet[code]; hit {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, exception.CodePermissionDenied, "无权限访问", nil)
		c.Abort()
	}
}
// isAdminUserContext 与 menu.isAdminUser 语义保持一致。
func isAdminUserContext(uc *UserContext) bool {
	if uc == nil {
		return false
	}
	if strings.EqualFold(uc.Username, "admin") {
		return true
	}
	for _, r := range uc.Roles {
		lr := strings.ToLower(r)
		if lr == "admin" || lr == "super_admin" {
			return true
		}
	}
	return false
}
