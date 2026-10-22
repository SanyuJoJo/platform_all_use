package models

import (
	"time"

	"gorm.io/gorm"
)

// AuthRoleMenu 角色-菜单关联表（auth_role_menu）。
//
// 语义：
//   - 一个角色勾选了哪些菜单，即拥有这些菜单（及其子孙）的 perms；
//   - 权限计算时从勾选的菜单出发双向展开（见 auth.permission_calc.go）；
//   - 与 auth_role_permission 的关系：方案 A 上线后，后者仅做归档，
//     代码不再读写。
type AuthRoleMenu struct {
	RoleID    uint      `gorm:"primaryKey;not null" json:"roleId"`
	MenuID    uint      `gorm:"primaryKey;not null;index" json:"menuId"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName 指定表名。
func (AuthRoleMenu) TableName() string { return "auth_role_menu" }

// EnsureAuthRoleMenuTable 开发环境确保表存在（不含外键，供 AutoMigrate 使用）。
func EnsureAuthRoleMenuTable(db *gorm.DB) error {
	return db.AutoMigrate(&AuthRoleMenu{})
}
