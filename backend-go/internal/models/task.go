package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 异步任务表（platform_task）。
//
// 用于密码操作模块的长任务封装：
//   - task_id 由平台生成，贯穿任务全生命周期；
//   - request_id 关联平台请求与 core 调用；
//   - 状态机：PENDING → RUNNING → SUCCESS/FAILED/TIMEOUT/CANCELLED。
type Task struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID       string     `gorm:"size:36;not null;uniqueIndex" json:"task_id"`
	OperationID  string     `gorm:"size:50;not null;index" json:"operation_id"`
	RequestID    string     `gorm:"size:36;not null;index" json:"request_id"`
	ActorID      string     `gorm:"size:50;not null;index" json:"actor_id"`
	Status       string     `gorm:"size:20;not null;default:PENDING;index" json:"status"`
	Progress     *int       `json:"progress"`
	ResultRef    *string    `gorm:"size:255" json:"result_ref"`
	ErrorCode    *string    `gorm:"size:50" json:"error_code"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message"`
	TimeoutMs    *int       `json:"timeout_ms"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	UpdatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名。
func (Task) TableName() string { return "platform_task" }

// EnsureTaskTable 确保密码操作任务表存在（仅供开发/测试使用）。
//
// 生产环境请使用 goose 迁移脚本 migrations/00006_platform_task.sql。
func EnsureTaskTable(db *gorm.DB) error {
	return db.AutoMigrate(&Task{})
}
