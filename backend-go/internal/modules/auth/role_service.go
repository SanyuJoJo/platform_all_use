package auth
import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"gorm.io/gorm"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)
var (
	roleCodePattern       = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	permissionCodePattern = regexp.MustCompile(`^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$`)
)
const (
	roleNameMaxLen    = 50
	roleCodeMaxLen    = 50
	descriptionMaxLen = 255
)
func serializeRoleOut(role *models.Role) *RoleOut {
	codes := make([]string, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		codes = append(codes, p.Code)
	}
	sort.Strings(codes)
	return &RoleOut{
		ID:              role.ID,
		Name:            role.Name,
		Code:            role.Code,
		Description:     role.Description,
		IsSystem:        role.IsSystem,
		PermissionCodes: codes,
		CreatedAt:       formatDateTime(role.CreatedAt),
		UpdatedAt:       formatDateTime(role.UpdatedAt),
	}
}
func validateRoleName(name string) error {
	if len(name) < 1 || len(name) > roleNameMaxLen {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("角色名称长度必须为 1-%d 位", roleNameMaxLen), 400, nil)
	}
	return nil
}
func validateRoleCode(code string) error {
	if len(code) < 1 || len(code) > roleCodeMaxLen {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("角色编码长度必须为 1-%d 位", roleCodeMaxLen), 400, nil)
	}
	if !roleCodePattern.MatchString(code) {
		return exception.New(exception.CodeParamInvalid,
			"角色编码格式无效（小写字母开头，仅允许小写字母、数字、下划线）", 400, nil)
	}
	return nil
}
func validateRoleDescription(desc *string) error {
	if desc == nil {
		return nil
	}
	if len(*desc) > descriptionMaxLen {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("角色描述长度不得超过 %d 位", descriptionMaxLen), 400, nil)
	}
	return nil
}
func normalizePermissionCodes(codes []string) []string {
	if len(codes) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(codes))
	result := make([]string, 0, len(codes))
	for _, c := range codes {
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			result = append(result, c)
		}
	}
	return result
}
func validatePermissionCodeFormat(codes []string) error {
	for _, c := range codes {
		if !permissionCodePattern.MatchString(c) {
			return exception.New(exception.CodeAuthPermissionCodeFormat,
				fmt.Sprintf("权限编码格式无效：%s", c), 400, nil)
		}
	}
	return nil
}
func loadRoleWithPermissions(db *gorm.DB, roleID uint) (*models.Role, error) {
	var role models.Role
	if err := db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
func mapRoleIntegrityError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "auth_role.code") ||
		strings.Contains(msg, "auth_role_code") ||
		strings.Contains(msg, "ix_auth_role_code") ||
		strings.Contains(msg, "uq_auth_role_code") {
		return exception.New(exception.CodeAuthRoleCodeExists, "角色编码已存在", 400, nil)
	}
	return exception.New(exception.CodeDataConflict, "数据冲突", 409, nil)
}
func (s *Service) ListRoles(query RoleListQuery) (map[string]interface{}, error) {
	var conditions []string
	var args []interface{}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		conditions = append(conditions, "(name LIKE ? OR code LIKE ?)")
		args = append(args, like, like)
	}
	base := s.db.Model(&models.Role{})
	if len(conditions) > 0 {
		base = base.Where(strings.Join(conditions, " AND "), args...)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询角色失败", 500, nil)
	}
	var roles []models.Role
	q := s.db.Preload("Permissions")
	if len(conditions) > 0 {
		q = q.Where(strings.Join(conditions, " AND "), args...)
	}
	if err := q.Order("id ASC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&roles).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询角色失败", 500, nil)
	}
	items := make([]*RoleOut, 0, len(roles))
	for i := range roles {
		items = append(items, serializeRoleOut(&roles[i]))
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
func (s *Service) GetRoleDetail(roleID uint) (*RoleOut, error) {
	role, err := loadRoleWithPermissions(s.db, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeAuthRoleNotFound, "角色不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询角色失败", 500, nil)
	}
	return serializeRoleOut(role), nil
}
func (s *Service) CreateRole(req RoleCreate, operatorID uint, operatorName string) (*RoleOut, error) {
	if err := validateRoleName(req.Name); err != nil {
		return nil, err
	}
	if err := validateRoleCode(req.Code); err != nil {
		return nil, err
	}
	if err := validateRoleDescription(req.Description); err != nil {
		return nil, err
	}
	codes := normalizePermissionCodes(req.PermissionCodes)
	if err := validatePermissionCodeFormat(codes); err != nil {
		return nil, err
	}
	var count int64
	if err := s.db.Model(&models.Role{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询角色失败", 500, nil)
	}
	if count > 0 {
		logAuthEvent("role_create", &operatorID, operatorName, "fail",
			intPtr(exception.CodeAuthRoleNameExists), "", "", fmt.Sprintf("角色名称已存在：%s", req.Name))
		return nil, exception.New(exception.CodeAuthRoleNameExists, "角色名称已存在", 400, nil)
	}
	if err := s.db.Model(&models.Role{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询角色失败", 500, nil)
	}
	if count > 0 {
		logAuthEvent("role_create", &operatorID, operatorName, "fail",
			intPtr(exception.CodeAuthRoleCodeExists), "", "", fmt.Sprintf("角色编码已存在：%s", req.Code))
		return nil, exception.New(exception.CodeAuthRoleCodeExists, "角色编码已存在", 400, nil)
	}
	var createdID uint
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var perms []models.Permission
		if len(codes) > 0 {
			if err := tx.Where("code IN ?", codes).Find(&perms).Error; err != nil {
				return err
			}
			if len(perms) != len(codes) {
				found := make(map[string]struct{}, len(perms))
				for _, p := range perms {
					found[p.Code] = struct{}{}
				}
				var missing []string
				for _, c := range codes {
					if _, ok := found[c]; !ok {
						missing = append(missing, c)
					}
				}
				return exception.New(exception.CodeAuthPermissionNotFound,
					fmt.Sprintf("权限不存在：%v", missing), 404, nil)
			}
		}
		role := models.Role{
			Name:        req.Name,
			Code:        req.Code,
			Description: req.Description,
			IsSystem:    0,
			Permissions: perms,
		}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		createdID = role.ID
		return nil
	})
	if txErr != nil {
		if pe, ok := txErr.(*exception.PlatformError); ok {
			return nil, pe
		}
		return nil, mapRoleIntegrityError(txErr)
	}
	full, err := loadRoleWithPermissions(s.db, createdID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载角色失败", 500, nil)
	}
	logAuthEvent("role_create", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("创建角色 %s（id=%d, code=%s）", req.Name, createdID, req.Code))
	return serializeRoleOut(full), nil
}
func (s *Service) UpdateRole(roleID uint, req RoleUpdate, operatorID uint, operatorName string) (*RoleOut, error) {
	if req.Name != nil {
		if err := validateRoleName(*req.Name); err != nil {
			return nil, err
		}
	}
	if req.Description != nil {
		if err := validateRoleDescription(req.Description); err != nil {
			return nil, err
		}
	}
	var permissionCodes []string
	hasPermissionUpdate := false
	if req.PermissionCodes != nil {
		permissionCodes = normalizePermissionCodes(*req.PermissionCodes)
		if err := validatePermissionCodeFormat(permissionCodes); err != nil {
			return nil, err
		}
		hasPermissionUpdate = true
	}
	var roleCode string
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return err
		}
		roleCode = role.Code
		if req.Name != nil {
			var count int64
			if err := tx.Model(&models.Role{}).
				Where("name = ? AND id != ?", *req.Name, roleID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return exception.New(exception.CodeAuthRoleNameExists, "角色名称已存在", 400, nil)
			}
			if err := tx.Model(&role).Update("name", *req.Name).Error; err != nil {
				return err
			}
		}
		if req.Description != nil {
			if err := tx.Model(&role).Update("description", req.Description).Error; err != nil {
				return err
			}
		}
		if hasPermissionUpdate {
			var perms []models.Permission
			if len(permissionCodes) > 0 {
				if err := tx.Where("code IN ?", permissionCodes).Find(&perms).Error; err != nil {
					return err
				}
				if len(perms) != len(permissionCodes) {
					found := make(map[string]struct{}, len(perms))
					for _, p := range perms {
						found[p.Code] = struct{}{}
					}
					var missing []string
					for _, c := range permissionCodes {
						if _, ok := found[c]; !ok {
							missing = append(missing, c)
						}
					}
					return exception.New(exception.CodeAuthPermissionNotFound,
						fmt.Sprintf("权限不存在：%v", missing), 404, nil)
				}
			}
			if err := tx.Model(&role).Association("Permissions").Replace(perms); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeAuthRoleNotFound, "角色不存在", 404, nil)
		}
		if pe, ok := txErr.(*exception.PlatformError); ok {
			return nil, pe
		}
		return nil, mapRoleIntegrityError(txErr)
	}
	full, err := loadRoleWithPermissions(s.db, roleID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载角色失败", 500, nil)
	}
	logAuthEvent("role_update", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("更新角色 id=%d（code=%s）", roleID, roleCode))
	return serializeRoleOut(full), nil
}
func (s *Service) DeleteRole(roleID uint, operatorID uint, operatorName string) error {
	var roleCode, roleName string
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var role models.Role
		if err := tx.Preload("Permissions").First(&role, roleID).Error; err != nil {
			return err
		}
		if role.IsSystem == 1 {
			return exception.New(exception.CodeAuthSystemRoleProtected,
				"不能删除系统内置角色", 403, nil)
		}
		roleCode = role.Code
		roleName = role.Name
		var userCount int64
		if err := tx.Model(&models.UserRole{}).
			Where("role_id = ?", roleID).
			Count(&userCount).Error; err != nil {
			return err
		}
		if userCount > 0 {
			return exception.New(exception.CodeAuthRoleInUse,
				"角色已被用户使用，请先解除关联", 409, nil)
		}
		if err := tx.Model(&role).Association("Permissions").Clear(); err != nil {
			return err
		}
		if err := tx.Delete(&models.Role{}, roleID).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeAuthRoleNotFound, "角色不存在", 404, nil)
		}
		if pe, ok := txErr.(*exception.PlatformError); ok {
			if pe.Code == exception.CodeAuthRoleInUse {
				logAuthEvent("role_delete", &operatorID, operatorName, "fail",
					intPtr(exception.CodeAuthRoleInUse), "", "",
					fmt.Sprintf("角色被使用，id=%d", roleID))
			}
			return pe
		}
		return mapRoleIntegrityError(txErr)
	}
	logAuthEvent("role_delete", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("删除角色 id=%d（code=%s, name=%s）", roleID, roleCode, roleName))
	return nil
}
