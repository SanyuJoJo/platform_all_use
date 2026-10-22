// backend-go/internal/modules/auth/permission_calc.go
//
// 方案 A 的权限计算入口。实际逻辑委托给 internal/rbac，
// 保留 Service 方法是为了已有调用方（serializeUser / GetUserInfo /
// Login / Refresh）零改动。
package auth

import (
	"backend-go/internal/rbac"
)

// LoadUserPermissions 计算用户的权限码集合（方案 A）。
//
// 委托 rbac.LoadUserPermissions：
//   - admin / super_admin：返回 platform_menu 全部非空 perms（不过滤 status）；
//   - 普通用户：auth_role_menu → 双向展开 → platform_menu.perms（status=0）。
func (s *Service) LoadUserPermissions(userID uint) ([]string, error) {
	return rbac.LoadUserPermissions(s.db, userID)
}

// expandMenuIDs 保留为方法（兼容潜在外部调用），委托给 rbac 包。
func (s *Service) expandMenuIDs(rootIDs []uint) ([]uint, error) {
	return rbac.ExpandMenuIDs(s.db, rootIDs)
}
