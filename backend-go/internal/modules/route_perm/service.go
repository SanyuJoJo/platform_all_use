package route_perm
import (
	"strings"
	"sync"
	"time"
	"gorm.io/gorm"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)
// Service 路由权限服务。
//
// 缓存策略：
//   - 内存缓存 key = method + " " + pattern，value = []权限码；
//   - 默认 TTL 30 秒，到期自动重载；
//   - 提供 Invalidate() 供菜单/路由权限变更时主动失效。
type Service struct {
	db     *gorm.DB
	mu     sync.RWMutex
	cache  map[string][]string
	loaded time.Time
	ttl    time.Duration
}
// NewService 构造路由权限服务。
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:    db,
		cache: make(map[string][]string),
		ttl:   30 * time.Second,
	}
}
// Invalidate 清空缓存，下次查询时重载。
func (s *Service) Invalidate() {
	s.mu.Lock()
	s.cache = make(map[string][]string)
	s.loaded = time.Time{}
	s.mu.Unlock()
}
// LookupCodes 查询路由所需权限码。
//
// 返回 nil 表示"未配置映射"，调用方应放行（兼容旧系统）。
// 返回非空切片表示"已配置映射"，调用方必须校验用户权限。
func (s *Service) LookupCodes(method, pattern string) ([]string, error) {
	key := strings.ToUpper(method) + " " + pattern
	s.mu.RLock()
	fresh := time.Since(s.loaded) < s.ttl
	if fresh {
		if v, ok := s.cache[key]; ok {
			codes := append([]string(nil), v...)
			s.mu.RUnlock()
			return codes, nil
		}
		s.mu.RUnlock()
		return nil, nil
	}
	s.mu.RUnlock()
	if err := s.reload(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.cache[key]; ok {
		return append([]string(nil), v...), nil
	}
	return nil, nil
}
// reload 重新加载全表到内存。
func (s *Service) reload() error {
	var rows []models.RoutePermission
	if err := s.db.Where("status = ?", 0).Find(&rows).Error; err != nil {
		return exception.New(exception.CodeInternalError, "加载路由权限失败："+err.Error(), 500, nil)
	}
	next := make(map[string][]string)
	for _, r := range rows {
		if strings.TrimSpace(r.PermissionCode) == "" {
			continue
		}
		key := strings.ToUpper(r.HTTPMethod) + " " + r.RoutePattern
		next[key] = append(next[key], r.PermissionCode)
	}
	s.mu.Lock()
	s.cache = next
	s.loaded = time.Now()
	s.mu.Unlock()
	return nil
}
// List 查询路由权限列表。
func (s *Service) List(moduleID string) ([]*models.RoutePermission, error) {
	tx := s.db.Model(&models.RoutePermission{})
	if moduleID != "" {
		tx = tx.Where("module_id = ?", moduleID)
	}
	var list []*models.RoutePermission
	if err := tx.Order("route_pattern ASC, http_method ASC").Find(&list).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询路由权限失败："+err.Error(), 500, nil)
	}
	return list, nil
}
// CreateRequest 新增/修改路由权限请求。
type CreateRequest struct {
	HTTPMethod     string `json:"httpMethod"`
	RoutePattern   string `json:"routePattern"`
	PermissionCode string `json:"permissionCode"`
	ModuleID       string `json:"moduleId"`
	Description    string `json:"description"`
}
// Create 新增路由权限。
func (s *Service) Create(req *CreateRequest) (*models.RoutePermission, error) {
	if strings.TrimSpace(req.HTTPMethod) == "" ||
		strings.TrimSpace(req.RoutePattern) == "" ||
		strings.TrimSpace(req.PermissionCode) == "" {
		return nil, exception.New(exception.CodeParamInvalid, "httpMethod/routePattern/permissionCode 必填", 400, nil)
	}
	method := strings.ToUpper(req.HTTPMethod)
	var count int64
	if err := s.db.Model(&models.RoutePermission{}).
		Where("http_method = ? AND route_pattern = ? AND permission_code = ?",
			method, req.RoutePattern, req.PermissionCode).
		Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "校验失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return nil, exception.New(exception.CodeRoutePermConflict, "该路由权限映射已存在", 400, nil)
	}
	rp := &models.RoutePermission{
		HTTPMethod:     method,
		RoutePattern:   req.RoutePattern,
		PermissionCode: req.PermissionCode,
		ModuleID:       req.ModuleID,
		Description:    req.Description,
		Status:         0,
	}
	if err := s.db.Create(rp).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "创建失败："+err.Error(), 500, nil)
	}
	s.Invalidate()
	return rp, nil
}
// Update 修改路由权限。
func (s *Service) Update(id uint, req *CreateRequest, status *int8) (*models.RoutePermission, error) {
	var rp models.RoutePermission
	if err := s.db.First(&rp, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeRoutePermNotFound, "路由权限不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询失败："+err.Error(), 500, nil)
	}
	if strings.TrimSpace(req.HTTPMethod) != "" {
		rp.HTTPMethod = strings.ToUpper(req.HTTPMethod)
	}
	if strings.TrimSpace(req.RoutePattern) != "" {
		rp.RoutePattern = req.RoutePattern
	}
	if strings.TrimSpace(req.PermissionCode) != "" {
		rp.PermissionCode = req.PermissionCode
	}
	rp.ModuleID = req.ModuleID
	rp.Description = req.Description
	if status != nil {
		rp.Status = *status
	}
	if err := s.db.Save(&rp).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "更新失败："+err.Error(), 500, nil)
	}
	s.Invalidate()
	return &rp, nil
}
// Delete 删除路由权限。
func (s *Service) Delete(id uint) error {
	if err := s.db.Delete(&models.RoutePermission{}, id).Error; err != nil {
		return exception.New(exception.CodeInternalError, "删除失败："+err.Error(), 500, nil)
	}
	s.Invalidate()
	return nil
}
