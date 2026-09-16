package models
import (
	"database/sql/driver"
	"encoding/json"
	"time"
	"gorm.io/gorm"
)
// StringArray 字符串数组，序列化为 JSON 存储。
//
// 对应 Python SQLAlchemy 的 JSON 字段（authorized_modules）。
// 支持 nil → SQL NULL，非 nil → JSON 数组字符串。
type StringArray []string
// Value 实现 driver.Valuer。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}
// Scan 实现 sql.Scanner。
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
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
		*a = nil
		return nil
	}
	return json.Unmarshal(bytes, a)
}
// License License 表（license_license）。
//
// 严格对应 Python SQLAlchemy 模型 src/modules/license/models.py。
// 部分唯一索引 uq_license_license_active（WHERE is_active = 1）由 goose
// 迁移脚本创建；GORM AutoMigrate 不支持带 WHERE 子句的唯一索引。
type License struct {
	ID                uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	LicenseKey        string      `gorm:"size:255;not null;uniqueIndex:ix_license_license_license_key" json:"license_key"`
	LicenseType       string      `gorm:"size:50;not null" json:"license_type"`
	MaxUsers          *int        `json:"max_users"`
	AuthorizedModules StringArray `gorm:"type:json" json:"authorized_modules"`
	MachineCode       *string     `gorm:"size:255" json:"machine_code"`
	IssuedAt          time.Time   `gorm:"not null" json:"issued_at"`
	ExpiresAt         time.Time   `gorm:"not null;index:ix_license_license_expires_at" json:"expires_at"`
	IsActive          int8        `gorm:"not null;default:1" json:"is_active"`
	ActivatedAt       *time.Time  `json:"activated_at"`
	CreatedAt         time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
// TableName 指定表名。
func (License) TableName() string { return "license_license" }
// EnsureLicenseTable 开发环境确保 License 表存在。
//
// 注意：部分唯一索引 uq_license_license_active 无法通过 GORM AutoMigrate
// 创建，必须通过 goose 迁移脚本创建。
func EnsureLicenseTable(db *gorm.DB) error {
	return db.AutoMigrate(&License{})
}
