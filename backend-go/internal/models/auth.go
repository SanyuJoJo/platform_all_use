package models
import (
	"time"
	"gorm.io/gorm"
)
// User 用户表（auth_user）。
// 严格对应 Python SQLAlchemy 模型。
type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"size:50;not null;uniqueIndex" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	Nickname     string     `gorm:"size:50;not null" json:"nickname"`
	Email        *string    `gorm:"size:100;index" json:"email"`
	Avatar       *string    `gorm:"size:255" json:"avatar"`
	Status       int8       `gorm:"not null;default:1" json:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	LastLoginIP  *string    `gorm:"size:45" json:"lastLoginIp"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	Roles []Role `gorm:"many2many:auth_user_role;" json:"roles,omitempty"`
}
// TableName 指定表名。
func (User) TableName() string { return "auth_user" }
// Role 角色表（auth_role）。
type Role struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Code        string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Description *string   `gorm:"size:255" json:"description"`
	IsSystem    int8      `gorm:"not null;default:0" json:"isSystem"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	Permissions []Permission `gorm:"many2many:auth_role_permission;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:auth_user_role;" json:"users,omitempty"`
}
// TableName 指定表名。
func (Role) TableName() string { return "auth_role" }
// Permission 权限表（auth_permission）。
type Permission struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"size:100;not null;uniqueIndex" json:"code"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	ModuleID  string    `gorm:"size:50;not null;index" json:"moduleId"`
	Resource  string    `gorm:"size:50;not null" json:"resource"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	Roles []Role `gorm:"many2many:auth_role_permission;" json:"roles,omitempty"`
}
// TableName 指定表名。
func (Permission) TableName() string { return "auth_permission" }
// UserRole 用户-角色关联表（auth_user_role）。
type UserRole struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uq_auth_user_role_user_role" json:"userId"`
	RoleID    uint      `gorm:"not null;uniqueIndex:uq_auth_user_role_user_role" json:"roleId"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
}
// TableName 指定表名。
func (UserRole) TableName() string { return "auth_user_role" }
// RolePermission 角色-权限关联表（auth_role_permission）。
type RolePermission struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID       uint      `gorm:"not null;uniqueIndex:uq_auth_role_permission_role_perm" json:"roleId"`
	PermissionID uint      `gorm:"not null;uniqueIndex:uq_auth_role_permission_role_perm" json:"permissionId"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
}
// TableName 指定表名。
func (RolePermission) TableName() string { return "auth_role_permission" }
// RefreshToken Refresh Token 持久化表（auth_refresh_token）。
type RefreshToken struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"not null;index" json:"userId"`
	TokenHash string     `gorm:"size:64;not null;uniqueIndex" json:"tokenHash"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expiresAt"`
	Revoked   int8       `gorm:"not null;default:0" json:"revoked"`
	RevokedAt *time.Time `json:"revokedAt"`
	IP        *string    `gorm:"size:45" json:"ip"`
	UserAgent *string    `gorm:"size:255" json:"userAgent"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}
// TableName 指定表名。
func (RefreshToken) TableName() string { return "auth_refresh_token" }
// AutoMigrate 自动迁移认证模块表（仅供开发/测试使用，生产使用 goose）。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&UserRole{},
		&RolePermission{},
		&RefreshToken{},
	)
}
