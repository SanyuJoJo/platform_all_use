package models

import (
	"time"

	"gorm.io/gorm"
)

// Menu 菜单/路由元数据表（platform_menu）。
//
// 对标若依 sys_menu，是前端动态路由、侧边栏菜单与按钮权限的单一数据源。
//
// 字段语义：
//   - MenuType：M=目录，C=菜单，F=按钮；
//   - Name    ：英文路由名（PascalCase），全局唯一，供 Vue Router 使用；
//   - MicroApp：所属微应用 id（main/auth/module_manager/audit_log/license/
//               crypto_console/ssl_proxy），用于 qiankun 子应用路由分发；
//   - Perms   ：与 auth_permission.code 对齐的三段式权限码，
//              为空表示所有登录用户可见；
//   - Path    ：路由地址（M/C 必填，F 忽略）；
//   - Component：Vue 组件相对路径（相对 views/），M 目录固定为 Layout；
//   - IsFrame ：是否外链，0=是，1=否；
//   - IsCache ：是否 keep-alive，0=缓存，1=不缓存；
//   - Visible ：侧边栏是否显示，0=显示，1=隐藏；
//   - Status  ：菜单状态，0=正常，1=停用（停用后 getRouters 不返回）。
//
// 层级关系由 ParentID 自关联；删除由服务层校验无子节点。
type Menu struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID   uint      `gorm:"not null;default:0;index" json:"parentId"`
	MenuName   string    `gorm:"size:50;not null" json:"menuName"`
	Name       string    `gorm:"size:100;not null;default:'';index" json:"name"`
	MicroApp   string    `gorm:"size:50;not null;default:'main';index" json:"microApp"`
	OrderNum   int       `gorm:"not null;default:0" json:"orderNum"`
	Path       string    `gorm:"size:200;not null;default:''" json:"path"`
	Component  string    `gorm:"size:255;not null;default:''" json:"component"`
	Query      string    `gorm:"size:255;not null;default:''" json:"query"`
	IsFrame    int8      `gorm:"not null;default:1" json:"isFrame"`
	IsCache    int8      `gorm:"not null;default:0" json:"isCache"`
	MenuType   string    `gorm:"size:1;not null;index" json:"menuType"`
	Visible    int8      `gorm:"not null;default:0" json:"visible"`
	Status     int8      `gorm:"not null;default:0;index" json:"status"`
	Perms      string    `gorm:"size:100;not null;default:'';index" json:"perms"`
	Icon       string    `gorm:"size:100;not null;default:'#'" json:"icon"`
	ModuleID   string    `gorm:"size:50;not null;default:'';index" json:"moduleId"`
	CreateBy   string    `gorm:"size:64;not null;default:''" json:"createBy"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"createTime"`
	UpdateBy   string    `gorm:"size:64;not null;default:''" json:"updateBy"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updateTime"`
	Remark     string    `gorm:"size:500;not null;default:''" json:"remark"`

	// Children 仅内存使用，由服务层构建。
	Children []*Menu `gorm:"-" json:"children,omitempty"`
}

// TableName 指定表名。
func (Menu) TableName() string { return "platform_menu" }

// EnsureMenuSchema 开发环境确保菜单表存在。
func EnsureMenuSchema(db *gorm.DB) error {
	return db.AutoMigrate(&Menu{})
}
