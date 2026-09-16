package auth
import (
	"path/filepath"
	"sync"
	"testing"
	"time"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"backend-go/internal/config"
	"backend-go/internal/exception"
	"backend-go/internal/models"
	"backend-go/internal/security"
)
// setupTestDB 创建临时文件 SQLite 数据库并自动迁移。
//
// v1.2（P2-NEW-03）：使用 t.TempDir() 下的文件库替代内存库，
//   并设置 busy_timeout(5000) + WAL 模式，保证并发写场景稳定。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test.db")
	dsn := tmpFile + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("migrate test db failed: %v", err)
	}
	return db
}
// setupTestService 创建测试用 Service 并初始化种子数据。
func setupTestService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	if err := EnsureAuthSeedData(db); err != nil {
		t.Fatalf("seed data failed: %v", err)
	}
	cfg := &config.Config{
		SecretKey: "test-secret-key-for-unit-tests",
	}
	return NewService(db, cfg), db
}
// ---------------------------------------------------------------------------
// 安全模块基础测试
// ---------------------------------------------------------------------------
func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := security.HashPassword("123456")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if !security.VerifyPassword("123456", hash) {
		t.Fatal("VerifyPassword should return true for correct password")
	}
	if security.VerifyPassword("wrong", hash) {
		t.Fatal("VerifyPassword should return false for wrong password")
	}
}
func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-key"
	claims := security.JWTClaims{
		Username:    "admin",
		Roles:       []string{"admin"},
		Permissions: []string{"auth:user:view"},
		Type:        "access",
	}
	token, err := security.GenerateToken(secret, "1", claims, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	parsed, err := security.ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if parsed.Subject != "1" {
		t.Fatalf("expected subject=1, got %s", parsed.Subject)
	}
	if parsed.Username != "admin" {
		t.Fatalf("expected username=admin, got %s", parsed.Username)
	}
	if parsed.Type != "access" {
		t.Fatalf("expected type=access, got %s", parsed.Type)
	}
}
func TestParseTokenWithWrongSecret(t *testing.T) {
	claims := security.JWTClaims{Type: "access"}
	token, _ := security.GenerateToken("secret-a", "1", claims, time.Now().Add(time.Hour))
	_, err := security.ParseToken("secret-b", token)
	if err == nil {
		t.Fatal("ParseToken should fail with wrong secret")
	}
}
// ---------------------------------------------------------------------------
// Service 层集成测试
// ---------------------------------------------------------------------------
func TestAuthenticateUser(t *testing.T) {
	svc, _ := setupTestService(t)
	userInfo, err := svc.AuthenticateUser("admin", "123456", "127.0.0.1", "pytest")
	if err != nil {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}
	if userInfo.Username != "admin" {
		t.Fatalf("expected username=admin, got %s", userInfo.Username)
	}
	// 密码错误
	_, err = svc.AuthenticateUser("admin", "wrong", "", "")
	if err == nil {
		t.Fatal("AuthenticateUser should fail with wrong password")
	}
	pe, ok := err.(*exception.PlatformError)
	if !ok || pe.Code != exception.CodeAuthUnauthorized {
		t.Fatalf("expected code=%d, got %v", exception.CodeAuthUnauthorized, err)
	}
	// 用户不存在
	_, err = svc.AuthenticateUser("nouser", "123456", "", "")
	if err == nil {
		t.Fatal("AuthenticateUser should fail with unknown user")
	}
}
func TestCreateTokensForUser(t *testing.T) {
	svc, db := setupTestService(t)
	userInfo, err := svc.AuthenticateUser("admin", "123456", "", "")
	if err != nil {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}
	tokens, err := svc.CreateTokensForUser(userInfo, "127.0.0.1", "pytest")
	if err != nil {
		t.Fatalf("CreateTokensForUser failed: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("tokens should not be empty")
	}
	// 检查 Refresh Token 已写入
	var count int64
	db.Model(&models.RefreshToken{}).Where("user_id = ?", userInfo.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 refresh token record, got %d", count)
	}
}
func TestRefreshAccessToken(t *testing.T) {
	svc, _ := setupTestService(t)
	userInfo, err := svc.AuthenticateUser("admin", "123456", "", "")
	if err != nil {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}
	tokens, err := svc.CreateTokensForUser(userInfo, "", "")
	if err != nil {
		t.Fatalf("CreateTokensForUser failed: %v", err)
	}
	oldRefresh := tokens.RefreshToken
	// 刷新：应返回新的 refresh_token
	resp, err := svc.RefreshAccessToken(oldRefresh)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}
	if resp.RefreshToken == oldRefresh {
		t.Fatal("refresh token should be rotated")
	}
	// 旧 token 复用应失败
	_, err = svc.RefreshAccessToken(oldRefresh)
	if err == nil {
		t.Fatal("old refresh token should be invalid after rotation")
	}
	pe, ok := err.(*exception.PlatformError)
	if !ok || pe.Code != exception.CodeAuthUnauthorized {
		t.Fatalf("expected code=%d, got %v", exception.CodeAuthUnauthorized, err)
	}
	// 新 token 可继续刷新
	resp2, err := svc.RefreshAccessToken(resp.RefreshToken)
	if err != nil {
		t.Fatalf("new refresh token should be valid: %v", err)
	}
	if resp2.AccessToken == "" {
		t.Fatal("access token should not be empty")
	}
}
// TestRefreshAccessTokenConcurrent 验证并发刷新场景。
//
// v1.2（P2-NEW-03）：使用 sync.WaitGroup + go 真正并发调用；
//   busy_timeout + WAL 保证 SQLite 文件库并发稳定性；
//   断言"恰好 1 个成功，1 个失败"。
func TestRefreshAccessTokenConcurrent(t *testing.T) {
	svc, _ := setupTestService(t)
	userInfo, err := svc.AuthenticateUser("admin", "123456", "", "")
	if err != nil {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}
	tokens, err := svc.CreateTokensForUser(userInfo, "", "")
	if err != nil {
		t.Fatalf("CreateTokensForUser failed: %v", err)
	}
	const concurrency = 2
	var wg sync.WaitGroup
	results := make([]error, concurrency)
	start := make(chan struct{})
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start // 所有 goroutine 同时起跑，最大化并发度
			_, results[idx] = svc.RefreshAccessToken(tokens.RefreshToken)
		}(i)
	}
	close(start)
	wg.Wait()
	successCount := 0
	failCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 success, got %d (fail=%d)", successCount, failCount)
	}
	if failCount != 1 {
		t.Fatalf("expected exactly 1 failure, got %d (success=%d)", failCount, successCount)
	}
}
func TestChangePassword(t *testing.T) {
	svc, db := setupTestService(t)
	userInfo, _ := svc.AuthenticateUser("admin", "123456", "", "")
	// 两次密码不一致
	err := svc.ChangePassword(userInfo.ID, "123456", "654321", "abcdef", "127.0.0.1", "pytest")
	if err == nil {
		t.Fatal("mismatched passwords should fail")
	}
	// 原密码错误
	err = svc.ChangePassword(userInfo.ID, "wrong", "654321", "654321", "127.0.0.1", "pytest")
	if err == nil {
		t.Fatal("wrong old password should fail")
	}
	// 密码过短
	err = svc.ChangePassword(userInfo.ID, "123456", "x", "x", "127.0.0.1", "pytest")
	if err == nil {
		t.Fatal("short password should fail")
	}
	// 成功
	err = svc.ChangePassword(userInfo.ID, "123456", "abcdef", "abcdef", "127.0.0.1", "pytest")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}
	// 用新密码登录
	_, err = svc.AuthenticateUser("admin", "abcdef", "", "")
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	// 检查所有 Refresh Token 已撤销
	var count int64
	db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = 0", userInfo.ID).
		Count(&count)
	if count != 0 {
		t.Fatalf("expected all refresh tokens revoked, got %d active", count)
	}
}
func TestSeedDataIdempotent(t *testing.T) {
	db := setupTestDB(t)
	// 多次执行种子数据
	for i := 0; i < 3; i++ {
		if err := EnsureAuthSeedData(db); err != nil {
			t.Fatalf("seed data failed on iteration %d: %v", i, err)
		}
	}
	// 检查用户数
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount != 2 {
		t.Fatalf("expected 2 users, got %d", userCount)
	}
	// 检查角色数
	var roleCount int64
	db.Model(&models.Role{}).Count(&roleCount)
	if roleCount != 2 {
		t.Fatalf("expected 2 roles, got %d", roleCount)
	}
	// 检查用户-角色关联数
	var urCount int64
	db.Model(&models.UserRole{}).Count(&urCount)
	if urCount != 2 {
		t.Fatalf("expected 2 user-role links, got %d", urCount)
	}
}
func TestRevokeRefreshTokenIdempotent(t *testing.T) {
	svc, _ := setupTestService(t)
	userInfo, _ := svc.AuthenticateUser("admin", "123456", "", "")
	tokens, _ := svc.CreateTokensForUser(userInfo, "", "")
	// 第一次撤销
	if err := svc.RevokeRefreshToken(tokens.RefreshToken); err != nil {
		t.Fatalf("first revoke failed: %v", err)
	}
	// 第二次撤销应幂等
	if err := svc.RevokeRefreshToken(tokens.RefreshToken); err != nil {
		t.Fatalf("second revoke should be idempotent, got: %v", err)
	}
	// 撤销后刷新应失败
	_, err := svc.RefreshAccessToken(tokens.RefreshToken)
	if err == nil {
		t.Fatal("revoked token should not refresh")
	}
}
