package models
import (
	"time"
	"gorm.io/gorm"
)
// AuditLogOperation 操作日志表（audit_log_operation）。
//
// 严格对应 Python SQLAlchemy 模型 src/modules/audit_log/models.py。
// 审计日志不可修改，因此不含 updated_at 字段。
//
// 外键说明：
//   - user_id → auth_user.id ON DELETE SET NULL
//   - 生产环境由 goose 迁移脚本 migrations/00004_audit_log.sql 创建；
//   - 开发环境 EnsureAuditLogTable 仅调用 AutoMigrate 保证表存在，
//     **不补充外键**（GORM AutoMigrate 不创建外键）。
//     开发环境如需完整外键，请执行：make migrate-up
type AuditLogOperation struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     *uint     `gorm:"index" json:"user_id"`
	Username   *string   `gorm:"size:50" json:"username"`
	ModuleID   string    `gorm:"size:50;not null;index" json:"module_id"`
	Action     string    `gorm:"size:50;not null;index" json:"action"`
	Resource   *string   `gorm:"size:50" json:"resource"`
	ResourceID *string   `gorm:"size:50" json:"resource_id"`
	Detail     *string   `gorm:"type:text" json:"detail"`
	IP         *string   `gorm:"size:45" json:"ip"`
	UserAgent  *string   `gorm:"size:255" json:"user_agent"`
	Status     string    `gorm:"size:20;not null;default:success" json:"status"`
	ErrorCode  *int      `json:"error_code"`
	RequestID  *string   `gorm:"size:36" json:"request_id"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}
// TableName 指定表名。
func (AuditLogOperation) TableName() string { return "audit_log_operation" }
// EnsureAuditLogTable 确保审计日志表存在（仅供开发/测试使用）。
//
// v1.2（P1-NEW-2）：
//   - 删除 v1.1 中的无效空操作代码（仅查询 pragma_foreign_key_list
//     但未做任何实际 DDL 补充，属于死代码）；
//   - 本函数只保证表存在，**不补充外键**；
//   - 生产环境请使用 goose 迁移脚本 migrations/00004_audit_log.sql
//     以保证外键 user_id → auth_user.id ON DELETE SET NULL 与 Python 版一致；
//   - 开发环境如需完整外键，请执行：make migrate-up
func EnsureAuditLogTable(db *gorm.DB) error {
	return db.AutoMigrate(&AuditLogOperation{})
}
