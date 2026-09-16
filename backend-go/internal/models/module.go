package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Module 模块主表（module_manager_module）。
//
// 严格对应 Python SQLAlchemy 模型 src/modules/module_manager/models.py。
//
// BUG-08 修复：移除 `constraint:OnDelete:CASCADE` 标签。
// 原因：glebarez/sqlite 驱动在处理该标签时会尝试重建表并添加外键，
// 但 SQLite 不支持原生 ALTER TABLE ADD CONSTRAINT，GORM 在拼接
// 临时表列定义时会将 FOREIGN KEY 误识别为列名，导致
// "table module_manager_dependency__temp has no column named FOREIGN"。
//
// 外键与 ON DELETE CASCADE 由 goose 迁移脚本
// migrations/00003_module_manager.sql 保证，GORM 层无需重复声明。
type Module struct {
	ID            string    `gorm:"primaryKey;size:50" json:"id"`
	Name          string    `gorm:"size:50;not null" json:"name"`
	Version       string    `gorm:"size:20;not null" json:"version"`
	Description   *string   `gorm:"type:text" json:"description"`
	Author        *string   `gorm:"size:100" json:"author"`
	Homepage      *string   `gorm:"size:255" json:"homepage"`
	Status        string    `gorm:"size:20;not null;default:inactive;index" json:"status"`
	EntryBackend  string    `gorm:"size:100;not null" json:"entry_backend"`
	EntryFrontend *string   `gorm:"size:255" json:"entry_frontend"`
	Config        JSONMap   `gorm:"type:json" json:"config"`
	Manifest      JSONMap   `gorm:"type:json" json:"manifest"`
	InstalledAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"installed_at"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Dependencies 模块依赖列表。
	//
	// BUG-08 修复：保留 foreignKey / references 供 GORM Preload 使用，
	// 移除 constraint:OnDelete:CASCADE（由 goose 迁移脚本负责）。
	Dependencies []ModuleDependency `gorm:"foreignKey:ModuleID;references:ID" json:"-"`
}

// TableName 指定表名。
func (Module) TableName() string { return "module_manager_module" }

// ModuleDependency 模块依赖表（module_manager_dependency）。
type ModuleDependency struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ModuleID     string    `gorm:"size:50;not null;index" json:"module_id"`
	DependencyID string    `gorm:"size:50;not null;index" json:"dependency_id"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名。
func (ModuleDependency) TableName() string { return "module_manager_dependency" }

// JSONMap 通用 JSON 字段类型。
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner。
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	if len(bytes) == 0 {
		*m = nil
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// EnsureModuleTables 确保模块管理表存在（仅供开发/测试使用）。
//
// BUG-08 修复：分开迁移两张表，避免 GORM 在处理关联表时
// 尝试重建 module_manager_dependency 添加外键。
func EnsureModuleTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&Module{}); err != nil {
		return err
	}
	return db.AutoMigrate(&ModuleDependency{})
}
