// backend-go/internal/rbac/permission.go
//
// 用户权限码计算（方案 A：角色-菜单绑定）。
//
// 放在独立包中是因为：
//   - auth 包需要它（登录/刷新/用户信息）
//   - middleware 包需要它（每个受保护请求构建 UserContext）
//   - 但 auth 与 middleware 互相依赖会形成循环，故抽出中立包。
package rbac

import (
	"backend-go/internal/exception"
	"backend-go/internal/models"

	"gorm.io/gorm"
)

// LoadUserPermissions 计算用户的权限码集合（方案 A）。
//
// 两个分支：
//  1. admin / super_admin：返回 platform_menu 里所有非空 perms，
//     **不过滤 status**——超管应当拥有全部权限（包括已停用的按钮），
//     保证系统在任何数据状态下都能被管理员修复。
//  2. 普通用户：auth_role_menu → ExpandMenuIDs → platform_menu.perms，
//     过滤 status=0。
func LoadUserPermissions(db *gorm.DB, userID uint) ([]string, error) {
	// ---- 分支 1：admin / super_admin ----
	if IsAdminUser(db, userID) {
		var perms []string
		err := db.Model(&models.Menu{}).
			Where("perms <> ''"). // 不过滤 status
			Distinct().
			Pluck("perms", &perms).Error
		if err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				"查询超管权限码失败："+err.Error(), 500, nil,
			)
		}
		if perms == nil {
			perms = []string{}
		}
		return perms, nil
	}

	// ---- 分支 2：普通用户 ----
	var menuIDs []uint
	err := db.Table("auth_role_menu").
		Select("DISTINCT auth_role_menu.menu_id").
		Joins("JOIN auth_user_role ON auth_user_role.role_id = auth_role_menu.role_id").
		Where("auth_user_role.user_id = ?", userID).
		Pluck("auth_role_menu.menu_id", &menuIDs).Error
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			"查询用户菜单失败："+err.Error(), 500, nil,
		)
	}
	if len(menuIDs) == 0 {
		return []string{}, nil
	}

	allIDs, err := ExpandMenuIDs(db, menuIDs)
	if err != nil {
		return nil, err
	}
	if len(allIDs) == 0 {
		return []string{}, nil
	}

	var perms []string
	err = db.Model(&models.Menu{}).
		Where("id IN ? AND perms <> '' AND status = 0", allIDs).
		Distinct().
		Pluck("perms", &perms).Error
	if err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			"查询权限码失败："+err.Error(), 500, nil,
		)
	}
	if perms == nil {
		perms = []string{}
	}
	return perms, nil
}

// IsAdminUser 判断用户是否属于 admin / super_admin 角色。
func IsAdminUser(db *gorm.DB, userID uint) bool {
	var count int64
	if err := db.Table("auth_user_role AS ur").
		Joins("JOIN auth_role AS r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.code IN ?", userID, []string{"admin", "super_admin"}).
		Count(&count).Error; err == nil && count > 0 {
		return true
	}
	var username string
	if err := db.Model(&models.User{}).
		Where("id = ?", userID).
		Pluck("username", &username).Error; err == nil {
		if username == "admin" {
			return true
		}
	}
	return false
}

// ExpandMenuIDs 从勾选的菜单出发，计算所有相关的菜单 ID。
//
// 语义（v5）：
//   - 向下展开：**仅 M（目录）类型**向下展开一层到 C，不再继续展开 C 下的 F。
//     勾 M = 该目录下所有页面可见（view），但无具体操作按钮权限。
//   - C（菜单）和 F（按钮）不向下展开：
//     勾 C = 只有该页面的 view 权限，无操作按钮权限；
//     勾 F = 只有该按钮权限（+ 向上补 C 的 view）。
//   - 向上展开：任何节点都向上补直接祖先链的 perms（补 C 的 view）。
//     **向上不展开祖先的其他子孙**，避免"勾一个叶子节点扩散到整个树"的权限泄漏。
//
// 设计意图：
//   管理员勾选菜单时，语义是"授予该菜单自身的权限"。
//   要"页面可见 + 该页所有按钮可用"，需同时勾选页面（C）及其下所有按钮（F）。
//   这避免了"勾 C 自动获得所有 F"导致无法实现"只读角色"
//   （只有 view 没有编辑/删除按钮）。
//
// 若前端希望"勾 C 时自动勾选所有 F"作为 UX 增强，应在树形组件层面处理
// （el-tree 的 check-strictly=false 级联），后端语义保持精确。
func ExpandMenuIDs(db *gorm.DB, rootIDs []uint) ([]uint, error) {
	if len(rootIDs) == 0 {
		return []uint{}, nil
	}

	visited := make(map[uint]bool, len(rootIDs))
	result := make([]uint, 0, len(rootIDs))

	// 阶段 1：向下展开（仅 M 类型）
	queue := append([]uint(nil), rootIDs...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		visited[id] = true
		result = append(result, id)

		// 检查节点类型：只有 M（目录）向下展开
		var menuType string
		if err := db.Model(&models.Menu{}).
			Select("menu_type").
			Where("id = ?", id).
			Scan(&menuType).Error; err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				"查询菜单类型失败："+err.Error(), 500, nil,
			)
		}
		if menuType != "M" {
			continue // C/F 不向下展开
		}

		var childIDs []uint
		if err := db.Model(&models.Menu{}).
			Where("parent_id = ?", id).
			Pluck("id", &childIDs).Error; err != nil {
			return nil, exception.New(
				exception.CodeInternalError,
				"展开子菜单失败："+err.Error(), 500, nil,
			)
		}
		queue = append(queue, childIDs...)
	}

	// 阶段 2：向上补祖先自身（不展开祖先的其他子孙）
	for _, rootID := range rootIDs {
		cur := rootID
		for {
			var parentID uint
			if err := db.Model(&models.Menu{}).
				Select("parent_id").
				Where("id = ?", cur).
				Scan(&parentID).Error; err != nil {
				return nil, exception.New(
					exception.CodeInternalError,
					"查询父菜单失败："+err.Error(), 500, nil,
				)
			}
			if parentID == 0 {
				break
			}
			if visited[parentID] {
				break // 已经处理过
			}
			visited[parentID] = true
			result = append(result, parentID)
			cur = parentID
		}
	}

	return result, nil
}
