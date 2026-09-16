package module_manager
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/modules/auth"
)
// Service 模块管理服务。
type Service struct {
	db     *gorm.DB
	cfg    *config.Config
	loader *Loader // P1-NEW-02：可选注入
}
// NewService 创建模块管理服务。
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}
// SetLoader 注入 Loader（P1-NEW-02）。
//
// 单元测试场景可不注入；此时所有 MarkLoaded/UnmarkLoaded 调用被忽略。
func (s *Service) SetLoader(loader *Loader) {
	s.loader = loader
}
// ---------------------------------------------------------------------------
// 工具函数
// ---------------------------------------------------------------------------
func (s *Service) moduleRoot() string {
	root := s.cfg.ModulesDir
	if root == "" {
		root = "./src/modules"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}
func (s *Service) uploadRoot() string {
	root := s.cfg.ModuleUploadDir
	if root == "" {
		root = "./uploads/modules"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}
func (s *Service) getModuleDir(moduleID string) string {
	return filepath.Join(s.moduleRoot(), "module_"+moduleID)
}
func safeRmtreeQuiet(path string) {
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err != nil {
		return
	}
	if err := os.RemoveAll(path); err != nil {
		log.Warn().Err(err).Str("path", path).Msg("清理目录失败（忽略）")
	}
}
func (s *Service) isCoreModule(moduleID string) bool {
	_, ok := CORE_MODULE_IDS[moduleID]
	return ok
}
func (s *Service) loadModuleWithDeps(tx *gorm.DB, moduleID string) (*models.Module, error) {
	var m models.Module
	err := tx.Preload("Dependencies").First(&m, "id = ?", moduleID).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (s *Service) checkDependencies(tx *gorm.DB, dependencies []string) error {
	for _, dep := range dependencies {
		var m models.Module
		err := tx.First(&m, "id = ?", dep).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeModuleDependencyMissing,
				fmt.Sprintf("依赖模块缺失：%s", dep), 400, nil)
		}
		if err != nil {
			return exception.New(exception.CodeInternalError, "查询依赖失败", 500, nil)
		}
		if m.Status != "active" {
			return exception.New(exception.CodeModuleDependencyMissing,
				fmt.Sprintf("依赖模块未启用：%s", dep), 400, nil)
		}
	}
	return nil
}
// strPtrOrNil 空字符串转 nil，非空返回指针。
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
// intPtr 返回 int 指针。
func intPtr(v int) *int { return &v }
// ---------------------------------------------------------------------------
// 序列化
// ---------------------------------------------------------------------------
func serializeModule(m *models.Module) *ModuleOut {
	deps := make([]string, 0, len(m.Dependencies))
	for _, d := range m.Dependencies {
		deps = append(deps, d.DependencyID)
	}
	sort.Strings(deps)
	menus := make([]MenuItem, 0)
	if m.Manifest != nil {
		if rawMenus, ok := m.Manifest["menus"].([]interface{}); ok {
			for _, item := range rawMenus {
				menuMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				mi := MenuItem{
					ID:        toString(menuMap["id"]),
					Title:     toString(menuMap["title"]),
					Path:      toString(menuMap["path"]),
					Component: toString(menuMap["component"]),
					Order:     toInt(menuMap["order"]),
				}
				if v, ok := menuMap["parent_id"].(string); ok && v != "" {
					mi.ParentID = &v
				}
				if v, ok := menuMap["icon"].(string); ok && v != "" {
					mi.Icon = &v
				}
				if v, ok := menuMap["permission"].(string); ok && v != "" {
					mi.Permission = &v
				}
				menus = append(menus, mi)
			}
		}
	}
	config := map[string]interface{}{}
	if m.Config != nil {
		config = m.Config
	}
	return &ModuleOut{
		ID:            m.ID,
		Name:          m.Name,
		Version:       m.Version,
		Description:   m.Description,
		Author:        m.Author,
		Homepage:      m.Homepage,
		Status:        m.Status,
		EntryBackend:  m.EntryBackend,
		EntryFrontend: m.EntryFrontend,
		Dependencies:  deps,
		Menus:         menus,
		Config:        config,
		InstalledAt:   formatDateTime(m.InstalledAt),
		UpdatedAt:     formatDateTime(m.UpdatedAt),
	}
}
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
func toInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	}
	return 0
}
// filterMenusByPermission 按权限递归过滤菜单。
func filterMenusByPermission(menus []MenuItem, userPerms map[string]struct{}) []MenuItem {
	if len(menus) == 0 {
		return menus
	}
	byID := make(map[string]MenuItem, len(menus))
	for _, m := range menus {
		byID[m.ID] = m
	}
	children := make(map[string][]MenuItem)
	var roots []MenuItem
	for _, m := range menus {
		if m.ParentID != nil && *m.ParentID != "" {
			if _, ok := byID[*m.ParentID]; ok {
				children[*m.ParentID] = append(children[*m.ParentID], m)
				continue
			}
		}
		roots = append(roots, m)
	}
	allowed := make(map[string]struct{})
	var visit func(node MenuItem)
	visit = func(node MenuItem) {
		if node.Permission != nil && *node.Permission != "" {
			if _, ok := userPerms[*node.Permission]; !ok {
				return
			}
		}
		allowed[node.ID] = struct{}{}
		for _, child := range children[node.ID] {
			visit(child)
		}
	}
	for _, root := range roots {
		visit(root)
	}
	result := make([]MenuItem, 0, len(menus))
	for _, m := range menus {
		if _, ok := allowed[m.ID]; ok {
			result = append(result, m)
		}
	}
	return result
}
// ---------------------------------------------------------------------------
// 列表
// ---------------------------------------------------------------------------
// ListModules 查询模块列表。
func (s *Service) ListModules(query ModuleListQuery, userPerms []string) (map[string]interface{}, error) {
	q := s.db.Model(&models.Module{})
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		q = q.Where("id LIKE ? OR name LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询模块失败", 500, nil)
	}
	var modules []models.Module
	if err := q.Preload("Dependencies").
		Order("installed_at ASC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&modules).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询模块失败", 500, nil)
	}
	items := make([]*ModuleOut, 0, len(modules))
	for i := range modules {
		item := serializeModule(&modules[i])
		if query.FilterMenus {
			permSet := make(map[string]struct{}, len(userPerms))
			for _, p := range userPerms {
				permSet[p] = struct{}{}
			}
			item.Menus = filterMenusByPermission(item.Menus, permSet)
		}
		items = append(items, item)
	}
	pages := 0
	if query.PageSize > 0 {
		pages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}
	return map[string]interface{}{
		"items":     items,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
		"pages":     pages,
	}, nil
}
// ---------------------------------------------------------------------------
// FS 三工具函数
// ---------------------------------------------------------------------------
func (s *Service) prepareNewDir(source, target string) (string, error) {
	newDir := target + ".new"
	safeRmtreeQuiet(newDir)
	if err := copyDir(source, newDir); err != nil {
		safeRmtreeQuiet(newDir)
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("模块源码复制失败：%v", err), 500, nil)
	}
	if _, err := os.Stat(filepath.Join(newDir, "manifest.json")); err != nil {
		safeRmtreeQuiet(newDir)
		return "", exception.New(exception.CodeModuleManifestInvalid,
			"模块源码不完整（缺少 manifest.json）", 400, nil)
	}
	return newDir, nil
}
func (s *Service) swapNewToTarget(newDir, target string) (string, error) {
	oldBackup := target + ".old"
	safeRmtreeQuiet(oldBackup)
	_, statErr := os.Stat(target)
	targetExisted := statErr == nil
	if targetExisted {
		if err := os.Rename(target, oldBackup); err != nil {
			safeRmtreeQuiet(newDir)
			return "", exception.New(exception.CodeInternalError,
				fmt.Sprintf("备份旧目录失败：%v", err), 500, nil)
		}
	}
	if err := os.Rename(newDir, target); err != nil {
		if targetExisted {
			_ = os.Rename(oldBackup, target)
		}
		safeRmtreeQuiet(newDir)
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("模块目录替换失败：%v", err), 500, nil)
	}
	if targetExisted {
		return oldBackup, nil
	}
	return "", nil
}
func (s *Service) finalizeSwap(oldBackup string) {
	if oldBackup != "" {
		safeRmtreeQuiet(oldBackup)
	}
}
func (s *Service) rollbackSwap(target, oldBackup string) {
	safeRmtreeQuiet(target)
	if oldBackup != "" {
		if err := os.Rename(oldBackup, target); err != nil {
			log.Error().Err(err).
				Str("old", oldBackup).Str("target", target).
				Msg("恢复旧目录失败，需人工介入")
		}
	}
}
// copyDir 复制目录，忽略 CopyIgnorePatterns 中的模式。
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		for _, pattern := range CopyIgnorePatterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if rel == pattern {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		targetPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		return copyFile(path, targetPath)
	})
}
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := out.ReadFrom(in); err != nil {
		return err
	}
	return nil
}
// ---------------------------------------------------------------------------
// 路径校验
// ---------------------------------------------------------------------------
func (s *Service) validateSourcePath(sourcePath string) (string, error) {
	raw := sourcePath
	if !filepath.IsAbs(raw) {
		abs, err := filepath.Abs(raw)
		if err != nil {
			return "", exception.New(exception.CodeParamInvalid, "source_path 无效", 400, nil)
		}
		raw = abs
	}
	resolved, err := filepath.EvalSymlinks(raw)
	if err != nil {
		resolved = raw
	}
	uploadRoot, _ := filepath.EvalSymlinks(s.uploadRoot())
	if uploadRoot == "" {
		uploadRoot = s.uploadRoot()
	}
	rel, err := filepath.Rel(uploadRoot, resolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("source_path 必须位于 %s 下", s.uploadRoot()), 400, nil)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", exception.New(exception.CodeParamInvalid,
			"source_path 必须为已存在目录", 400, nil)
	}
	return resolved, nil
}
// ---------------------------------------------------------------------------
// 安装
// ---------------------------------------------------------------------------
// InstallModule 安装模块（延迟提交）。
func (s *Service) InstallModule(
	installType, filePath, sourcePath string,
	operatorID uint, operatorName string,
) (*ModuleOut, error) {
	tempDir, err := os.MkdirTemp("", "module_install_")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "创建临时目录失败", 500, nil)
	}
	defer os.RemoveAll(tempDir)
	var moduleID string
	var targetDir, newDir, oldBackup string
	fsSwapped := false
	zipCleaned := false
	cleanupZip := func() {
		if zipCleaned {
			return
		}
		zipCleaned = true
		if installType == "zip" && filePath != "" {
			s.cleanupUploadedZip(filePath)
		}
	}
	defer func() {
		if newDir != "" {
			safeRmtreeQuiet(newDir)
		}
		cleanupZip()
	}()
	// ---- 阶段 1：staging 与校验 ----
	sourceRoot, err := s.resolveModuleSource(installType, filePath, sourcePath, tempDir)
	if err != nil {
		return nil, err
	}
	rawManifest, err := LoadManifestFromDir(sourceRoot)
	if err != nil {
		return nil, err
	}
	manifest, err := ValidateManifestDict(rawManifest)
	if err != nil {
		return nil, err
	}
	moduleID = manifest.ID
	if s.isCoreModule(moduleID) {
		return nil, exception.New(exception.CodeModuleExists,
			"模块 ID 与核心模块冲突", 409, nil)
	}
	var existing models.Module
	if err := s.db.First(&existing, "id = ?", moduleID).Error; err == nil {
		return nil, exception.New(exception.CodeModuleExists, "模块已存在", 409, nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.New(exception.CodeInternalError, "查询模块失败", 500, nil)
	}
	if err := s.checkDependencies(s.db, manifest.Dependencies); err != nil {
		return nil, err
	}
	targetDir = s.getModuleDir(moduleID)
	if _, err := os.Stat(targetDir); err == nil {
		return nil, exception.New(exception.CodeModuleExists,
			fmt.Sprintf("模块目录已存在但无 DB 记录：%s", targetDir), 409, nil)
	}
	// ---- 阶段 2：准备 new_dir ----
	newDir, err = s.prepareNewDir(sourceRoot, targetDir)
	if err != nil {
		return nil, err
	}
	// ---- 阶段 3：FS 替换 ----
	oldBackup, err = s.swapNewToTarget(newDir, targetDir)
	if err != nil {
		return nil, err
	}
	fsSwapped = true
	// ---- 阶段 4：DB 修改 ----
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		module := models.Module{
			ID:            moduleID,
			Name:          manifest.Name,
			Version:       manifest.Version,
			Description:   strPtrOrNil(manifest.Description),
			Author:        manifest.Author,
			Homepage:      manifest.Homepage,
			Status:        "inactive",
			EntryBackend:  manifest.EntryBackend,
			EntryFrontend: manifest.EntryFrontend,
			Config:        models.JSONMap{},
			Manifest:      models.JSONMap(rawManifest),
		}
		if err := tx.Create(&module).Error; err != nil {
			return err
		}
		for _, dep := range manifest.Dependencies {
			if err := tx.Create(&models.ModuleDependency{
				ModuleID:     moduleID,
				DependencyID: dep,
			}).Error; err != nil {
				return err
			}
		}
		if len(manifest.Permissions) > 0 {
			items := make([]auth.PermissionRegisterItem, 0, len(manifest.Permissions))
			for _, p := range manifest.Permissions {
				items = append(items, auth.PermissionRegisterItem{
					Code:     toString(p["code"]),
					Name:     toString(p["name"]),
					Resource: toString(p["resource"]),
					Action:   toString(p["action"]),
				})
			}
			if err := auth.RegisterPermissionsTx(tx, moduleID, items); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		if fsSwapped {
			s.rollbackSwap(targetDir, oldBackup)
		}
		cleanupZip()
		auth.LogAuthEvent("module_install", &operatorID, operatorName, "fail",
			intPtr(errCode(txErr)), "", "", fmt.Sprintf("安装模块失败：%v", txErr))
		return nil, txErr
	}
	// ---- 阶段 5：finalize ----
	s.finalizeSwap(oldBackup)
	fsSwapped = false
	cleanupZip()
	full, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载模块失败", 500, nil)
	}
	auth.LogAuthEvent("module_install", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("安装模块 %s v%s", moduleID, manifest.Version))
	return serializeModule(full), nil
}
func (s *Service) resolveModuleSource(
	installType, filePath, sourcePath, tempDir string,
) (string, error) {
	switch installType {
	case "zip":
		if filePath == "" {
			return "", exception.New(exception.CodeParamInvalid, "file_path 不能为空", 400, nil)
		}
		abs := filePath
		if !filepath.IsAbs(abs) {
			a, err := filepath.Abs(abs)
			if err != nil {
				return "", exception.New(exception.CodeParamInvalid, "file_path 无效", 400, nil)
			}
			abs = a
		}
		if _, err := os.Stat(abs); err != nil {
			return "", exception.New(exception.CodeModuleZipInvalid, "ZIP 包不存在", 400, nil)
		}
		maxSize := int64(s.cfg.ModuleZipMaxSize)
		if maxSize <= 0 {
			maxSize = DefaultZipMaxSize
		}
		maxTotal := int64(s.cfg.ModuleZipMaxTotal)
		if maxTotal <= 0 {
			maxTotal = DefaultZipMaxTotal
		}
		maxFiles := s.cfg.ModuleZipMaxFiles
		if maxFiles <= 0 {
			maxFiles = DefaultZipMaxFiles
		}
		if err := SafeExtractZip(abs, tempDir, maxSize, maxTotal, maxFiles); err != nil {
			return "", err
		}
		manifestPath, err := findUniqueManifest(tempDir)
		if err != nil {
			return "", err
		}
		return filepath.Dir(manifestPath), nil
	case "path":
		if sourcePath == "" {
			return "", exception.New(exception.CodeParamInvalid, "source_path 不能为空", 400, nil)
		}
		return s.validateSourcePath(sourcePath)
	default:
		return "", exception.New(exception.CodeParamInvalid,
			"install_type 必须为 zip 或 path", 400, nil)
	}
}
func findUniqueManifest(root string) (string, error) {
	var found []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "manifest.json" {
			found = append(found, path)
		}
		return nil
	})
	if len(found) != 1 {
		return "", exception.New(exception.CodeModuleZipInvalid,
			"ZIP 包中未找到唯一 manifest.json", 400, nil)
	}
	return found[0], nil
}
func (s *Service) cleanupUploadedZip(filePath string) {
	if filePath == "" {
		return
	}
	abs := filePath
	if !filepath.IsAbs(abs) {
		a, err := filepath.Abs(abs)
		if err != nil {
			return
		}
		abs = a
	}
	resolved, _ := filepath.EvalSymlinks(abs)
	if resolved == "" {
		resolved = abs
	}
	uploadRoot := s.uploadRoot()
	rel, err := filepath.Rel(uploadRoot, resolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return
	}
	if err := os.Remove(resolved); err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Str("path", resolved).Msg("清理上传 ZIP 失败")
	}
}
// ---------------------------------------------------------------------------
// 升级
// ---------------------------------------------------------------------------
// UpgradeModule 升级模块（延迟提交）。
func (s *Service) UpgradeModule(
	moduleID, installType, filePath, sourcePath string,
	operatorID uint, operatorName string,
) (*ModuleOut, error) {
	if s.isCoreModule(moduleID) {
		return nil, exception.New(exception.CodeModuleCoreProtected,
			"核心模块不允许执行 upgrade 操作", 403, nil)
	}
	module, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeModuleNotFound, "模块不存在", 404, nil)
	}
	tempDir, err := os.MkdirTemp("", "module_upgrade_")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "创建临时目录失败", 500, nil)
	}
	defer os.RemoveAll(tempDir)
	var targetDir, newDir, oldBackup string
	fsSwapped := false
	defer func() {
		if newDir != "" {
			safeRmtreeQuiet(newDir)
		}
	}()
	// ---- 阶段 1：staging 与校验 ----
	sourceRoot, err := s.resolveModuleSource(installType, filePath, sourcePath, tempDir)
	if err != nil {
		return nil, err
	}
	rawManifest, err := LoadManifestFromDir(sourceRoot)
	if err != nil {
		return nil, err
	}
	manifest, err := ValidateManifestDict(rawManifest)
	if err != nil {
		return nil, err
	}
	if manifest.ID != moduleID {
		return nil, exception.New(exception.CodeModuleManifestInvalid,
			fmt.Sprintf("升级包模块 ID 不一致：%s != %s", manifest.ID, moduleID), 400, nil)
	}
	if compareVersion(manifest.Version, module.Version) <= 0 {
		return nil, exception.New(exception.CodeModuleVersionInvalid,
			fmt.Sprintf("新版本 %s 必须高于当前版本 %s", manifest.Version, module.Version), 400, nil)
	}
	if err := s.checkDependencies(s.db, manifest.Dependencies); err != nil {
		return nil, err
	}
	targetDir = s.getModuleDir(moduleID)
	// ---- 阶段 2：准备 new_dir ----
	newDir, err = s.prepareNewDir(sourceRoot, targetDir)
	if err != nil {
		return nil, err
	}
	// ---- 阶段 3：FS 替换 ----
	oldBackup, err = s.swapNewToTarget(newDir, targetDir)
	if err != nil {
		return nil, err
	}
	fsSwapped = true
	// ---- 阶段 4：DB 修改 ----
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":           manifest.Name,
			"version":        manifest.Version,
			"description":    strPtrOrNil(manifest.Description),
			"author":         manifest.Author,
			"homepage":       manifest.Homepage,
			"entry_backend":  manifest.EntryBackend,
			"entry_frontend": manifest.EntryFrontend,
			"manifest":       models.JSONMap(rawManifest),
			"updated_at":     time.Now().UTC(),
		}
		if err := tx.Model(&models.Module{}).
			Where("id = ?", moduleID).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("module_id = ?", moduleID).
			Delete(&models.ModuleDependency{}).Error; err != nil {
			return err
		}
		for _, dep := range manifest.Dependencies {
			if err := tx.Create(&models.ModuleDependency{
				ModuleID:     moduleID,
				DependencyID: dep,
			}).Error; err != nil {
				return err
			}
		}
		if len(manifest.Permissions) > 0 {
			items := make([]auth.PermissionRegisterItem, 0, len(manifest.Permissions))
			for _, p := range manifest.Permissions {
				items = append(items, auth.PermissionRegisterItem{
					Code:     toString(p["code"]),
					Name:     toString(p["name"]),
					Resource: toString(p["resource"]),
					Action:   toString(p["action"]),
				})
			}
			if err := auth.RegisterPermissionsTx(tx, moduleID, items); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		if fsSwapped {
			s.rollbackSwap(targetDir, oldBackup)
		}
		auth.LogAuthEvent("module_upgrade", &operatorID, operatorName, "fail",
			intPtr(errCode(txErr)), "", "", fmt.Sprintf("升级模块失败：%v", txErr))
		return nil, txErr
	}
	// ---- 阶段 5：finalize ----
	s.finalizeSwap(oldBackup)
	fsSwapped = false
	full, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载模块失败", 500, nil)
	}
	auth.LogAuthEvent("module_upgrade", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("升级模块 %s：%s → %s", moduleID, module.Version, manifest.Version))
	return serializeModule(full), nil
}
// compareVersion 语义化版本比较（含 pre-release 点分语义）。
func compareVersion(a, b string) int {
	pa := parseVersion(a)
	pb := parseVersion(b)
	if pa.major != pb.major {
		return sign(pa.major - pb.major)
	}
	if pa.minor != pb.minor {
		return sign(pa.minor - pb.minor)
	}
	if pa.patch != pb.patch {
		return sign(pa.patch - pb.patch)
	}
	return comparePreRelease(pa.pre, pb.pre)
}
func comparePreRelease(a, b string) int {
	if a == "" && b == "" {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}
	for i := 0; i < maxLen; i++ {
		if i >= len(partsA) {
			return -1
		}
		if i >= len(partsB) {
			return 1
		}
		segA := partsA[i]
		segB := partsB[i]
		numA, isNumA := strconv.Atoi(segA)
		numB, isNumB := strconv.Atoi(segB)
		switch {
		case isNumA == nil && isNumB == nil:
			if numA != numB {
				return sign(numA - numB)
			}
		case isNumA == nil && isNumB != nil:
			return -1
		case isNumA != nil && isNumB == nil:
			return 1
		default:
			if c := strings.Compare(segA, segB); c != 0 {
				return c
			}
		}
	}
	return 0
}
type parsedVersion struct {
	major, minor, patch int
	pre                 string
}
func parseVersion(v string) parsedVersion {
	var p parsedVersion
	if idx := strings.Index(v, "+"); idx >= 0 {
		v = v[:idx]
	}
	if idx := strings.Index(v, "-"); idx >= 0 {
		p.pre = v[idx+1:]
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		p.major = atoi(parts[0])
	}
	if len(parts) > 1 {
		p.minor = atoi(parts[1])
	}
	if len(parts) > 2 {
		p.patch = atoi(parts[2])
	}
	return p
}
func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}
// ---------------------------------------------------------------------------
// 卸载
// ---------------------------------------------------------------------------
// UninstallModule 卸载模块。
func (s *Service) UninstallModule(
	moduleID string, force, dropTables bool,
	operatorID uint, operatorName string,
) error {
	if s.isCoreModule(moduleID) {
		return exception.New(exception.CodeModuleCoreProtected,
			"核心模块不允许执行 uninstall 操作", 403, nil)
	}
	module, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return exception.New(exception.CodeModuleNotFound, "模块不存在", 404, nil)
	}
	if module.Status == "active" {
		return exception.New(exception.CodeModuleInUse,
			"模块正在运行，请先停用", 409, nil)
	}
	if !force {
		var count int64
		if err := s.db.Model(&models.ModuleDependency{}).
			Joins("JOIN module_manager_module ON module_manager_module.id = module_manager_dependency.module_id").
			Where("module_manager_dependency.dependency_id = ? AND module_manager_module.status = ?",
				moduleID, "active").
			Count(&count).Error; err != nil {
			return exception.New(exception.CodeInternalError, "检查依赖失败", 500, nil)
		}
		if count > 0 {
			return exception.New(exception.CodeModuleDependencyConflict,
				"模块被 active 模块依赖", 400, nil)
		}
	}
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := auth.UnregisterPermissionsTx(tx, moduleID); err != nil {
			return err
		}
		if dropTables && module.Manifest != nil {
			tables := extractStringList(module.Manifest["database_tables"])
			for _, tableName := range tables {
				if tableName == "" {
					continue
				}
				if !tableNamePattern.MatchString(tableName) {
					log.Warn().Str("table", tableName).Msg("跳过非法表名")
					continue
				}
				if !strings.HasPrefix(tableName, moduleID+"_") {
					log.Warn().Str("table", tableName).Msg("跳过不合规表名")
					continue
				}
				if err := tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, tableName)).Error; err != nil {
					log.Warn().Err(err).Str("table", tableName).Msg("删除表失败")
				}
			}
		}
		if err := tx.Where("module_id = ?", moduleID).
			Delete(&models.ModuleDependency{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Module{}, "id = ?", moduleID).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	safeRmtreeQuiet(s.getModuleDir(moduleID))
	if s.loader != nil {
		s.loader.UnmarkLoaded(moduleID)
	}
	auth.LogAuthEvent("module_uninstall", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("卸载模块 %s", moduleID))
	return nil
}
// extractStringList 兼容 []interface{} / []string。
func extractStringList(v interface{}) []string {
	switch t := v.(type) {
	case []interface{}:
		result := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return t
	}
	return nil
}
// ---------------------------------------------------------------------------
// 启用 / 停用
// ---------------------------------------------------------------------------
// EnableModule 启用模块。
func (s *Service) EnableModule(
	moduleID string, operatorID uint, operatorName string,
) (*ModuleOut, error) {
	if s.isCoreModule(moduleID) {
		return nil, exception.New(exception.CodeModuleCoreProtected,
			"核心模块不允许执行 enable 操作", 403, nil)
	}
	module, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeModuleNotFound, "模块不存在", 404, nil)
	}
	if module.Status == "active" {
		return serializeModule(module), nil
	}
	var deps []string
	for _, d := range module.Dependencies {
		deps = append(deps, d.DependencyID)
	}
	if err := s.checkDependencies(s.db, deps); err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.Module{}).
		Where("id = ?", moduleID).
		Updates(map[string]interface{}{
			"status":     "active",
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "更新状态失败", 500, nil)
	}
	if s.loader != nil {
		s.loader.MarkLoaded(moduleID)
	}
	full, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载模块失败", 500, nil)
	}
	auth.LogAuthEvent("module_enable", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("启用模块 %s", moduleID))
	return serializeModule(full), nil
}
// DisableModule 停用模块。
func (s *Service) DisableModule(
	moduleID string, operatorID uint, operatorName string,
) (*ModuleOut, error) {
	if s.isCoreModule(moduleID) {
		return nil, exception.New(exception.CodeModuleCoreProtected,
			"核心模块不允许执行 disable 操作", 403, nil)
	}
	module, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeModuleNotFound, "模块不存在", 404, nil)
	}
	if module.Status == "inactive" {
		return serializeModule(module), nil
	}
	var count int64
	if err := s.db.Model(&models.ModuleDependency{}).
		Joins("JOIN module_manager_module ON module_manager_module.id = module_manager_dependency.module_id").
		Where("module_manager_dependency.dependency_id = ? AND module_manager_module.status = ?",
			moduleID, "active").
		Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "检查依赖失败", 500, nil)
	}
	if count > 0 {
		return nil, exception.New(exception.CodeModuleDependencyConflict,
			"模块被其他 active 模块依赖，无法停用", 400, nil)
	}
	if err := s.db.Model(&models.Module{}).
		Where("id = ?", moduleID).
		Updates(map[string]interface{}{
			"status":     "inactive",
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "更新状态失败", 500, nil)
	}
	if s.loader != nil {
		s.loader.UnmarkLoaded(moduleID)
	}
	full, err := s.loadModuleWithDeps(s.db, moduleID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载模块失败", 500, nil)
	}
	auth.LogAuthEvent("module_disable", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("停用模块 %s", moduleID))
	return serializeModule(full), nil
}
// ---------------------------------------------------------------------------
// 配置
// ---------------------------------------------------------------------------
// GetModuleConfig 获取模块配置。
func (s *Service) GetModuleConfig(moduleID string) (map[string]interface{}, error) {
	var module models.Module
	if err := s.db.First(&module, "id = ?", moduleID).Error; err != nil {
		return nil, exception.New(exception.CodeModuleNotFound, "模块不存在", 404, nil)
	}
	if module.Config == nil {
		return map[string]interface{}{}, nil
	}
	return module.Config, nil
}
// UpdateModuleConfig 更新模块配置（全量覆盖）。
func (s *Service) UpdateModuleConfig(
	moduleID string, config map[string]interface{},
	operatorID uint, operatorName string,
) (map[string]interface{}, error) {
	var module models.Module
	if err := s.db.First(&module, "id = ?", moduleID).Error; err != nil {
		return nil, exception.New(exception.CodeModuleNotFound, "模块不存在", 404, nil)
	}
	if err := validateConfig(module.Manifest, config); err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.Module{}).
		Where("id = ?", moduleID).
		Updates(map[string]interface{}{
			"config":     models.JSONMap(config),
			"updated_at": time.Now().UTC(),
		}).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "更新配置失败", 500, nil)
	}
	auth.LogAuthEvent("module_config_update", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("更新模块 %s 配置", moduleID))
	return config, nil
}
// validateConfig 按 manifest 中的 config_schema 校验配置。
func validateConfig(manifest models.JSONMap, config map[string]interface{}) error {
	if manifest == nil {
		return nil
	}
	schemaRaw, ok := manifest["config_schema"]
	if !ok {
		return nil
	}
	schema, ok := schemaRaw.(map[string]interface{})
	if !ok || len(schema) == 0 {
		return nil
	}
	for _, field := range extractStringList(schema["required"]) {
		if field == "" {
			continue
		}
		if _, ok := config[field]; !ok {
			return exception.New(exception.CodeModuleConfigInvalid,
				fmt.Sprintf("配置缺少必填字段：%s", field), 400, nil)
		}
	}
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}
	for key, value := range config {
		prop, ok := properties[key].(map[string]interface{})
		if !ok {
			continue
		}
		expectedType, _ := prop["type"].(string)
		if expectedType == "" {
			continue
		}
		if !checkType(value, expectedType) {
			return exception.New(exception.CodeModuleConfigInvalid,
				fmt.Sprintf("配置字段 %s 类型无效，期望 %s", key, expectedType), 400, nil)
		}
	}
	return nil
}
func checkType(value interface{}, expected string) bool {
	switch expected {
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "integer":
		switch v := value.(type) {
		case int, int64:
			return true
		case float64:
			return v == float64(int64(v))
		}
		return false
	case "number":
		switch value.(type) {
		case int, int64, float64:
			return true
		}
		return false
	case "string":
		_, ok := value.(string)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	}
	return true
}
// SaveUploadedZip 保存上传的 ZIP 到 MODULE_UPLOAD_DIR，返回保存路径。
func (s *Service) SaveUploadedZip(fileName string, content []byte) (string, error) {
	if !strings.HasSuffix(strings.ToLower(fileName), ".zip") {
		return "", exception.New(exception.CodeModuleZipInvalid, "仅支持 .zip 文件", 400, nil)
	}
	maxSize := int64(s.cfg.ModuleZipMaxSize)
	if maxSize <= 0 {
		maxSize = DefaultZipMaxSize
	}
	if int64(len(content)) > maxSize {
		return "", exception.New(exception.CodeModuleZipInvalid,
			fmt.Sprintf("ZIP 文件大小超过上限 %d", maxSize), 400, nil)
	}
	root := s.uploadRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", exception.New(exception.CodeInternalError, "创建上传目录失败", 500, nil)
	}
	target := filepath.Join(root, fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileName))
	if err := os.WriteFile(target, content, 0o644); err != nil {
		return "", exception.New(exception.CodeInternalError, "保存 ZIP 失败", 500, nil)
	}
	return target, nil
}
// errCode 从 error 中提取错误码（支持 PlatformError）。
func errCode(err error) int {
	var pe *exception.PlatformError
	if errors.As(err, &pe) {
		return pe.Code
	}
	return exception.CodeInternalError
}
