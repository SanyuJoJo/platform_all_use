package crypto

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend-go/internal/exception"
)

// paginateQuery 通用分页查询：先 Count 再 Find。
func paginateQuery(q *gorm.DB, slicePtr interface{}, page, pageSize int) (map[string]interface{}, error) {
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if err := q.Session(&gorm.Session{}).
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(slicePtr).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	pages := 0
	if pageSize > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return map[string]interface{}{
		"items":     slicePtr,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     pages,
	}, nil
}

// parsePageQuery 解析 page / page_size 查询参数。
func parsePageQuery(c *gin.Context) (int, int) {
	page, pageSize := 1, 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}
	return page, pageSize
}
