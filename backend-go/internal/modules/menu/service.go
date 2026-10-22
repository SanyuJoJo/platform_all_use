package menu

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// microAppPattern microApp 合法格式：小写字母开头，
// 后续仅允许小写字母/数字/下划线/连字符，总长 ≤50。
//
// 包级编译一次，避免 validateMicroApp 每次调用都重新编译正则。
var microAppPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`)

// Service 菜单服务。
type Service struct {
	db *gorm.DB
}

// NewService 构造菜单服务。
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// AdminRoleCodes 超级管理员角色 code。
var AdminRoleCodes = map[string]struct{}{
	"admin":       {},
	"super_admin": {},
}

// Create 新增菜单。
//
// microApp 处理（v1.7 更新）：
//   - 格式校验：任何非空值都必须满足 microAppPattern；
//   - 存在性校验：不做（配置先行，允许应用尚未注册）；
//   - 继承：请求为空时，父节点存在则继承父节点的 microApp（若非 main）。
func (s *Service) Create(req *CreateMenuRequest, operator string) (*models.Menu, error) {
	if err := s.validateMenuRequest(req.MenuType, req.ParentID, req.Path, req.Perms); err != nil {
		return nil, err
	}
	if err := s.checkNameUnique(req.ParentID, req.MenuName, 0); err != nil {
		return nil, err
	}
	if err := s.checkPathUnique(req.Path, req.MenuType, 0); err != nil {
		return nil, err
	}
	// ★ 新增：microApp 格式校验（不校验存在性）
	if err := s.validateMicroApp(req.MicroApp); err != nil {
		return nil, err
	}

	routeName := strings.TrimSpace(req.Name)
	if routeName == "" {
		routeName = deriveRouteName(req.Path, 0)
		routeName = s.ensureUniqueRouteName(routeName, 0)
	} else if err := s.checkRouteNameUnique(routeName, 0); err != nil {
		return nil, err
	}

	microApp := s.resolveCreateMicroApp(req.ParentID, req.MicroApp)

	now := time.Now()
	m := &models.Menu{
		ParentID:   req.ParentID,
		MenuName:   req.MenuName,
		Name:       routeName,
		MicroApp:   microApp,
		OrderNum:   req.OrderNum,
		Path:       req.Path,
		Component:  req.Component,
		Query:      req.Query,
		IsFrame:    derefInt8(req.IsFrame, 1),
		IsCache:    derefInt8(req.IsCache, 0),
		MenuType:   req.MenuType,
		Visible:    derefInt8(req.Visible, 0),
		Status:     derefInt8(req.Status, 0),
		Perms:      req.Perms,
		Icon:       req.Icon,
		ModuleID:   req.ModuleID,
		CreateBy:   operator,
		CreateTime: now,
		UpdateBy:   operator,
		UpdateTime: now,
		Remark:     req.Remark,
	}
	if m.Icon == "" {
		m.Icon = "#"
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "创建菜单失败："+err.Error(), 500, nil)
	}
	return m, nil
}

// resolveCreateMicroApp 决定新建菜单的 microApp。
//
// 规则：
//   - 请求显式非空且非 "main" → 用请求值；
//   - 请求空且父节点存在 → 继承父节点的 microApp（若非空非 main）；
//   - 否则 → "main"。
func (s *Service) resolveCreateMicroApp(parentID uint, reqMicro string) string {
	req := strings.TrimSpace(reqMicro)
	if req != "" && req != "main" {
		return req
	}
	if parentID != 0 {
		if parent, err := s.GetByID(parentID); err == nil {
			pm := strings.TrimSpace(parent.MicroApp)
			if pm != "" && pm != "main" {
				return pm
			}
		}
	}
	if req != "" {
		return req
	}
	return "main"
}

// Update 修改菜单。
//
// microApp 处理（v1.7 更新）：同 Create，只做格式校验。
func (s *Service) Update(id uint, req *UpdateMenuRequest, operator string) (*models.Menu, error) {
	old, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.ParentID == id {
		return nil, exception.New(exception.CodeMenuParentInvalid, "父菜单不能是自己", 400, nil)
	}
	if err := s.checkCycle(id, req.ParentID); err != nil {
		return nil, err
	}
	if err := s.validateMenuRequest(req.MenuType, req.ParentID, req.Path, req.Perms); err != nil {
		return nil, err
	}
	if err := s.checkNameUnique(req.ParentID, req.MenuName, id); err != nil {
		return nil, err
	}
	if err := s.checkPathUnique(req.Path, req.MenuType, id); err != nil {
		return nil, err
	}
	// ★ 新增：microApp 格式校验（不校验存在性）
	if err := s.validateMicroApp(req.MicroApp); err != nil {
		return nil, err
	}

	routeName := strings.TrimSpace(req.Name)
	if routeName == "" {
		routeName = old.Name
	}
	if routeName == "" {
		routeName = deriveRouteName(req.Path, old.ID)
		routeName = s.ensureUniqueRouteName(routeName, old.ID)
	}
	if routeName != old.Name {
		if err := s.checkRouteNameUnique(routeName, old.ID); err != nil {
			return nil, err
		}
	}

	microApp := s.resolveUpdateMicroApp(old, req)

	old.ParentID = req.ParentID
	old.MenuName = req.MenuName
	old.Name = routeName
	old.MicroApp = microApp
	old.OrderNum = req.OrderNum
	old.Path = req.Path
	old.Component = req.Component
	old.Query = req.Query
	old.IsFrame = derefInt8(req.IsFrame, old.IsFrame)
	old.IsCache = derefInt8(req.IsCache, old.IsCache)
	old.MenuType = req.MenuType
	old.Visible = derefInt8(req.Visible, old.Visible)
	old.Status = derefInt8(req.Status, old.Status)
	old.Perms = req.Perms
	old.Icon = req.Icon
	if old.Icon == "" {
		old.Icon = "#"
	}
	old.ModuleID = req.ModuleID
	old.Remark = req.Remark
	old.UpdateBy = operator
	old.UpdateTime = time.Now()
	if err := s.db.Save(old).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "更新菜单失败："+err.Error(), 500, nil)
	}
	return old, nil
}

// resolveUpdateMicroApp 决定修改菜单时最终的 microApp。
func (s *Service) resolveUpdateMicroApp(old *models.Menu, req *UpdateMenuRequest) string {
	reqMicro := strings.TrimSpace(req.MicroApp)
	if reqMicro != "" {
		return reqMicro
	}
	if req.ParentID != old.ParentID && req.ParentID != 0 {
		if parent, err := s.GetByID(req.ParentID); err == nil {
			pm := strings.TrimSpace(parent.MicroApp)
			if pm != "" && pm != "main" {
				return pm
			}
		}
	}
	oldMicro := strings.TrimSpace(old.MicroApp)
	if oldMicro == "" {
		return "main"
	}
	return oldMicro
}

// Delete 删除菜单（要求无子节点）。
func (s *Service) Delete(id uint) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	var count int64
	if err := s.db.Model(&models.Menu{}).Where("parent_id = ?", id).Count(&count).Error; err != nil {
		return exception.New(exception.CodeInternalError, "统计子菜单失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return exception.New(exception.CodeMenuHasChildren, "存在子菜单，无法删除", 400, nil)
	}
	if err := s.db.Delete(&models.Menu{}, id).Error; err != nil {
		return exception.New(exception.CodeInternalError, "删除菜单失败："+err.Error(), 500, nil)
	}
	return nil
}

// GetByID 查询单个菜单。
func (s *Service) GetByID(id uint) (*models.Menu, error) {
	var m models.Menu
	if err := s.db.First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeMenuNotFound, "菜单不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询菜单失败："+err.Error(), 500, nil)
	}
	return &m, nil
}

// List 查询菜单列表（平铺，按 parent_id, order_num 排序）。
func (s *Service) List(q *QueryMenuRequest) ([]*models.Menu, error) {
	tx := s.db.Model(&models.Menu{})
	if q.MenuName != "" {
		tx = tx.Where("menu_name LIKE ?", "%"+q.MenuName+"%")
	}
	if q.MenuType != "" {
		tx = tx.Where("menu_type = ?", q.MenuType)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.ModuleID != "" {
		tx = tx.Where("module_id = ?", q.ModuleID)
	}
	if q.MicroApp != "" {
		tx = tx.Where("micro_app = ?", q.MicroApp)
	}
	var list []*models.Menu
	if err := tx.Order("parent_id ASC, order_num ASC, id ASC").Find(&list).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询菜单列表失败："+err.Error(), 500, nil)
	}
	return list, nil
}

// Tree 完整菜单树（管理用，不做权限过滤）。
func (s *Service) Tree() ([]*models.Menu, error) {
	list, err := s.List(&QueryMenuRequest{})
	if err != nil {
		return nil, err
	}
	return buildTree(list), nil
}

// BuildRouters 构建当前登录用户可见的前端路由树。
func (s *Service) BuildRouters(username string, roles []string, perms []string) ([]RouterVO, error) {
	admin := isAdminUser(username, roles)
	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permSet[p] = struct{}{}
	}

	var rows []*models.Menu
	if err := s.db.
		Where("status = ?", 0).
		Where("menu_type IN ?", []string{"M", "C"}).
		Order("parent_id ASC, order_num ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询路由失败："+err.Error(), 500, nil)
	}

	filtered := make([]*models.Menu, 0, len(rows))
	for _, m := range rows {
		if !admin && m.Perms != "" {
			if _, ok := permSet[m.Perms]; !ok {
				continue
			}
		}
		filtered = append(filtered, m)
	}

	tree := buildTree(filtered)
	return toRouterVOs(tree), nil
}

// validateMenuRequest 校验菜单字段组合。
func (s *Service) validateMenuRequest(menuType string, parentID uint, path, perms string) error {
	switch menuType {
	case "M", "C":
		if strings.TrimSpace(path) == "" {
			return exception.New(exception.CodeMenuTypeInvalid, "目录/菜单类型必须填写路由地址", 400, nil)
		}
	case "F":
		if strings.TrimSpace(perms) == "" {
			return exception.New(exception.CodeMenuTypeInvalid, "按钮类型必须填写权限标识", 400, nil)
		}
	default:
		return exception.New(exception.CodeMenuTypeInvalid, "菜单类型无效（M/C/F）", 400, nil)
	}
	if parentID != 0 {
		if _, err := s.GetByID(parentID); err != nil {
			return exception.New(exception.CodeMenuParentInvalid, "父菜单不存在", 400, nil)
		}
	}
	return nil
}

// checkNameUnique 同级菜单名唯一。
func (s *Service) checkNameUnique(parentID uint, name string, excludeID uint) error {
	var count int64
	tx := s.db.Model(&models.Menu{}).
		Where("parent_id = ? AND menu_name = ?", parentID, name)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	if err := tx.Count(&count).Error; err != nil {
		return exception.New(exception.CodeInternalError, "校验菜单名失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return exception.New(exception.CodeMenuNameExists, "同级菜单名已存在", 400, nil)
	}
	return nil
}

// checkPathUnique 路由地址唯一（目录/菜单），按钮忽略 path。
func (s *Service) checkPathUnique(path, menuType string, excludeID uint) error {
	if menuType == "F" || strings.TrimSpace(path) == "" {
		return nil
	}
	var count int64
	tx := s.db.Model(&models.Menu{}).
		Where("path = ? AND menu_type IN ?", path, []string{"M", "C"})
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	if err := tx.Count(&count).Error; err != nil {
		return exception.New(exception.CodeInternalError, "校验路由地址失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return exception.New(exception.CodeMenuPathExists, "路由地址已存在", 400, nil)
	}
	return nil
}

// checkRouteNameUnique 英文路由名全局唯一。
func (s *Service) checkRouteNameUnique(name string, excludeID uint) error {
	var count int64
	tx := s.db.Model(&models.Menu{}).Where("name = ?", name)
	if excludeID > 0 {
		tx = tx.Where("id <> ?", excludeID)
	}
	if err := tx.Count(&count).Error; err != nil {
		return exception.New(exception.CodeInternalError, "校验路由名失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return exception.New(exception.CodeMenuNameExists, "路由名已存在："+name, 400, nil)
	}
	return nil
}

// ensureUniqueRouteName 派生 name 冲突时追加数字后缀，最多重试 999 次。
func (s *Service) ensureUniqueRouteName(base string, excludeID uint) string {
	candidate := base
	for i := 2; i < 1000; i++ {
		var count int64
		tx := s.db.Model(&models.Menu{}).Where("name = ?", candidate)
		if excludeID > 0 {
			tx = tx.Where("id <> ?", excludeID)
		}
		if err := tx.Count(&count).Error; err != nil {
			break
		}
		if count == 0 {
			return candidate
		}
		candidate = fmt.Sprintf("%s%d", base, i)
	}
	return fmt.Sprintf("%s%d", base, time.Now().Unix())
}

// checkCycle 防止父级挂到自己子孙下。
func (s *Service) checkCycle(id, newParentID uint) error {
	if newParentID == 0 {
		return nil
	}
	cur := newParentID
	for cur != 0 {
		if cur == id {
			return exception.New(exception.CodeMenuParentInvalid, "父菜单不能是自身或子孙节点", 400, nil)
		}
		var m models.Menu
		if err := s.db.Select("id", "parent_id").First(&m, cur).Error; err != nil {
			return nil
		}
		cur = m.ParentID
	}
	return nil
}

// isAdminUser 判断是否为超级管理员。
func isAdminUser(username string, roles []string) bool {
	if strings.EqualFold(username, "admin") {
		return true
	}
	for _, r := range roles {
		if _, ok := AdminRoleCodes[strings.ToLower(r)]; ok {
			return true
		}
	}
	return false
}

// buildTree 平铺列表 → 树。
func buildTree(list []*models.Menu) []*models.Menu {
	index := make(map[uint]*models.Menu, len(list))
	for _, m := range list {
		m.Children = nil
		index[m.ID] = m
	}
	roots := make([]*models.Menu, 0, len(list))
	for _, m := range list {
		if m.ParentID == 0 {
			roots = append(roots, m)
			continue
		}
		if parent, ok := index[m.ParentID]; ok {
			parent.Children = append(parent.Children, m)
		}
	}
	sortMenus(roots)
	return roots
}

func sortMenus(list []*models.Menu) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].OrderNum == list[j].OrderNum {
			return list[i].ID < list[j].ID
		}
		return list[i].OrderNum < list[j].OrderNum
	})
	for _, m := range list {
		if len(m.Children) > 0 {
			sortMenus(m.Children)
		}
	}
}

// toRouterVOs 菜单树 → RouterVO 树。
func toRouterVOs(list []*models.Menu) []RouterVO {
	out := make([]RouterVO, 0, len(list))
	for _, m := range list {
		out = append(out, toRouterVO(m, nil, ""))
	}
	return out
}

// toRouterVO 单节点转换。
func toRouterVO(m *models.Menu, parentID *string, parentMicroApp string) RouterVO {
	idStr := fmt.Sprintf("%d", m.ID)
	nodeType := "menu"
	component := m.Component
	if m.MenuType == "M" {
		nodeType = "directory"
		component = "Layout"
	}

	name := strings.TrimSpace(m.Name)
	if name == "" {
		name = deriveRouteName(m.Path, m.ID)
	}

	microApp := resolveMicroApp(m.MicroApp, parentMicroApp)

	meta := &MetaVO{
		Title:      m.MenuName,
		Icon:       m.Icon,
		Permission: m.Perms,
		Hidden:     m.Visible == 1,
		NoCache:    m.IsCache == 1,
	}
	if m.IsFrame == 0 && strings.HasPrefix(m.Path, "http") {
		meta.Link = m.Path
	}

	rv := RouterVO{
		ID:        idStr,
		ParentID:  parentID,
		Name:      name,
		Path:      m.Path,
		Component: component,
		MicroApp:  microApp,
		Type:      nodeType,
		Order:     m.OrderNum,
		Meta:      meta,
	}

	if len(m.Children) > 0 {
		children := make([]RouterVO, 0, len(m.Children))
		for _, c := range m.Children {
			cp := idStr
			children = append(children, toRouterVO(c, &cp, microApp))
		}
		rv.Children = children
	}
	return rv
}

// resolveMicroApp 解析节点最终使用的 microApp。
//
// 继承规则：
//  1. 自身非空且非 "main" → 自身（真正"明确指定"）；
//  2. 自身为空或 "main"，父节点非空且非 "main" → 继承父节点；
//  3. 否则 → "main"。
func resolveMicroApp(self, parent string) string {
	s := strings.TrimSpace(self)
	p := strings.TrimSpace(parent)
	if s != "" && s != "main" {
		return s
	}
	if p != "" && p != "main" {
		return p
	}
	return "main"
}

// deriveRouteName 按 path 派生 PascalCase 路由名。
func deriveRouteName(path string, id uint) string {
	p := strings.Trim(strings.TrimSpace(path), "/")
	if p == "" {
		return fmt.Sprintf("Route%d", id)
	}
	parts := strings.FieldsFunc(p, func(r rune) bool {
		return r == '/' || r == '-' || r == '_'
	})
	var sb strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		sb.WriteString(strings.ToUpper(string(runes[0])))
		if len(runes) > 1 {
			sb.WriteString(string(runes[1:]))
		}
	}
	if sb.Len() == 0 {
		return fmt.Sprintf("Route%d", id)
	}
	return sb.String()
}

func derefInt8(p *int8, def int8) int8 {
	if p == nil {
		return def
	}
	return *p
}

// =============================================================================
// 微应用列表（菜单管理下拉数据源）+ 格式校验
// =============================================================================

// MicroAppOption 微应用选项。
//
// Source 语义：
//   - builtin     ：主应用固定项（main/platform）
//   - module      ：来自 module_manager_module 的已注册应用（status=active）
//   - menu_history：来自 platform_menu 里出现过的历史 micro_app
type MicroAppOption struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

// ListKnownMicroApps 返回所有已知微应用（并集），供菜单管理下拉使用。
//
// 不要求微应用已注册，历史菜单里出现过的 id 也返回，
// 这样即使应用尚未部署，管理员也能看到并复用已有 id。
//
// 三来源顺序：builtin → module → menu_history。
// 同一 id 只返回第一次出现的来源（builtin > module > menu_history）。
func (s *Service) ListKnownMicroApps() ([]MicroAppOption, error) {
	result := []MicroAppOption{
		{ID: "main", Name: "主应用", Source: "builtin"},
		{ID: "platform", Name: "平台基础", Source: "builtin"},
	}
	seen := map[string]bool{"main": true, "platform": true}

	// 来源 1：已注册模块（只取 active）
	// 注意：models.Module.TableName() 返回 "module_manager_module"，
	// 不是 "platform_module"。
	type moduleRow struct {
		ID   string
		Name string
	}
	var mods []moduleRow
	if err := s.db.Table("module_manager_module").
		Select("id, name").
		Where("status = ?", "active").
		Order("id ASC").
		Find(&mods).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			"查询模块失败："+err.Error(), 500, nil,
		)
	}
	for _, m := range mods {
		if seen[m.ID] || strings.TrimSpace(m.ID) == "" {
			continue
		}
		seen[m.ID] = true
		name := strings.TrimSpace(m.Name)
		if name == "" {
			name = m.ID
		}
		result = append(result, MicroAppOption{
			ID: m.ID, Name: name, Source: "module",
		})
	}

	// 来源 2：已有菜单里出现过的 micro_app
	var histIDs []string
	if err := s.db.Model(&models.Menu{}).
		Where("micro_app <> '' AND micro_app NOT IN ?", []string{"main", "platform"}).
		Distinct().
		Pluck("micro_app", &histIDs).Error; err != nil {
		return nil, exception.New(
			exception.CodeInternalError,
			"查询历史微应用失败："+err.Error(), 500, nil,
		)
	}
	for _, id := range histIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, MicroAppOption{
			ID: id, Name: id, Source: "menu_history",
		})
	}

	return result, nil
}

// validateMicroApp 只做格式校验，不要求存在于 module_manager_module。
//
// 设计理由：
//   - 菜单是"配置先行"的元数据，新应用可以在部署前先建菜单；
//   - 菜单点击时如果应用未部署，前端会 404，属于可接受的"配置了但未上线"；
//   - 格式约束保证前端 router/qiankun 能正确路由。
func (s *Service) validateMicroApp(microApp string) error {
	v := strings.TrimSpace(microApp)
	if v == "" {
		return nil // 允许空，由 resolveMicroApp / resolveCreateMicroApp 归一化为 "main"
	}
	if !microAppPattern.MatchString(v) {
		return exception.New(
			exception.CodeParamInvalid,
			"microApp 格式非法：需小写字母开头，仅含小写字母/数字/下划线/连字符，长度 ≤50",
			400, nil,
		)
	}
	return nil
}
