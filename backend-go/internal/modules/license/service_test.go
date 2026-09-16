package license
import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"backend-go/internal/config"
	"backend-go/internal/models"
)
// setupTestDB 创建测试用 SQLite 文件库。
//
// 使用文件库而非内存库，便于部分唯一索引在多次事务间生效。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test_license.db")
	db, err := gorm.Open(sqlite.Open(tmpFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.License{},
		&models.User{},
		&models.Module{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	// 部分唯一索引（模拟 goose 迁移创建）
	if err := db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS uq_license_license_active " +
			"ON license_license (is_active) WHERE is_active = 1",
	).Error; err != nil {
		t.Fatalf("create partial unique index failed: %v", err)
	}
	return db
}
// ---------------------------------------------------------------------------
// 签名与 canonical_json
// ---------------------------------------------------------------------------
func TestCanonicalJSONNoHTMLEscape(t *testing.T) {
	payload := map[string]interface{}{
		"b": "a<b>c&d",
		"a": 1,
	}
	got, err := CanonicalJSON(payload)
	if err != nil {
		t.Fatalf("CanonicalJSON failed: %v", err)
	}
	// 期望：紧凑格式（无空格），key 按字母序 a、b，HTML 字符不转义
	want := `{"a":1,"b":"a<b>c&d"}`
	if got != want {
		t.Fatalf("CanonicalJSON mismatch:\n got=%s\nwant=%s", got, want)
	}
}
func TestSignatureRoundTrip(t *testing.T) {
	secret := "test-secret"
	payload := map[string]interface{}{
		"license_key":  "LIC-TEST-001",
		"license_type": "enterprise",
		"issued_at":    "2026-09-01T00:00:00",
		"expires_at":   "2027-09-01T00:00:00",
	}
	sig, err := ComputeSignature(payload, secret)
	if err != nil {
		t.Fatalf("ComputeSignature failed: %v", err)
	}
	ok, err := VerifySignature(payload, sig, secret)
	if err != nil {
		t.Fatalf("VerifySignature failed: %v", err)
	}
	if !ok {
		t.Fatal("signature should verify")
	}
	// 篡改签名
	ok, _ = VerifySignature(payload, "0"+sig[1:], secret)
	if ok {
		t.Fatal("tampered signature should fail")
	}
}
// ---------------------------------------------------------------------------
// 机器码
// ---------------------------------------------------------------------------
func TestMachineCodeOverride(t *testing.T) {
	// 保存/恢复 config.C，避免污染其他测试
	savedC := config.C
	config.C = &config.Config{LicenseMachineCodeOverride: "TEST-MACHINE-CODE"}
	defer func() { config.C = savedC }()
	if got := GetMachineCode(); got != "TEST-MACHINE-CODE" {
		t.Fatalf("expected override, got %s", got)
	}
}
// TestMachineCodeDefault 验证默认机器码生成。
//
// v1.1（Go-P2-02）：显式将 config.C 置 nil，避免前序测试设置的
// LicenseMachineCodeOverride 影响本测试；使用 defer 恢复。
func TestMachineCodeDefault(t *testing.T) {
	savedC := config.C
	config.C = nil
	defer func() { config.C = savedC }()
	code := GetMachineCode()
	if len(code) != 32 {
		t.Fatalf("machine code should be 32 chars, got %d: %s", len(code), code)
	}
	if code != strings.ToUpper(code) {
		t.Fatalf("machine code should be uppercase, got %s", code)
	}
}
// ---------------------------------------------------------------------------
// 状态查询
// ---------------------------------------------------------------------------
func TestGetLicenseStatusNoLicense(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	_, err := svc.GetLicenseStatus()
	if err == nil {
		t.Fatal("expected error when no license")
	}
	if !strings.Contains(err.Error(), "License不存在") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestGetLicenseStatusSuccess(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	lic := models.License{
		LicenseKey:        "LIC-STATUS-001",
		LicenseType:       "enterprise",
		MaxUsers:          intPtr(100),
		AuthorizedModules: models.StringArray{"auth", "customer_relation"},
		IssuedAt:          now,
		ExpiresAt:         expires,
		IsActive:          1,
	}
	if err := db.Create(&lic).Error; err != nil {
		t.Fatalf("create license failed: %v", err)
	}
	status, err := svc.GetLicenseStatus()
	if err != nil {
		t.Fatalf("GetLicenseStatus failed: %v", err)
	}
	if !status.IsValid {
		t.Fatal("license should be valid")
	}
	if status.IsExpired {
		t.Fatal("license should not be expired")
	}
	if status.IsExpiringSoon {
		t.Fatal("license should not be expiring soon")
	}
	if status.LicenseType == nil || *status.LicenseType != "enterprise" {
		t.Fatalf("license_type mismatch: %v", status.LicenseType)
	}
	if len(status.AuthorizedModules) != 2 {
		t.Fatalf("authorized_modules mismatch: %v", status.AuthorizedModules)
	}
}
// TestAuthorizedModulesNilAsEmptyArray 验证 Go-P1-01 修复。
//
// DB 中 authorized_modules = NULL 时：
//   - models.StringArray 为 nil；
//   - 归一化后 AuthorizedModules 应为 []string{}（非 nil）；
//   - JSON 序列化后应输出 []，而非 null。
func TestAuthorizedModulesNilAsEmptyArray(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	lic := models.License{
		LicenseKey:        "LIC-NULL-MODULES",
		LicenseType:       "enterprise",
		AuthorizedModules: nil, // 显式 nil
		IssuedAt:          now,
		ExpiresAt:         now.Add(365 * 24 * time.Hour),
		IsActive:          1,
	}
	if err := db.Create(&lic).Error; err != nil {
		t.Fatalf("create license failed: %v", err)
	}
	status, err := svc.GetLicenseStatus()
	if err != nil {
		t.Fatalf("GetLicenseStatus failed: %v", err)
	}
	// 断言 1：AuthorizedModules 非 nil
	if status.AuthorizedModules == nil {
		t.Fatal("authorized_modules should be []string{}, not nil")
	}
	// 断言 2：长度为 0
	if len(status.AuthorizedModules) != 0 {
		t.Fatalf("expected empty slice, got %v", status.AuthorizedModules)
	}
	// 断言 3：JSON 序列化为 []
	body, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if !strings.Contains(string(body), `"authorized_modules":[]`) {
		t.Fatalf("expected JSON `[]`, got: %s", body)
	}
	if strings.Contains(string(body), `"authorized_modules":null`) {
		t.Fatalf("JSON should not contain null: %s", body)
	}
}
// ---------------------------------------------------------------------------
// 导入
// ---------------------------------------------------------------------------
func TestImportLicenseSuccess(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{LicenseSecretKey: "test-secret"}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	payload := map[string]interface{}{
		"license_key":        "LIC-IMPORT-001",
		"license_type":       "enterprise",
		"max_users":          float64(100),
		"authorized_modules": []interface{}{"auth", "customer_relation"},
		"machine_code":       nil,
		"issued_at":          now.Format("2006-01-02T15:04:05"),
		"expires_at":         expires.Format("2006-01-02T15:04:05"),
	}
	sig, _ := ComputeSignature(payload, cfg.LicenseSecretKey)
	content, _ := json.Marshal(map[string]interface{}{
		"payload":   payload,
		"signature": sig,
	})
	lic, err := svc.ImportLicense(content, "", 1, "admin")
	if err != nil {
		t.Fatalf("ImportLicense failed: %v", err)
	}
	if lic.LicenseKey != "LIC-IMPORT-001" {
		t.Fatalf("license_key mismatch: %s", lic.LicenseKey)
	}
	if lic.IsActive != 1 {
		t.Fatal("license should be active")
	}
	if lic.MaxUsers == nil || *lic.MaxUsers != 100 {
		t.Fatalf("max_users mismatch: %v", lic.MaxUsers)
	}
}
func TestImportLicenseExpired(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{LicenseSecretKey: "test-secret"}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	issued := now.Add(-365 * 24 * time.Hour)
	expires := now.Add(-24 * time.Hour)
	payload := map[string]interface{}{
		"license_key":  "LIC-EXPIRED",
		"license_type": "enterprise",
		"issued_at":    issued.Format("2006-01-02T15:04:05"),
		"expires_at":   expires.Format("2006-01-02T15:04:05"),
	}
	sig, _ := ComputeSignature(payload, cfg.LicenseSecretKey)
	content, _ := json.Marshal(map[string]interface{}{
		"payload":   payload,
		"signature": sig,
	})
	_, err := svc.ImportLicense(content, "", 1, "admin")
	if err == nil {
		t.Fatal("expected error for expired license")
	}
	if !strings.Contains(err.Error(), "已过期") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestImportLicenseSignatureFailed(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{LicenseSecretKey: "test-secret"}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	payload := map[string]interface{}{
		"license_key":  "LIC-BAD-SIG",
		"license_type": "enterprise",
		"issued_at":    now.Format("2006-01-02T15:04:05"),
		"expires_at":   expires.Format("2006-01-02T15:04:05"),
	}
	content, _ := json.Marshal(map[string]interface{}{
		"payload":   payload,
		"signature": strings.Repeat("0", 64),
	})
	_, err := svc.ImportLicense(content, "", 1, "admin")
	if err == nil {
		t.Fatal("expected signature failure")
	}
	if !strings.Contains(err.Error(), "签名验证失败") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestImportLicenseDuplicate(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{LicenseSecretKey: "test-secret"}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	payload := map[string]interface{}{
		"license_key":  "LIC-DUP-001",
		"license_type": "enterprise",
		"issued_at":    now.Format("2006-01-02T15:04:05"),
		"expires_at":   expires.Format("2006-01-02T15:04:05"),
	}
	sig, _ := ComputeSignature(payload, cfg.LicenseSecretKey)
	content, _ := json.Marshal(map[string]interface{}{
		"payload":   payload,
		"signature": sig,
	})
	_, err := svc.ImportLicense(content, "", 1, "admin")
	if err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	_, err = svc.ImportLicense(content, "", 1, "admin")
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "License 已激活") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestImportLicenseMachineCodeMismatch(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{LicenseSecretKey: "test-secret"}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	payload := map[string]interface{}{
		"license_key":  "LIC-MC-MISMATCH",
		"license_type": "enterprise",
		"machine_code": "OTHER-MACHINE-CODE",
		"issued_at":    now.Format("2006-01-02T15:04:05"),
		"expires_at":   expires.Format("2006-01-02T15:04:05"),
	}
	sig, _ := ComputeSignature(payload, cfg.LicenseSecretKey)
	content, _ := json.Marshal(map[string]interface{}{
		"payload":   payload,
		"signature": sig,
	})
	_, err := svc.ImportLicense(content, "", 1, "admin")
	if err == nil {
		t.Fatal("expected machine code mismatch")
	}
	if !strings.Contains(err.Error(), "机器码不匹配") {
		t.Fatalf("unexpected error: %v", err)
	}
}
// ---------------------------------------------------------------------------
// 在线激活
// ---------------------------------------------------------------------------
// TestActivateLicenseNoService 验证未配置 LICENSE_ACTIVATION_URL 时返回 50005。
//
// v1.1（Go-P3-05）新增。
func TestActivateLicenseNoService(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		LicenseSecretKey:     "test-secret",
		LicenseActivationURL: "",
	}
	svc := NewService(db, cfg)
	_, err := svc.ActivateLicense("ABCD-1234", GetMachineCode(), 1, "admin")
	if err == nil {
		t.Fatal("expected 50005 error")
	}
	if !strings.Contains(err.Error(), "未配置在线激活服务") {
		t.Fatalf("unexpected error: %v", err)
	}
}
// ---------------------------------------------------------------------------
// 部分唯一索引（R-3）
// ---------------------------------------------------------------------------
func TestPartialUniqueIndexPreventsMultipleActive(t *testing.T) {
	db := setupTestDB(t)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	// 第一条 active
	lic1 := models.License{
		LicenseKey: "LIC-IDX-A",
		LicenseType: "enterprise",
		IssuedAt:    now,
		ExpiresAt:   expires,
		IsActive:    1,
	}
	if err := db.Create(&lic1).Error; err != nil {
		t.Fatalf("create first license failed: %v", err)
	}
	// 第二条 active 应失败
	lic2 := models.License{
		LicenseKey: "LIC-IDX-B",
		LicenseType: "enterprise",
		IssuedAt:    now,
		ExpiresAt:   expires,
		IsActive:    1,
	}
	err := db.Create(&lic2).Error
	if err == nil {
		t.Fatal("expected partial unique index conflict")
	}
	if !isUniqueConstraintError(err) {
		t.Fatalf("expected unique constraint error, got: %v", err)
	}
}
// ---------------------------------------------------------------------------
// 模块授权 / 用户配额
// ---------------------------------------------------------------------------
func TestCheckModuleAuthorizedCoreBypass(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	for core := range CoreModulesBypass {
		if err := svc.CheckModuleAuthorized(core); err != nil {
			t.Fatalf("core module %s should bypass, got: %v", core, err)
		}
	}
}
func TestCheckModuleAuthorizedNoLicense(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	err := svc.CheckModuleAuthorized("customer_relation")
	if err == nil {
		t.Fatal("expected 50009 error")
	}
	if !strings.Contains(err.Error(), "License不存在") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestCheckModuleAuthorizedDenied(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	lic := models.License{
		LicenseKey:        "LIC-AUTH-ONLY",
		LicenseType:       "enterprise",
		AuthorizedModules: models.StringArray{"auth"},
		IssuedAt:          now,
		ExpiresAt:         expires,
		IsActive:          1,
	}
	if err := db.Create(&lic).Error; err != nil {
		t.Fatalf("create license failed: %v", err)
	}
	err := svc.CheckModuleAuthorized("customer_relation")
	if err == nil {
		t.Fatal("expected 50007 error")
	}
	if !strings.Contains(err.Error(), "模块未授权") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestCheckUserQuotaExceeded(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	// max_users=1，当前用户数=1
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	lic := models.License{
		LicenseKey:  "LIC-QUOTA",
		LicenseType: "enterprise",
		MaxUsers:    intPtr(1),
		IssuedAt:    now,
		ExpiresAt:   expires,
		IsActive:    1,
	}
	if err := db.Create(&lic).Error; err != nil {
		t.Fatalf("create license failed: %v", err)
	}
	if err := db.Create(&models.User{
		Username: "u1", PasswordHash: "x", Nickname: "u1", Status: 1,
	}).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	err := svc.CheckUserQuota()
	if err == nil {
		t.Fatal("expected 50008 error")
	}
	if !strings.Contains(err.Error(), "用户数超限") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestCheckUserQuotaOK(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{}
	svc := NewService(db, cfg)
	now := time.Now().UTC()
	expires := now.Add(365 * 24 * time.Hour)
	lic := models.License{
		LicenseKey:  "LIC-QUOTA-OK",
		LicenseType: "enterprise",
		MaxUsers:    intPtr(100),
		IssuedAt:    now,
		ExpiresAt:   expires,
		IsActive:    1,
	}
	if err := db.Create(&lic).Error; err != nil {
		t.Fatalf("create license failed: %v", err)
	}
	if err := svc.CheckUserQuota(); err != nil {
		t.Fatalf("CheckUserQuota should pass: %v", err)
	}
}
// ---------------------------------------------------------------------------
// 辅助
// ---------------------------------------------------------------------------
func intPtr(v int) *int { return &v }
