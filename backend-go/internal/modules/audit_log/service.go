package audit_log
import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)
// Service 日志审计服务。
type Service struct {
	db  *gorm.DB
	cfg *config.Config
	wg  sync.WaitGroup
}
// NewService 创建日志审计服务。
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}
// Wait 等待所有异步写入任务完成（供优雅关闭调用）。
func (s *Service) Wait() {
	s.wg.Wait()
}
// WriteOperationLogAsync 异步写入一条操作日志。
//
// 调用方负责将空字符串字段归一化为 nil（通过 strPtrOrNil）。
func (s *Service) WriteOperationLogAsync(entry *models.AuditLogOperation) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
			log.Error().Err(err).Msg("写入审计日志失败")
		}
	}()
}
// WriteOperationLog 同步写入一条操作日志（供测试使用）。
func (s *Service) WriteOperationLog(entry *models.AuditLogOperation) error {
	return s.db.Create(entry).Error
}
// ListAuditLogs 分页查询操作日志。
//
// v1.1（P1-3）：Count 后使用 Session(&gorm.Session{}) 隔离，避免
// Statement 污染导致的 Order / Limit 失效。
//
// v1.2（P2-NEW-1）：分页参数从 Handler 层传入的非 nil 值展开。
func (s *Service) ListAuditLogs(q AuditLogListQuery) (map[string]interface{}, error) {
	page := 1
	if q.Page != nil {
		page = *q.Page
	}
	pageSize := 20
	if q.PageSize != nil {
		pageSize = *q.PageSize
	}
	startTime, err := ParseTime(q.StartTime)
	if err != nil {
		return nil, exception.New(exception.CodeValidationFail, "请求参数校验失败", 422,
			map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"type": "datetime_parsing",
						"loc":  []string{"query", "start_time"},
						"msg":  fmt.Sprintf("时间格式无效：%s", q.StartTime),
					},
				},
			})
	}
	endTime, err := ParseTime(q.EndTime)
	if err != nil {
		return nil, exception.New(exception.CodeValidationFail, "请求参数校验失败", 422,
			map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"type": "datetime_parsing",
						"loc":  []string{"query", "end_time"},
						"msg":  fmt.Sprintf("时间格式无效：%s", q.EndTime),
					},
				},
			})
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, exception.New(exception.CodeAuditLogInvalidTimeRange,
			"start_time 不能晚于 end_time", 400, nil)
	}
	query := s.db.Model(&models.AuditLogOperation{})
	if q.ModuleID != "" {
		query = query.Where("module_id = ?", q.ModuleID)
	}
	if q.UserID != nil {
		query = query.Where("user_id = ?", *q.UserID)
	}
	if q.Action != "" {
		query = query.Where("action = ?", q.Action)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		query = query.Where(
			"detail LIKE ? OR username LIKE ? OR resource LIKE ?",
			like, like, like,
		)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询日志失败", 500, nil)
	}
	var logs []models.AuditLogOperation
	// v1.1（P1-3）：Session 隔离，避免 Count 修改 Statement 影响 Find
	if err := query.Session(&gorm.Session{}).
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询日志失败", 500, nil)
	}
	items := make([]AuditLogOut, 0, len(logs))
	for i := range logs {
		items = append(items, *serializeAuditLog(&logs[i]))
	}
	pages := 0
	if pageSize > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     pages,
	}, nil
}
// GetAuditLogDetail 获取日志详情。
func (s *Service) GetAuditLogDetail(logID uint) (*AuditLogOut, error) {
	var entry models.AuditLogOperation
	if err := s.db.First(&entry, logID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeAuditLogNotFound, "日志不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询日志失败", 500, nil)
	}
	return serializeAuditLog(&entry), nil
}
// ExportAuditLogs 导出操作日志。
// 返回：内容、MediaType、文件名。
func (s *Service) ExportAuditLogs(q AuditLogExportQuery) ([]byte, string, string, error) {
	// v1.1（P2-4）：统一小写
	format := strings.ToLower(q.Format)
	if format != ExportFormatCSV && format != ExportFormatJSON {
		return nil, "", "", exception.New(
			exception.CodeAuditLogExportFormat,
			fmt.Sprintf("导出格式不支持：%s，仅支持 csv/json", q.Format),
			400,
			nil,
		)
	}
	startTime, err := ParseTime(q.StartTime)
	if err != nil {
		return nil, "", "", exception.New(exception.CodeValidationFail, "请求参数校验失败", 422,
			map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"type": "datetime_parsing",
						"loc":  []string{"query", "start_time"},
						"msg":  fmt.Sprintf("时间格式无效：%s", q.StartTime),
					},
				},
			})
	}
	endTime, err := ParseTime(q.EndTime)
	if err != nil {
		return nil, "", "", exception.New(exception.CodeValidationFail, "请求参数校验失败", 422,
			map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"type": "datetime_parsing",
						"loc":  []string{"query", "end_time"},
						"msg":  fmt.Sprintf("时间格式无效：%s", q.EndTime),
					},
				},
			})
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, "", "", exception.New(exception.CodeAuditLogInvalidTimeRange,
			"start_time 不能晚于 end_time", 400, nil)
	}
	query := s.db.Model(&models.AuditLogOperation{})
	if q.ModuleID != "" {
		query = query.Where("module_id = ?", q.ModuleID)
	}
	if q.UserID != nil {
		query = query.Where("user_id = ?", *q.UserID)
	}
	if q.Action != "" {
		query = query.Where("action = ?", q.Action)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		query = query.Where(
			"detail LIKE ? OR username LIKE ? OR resource LIKE ?",
			like, like, like,
		)
	}
	var logs []models.AuditLogOperation
	if err := query.Session(&gorm.Session{}).
		Order("created_at DESC, id DESC").
		Limit(ExportMaxRows).
		Find(&logs).Error; err != nil {
		return nil, "", "", exception.New(exception.CodeAuditLogExportFailed, "日志导出失败", 500, nil)
	}
	timestamp := time.Now().UTC().Format("20060102_150405")
	if format == ExportFormatCSV {
		content, err := toCSV(logs)
		if err != nil {
			return nil, "", "", exception.New(
				exception.CodeAuditLogExportFailed,
				fmt.Sprintf("日志导出失败：%v", err),
				500, nil,
			)
		}
		return content, "text/csv; charset=utf-8", fmt.Sprintf("audit_logs_%s.csv", timestamp), nil
	}
	content, err := toJSON(logs)
	if err != nil {
		return nil, "", "", exception.New(
			exception.CodeAuditLogExportFailed,
			fmt.Sprintf("日志导出失败：%v", err),
			500, nil,
		)
	}
	return content, "application/json; charset=utf-8", fmt.Sprintf("audit_logs_%s.json", timestamp), nil
}
// ParseTime 解析时间字符串，支持 RFC3339 与无时区 ISO 格式。
func ParseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			t = t.UTC()
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid time format: %s", s)
}
// serializeAuditLog 将 ORM 对象序列化为响应结构。
func serializeAuditLog(entry *models.AuditLogOperation) *AuditLogOut {
	return &AuditLogOut{
		ID:         entry.ID,
		UserID:     entry.UserID,
		Username:   entry.Username,
		ModuleID:   entry.ModuleID,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		Detail:     entry.Detail,
		IP:         entry.IP,
		UserAgent:  entry.UserAgent,
		Status:     entry.Status,
		ErrorCode:  entry.ErrorCode,
		RequestID:  entry.RequestID,
		CreatedAt:  formatDateTime(entry.CreatedAt),
	}
}
// toCSV 生成 CSV 内容（含 UTF-8 BOM），全部交由 encoding/csv 按 RFC 4180 转义。
//
// v1.1（P2-2）：逐条检查 w.Write 返回错误。
func toCSV(logs []models.AuditLogOperation) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(CSV_BOM)
	w := csv.NewWriter(buf)
	if err := w.Write([]string{
		"id", "user_id", "username", "module_id", "action", "resource",
		"resource_id", "detail", "ip", "user_agent", "status",
		"error_code", "request_id", "created_at",
	}); err != nil {
		return nil, err
	}
	for _, entry := range logs {
		record := []string{
			strconv.FormatUint(uint64(entry.ID), 10),
			uintPtrToString(entry.UserID),
			strPtrToString(entry.Username),
			entry.ModuleID,
			entry.Action,
			strPtrToString(entry.Resource),
			strPtrToString(entry.ResourceID),
			strPtrToString(entry.Detail),
			strPtrToString(entry.IP),
			strPtrToString(entry.UserAgent),
			entry.Status,
			intPtrToString(entry.ErrorCode),
			strPtrToString(entry.RequestID),
			formatDateTime(entry.CreatedAt),
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
// toJSON 生成 JSON 内容。
func toJSON(logs []models.AuditLogOperation) ([]byte, error) {
	items := make([]AuditLogOut, 0, len(logs))
	for i := range logs {
		items = append(items, *serializeAuditLog(&logs[i]))
	}
	return json.MarshalIndent(map[string]interface{}{
		"total": len(items),
		"items": items,
	}, "", "  ")
}
func strPtrToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func uintPtrToString(p *uint) string {
	if p == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*p), 10)
}
func intPtrToString(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}
