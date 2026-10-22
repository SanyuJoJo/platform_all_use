package auth

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// ListRoles 分页查询角色列表，附 menuCount（不返回 menuIds）。
func (s *Service) ListRoles(page, pageSize int, keyword string) (*RoleListData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	type roleRow struct {
		ID          uint      `gorm:"column:id"`
		Name        string    `gorm:"column:name"`
		Code        string    `gorm:"column:code"`
		Description *string   `gorm:"column:description"`
		IsSystem    int8      `gorm:"column:is_system"`
		MenuCount   int64     `gorm:"column:menu_count"`
		CreatedAt   time.Time `gorm:"column:created_at"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
	}

	kw := strings.TrimSpace(keyword)

	var total int64
	countTx := s.db.Table("auth_role AS r")
	if kw != "" {
		like := "%" + kw + "%"
		countTx = countTx.Where("r.name LIKE ? OR r.code LIKE ?", like, like)
	}
	if err := countTx.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "统计角色数失败："+err.Error(), 500, nil)
	}

	var rows []roleRow
	q := s.db.Table("auth_role AS r").
		Select(`r.id, r.name, r.code, r.description, r.is_system,
		        r.created_at, r.updated_at,
		        COUNT(rm.menu_id) AS menu_count`).
		Joins("LEFT JOIN auth_role_menu AS rm ON rm.role_id = r.id")
	if kw != "" {
		like := "%" + kw + "%"
		q = q.Where("r.name LIKE ? OR r.code LIKE ?", like, like)
	}
	err := q.Group("r.id, r.name, r.code, r.description, r.is_system, r.created_at, r.updated_at").
		Order("r.id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询角色列表失败："+err.Error(), 500, nil)
	}

	items := make([]RoleListItem, 0, len(rows))
	for _, r := range rows {
		desc := ""
		if r.Description != nil {
			desc = *r.Description
		}
		items = append(items, RoleListItem{
			ID:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Description: desc,
			IsSystem:    r.IsSystem == 1,
			MenuCount:   r.MenuCount,
			CreatedAt:   formatISO(r.CreatedAt),
			UpdatedAt:   formatISO(r.UpdatedAt),
		})
	}

	pages := 0
	if total > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &RoleListData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}, nil
}

// GetRoleDetail 查询角色详情（含 menuIds）。
func (s *Service) GetRoleDetail(id uint) (*RoleDetail, error) {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeAuthRoleNotFound, "角色不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询角色失败："+err.Error(), 500, nil)
	}

	var menuIDs []uint
	if err := s.db.Model(&models.AuthRoleMenu{}).
		Where("role_id = ?", id).
		Order("menu_id ASC").
		Pluck("menu_id", &menuIDs).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询角色菜单失败："+err.Error(), 500, nil)
	}
	if menuIDs == nil {
		menuIDs = []uint{}
	}

	desc := ""
	if role.Description != nil {
		desc = *role.Description
	}

	return &RoleDetail{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: desc,
		IsSystem:    role.IsSystem == 1,
		MenuIDs:     menuIDs,
		CreatedAt:   formatISO(role.CreatedAt),
		UpdatedAt:   formatISO(role.UpdatedAt),
	}, nil
}

// CreateRole 新增角色（事务：插入角色 + 校验并插入菜单关联）。
func (s *Service) CreateRole(req *CreateRoleRequest) (*RoleDetail, error) {
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(req.Code)
	if name == "" || code == "" {
		return nil, exception.New(exception.CodeParamInvalid, "name/code 必填", 400, nil)
	}
	if err := validateRoleName(name); err != nil {
		return nil, err
	}
	if err := validateRoleCode(code); err != nil {
		return nil, err
	}

	var count int64
	if err := s.db.Model(&models.Role{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "校验角色名失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return nil, exception.New(exception.CodeAuthRoleNameExists, "角色名已存在", 400, nil)
	}
	if err := s.db.Model(&models.Role{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "校验角色码失败："+err.Error(), 500, nil)
	}
	if count > 0 {
		return nil, exception.New(exception.CodeAuthRoleCodeExists, "角色码已存在", 400, nil)
	}

	desc := req.Description
	role := &models.Role{
		Name:        name,
		Code:        code,
		Description: &desc,
		IsSystem:    0,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return exception.New(exception.CodeInternalError, "创建角色失败："+err.Error(), 500, nil)
		}
		if len(req.MenuIDs) > 0 {
			cleanIDs := dedupeUint(req.MenuIDs)
			if err := ensureMenusExist(tx, cleanIDs); err != nil {
				return err
			}
			rows := make([]models.AuthRoleMenu, 0, len(cleanIDs))
			for _, mid := range cleanIDs {
				rows = append(rows, models.AuthRoleMenu{RoleID: role.ID, MenuID: mid})
			}
			if err := tx.CreateInBatches(rows, 100).Error; err != nil {
				return exception.New(exception.CodeInternalError, "写入角色菜单失败："+err.Error(), 500, nil)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetRoleDetail(role.ID)
}

// UpdateRole 修改角色（事务：更新字段 + 全量替换菜单关联）。
func (s *Service) UpdateRole(id uint, req *UpdateRoleRequest) (*RoleDetail, error) {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.New(exception.CodeAuthRoleNotFound, "角色不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询角色失败："+err.Error(), 500, nil)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, exception.New(exception.CodeParamInvalid, "name 必填", 400, nil)
	}
	if err := validateRoleName(name); err != nil {
		return nil, err
	}
	if name != role.Name {
		var count int64
		if err := s.db.Model(&models.Role{}).
			Where("name = ? AND id <> ?", name, id).
			Count(&count).Error; err != nil {
			return nil, exception.New(exception.CodeInternalError, "校验角色名失败："+err.Error(), 500, nil)
		}
		if count > 0 {
			return nil, exception.New(exception.CodeAuthRoleNameExists, "角色名已存在", 400, nil)
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"name":        name,
			"description": req.Description,
			"updated_at":  time.Now(),
		}
		if err := tx.Model(&models.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return exception.New(exception.CodeInternalError, "更新角色失败："+err.Error(), 500, nil)
		}

		if req.MenuIDs == nil {
			return nil // 不更新菜单
		}
		cleanIDs := dedupeUint(*req.MenuIDs)
		if len(cleanIDs) > 0 {
			if err := ensureMenusExist(tx, cleanIDs); err != nil {
				return err
			}
		}
		if err := tx.Where("role_id = ?", id).Delete(&models.AuthRoleMenu{}).Error; err != nil {
			return exception.New(exception.CodeInternalError, "清除旧菜单关联失败："+err.Error(), 500, nil)
		}
		if len(cleanIDs) == 0 {
			return nil
		}
		rows := make([]models.AuthRoleMenu, 0, len(cleanIDs))
		for _, mid := range cleanIDs {
			rows = append(rows, models.AuthRoleMenu{RoleID: id, MenuID: mid})
		}
		if err := tx.CreateInBatches(rows, 100).Error; err != nil {
			return exception.New(exception.CodeInternalError, "写入角色菜单失败："+err.Error(), 500, nil)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetRoleDetail(id)
}

// DeleteRole 删除角色。
func (s *Service) DeleteRole(id uint) error {
	var role models.Role
	if err := s.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return exception.New(exception.CodeAuthRoleNotFound, "角色不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError, "查询角色失败："+err.Error(), 500, nil)
	}
	if role.IsSystem == 1 {
		return exception.New(exception.CodeAuthSystemRoleProtected, "系统内置角色不可删除", 400, nil)
	}

	var userCount int64
	if err := s.db.Model(&models.UserRole{}).Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		return exception.New(exception.CodeInternalError, "校验角色使用情况失败："+err.Error(), 500, nil)
	}
	if userCount > 0 {
		return exception.New(exception.CodeAuthRoleInUse, "该角色下存在用户，无法删除", 400, nil)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", id).Delete(&models.AuthRoleMenu{}).Error; err != nil {
			return exception.New(exception.CodeInternalError, "清除角色菜单失败："+err.Error(), 500, nil)
		}
		if err := tx.Delete(&models.Role{}, id).Error; err != nil {
			return exception.New(exception.CodeInternalError, "删除角色失败："+err.Error(), 500, nil)
		}
		return nil
	})
}

// ensureMenusExist 校验所有 menu_id 都存在于 platform_menu。
func ensureMenusExist(tx *gorm.DB, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.Menu{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return exception.New(exception.CodeInternalError, "校验菜单存在性失败："+err.Error(), 500, nil)
	}
	if int(count) != len(ids) {
		return exception.New(
			exception.CodeParamInvalid,
			fmt.Sprintf("menuIds 中存在无效菜单（有效 %d / 提交 %d）", count, len(ids)),
			400, nil,
		)
	}
	return nil
}

// validateRoleName 角色名规则：1~50 字符。
func validateRoleName(name string) error {
	if len(name) < 1 || len(name) > 50 {
		return exception.New(exception.CodeParamInvalid, "角色名长度须在 1~50 之间", 400, nil)
	}
	return nil
}

// validateRoleCode 角色码规则：^[a-z][a-z0-9_]*$，2~50 字符。
func validateRoleCode(code string) error {
	if len(code) < 2 || len(code) > 50 {
		return exception.New(exception.CodeParamInvalid, "角色码长度须在 2~50 之间", 400, nil)
	}
	if code[0] < 'a' || code[0] > 'z' {
		return exception.New(exception.CodeAuthUsernameFormat, "角色码须以小写字母开头", 400, nil)
	}
	for i := 1; i < len(code); i++ {
		c := code[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			continue
		}
		return exception.New(exception.CodeAuthUsernameFormat, "角色码仅允许小写字母/数字/下划线", 400, nil)
	}
	return nil
}

// dedupeUint 去重并保持稳定顺序。
func dedupeUint(in []uint) []uint {
	seen := make(map[uint]struct{}, len(in))
	out := make([]uint, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// formatISO 输出与 Python datetime.isoformat() 一致的时间字符串。
func formatISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	u := t.UTC()
	if u.Nanosecond()/1000 == 0 {
		return u.Format("2006-01-02T15:04:05+00:00")
	}
	return u.Format("2006-01-02T15:04:05.000000+00:00")
}
