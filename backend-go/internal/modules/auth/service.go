package auth
import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/security"
)
const (
	AccessTokenExpireMinutes = 1440
	RefreshTokenExpireDays   = 7
)
// Service 认证服务。
type Service struct {
	db  *gorm.DB
	cfg *config.Config
}
// NewService 创建认证服务。
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}
// errRefreshConflict 刷新 Token 并发冲突哨兵错误。
var errRefreshConflict = errors.New("refresh token concurrent conflict")
// hashToken 计算 Token 的 SHA-256 摘要。
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
// utcNowNaive 返回 UTC naive 时间。
func utcNowNaive() time.Time {
	return time.Now().UTC()
}
// strPtrOrNil 空字符串转 nil；非空字符串转指针。
//
// v1.1（P2-04）：与 Python 存 NULL 的语义对齐，避免空字符串落库。
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
// ptrToString 指针字符串转字符串；nil 转空串。
//
// v1.2（建议-5）：`RefreshAccessToken` 成功日志复用 record 的 IP/UA 时使用。
func ptrToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
// truncate 截断字符串到指定长度。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
// logAuthEvent 记录认证事件（结构化日志）。
//
// v1.1（P2-05）：扩展签名，新增 ip / userAgent 字段。
// v1.2（P2-NEW-02）：调用方传真实 IP/UA，不再传空串。
func logAuthEvent(action string, userID *uint, username, status string, errorCode *int, ip, userAgent, detail string) {
	evt := log.Info()
	if status == "fail" {
		evt = log.Warn()
	}
	evt.
		Str("event", "AUTH_EVENT").
		Str("action", action).
		Interface("user_id", userID).
		Str("username", username).
		Str("status", status).
		Interface("error_code", errorCode).
		Str("ip", ip).
		Str("user_agent", truncate(userAgent, 120)).
		Str("detail", detail).
		Msg("auth_event")
}
// loadUserWithRBAC 按 ID 加载用户并预加载角色与权限。
func (s *Service) loadUserWithRBAC(userID uint) (*models.User, error) {
	var user models.User
	err := s.db.
		Preload("Roles.Permissions").
		First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
// loadUserByUsername 按用户名加载用户。
func (s *Service) loadUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.
		Preload("Roles.Permissions").
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
// serializeUser 将 User 序列化为 UserInfo。
//
// v1.1（P2-08 / P2-09）：角色与权限列表使用 sort.Strings 排序，
//   保证响应字段顺序稳定；替代手写冒泡排序。
func serializeUser(user *models.User) *UserInfo {
	roles := make([]string, 0, len(user.Roles))
	permSet := make(map[string]struct{})
	for _, role := range user.Roles {
		roles = append(roles, role.Code)
		for _, perm := range role.Permissions {
			permSet[perm.Code] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for code := range permSet {
		permissions = append(permissions, code)
	}
	sort.Strings(roles)
	sort.Strings(permissions)
	return &UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Status:      user.Status,
		Roles:       roles,
		Permissions: permissions,
	}
}
// AuthenticateUser 校验用户名密码，更新最后登录信息，返回用户信息。
//
// v1.1（P2-03）：使用显式 Update 更新 last_login_*，替代 db.Save(user)，
//   避免因 Save 保存所有字段而误写关联。
func (s *Service) AuthenticateUser(username, password, ip, userAgent string) (*UserInfo, error) {
	user, err := s.loadUserByUsername(username)
	if err != nil || !security.VerifyPassword(password, user.PasswordHash) {
		logAuthEvent("login", nil, username, "fail",
			intPtr(exception.CodeAuthUnauthorized), ip, userAgent, "用户名或密码错误")
		return nil, exception.New(exception.CodeAuthUnauthorized, "用户名或密码错误", 401, nil)
	}
	if user.Status != 1 {
		logAuthEvent("login", &user.ID, user.Username, "fail",
			intPtr(exception.CodeAuthForbidden), ip, userAgent, "用户已被禁用")
		return nil, exception.New(exception.CodeAuthForbidden, "用户已被禁用", 403, nil)
	}
	// 仅更新最后登录字段，避免 Save 误保存关联
	now := utcNowNaive()
	updates := map[string]interface{}{
		"last_login_at": now,
		"last_login_ip": strPtrOrNil(ip),
	}
	if err := s.db.Model(&models.User{}).Where("id = ?", user.ID).
		Updates(updates).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "更新登录信息失败", 500, nil)
	}
	// 重新加载以获取完整 RBAC
	user, err = s.loadUserWithRBAC(user.ID)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "加载用户信息失败", 500, nil)
	}
	logAuthEvent("login", &user.ID, user.Username, "success", nil,
		ip, userAgent, "登录成功")
	return serializeUser(user), nil
}
// CreateTokensForUser 为用户签发 access_token 与 refresh_token，并持久化 Refresh Token。
//
// v1.1（P2-04）：空字符串 IP / User-Agent 转 NULL。
func (s *Service) CreateTokensForUser(userInfo *UserInfo, ip, userAgent string) (*TokenResp, error) {
	accessToken, err := security.GenerateToken(
		s.cfg.SecretKey,
		fmt.Sprintf("%d", userInfo.ID),
		security.JWTClaims{
			Username:    userInfo.Username,
			Roles:       userInfo.Roles,
			Permissions: userInfo.Permissions,
			Type:        "access",
		},
		time.Now().Add(time.Duration(AccessTokenExpireMinutes)*time.Minute),
	)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签发 Access Token 失败", 500, nil)
	}
	refreshToken, err := security.GenerateToken(
		s.cfg.SecretKey,
		fmt.Sprintf("%d", userInfo.ID),
		security.JWTClaims{
			Type: "refresh",
		},
		time.Now().Add(time.Duration(RefreshTokenExpireDays)*24*time.Hour),
	)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签发 Refresh Token 失败", 500, nil)
	}
	record := models.RefreshToken{
		UserID:    userInfo.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(time.Duration(RefreshTokenExpireDays) * 24 * time.Hour),
		IP:        strPtrOrNil(ip),
		UserAgent: strPtrOrNil(truncate(userAgent, 255)),
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "持久化 Refresh Token 失败", 500, nil)
	}
	return &TokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    AccessTokenExpireMinutes,
		User:         userInfo,
	}, nil
}
// RefreshAccessToken 使用 refresh_token 换取新的 access_token。
//
// v1.1（P0-02）：撤销旧 Token 与创建新 Token 使用同一事务，保证原子性；
//   冲突时通过 errRefreshConflict 哨兵错误返回 10001。
// v1.2（建议-5）：成功日志复用 record.IP / record.UserAgent，保持审计完整。
func (s *Service) RefreshAccessToken(refreshToken string) (*RefreshResp, error) {
	// ---- 1. 解码（只读，事务外） ----
	claims, err := security.ParseToken(s.cfg.SecretKey, refreshToken)
	if err != nil {
		logAuthEvent("refresh", nil, "", "fail",
			intPtr(exception.CodeAuthUnauthorized), "", "", "Token 无法解析")
		return nil, exception.New(exception.CodeAuthUnauthorized, "Token无效或已过期", 401, nil)
	}
	if claims.Type != "refresh" {
		logAuthEvent("refresh", nil, "", "fail",
			intPtr(exception.CodeAuthUnauthorized), "", "", "Token 类型错误")
		return nil, exception.New(exception.CodeAuthUnauthorized, "无效的refresh token", 401, nil)
	}
	var uid uint
	if _, err := fmt.Sscanf(claims.Subject, "%d", &uid); err != nil {
		return nil, exception.New(exception.CodeAuthUnauthorized, "无效的refresh token", 401, nil)
	}
	tokenHash := hashToken(refreshToken)
	// ---- 2. 只读预检查（事务外，减少事务持有时间） ----
	var record models.RefreshToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&record).Error; err != nil {
		logAuthEvent("refresh", nil, "", "fail",
			intPtr(exception.CodeAuthUnauthorized), "", "", "Refresh Token 已失效")
		return nil, exception.New(exception.CodeAuthUnauthorized, "Refresh token 已失效", 401, nil)
	}
	if record.Revoked == 1 {
		logAuthEvent("refresh", nil, "", "fail",
			intPtr(exception.CodeAuthUnauthorized), ptrToString(record.IP), ptrToString(record.UserAgent), "Refresh Token 已撤销")
		return nil, exception.New(exception.CodeAuthUnauthorized, "Refresh token 已失效", 401, nil)
	}
	if record.UserID != uid {
		logAuthEvent("refresh", &record.UserID, "", "fail",
			intPtr(exception.CodeAuthUnauthorized), ptrToString(record.IP), ptrToString(record.UserAgent), "Refresh Token 与用户不匹配")
		return nil, exception.New(exception.CodeAuthUnauthorized, "无效的refresh token", 401, nil)
	}
	if record.ExpiresAt.Before(utcNowNaive()) {
		logAuthEvent("refresh", &record.UserID, "", "fail",
			intPtr(exception.CodeAuthUnauthorized), ptrToString(record.IP), ptrToString(record.UserAgent), "Refresh Token 已过期")
		return nil, exception.New(exception.CodeAuthUnauthorized, "Refresh token 已过期", 401, nil)
	}
	// ---- 3. 加载用户（事务外） ----
	user, err := s.loadUserWithRBAC(uid)
	if err != nil {
		return nil, exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
	}
	if user.Status != 1 {
		return nil, exception.New(exception.CodeAuthForbidden, "用户已被禁用", 403, nil)
	}
	userInfo := serializeUser(user)
	// ---- 4. 预生成新 Token（纯计算，事务外） ----
	accessToken, err := security.GenerateToken(
		s.cfg.SecretKey,
		fmt.Sprintf("%d", userInfo.ID),
		security.JWTClaims{
			Username:    userInfo.Username,
			Roles:       userInfo.Roles,
			Permissions: userInfo.Permissions,
			Type:        "access",
		},
		time.Now().Add(time.Duration(AccessTokenExpireMinutes)*time.Minute),
	)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签发 Access Token 失败", 500, nil)
	}
	newRefreshToken, err := security.GenerateToken(
		s.cfg.SecretKey,
		fmt.Sprintf("%d", userInfo.ID),
		security.JWTClaims{
			Type: "refresh",
		},
		time.Now().Add(time.Duration(RefreshTokenExpireDays)*24*time.Hour),
	)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError, "签发 Refresh Token 失败", 500, nil)
	}
	// ---- 5. 事务：撤销旧 Token + 创建新 Token（原子） ----
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		now := utcNowNaive()
		result := tx.Model(&models.RefreshToken{}).
			Where("token_hash = ? AND revoked = 0", tokenHash).
			Updates(map[string]interface{}{
				"revoked":    1,
				"revoked_at": now,
			})
		if result.Error != nil {
			return exception.New(exception.CodeInternalError, "撤销 Refresh Token 失败", 500, nil)
		}
		if result.RowsAffected == 0 {
			return errRefreshConflict
		}
		newRecord := models.RefreshToken{
			UserID:    userInfo.ID,
			TokenHash: hashToken(newRefreshToken),
			ExpiresAt: now.Add(time.Duration(RefreshTokenExpireDays) * 24 * time.Hour),
			IP:        record.IP,
			UserAgent: record.UserAgent,
		}
		if err := tx.Create(&newRecord).Error; err != nil {
			return exception.New(exception.CodeInternalError, "持久化 Refresh Token 失败", 500, nil)
		}
		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, errRefreshConflict) {
			logAuthEvent("refresh", &uid, userInfo.Username, "fail",
				intPtr(exception.CodeAuthUnauthorized), ptrToString(record.IP), ptrToString(record.UserAgent),
				"Refresh Token 已被并发撤销")
			return nil, exception.New(exception.CodeAuthUnauthorized, "Refresh token 已失效", 401, nil)
		}
		return nil, txErr
	}
	// v1.2（建议-5）：成功日志复用 record 的 IP/UA
	logAuthEvent("refresh", &userInfo.ID, userInfo.Username, "success", nil,
		ptrToString(record.IP), ptrToString(record.UserAgent), "")
	return &RefreshResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    AccessTokenExpireMinutes,
	}, nil
}
// RevokeRefreshToken 撤销指定的 Refresh Token（幂等）。
func (s *Service) RevokeRefreshToken(refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	now := utcNowNaive()
	return s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked = 0", tokenHash).
		Updates(map[string]interface{}{
			"revoked":    1,
			"revoked_at": now,
		}).Error
}
// RevokeAllUserTokens 撤销某用户的全部 Refresh Token，返回撤销数量。
func (s *Service) RevokeAllUserTokens(userID uint) (int64, error) {
	now := utcNowNaive()
	result := s.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = 0", userID).
		Updates(map[string]interface{}{
			"revoked":    1,
			"revoked_at": now,
		})
	return result.RowsAffected, result.Error
}
// GetUserInfo 获取用户完整信息。
func (s *Service) GetUserInfo(userID uint) (*UserInfo, error) {
	user, err := s.loadUserWithRBAC(userID)
	if err != nil {
		return nil, exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
	}
	return serializeUser(user), nil
}
// ChangePassword 修改当前用户密码。成功后批量撤销该用户全部 Refresh Token。
//
// v1.1：更新密码与撤销 Refresh Token 在同一事务中完成，保证原子性。
// v1.2（P2-NEW-02）：签名扩展为接收 ip, userAgent，供审计日志使用；
//   所有 logAuthEvent 调用点传真实 IP/UA。
func (s *Service) ChangePassword(
	userID uint,
	oldPassword, newPassword, confirmPassword string,
	ip, userAgent string,
) error {
	if newPassword != confirmPassword {
		logAuthEvent("change_password", &userID, "", "fail",
			intPtr(exception.CodeAuthPasswordDiff), ip, userAgent, "两次密码不一致")
		return exception.New(exception.CodeAuthPasswordDiff, "两次密码不一致", 400, nil)
	}
	if len(newPassword) < 6 || len(newPassword) > 20 {
		logAuthEvent("change_password", &userID, "", "fail",
			intPtr(exception.CodeAuthPasswordFormat), ip, userAgent, "新密码长度不符合要求")
		return exception.New(exception.CodeAuthPasswordFormat, "密码格式无效", 400, nil)
	}
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return exception.New(exception.CodeAuthUserNotFound, "用户不存在", 404, nil)
	}
	if !security.VerifyPassword(oldPassword, user.PasswordHash) {
		logAuthEvent("change_password", &userID, user.Username, "fail",
			intPtr(exception.CodeAuthOldPasswordErr), ip, userAgent, "原密码错误")
		return exception.New(exception.CodeAuthOldPasswordErr, "原密码错误", 400, nil)
	}
	hashed, err := security.HashPassword(newPassword)
	if err != nil {
		return exception.New(exception.CodeInternalError, "密码哈希失败", 500, nil)
	}
	// 事务：更新密码 + 撤销全部 Refresh Token
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).
			Update("password_hash", hashed).Error; err != nil {
			return exception.New(exception.CodeInternalError, "更新密码失败", 500, nil)
		}
		now := utcNowNaive()
		if err := tx.Model(&models.RefreshToken{}).
			Where("user_id = ? AND revoked = 0", userID).
			Updates(map[string]interface{}{
				"revoked":    1,
				"revoked_at": now,
			}).Error; err != nil {
			return exception.New(exception.CodeInternalError, "撤销 Refresh Token 失败", 500, nil)
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	logAuthEvent("change_password", &userID, user.Username, "success", nil,
		ip, userAgent, "")
	return nil
}
// intPtr 返回 int 指针。
func intPtr(v int) *int {
	return &v
}
// IsNotFound 判断错误是否为 gorm.ErrRecordNotFound。
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
