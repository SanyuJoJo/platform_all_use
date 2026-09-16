package module_manager
import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/models"
)
// Loader 模块加载器（P3 阶段为元数据接口，P6 阶段替换为子进程方案）。
//
// v1.3（P2-NEW-08）：作为"唯一真实源"，所有 MarkLoaded / UnmarkLoaded
// 调用统一通过 Service.loader 引用，避免多实例状态不一致。
type Loader struct {
	mu     sync.Mutex
	loaded map[string]struct{}
}
// NewLoader 创建加载器。
func NewLoader() *Loader {
	return &Loader{
		loaded: make(map[string]struct{}),
	}
}
// MarkLoaded 标记模块已加载。
func (l *Loader) MarkLoaded(moduleID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loaded[moduleID] = struct{}{}
}
// UnmarkLoaded 取消模块加载标记。
func (l *Loader) UnmarkLoaded(moduleID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.loaded, moduleID)
}
// IsLoaded 判断模块是否已加载。
func (l *Loader) IsLoaded(moduleID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, ok := l.loaded[moduleID]
	return ok
}
// LoadedModules 返回当前所有已加载模块的副本（供测试与监控使用）。
func (l *Loader) LoadedModules() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := make([]string, 0, len(l.loaded))
	for id := range l.loaded {
		result = append(result, id)
	}
	return result
}
// CleanupResidue 清理模块目录残留（.new / .old / .bak）。
func (s *Service) CleanupResidue() {
	root := s.moduleRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	var cleaned []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "module_") {
			continue
		}
		if strings.HasSuffix(name, ".new") ||
			strings.HasSuffix(name, ".old") ||
			strings.HasSuffix(name, ".bak") {
			safeRmtreeQuiet(filepath.Join(root, name))
			cleaned = append(cleaned, name)
		}
	}
	if len(cleaned) > 0 {
		log.Info().Strs("cleaned", cleaned).Msg("已清理模块残留目录")
	}
}
// LoadActiveModules 按依赖拓扑顺序加载 active 模块。
//
// v1.3（P2-NEW-08）：移除 `loader` 参数，统一使用 s.loader（唯一真实源）。
// 若 s.loader == nil（单元测试未注入），所有标记操作静默跳过。
func (s *Service) LoadActiveModules(db *gorm.DB) ([]string, error) {
	var modules []models.Module
	if err := db.Preload("Dependencies").
		Where("status = ?", "active").
		Find(&modules).Error; err != nil {
		return nil, err
	}
	byID := make(map[string]*models.Module, len(modules))
	for i := range modules {
		byID[modules[i].ID] = &modules[i]
	}
	visited := make(map[string]bool)
	var order []string
	var visit func(m *models.Module)
	visit = func(m *models.Module) {
		if visited[m.ID] {
			return
		}
		visited[m.ID] = true
		for _, dep := range m.Dependencies {
			if dm, ok := byID[dep.DependencyID]; ok {
				visit(dm)
			}
		}
		order = append(order, m.ID)
		if s.loader != nil {
			s.loader.MarkLoaded(m.ID)
		}
	}
	for i := range modules {
		visit(&modules[i])
	}
	return order, nil
}
