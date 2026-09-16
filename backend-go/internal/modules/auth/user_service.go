package auth
import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"gorm.io/gorm"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/security"
)
// 用户名格式：3-20 位，字母 / 数字 / 下划线
var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
const (
	passwordMinLen = 6
	passwordMaxLen = 20
	nicknameMinLen = 1
	nicknameMaxLen = 50
)
// ---------------------------------------------------------------------------
// 序列化与工具
// ---------------------------------------------------------------------------
func serializeUserOut(user *models.User) *UserOut {
	roles := make([]RoleBrief, 0, len(user.Roles))
	for _, r := range user.Roles {
		roles = append(roles, RoleBrief{ID: r.ID, Name: r.Name, Code: r.Code})
	}
	return &UserOut{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Status:    user.Status,
		Roles:     roles,
		CreatedAt: formatDateTime(user.CreatedAt),
		UpdatedAt: formatDateTime(user.UpdatedAt),
	}
}
func normalizeRoleIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return []uint{}
	}
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}
func normalizeEmail(email *string) *string {
	if email == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*email)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
func validateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return exception.New(exception.CodeAuthUsernameFormat,
			"用户名格式无效（3-20 位字母、数字或下划线）", 400, nil)
	}
	return nil
}
func validatePassword(password string) error {
	if len(password) < passwordMinLen || len(password) > passwordMaxLen {
		return exception.New(exception.CodeAuthPasswordFormat,
			fmt.Sprintf("密码长度必须为 %d-%d 位", passwordMinLen, passwordMaxLen), 400, nil)
	}
	return nil
}
func validateNickname(nickname *string) error {
	if nickname == nil {
		return nil
	}
	if len(*nickname) < nicknameMinLen || len(*nickname) > nicknameMaxLen {
		return exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("昵称长度必须为 %d-%d 位", nicknameMinLen, nicknameMaxLen), 400, nil)
	}
	return nil
}
func mapUserIntegrityError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "user_role") || strings.Contains(msg, "uq_auth_user_role") {
		return exception.New(exception.CodeDataConflict, "用户角色关联冲突", 409, nil)
	}
	if strings.Contains(msg, "auth_user.username") ||
		strings.Contains(msg, "auth_user_username") ||
		strings.Contains(msg, "username") {
		return exception.New(exception.CodeAuthUsernameExists, "用户名已存在", 400, nil)
	}
	if strings.Contains(msg, "auth_user.email") ||
		strings.Contains(msg, "auth_user_email") ||
		strings.Contains(msg, "email") {
		return exception.New(exception.CodeAuthEmailExists, "邮箱已被使用", 409, nil)
	}
	return exception.New(exception.CodeDataConflict, "数据冲突", 409, nil)
}
func loadUserWithRoles(db *gorm.DB, userID uint) (*models.User, error) {
	var user models.User
	err := db.Preload("Roles").First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
// ---------------------------------------------------------------------------
// 列表 / 详情
// ---------------------------------------------------------------------------
func (s *Service) ListUsers(query UserListQuery) (map[string]interface{}, error) {
	var conditions []string
	var args []interface{}
	if query.Keyword != "" {
		like := "%" + query.Keyword + "%"
		conditions = append(conditions,
			"(username LIKE ? OR nickname LIKE ? OR (email IS NOT NULL AND email LIKE ?))")
		args = append(args, like, like, like)
	}
	if query.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *query.Status)
	}
	if query.RoleID != nil {
		conditions = append(conditions,
			"id IN (SELECT user_id FROM auth_user_role WHERE role_id = ?)")
		args = append(args, *query.RoleID)
	}
	base := s.db.Model(&models.User{})
	if len(conditions) > 0 {
		base = base.Where(strings.Join(conditions, " AND "), args...)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询用户失败", 500, nil)
	}
	var users []models.User
	listQuery := s.db.Preload("Roles")
	if len(conditions) > 0 {
		listQuery = listQuery.Where(strings.Join(conditions, " AND "), args...)
	}
	if err := listQuery.
		Order("id ASC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&users).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询用户失败", 500, nil)
	}
	items := make([]*UserOut, 0, len(users))
	for i := range users {
		items = append(items, serializeUserOut(&users[i]))
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
func (s *Service) GetUserDetail(userID uint) (*UserOut, error) {
	user, err := loadUserWithRoles(s.db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询用户失败", 500, nil)
	}
	return serializeUserOut(user), nil
}
// ---------------------------------------------------------------------------
// 创建
// ---------------------------------------------------------------------------
func (s *Service) CreateUser(req UserCreate, operatorID uint, operatorName string) (*UserOut, error) {
	email := normalizeEmail(req.Email)
	roleIDs := normalizeRoleIDs(req.RoleIDs)
	if err := validateUsername(req.Username); err != nil {
		logAuthEvent("user_create", &operatorID, operatorName, "fail",
			intPtr(exception.CodeAuthUsernameFormat), "", "", fmt.Sprintf("用户名格式无效：%s", req.Username))
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		logAuthEvent("user_create", &operatorID, operatorName, "fail",
			intPtr(exception.CodeAuthPasswordFormat), "", "", "密码长度不符合要求")
		return nil, err
	}
	if err := validateNickname(&req.Nickname); err != nil {
		logAuthEvent("user_create", &operatorID, operatorName, "fail",
			intPtr(exception.CodeParamInvalid), "", "", "昵称长度不符合要求")
		return nil, err
	}
	// 用户名唯一
	var count int64
	if err := s.db.Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询用户失败", 500, nil)
	}
	if count > 0 {
		logAuthEvent("user_create", &operatorID, operatorName, "fail",
			intPtr(exception.CodeAuthUsernameExists), "", "", fmt.Sprintf("用户名已存在：%s", req.Username))
		return nil, exception.New(exception.CodeAuthUsernameExists, "用户名已存在", 400, nil)
	}
	// 邮箱唯一
	if email != nil {
		if err := s.db.Model(&models.User{}).Where("email = ?", *email).Count(&count).Error; err != nil {
			return nil, exception.New(exception.CodeInternalError, "查询用户失败", 500, nil)
		}
		if count > 0 {
			logAuthEvent("user_create", &operatorID, operatorName, "fail",
				intPtr(exception.CodeAuthEmailExists), "", "", fmt.Sprintf("邮箱已被使用：%s", *email))
			return nil, exception.New(exception.CodeAuthEmailExists, "邮箱已被使用", 409, nil)
		}
	}
	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "密码哈希失败", 500, nil)
	}
	// 默认值 1（启用），与 Python Pydantic Field(1, ge=0, le=1) 一致；
	// 删除静默改写逻辑；值范围已由 HTTP 绑定层校验。
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	var createdID uint
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var roles []models.Role
		if len(roleIDs) > 0 {
			if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return err
			}
			if len(roles) != len(roleIDs) {
				found := make(map[uint]struct{}, len(roles))
				for _, r := range roles {
					found[r.ID] = struct{}{}
				}
				var missing []uint
				for _, id := range roleIDs {
					if _, ok := found[id]; !ok {
						missing = append(missing, id)
					}
				}
				return exception.New(exception.CodeAuthRoleNotFound,
					fmt.Sprintf("角色不存在：%v", missing), 404, nil)
			}
		}
		user := models.User{
			Username:     req.Username,
			PasswordHash: hashed,
			Nickname:     req.Nickname,
			Email:        email,
			Status:       status,
			Roles:        roles,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		createdID = user.ID
		return nil
	})
	if txErr != nil {
		if pe, ok := txErr.(*exception.PlatformError); ok {
			return nil, pe
		}
		return nil, mapUserIntegrityError(txErr)
	}
	full, err := loadUserWithRoles(s.db, createdID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载用户失败", 500, nil)
	}
	logAuthEvent("user_create", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("创建用户 %s（id=%d）", req.Username, createdID))
	return serializeUserOut(full), nil
}
// ---------------------------------------------------------------------------
// 更新
// ---------------------------------------------------------------------------
func (s *Service) UpdateUser(userID uint, req UserUpdate, operatorID uint, operatorName string) (*UserOut, error) {
	if req.Nickname != nil {
		if err := validateNickname(req.Nickname); err != nil {
			return nil, err
		}
	}
	var roleIDs []uint
	hasRoleUpdate := false
	if req.RoleIDs != nil {
		roleIDs = normalizeRoleIDs(*req.RoleIDs)
		hasRoleUpdate = true
	}
	var newEmail *string
	emailChanged := false
	if req.Email != nil {
		newEmail = normalizeEmail(req.Email)
		emailChanged = true
	}
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{}
		if req.Nickname != nil {
			updates["nickname"] = *req.Nickname
		}
		if emailChanged {
			if newEmail != nil {
				var count int64
				if err := tx.Model(&models.User{}).
					Where("email = ? AND id != ?", *newEmail, userID).
					Count(&count).Error; err != nil {
					return err
				}
				if count > 0 {
					return exception.New(exception.CodeAuthEmailExists, "邮箱已被使用", 409, nil)
				}
			}
			updates["email"] = newEmail
		}
		if req.Status != nil {
			// 值范围已由 HTTP 绑定层校验（binding:"omitempty,min=0,max=1"）
			if userID == operatorID && *req.Status == 0 {
				return exception.New(exception.CodeAuthCannotDisableSelf, "不能禁用自己", 403, nil)
			}
			updates["status"] = *req.Status
		}
		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
		}
		if hasRoleUpdate {
			var newRoles []models.Role
			if len(roleIDs) > 0 {
				if err := tx.Where("id IN ?", roleIDs).Find(&newRoles).Error; err != nil {
					return err
				}
				if len(newRoles) != len(roleIDs) {
					found := make(map[uint]struct{}, len(newRoles))
					for _, r := range newRoles {
						found[r.ID] = struct{}{}
					}
					var missing []uint
					for _, id := range roleIDs {
						if _, ok := found[id]; !ok {
							missing = append(missing, id)
						}
					}
					return exception.New(exception.CodeAuthRoleNotFound,
						fmt.Sprintf("角色不存在：%v", missing), 404, nil)
				}
			}
			if err := tx.Model(&user).Association("Roles").Replace(newRoles); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
		}
		if pe, ok := txErr.(*exception.PlatformError); ok {
			return nil, pe
		}
		return nil, mapUserIntegrityError(txErr)
	}
	full, err := loadUserWithRoles(s.db, userID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载用户失败", 500, nil)
	}
	logAuthEvent("user_update", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("更新用户 id=%d", userID))
	return serializeUserOut(full), nil
}
// ---------------------------------------------------------------------------
// 删除
// ---------------------------------------------------------------------------
func (s *Service) DeleteUser(userID uint, operatorID uint, operatorName string) error {
	if userID == operatorID {
		return exception.New(exception.CodeAuthCannotDeleteSelf, "不能删除自己", 403, nil)
	}
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		if err := revokeAllUserTokensTx(tx, userID); err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.User{}, userID).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError, "删除用户失败", 500, nil)
	}
	logAuthEvent("user_delete", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("删除用户 id=%d", userID))
	return nil
}
// ---------------------------------------------------------------------------
// 启用/禁用
// ---------------------------------------------------------------------------
func (s *Service) UpdateUserStatus(userID uint, status int8, operatorID uint, operatorName string) (*UserOut, error) {
	if userID == operatorID && status == 0 {
		return nil, exception.New(exception.CodeAuthCannotDisableSelf, "不能禁用自己", 403, nil)
	}
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		return tx.Model(&user).Update("status", status).Error
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "更新状态失败", 500, nil)
	}
	full, err := loadUserWithRoles(s.db, userID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载用户失败", 500, nil)
	}
	logAuthEvent("user_status", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("用户 id=%d 状态变更为 %d", userID, status))
	return serializeUserOut(full), nil
}
// ---------------------------------------------------------------------------
// 重置密码
// ---------------------------------------------------------------------------
func (s *Service) ResetUserPassword(userID uint, newPassword string, operatorID uint, operatorName string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hashed, err := security.HashPassword(newPassword)
	if err != nil {
		return exception.New(exception.CodeInternalError, "密码哈希失败", 500, nil)
	}
	var username string
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		username = user.Username
		if err := tx.Model(&user).Update("password_hash", hashed).Error; err != nil {
			return err
		}
		return revokeAllUserTokensTx(tx, userID)
	})
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError, "重置密码失败", 500, nil)
	}
	logAuthEvent("user_password_reset", &operatorID, operatorName, "success", nil,
		"", "", fmt.Sprintf("管理员重置用户 id=%d（%s）的密码", userID, username))
	return nil
}
