package models
import (
	"time"
	"gorm.io/gorm"
)
// RoutePermission 路由-权限映射表（platform_route_permission）。
//
// 解决"接口权限校验与 Gin 路由注册解耦"的问题：
//   - 路由本身仍在编译期注册到 Gin；
//   - 但"某个路由需要什么权限"由本表决定，可在运行时增删改；
//   - DynamicPermission 中间件在每次请求时查表（带缓存）完成校验。
//
// 约定：
//   - HTTPMethod 大写（GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS）；
//   - RoutePattern 必须与 c.FullPath() 一致（含 :param 占位符）；
//   - PermissionCode 对应 auth_permission.code；
//   - 同 (method, pattern) 可配置多个权限码，满足"或"关系（任一即可）。
type RoutePermission struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	HTTPMethod     string    `gorm:"size:10;not null" json:"httpMethod"`
	RoutePattern   string    `gorm:"size:255;not null;index:idx_rp_method_pattern" json:"routePattern"`
	PermissionCode string    `gorm:"size:100;not null;index" json:"permissionCode"`
	ModuleID       string    `gorm:"size:50;not null;default:'';index" json:"moduleId"`
	Status         int8      `gorm:"not null;default:0;index" json:"status"`
	Description    string    `gorm:"size:255;not null;default:''" json:"description"`
	CreatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}
// TableName 指定表名。
func (RoutePermission) TableName() string { return "platform_route_permission" }
// EnsureRoutePermissionSchema 开发环境确保路由权限表存在。
func EnsureRoutePermissionSchema(db *gorm.DB) error {
	return db.AutoMigrate(&RoutePermission{})
}
