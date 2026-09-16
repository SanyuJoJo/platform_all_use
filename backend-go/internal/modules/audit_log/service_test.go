package audit_log
import (
	"fmt"
	"strings"
	"testing"
	"time"
	"backend-go/internal/models"
)
// TestParsePath 验证路径解析。
func TestParsePath(t *testing.T) {
	cases := []struct {
		path         string
		wantModule   string
		wantResource string
		wantResID    string
	}{
		{"/api/v1/auth/users", "auth", "users", ""},
		{"/api/v1/auth/users/123", "auth", "users", "123"},
		{"/api/v1/auth/login", "auth", "login", ""},
		{"/api/v1/modules", "module_manager", "modules", ""},
		{"/api/v1/audit-logs", "audit_log", "audit-logs", ""},
		{"/api/v1/audit-logs/export", "audit_log", "export", ""},
		{"/api/v1/audit-logs/1", "audit_log", "1", ""},
	}
	for _, c := range cases {
		moduleID, resource, resourceID := parsePath(c.path)
		if moduleID != c.wantModule {
			t.Errorf("parsePath(%q) moduleID = %q, want %q", c.path, moduleID, c.wantModule)
		}
		gotResource := ""
		if resource != nil {
			gotResource = *resource
		}
		if gotResource != c.wantResource {
			t.Errorf("parsePath(%q) resource = %q, want %q", c.path, gotResource, c.wantResource)
		}
		gotResID := ""
		if resourceID != nil {
			gotResID = *resourceID
		}
		if gotResID != c.wantResID {
			t.Errorf("parsePath(%q) resourceID = %q, want %q", c.path, gotResID, c.wantResID)
		}
	}
}
// TestResolveAction 验证 HTTP 方法 → action 映射。
func TestResolveAction(t *testing.T) {
	cases := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/api/v1/auth/users", "view"},
		{"POST", "/api/v1/auth/users", "create"},
		{"PUT", "/api/v1/auth/users/1", "update"},
		{"PATCH", "/api/v1/auth/users/1/status", "update"},
		{"DELETE", "/api/v1/auth/users/1", "delete"},
		{"GET", "/api/v1/audit-logs/export", "export"},
	}
	for _, c := range cases {
		if got := resolveAction(c.method, c.path); got != c.want {
			t.Errorf("resolveAction(%q, %q) = %q, want %q", c.method, c.path, got, c.want)
		}
	}
}
// TestShouldSkip 验证白名单精确匹配。
//
// v1.2（P1-NEW-1）修正：
//   - `/health-check`、`/healthz` 不在 `/api/v1/*` 下，被第二个条件（非 /api/v1/*
//     路径一律跳过）跳过，期望值应为 `true`（v1.1 错误地期望 `false`）；
//   - 新增 `/api/v1/health`、`/api/v1/health-check` 用例，验证白名单前缀
//     不会在 `/api/v1/*` 下生效（设计如此：仅跳过完全匹配 `/health` 或
//     `/health/` 开头的路径）。
func TestShouldSkip(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		// 白名单路径（精确匹配）
		{"/health", true},
		{"/health/", true},
		{"/docs", true},
		{"/docs/", true},
		{"/redoc", true},
		{"/openapi.json", true},
		{"/favicon.ico", true},
		{"/static", true},
		{"/static/foo.js", true},
		// 白名单路径的"近亲"——不在 /api/v1/* 下，仍跳过
		// v1.2（P1-NEW-1）修正：期望值从 false 改为 true
		{"/health-check", true},
		{"/healthz", true},
		{"/docs-old", true},
		// /api/v1/* 路径——不受白名单前缀影响，不跳过
		{"/api/v1/health", false},
		{"/api/v1/health-check", false},
		{"/api/v1/auth/users", false},
		{"/api/v1/audit-logs", false},
		{"/api/v1/", false},
		{"/api/v1", false},
		// 完全无关路径——跳过
		{"/other", true},
		{"/", true},
	}
	for _, c := range cases {
		if got := shouldSkip(c.path); got != c.want {
			t.Errorf("shouldSkip(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
// TestToCSVEscapesSpecialChars 验证 CSV 数据行对逗号、引号、换行的转义。
func TestToCSVEscapesSpecialChars(t *testing.T) {
	marker := "test-marker-123"
	// 同时包含逗号、引号、换行，且包含唯一 marker
	detail := fmt.Sprintf(`含逗号%s, 引号"q" 和
换行`, marker)
	entry := models.AuditLogOperation{
		ID:        1,
		ModuleID:  "test_module",
		Action:    "create",
		Detail:    &detail,
		Status:    "success",
		CreatedAt: time.Now(),
	}
	content, err := toCSV([]models.AuditLogOperation{entry})
	if err != nil {
		t.Fatalf("toCSV failed: %v", err)
	}
	body := string(content)
	// 1. marker 应存在于导出内容中
	if !strings.Contains(body, marker) {
		t.Fatalf("CSV 未包含 marker %q: %s", marker, body)
	}
	// 2. 逗号字段应存在
	if !strings.Contains(body, "含逗号") {
		t.Fatalf("CSV 未包含逗号字段: %s", body)
	}
	// 3. 引号应被转义为两个双引号
	if !strings.Contains(body, `""q""`) {
		t.Fatalf("CSV 引号未正确转义: %s", body)
	}
	// 4. 不应出现三重引号（双重转义的典型特征）
	if strings.Contains(body, `"""`) {
		t.Fatalf("CSV 出现三重引号，存在双重转义: %s", body)
	}
}
// TestStrPtrOrNil 验证空字符串归一化。
func TestStrPtrOrNil(t *testing.T) {
	if p := strPtrOrNil(""); p != nil {
		t.Error("expected nil for empty string")
	}
	if p := strPtrOrNil("test"); p == nil || *p != "test" {
		t.Error("expected non-nil pointer with value 'test'")
	}
}
// TestParseTime 验证时间解析。
func TestParseTime(t *testing.T) {
	cases := []struct {
		input     string
		wantError bool
	}{
		{"", false},
		{"2026-09-01T00:00:00", false},
		{"2026-09-01T00:00:00Z", false},
		{"2026-09-01 00:00:00", false},
		{"invalid", true},
		{"2026-13-01", true},
	}
	for _, c := range cases {
		_, err := ParseTime(c.input)
		if c.wantError && err == nil {
			t.Errorf("ParseTime(%q) expected error, got nil", c.input)
		}
		if !c.wantError && err != nil {
			t.Errorf("ParseTime(%q) unexpected error: %v", c.input, err)
		}
	}
}
