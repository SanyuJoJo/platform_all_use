package auth
import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"gorm.io/gorm"
	"backend-go/internal/exception"
	"backend-go/internal/models"
)
var moduleIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var permissionCodePattern = regexp.MustCompile(`^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$`)
func serializePermissionOut(p *models.Permission) *PermissionOut {
	return &PermissionOut{
		ID:        p.ID,
		Code:      p.Code,
		Name:      p.Name,
		ModuleID:  p.ModuleID,
		Resource:  p.Resource,
		Action:    p.Action,
		CreatedAt: formatDateTime(p.CreatedAt),
	}
}
func validateModuleID(moduleID string) error {
	if moduleID == "" {
		return exception.New(exception.CodeParamInvalid, "module_id 必须为非空字符串", 400, nil)
	}
	if len(moduleID) > 50 || !moduleIDPattern.MatchString(moduleID) {
		return exception.New(exception.CodeParamInvalid,
			"module_id 格式无效（小写字母开头，仅允许小写字母、数字、下划线）", 400, nil)
	}
	return nil
}
func validatePermissionItem(moduleID string, item PermissionRegisterItem) error {
	if item.Code == "" || item.Name == "" || item.Resource == "" || item.Action == "" {
		return exception.New(exception.CodeParamInvalid, "权限字段必须为非空字符串", 400, nil)
	}
	if !permissionCodePattern.MatchString(item.Code) {
		return exception.New(exception.CodeAuthPermissionCodeFormat,
			fmt.Sprintf("权限编码格式无效（必须为小写 {module}:{resource}:{action}）：%s", item.Code), 400, nil)
	}
	parts := strings.Split(item.Code, ":")
	if len(parts) != 3 {
		return exception.New(exception.CodeAuthPermissionCodeFormat,
			fmt.Sprintf("权限编码格式无效：%s", item.Code), 400, nil)
	}
	if parts[0] != moduleID {
		return exception.New(exception.CodeAuthPermissionCodeFormat,
			fmt.Sprintf("权限编码模块前缀与 module_id 不一致：%s", item.Code), 400, nil)
	}
	if parts[1] != item.Resource || parts[2] != item.Action {
		return exception.New(exception.CodeAuthPermissionCodeFormat,
			fmt.Sprintf("权限编码与 resource/action 不一致：%s", item.Code), 400, nil)
	}
	if len(item.Name) < 1 || len(item.Name) > 50 {
		return exception.New(exception.CodeParamInvalid, "权限名称长度必须为 1-50 位", 400, nil)
	}
	return nil
}
func (s *Service) ListPermissions(moduleID, resource string) ([]*PermissionOut, error) {
	if moduleID != "" {
		if err := validateModuleID(moduleID); err != nil {
			return nil, err
		}
	}
	q := s.db.Model(&models.Permission{})
	if moduleID != "" {
		q = q.Where("module_id = ?", moduleID)
	}
	if resource != "" {
		q = q.Where("resource = ?", resource)
	}
	var perms []models.Permission
	if err := q.Order("module_id ASC, resource ASC, action ASC, id ASC").
		Find(&perms).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询权限失败", 500, nil)
	}
	result := make([]*PermissionOut, 0, len(perms))
	for i := range perms {
		result = append(result, serializePermissionOut(&perms[i]))
	}
	return result, nil
}
func (s *Service) GetPermissionsByModule(moduleID string) ([]*PermissionOut, error) {
	if err := validateModuleID(moduleID); err != nil {
		return nil, err
	}
	return s.ListPermissions(moduleID, "")
}
func (s *Service) AuditPermissionFormat() (map[string]interface{}, error) {
	var perms []models.Permission
	if err := s.db.Find(&perms).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询权限失败", 500, nil)
	}
	var invalid []map[string]interface{}
	for _, p := range perms {
		if !permissionCodePattern.MatchString(p.Code) {
			invalid = append(invalid, map[string]interface{}{
				"id": p.ID, "code": p.Code, "module_id": p.ModuleID,
				"resource": p.Resource, "action": p.Action,
			})
		}
	}
	return map[string]interface{}{
		"total":         len(perms),
		"invalid_count": len(invalid),
		"invalid":       invalid,
	}, nil
}
func (s *Service) RegisterPermissions(moduleID string, items []PermissionRegisterItem) (map[string]int, error) {
	if err := validateModuleID(moduleID); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return map[string]int{"created": 0, "updated": 0, "total": 0}, nil
	}
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if err := validatePermissionItem(moduleID, item); err != nil {
			return nil, err
		}
		if _, ok := seen[item.Code]; ok {
			return nil, exception.New(exception.CodeParamInvalid,
				fmt.Sprintf("权限编码重复：%s", item.Code), 400, nil)
		}
		seen[item.Code] = struct{}{}
	}
	created, updated := 0, 0
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var existing models.Permission
			err := tx.Where("code = ?", item.Code).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				perm := models.Permission{
					Code:     item.Code,
					Name:     item.Name,
					ModuleID: moduleID,
					Resource: item.Resource,
					Action:   item.Action,
				}
				if err := tx.Create(&perm).Error; err != nil {
					return err
				}
				created++
				continue
			} else if err != nil {
				return err
			}
			if existing.ModuleID != moduleID {
				return exception.New(exception.CodeAuthPermissionModuleConflict,
					fmt.Sprintf("权限编码已被模块 %s 占用：%s", existing.ModuleID, item.Code), 409, nil)
			}
			if err := tx.Model(&existing).Update("name", item.Name).Error; err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	if err != nil {
		errCode := exception.CodeInternalError
		var pe *exception.PlatformError
		if errors.As(err, &pe) {
			errCode = pe.Code
		}
		LogAuthEvent("permission_register", nil, "", "fail",
			intPtr(errCode), "", "",
			fmt.Sprintf("注册权限失败：module_id=%s err=%v", moduleID, err))
		return nil, err
	}
	LogAuthEvent("permission_register", nil, "", "success", nil, "", "",
		fmt.Sprintf("注册权限 module_id=%s created=%d updated=%d", moduleID, created, updated))
	return map[string]int{"created": created, "updated": updated, "total": created + updated}, nil
}
func (s *Service) UnregisterPermissions(moduleID string) (int, error) {
	if err := validateModuleID(moduleID); err != nil {
		return 0, err
	}
	var permIDs []uint
	if err := s.db.Model(&models.Permission{}).
		Where("module_id = ?", moduleID).
		Pluck("id", &permIDs).Error; err != nil {
		return 0, exception.New(exception.CodeInternalError, "查询权限失败", 500, nil)
	}
	if len(permIDs) == 0 {
		return 0, nil
	}
	deleted := 0
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("permission_id IN ?", permIDs).
			Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		res := tx.Where("id IN ?", permIDs).Delete(&models.Permission{})
		if res.Error != nil {
			return res.Error
		}
		deleted = int(res.RowsAffected)
		return nil
	})
	if err != nil {
		return 0, exception.New(exception.CodeInternalError, "清理权限失败", 500, nil)
	}
	LogAuthEvent("permission_unregister", nil, "", "success", nil, "", "",
		fmt.Sprintf("清理权限 module_id=%s deleted=%d", moduleID, deleted))
	return deleted, nil
}
// ---------------------------------------------------------------------------
// v1.1 新增：供模块管理复用的导出函数
// ---------------------------------------------------------------------------
// RegisterPermissionsTx 事务内权限注册（供模块管理使用）。
//
// 与 RegisterPermissions 的区别：
//   - 复用调用方传入的事务，不自行开启/提交事务；
//   - 不做日志记录（由调用方记录）；
//   - 校验逻辑与 RegisterPermissions 完全一致。
func RegisterPermissionsTx(tx *gorm.DB, moduleID string, items []PermissionRegisterItem) error {
	if err := validateModuleID(moduleID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if err := validatePermissionItem(moduleID, item); err != nil {
			return err
		}
		if _, ok := seen[item.Code]; ok {
			return exception.New(exception.CodeParamInvalid,
				fmt.Sprintf("权限编码重复：%s", item.Code), 400, nil)
		}
		seen[item.Code] = struct{}{}
	}
	for _, item := range items {
		var existing models.Permission
		err := tx.Where("code = ?", item.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			perm := models.Permission{
				Code:     item.Code,
				Name:     item.Name,
				ModuleID: moduleID,
				Resource: item.Resource,
				Action:   item.Action,
			}
			if err := tx.Create(&perm).Error; err != nil {
				return err
			}
			continue
		} else if err != nil {
			return err
		}
		if existing.ModuleID != moduleID {
			return exception.New(exception.CodeAuthPermissionModuleConflict,
				fmt.Sprintf("权限编码已被模块 %s 占用：%s", existing.ModuleID, item.Code), 409, nil)
		}
		if err := tx.Model(&existing).Update("name", item.Name).Error; err != nil {
			return err
		}
	}
	return nil
}
// UnregisterPermissionsTx 事务内权限清理（供模块管理使用）。
func UnregisterPermissionsTx(tx *gorm.DB, moduleID string) error {
	if err := validateModuleID(moduleID); err != nil {
		return err
	}
	var permIDs []uint
	if err := tx.Model(&models.Permission{}).
		Where("module_id = ?", moduleID).
		Pluck("id", &permIDs).Error; err != nil {
		return err
	}
	if len(permIDs) == 0 {
		return nil
	}
	if err := tx.Where("permission_id IN ?", permIDs).
		Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}
	return tx.Where("id IN ?", permIDs).Delete(&models.Permission{}).Error
}
