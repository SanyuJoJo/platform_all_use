package crypto
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)
// TaskScheduler 任务调度器。
type TaskScheduler struct {
	db       *gorm.DB
	mu       sync.RWMutex
	cancelMap map[string]context.CancelFunc
}
// NewTaskScheduler 创建任务调度器。
func NewTaskScheduler(db *gorm.DB) *TaskScheduler {
	return &TaskScheduler{
		db:        db,
		cancelMap: make(map[string]context.CancelFunc),
	}
}
// CreateTask 创建任务记录。
func (s *TaskScheduler) CreateTask(
	taskID, operationID, requestID, actorID string,
	timeoutMs int,
) (*models.Task, error) {
	task := &models.Task{
		TaskID:      taskID,
		OperationID: operationID,
		RequestID:   requestID,
		ActorID:     actorID,
		Status:      TaskStatusPending,
		Progress:    intPtr(0),
		TimeoutMs:   &timeoutMs,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.db.Create(task).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("创建任务失败: %v", err),
			500, nil,
		)
	}
	return task, nil
}
// UpdateTaskStatus 更新任务状态。
func (s *TaskScheduler) UpdateTaskStatus(
	taskID, status string,
	progress *int,
	resultRef, errorCode, errorMessage *string,
) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}
	if progress != nil {
		updates["progress"] = *progress
	}
	if resultRef != nil {
		updates["result_ref"] = *resultRef
	}
	if errorCode != nil {
		updates["error_code"] = *errorCode
	}
	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	}
	if status == TaskStatusRunning {
		now := time.Now().UTC()
		updates["started_at"] = now
	}
	if status == TaskStatusSuccess || status == TaskStatusFailed ||
		status == TaskStatusTimeout || status == TaskStatusCancelled {
		now := time.Now().UTC()
		updates["finished_at"] = now
	}
	return s.db.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}
// GetTask 查询任务。
func (s *TaskScheduler) GetTask(taskID string) (*models.Task, error) {
	var task models.Task
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(
				exception.CodeNotFound,
				"任务不存在",
				404, nil,
			)
		}
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("查询任务失败: %v", err),
			500, nil,
		)
	}
	return &task, nil
}
// CancelTask 取消任务。
func (s *TaskScheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	cancel, ok := s.cancelMap[taskID]
	s.mu.Unlock()
	if !ok {
		return exception.New(
			exception.CodeParamInvalid,
			"任务不存在或已结束",
			404, nil,
		)
	}
	cancel()
	// 更新状态为 CANCELLED
	errMsg := "用户取消"
	if err := s.UpdateTaskStatus(
		taskID, TaskStatusCancelled,
		nil, nil, nil, &errMsg,
	); err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("更新任务取消状态失败")
	}
	return nil
}
// RegisterCancel 注册取消函数。
func (s *TaskScheduler) RegisterCancel(taskID string, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelMap[taskID] = cancel
}
// UnregisterCancel 注销取消函数。
func (s *TaskScheduler) UnregisterCancel(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cancelMap, taskID)
}
// ListTasks 查询任务列表。
func (s *TaskScheduler) ListTasks(
	page, pageSize int,
	status string,
) (map[string]interface{}, error) {
	q := s.db.Model(&models.Task{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("查询任务失败: %v", err),
			500, nil,
		)
	}
	var tasks []models.Task
	if err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("查询任务失败: %v", err),
			500, nil,
		)
	}
	pages := 0
	if pageSize > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return map[string]interface{}{
		"items":     tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     pages,
	}, nil
}
