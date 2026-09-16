package audit_log
import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"backend-go/internal/exception"
	"backend-go/internal/middleware"
	"backend-go/internal/response"
)
// Handler 日志审计路由处理器。
type Handler struct {
	svc *Service
}
// NewHandler 创建日志审计路由处理器。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}
// RegisterRoutes 注册日志审计路由（受保护）。
//
// 路由顺序：/export 必须在 /:log_id 之前注册。
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	logs := r.Group("/audit-logs")
	logs.GET("", middleware.RequirePermission("audit_log:log:view"), h.ListAuditLogs)
	logs.GET("/export", middleware.RequirePermission("audit_log:log:export"), h.ExportAuditLogs)
	logs.GET("/:log_id", middleware.RequirePermission("audit_log:log:view"), h.GetAuditLogDetail)
}
// ListAuditLogs 查询操作日志。
//
// v1.2（P2-NEW-1）：分页参数改为 *int，显式判空后赋默认值：
//   - Page == nil → 使用 1；
//   - PageSize == nil → 使用 20；
//   - Page / PageSize 非 nil 时（含显式传 0）已在 binding 阶段校验 min=1。
func (h *Handler) ListAuditLogs(c *gin.Context) {
	var q AuditLogListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(&q, err, "query")})
		return
	}
	// 时间格式预校验
	if errs := validateTimeParams(q.StartTime, q.EndTime); len(errs) > 0 {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": errs})
		return
	}
	// 解引用分页参数
	page := 1
	if q.Page != nil {
		page = *q.Page
	}
	pageSize := 20
	if q.PageSize != nil {
		pageSize = *q.PageSize
	}
	// 转换为 Service 所需的非指针结构
	listQuery := listQueryFromRequest(q, page, pageSize)
	data, err := h.svc.ListAuditLogs(listQuery)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// listQueryFromRequest 将 Handler 层的指针分页参数展开为 Service 层结构。
func listQueryFromRequest(q AuditLogListQuery, page, pageSize int) AuditLogListQuery {
	return AuditLogListQuery{
		Page:      &page,
		PageSize:  &pageSize,
		ModuleID:  q.ModuleID,
		UserID:    q.UserID,
		Action:    q.Action,
		StartTime: q.StartTime,
		EndTime:   q.EndTime,
		Keyword:   q.Keyword,
	}
}
// GetAuditLogDetail 获取日志详情。
//
// v1.1（P1-6）：路径参数解析失败返回 422/90004（与 Python 版 Path(..., gt=0) 一致）。
func (h *Handler) GetAuditLogDetail(c *gin.Context) {
	logIDStr := c.Param("log_id")
	logID, err := strconv.ParseUint(logIDStr, 10, 64)
	if err != nil || logID == 0 {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": []map[string]interface{}{
				{
					"type": "type_error",
					"loc":  []string{"path", "log_id"},
					"msg":  "log_id 必须为正整数",
				},
			}})
		return
	}
	data, err := h.svc.GetAuditLogDetail(uint(logID))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, data, "success")
}
// ExportAuditLogs 导出操作日志。
func (h *Handler) ExportAuditLogs(c *gin.Context) {
	var q AuditLogExportQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": buildValidationErrors(&q, err, "query")})
		return
	}
	// 时间格式预校验
	if errs := validateTimeParams(q.StartTime, q.EndTime); len(errs) > 0 {
		response.Error(c, http.StatusUnprocessableEntity,
			exception.CodeValidationFail, "请求参数校验失败",
			gin.H{"errors": errs})
		return
	}
	if q.Format == "" {
		q.Format = ExportFormatCSV
	}
	content, mediaType, filename, err := h.svc.ExportAuditLogs(q)
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, mediaType, content)
}
// handleError 统一处理 PlatformError。
func handleError(c *gin.Context, err error) {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		response.Error(c, pe.HTTPStatus, pe.Code, pe.Message, pe.Data)
		return
	}
	response.Error(c, http.StatusInternalServerError,
		exception.CodeInternalError, "服务器内部错误", nil)
}
// validateTimeParams 校验时间参数格式。
//
// v1.1（P2-7）：返回结构化 errors，供 handler 直接返回 422/90004。
func validateTimeParams(startTime, endTime string) []map[string]interface{} {
	var errs []map[string]interface{}
	if startTime != "" {
		if _, err := ParseTime(startTime); err != nil {
			errs = append(errs, map[string]interface{}{
				"type": "datetime_parsing",
				"loc":  []string{"query", "start_time"},
				"msg":  fmt.Sprintf("时间格式无效：%s", startTime),
			})
		}
	}
	if endTime != "" {
		if _, err := ParseTime(endTime); err != nil {
			errs = append(errs, map[string]interface{}{
				"type": "datetime_parsing",
				"loc":  []string{"query", "end_time"},
				"msg":  fmt.Sprintf("时间格式无效：%s", endTime),
			})
		}
	}
	return errs
}
// buildValidationErrors 将绑定错误转换为结构化数组。
//
// v1.1（P1-4 / P1-5）：
//   - 字段校验错误 → validator.ValidationErrors，loc[1] 使用 form tag 字段名；
//   - query 类型错误 → type_error（区分 strconv.Parse* 错误）；
//   - 其他 → invalid_json。
func buildValidationErrors(req interface{}, err error, location string) []map[string]interface{} {
	var result []map[string]interface{}
	// 1. validator.ValidationErrors（字段校验失败）
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			fieldName := toFormFieldName(req, e.StructField())
			result = append(result, map[string]interface{}{
				"type":  e.Tag(),
				"loc":   []string{location, fieldName},
				"msg":   e.Error(),
				"input": e.Value(),
			})
		}
		return result
	}
	// 2. 类型转换错误（如 ?page=abc）
	if isTypeConversionError(err) {
		fieldName := extractFieldFromTypeError(err)
		result = append(result, map[string]interface{}{
			"type": "type_error",
			"loc":  []string{location, fieldName},
			"msg":  err.Error(),
		})
		return result
	}
	// 3. 其他（JSON body 解析失败等）
	result = append(result, map[string]interface{}{
		"type": "invalid_json",
		"msg":  err.Error(),
	})
	return result
}
// isTypeConversionError 判断错误是否为 query 参数的类型转换失败。
func isTypeConversionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "strconv.ParseInt") ||
		strings.Contains(msg, "strconv.ParseUint") ||
		strings.Contains(msg, "strconv.ParseFloat") ||
		strings.Contains(msg, "strconv.ParseBool")
}
// extractFieldFromTypeError 从类型转换错误中提取字段名。
//
// 已知限制：gin 在类型转换错误时返回的消息通常不含 `Key: 'Xxx.Field'` 前缀，
// 因此本函数通常返回 "unknown"。前端应以 errors[0].msg 为主进行错误提示。
func extractFieldFromTypeError(err error) string {
	msg := err.Error()
	const prefix = "Key: '"
	if idx := strings.Index(msg, prefix); idx >= 0 {
		rest := msg[idx+len(prefix):]
		if end := strings.Index(rest, "'"); end >= 0 {
			full := rest[:end]
			if dot := strings.LastIndex(full, "."); dot >= 0 {
				structField := full[dot+1:]
				return strings.ToLower(structField)
			}
		}
	}
	return "unknown"
}
// toFormFieldName 通过反射从结构体的 form tag（回退 json tag）提取字段名。
func toFormFieldName(req interface{}, structField string) string {
	t := reflect.TypeOf(req)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return strings.ToLower(structField)
	}
	f, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}
	for _, tagName := range []string{"form", "json"} {
		tag := f.Tag.Get(tagName)
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return strings.ToLower(structField)
}
