package module_manager
import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"gorm.io/gorm"
	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/models"
)
// ---------------------------------------------------------------------------
// 纯函数测试
// ---------------------------------------------------------------------------
// TestCompareVersion 验证语义化版本比较。
func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a, b     string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.99.99", 1},
		{"1.0.0-beta", "1.0.0", -1},
		{"1.0.0", "1.0.0-beta", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-beta", 0},
		{"1.0.0+build1", "1.0.0", 0},
		{"1.0.0-alpha.10", "1.0.0-alpha.2", 1},
		{"1.0.0-alpha.2", "1.0.0-alpha.10", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.1.1", -1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-1", "1.0.0-alpha", -1},
	}
	for _, c := range cases {
		if got := compareVersion(c.a, c.b); got != c.expected {
			t.Errorf("compareVersion(%q, %q) = %d, expected %d",
				c.a, c.b, got, c.expected)
		}
	}
}
// TestIsCoreModule 验证核心模块判定。
func TestIsCoreModule(t *testing.T) {
	svc := &Service{}
	for _, id := range []string{"auth", "platform", "module_manager", "audit_log", "license"} {
		if !svc.isCoreModule(id) {
			t.Errorf("expected %s to be a core module", id)
		}
	}
	if svc.isCoreModule("customer_relation") {
		t.Error("expected customer_relation NOT to be a core module")
	}
}
// TestFilterMenusByPermission 验证菜单权限过滤。
func TestFilterMenusByPermission(t *testing.T) {
	parentID := "auth:dashboard"
	perm := "auth:user:view"
	childPerm := "auth:role:view"
	menus := []MenuItem{
		{ID: "auth:dashboard", ParentID: nil, Title: "认证授权", Permission: &perm},
		{ID: "auth:users", ParentID: &parentID, Title: "用户管理", Permission: &perm},
		{ID: "auth:roles", ParentID: &parentID, Title: "角色管理", Permission: &childPerm},
	}
	userPerms := map[string]struct{}{"auth:user:view": {}}
	result := filterMenusByPermission(menus, userPerms)
	if len(result) != 2 {
		t.Fatalf("expected 2 menus, got %d", len(result))
	}
	for _, m := range result {
		if m.ID == "auth:roles" {
			t.Error("auth:roles should be filtered out")
		}
	}
}
// TestValidateManifest 验证 manifest 校验。
func TestValidateManifest(t *testing.T) {
	m := &Manifest{
		ID:           "test_module",
		Name:         "Test",
		Version:      "1.0.0",
		Description:  "test",
		Dependencies: []string{"auth"},
		Permissions: []map[string]interface{}{
			{"code": "test_module:demo:view", "name": "查看", "resource": "demo", "action": "view"},
		},
		Menus:          []map[string]interface{}{},
		EntryBackend:   "router:router",
		DatabaseTables: []string{},
	}
	if err := validateManifest(m); err != nil {
		t.Errorf("valid manifest rejected: %v", err)
	}
	m.Permissions[0]["code"] = "other_module:demo:view"
	if err := validateManifest(m); err == nil {
		t.Error("expected error for permission prefix mismatch")
	}
	m.Permissions[0]["code"] = "test_module:demo:view"
	m.EntryBackend = "router"
	if err := validateManifest(m); err == nil {
		t.Error("expected error for invalid entry_backend format")
	}
	m.EntryBackend = ":router"
	if err := validateManifest(m); err == nil {
		t.Error("expected error for empty package in entry_backend")
	}
}
// TestExtractDependencies 验证依赖提取兼容性。
func TestExtractDependencies(t *testing.T) {
	m1 := map[string]interface{}{
		"dependencies": []interface{}{"auth", "audit_log"},
	}
	deps1 := extractDependencies(m1)
	if len(deps1) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps1))
	}
	m2 := map[string]interface{}{
		"dependencies": []string{"auth"},
	}
	deps2 := extractDependencies(m2)
	if len(deps2) != 1 || deps2[0] != "auth" {
		t.Fatalf("expected [auth], got %v", deps2)
	}
	m3 := map[string]interface{}{}
	deps3 := extractDependencies(m3)
	if len(deps3) != 0 {
		t.Fatalf("expected 0 deps, got %d", len(deps3))
	}
}
// TestStrPtrOrNil 验证空字符串转 nil。
func TestStrPtrOrNil(t *testing.T) {
	if p := strPtrOrNil(""); p != nil {
		t.Error("expected nil for empty string")
	}
	if p := strPtrOrNil("test"); p == nil || *p != "test" {
		t.Error("expected non-nil pointer with value 'test'")
	}
}
// TestExtractStringList 验证类型兼容列表提取。
func TestExtractStringList(t *testing.T) {
	if got := extractStringList([]interface{}{"a", "b"}); len(got) != 2 {
		t.Errorf("expected 2 elements, got %d", len(got))
	}
	if got := extractStringList([]string{"a"}); len(got) != 1 || got[0] != "a" {
		t.Errorf("expected [a], got %v", got)
	}
	if got := extractStringList(nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
	if got := extractStringList([]interface{}{"a", 123, "b"}); len(got) != 2 {
		t.Errorf("expected 2 string elements, got %d", len(got))
	}
}
// TestValidateConfig 验证 config_schema 校验。
func TestValidateConfig(t *testing.T) {
	manifest := map[string]interface{}{
		"config_schema": map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"max_users"},
			"properties": map[string]interface{}{
				"max_users": map[string]interface{}{"type": "integer"},
				"enabled":   map[string]interface{}{"type": "boolean"},
			},
		},
	}
	if err := validateConfig(manifest, map[string]interface{}{
		"max_users": 100,
		"enabled":   true,
	}); err != nil {
		t.Errorf("valid config rejected: %v", err)
	}
	if err := validateConfig(manifest, map[string]interface{}{
		"enabled": true,
	}); err == nil {
		t.Error("expected error for missing required field")
	}
	if err := validateConfig(manifest, map[string]interface{}{
		"max_users": "not-a-number",
	}); err == nil {
		t.Error("expected error for invalid type")
	}
}
// ---------------------------------------------------------------------------
// ZIP 安全测试
// ---------------------------------------------------------------------------
// TestSafeExtractZipPathTraversal 验证 ZIP 路径穿越防护（基础）。
func TestSafeExtractZipPathTraversal(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("../../../evil.txt")
	if err != nil {
		t.Fatalf("create zip entry failed: %v", err)
	}
	if _, err := w.Write([]byte("evil")); err != nil {
		t.Fatalf("write zip entry failed: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip failed: %v", err)
	}
	tmpFile := filepath.Join(t.TempDir(), "evil.zip")
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write tmp zip failed: %v", err)
	}
	targetDir := t.TempDir()
	err = SafeExtractZip(tmpFile, targetDir, DefaultZipMaxSize, DefaultZipMaxTotal, DefaultZipMaxFiles)
	if err == nil {
		t.Fatal("expected error for path traversal zip")
	}
}
// TestSafeExtractZipCaseSensitiveTraversal 验证大小写敏感目录的路径穿越防护。
//
// v1.3（P1-REGRESSION-01）：新增，覆盖 v1.2 引入的 ToLower 漏洞。
func TestSafeExtractZipCaseSensitiveTraversal(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("../module/evil.txt")
	if err != nil {
		t.Fatalf("create zip entry failed: %v", err)
	}
	if _, err := w.Write([]byte("evil")); err != nil {
		t.Fatalf("write zip entry failed: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip failed: %v", err)
	}
	tmpFile := filepath.Join(t.TempDir(), "case_traversal.zip")
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write tmp zip failed: %v", err)
	}
	parent := t.TempDir()
	targetDir := filepath.Join(parent, "Module")
	err = SafeExtractZip(tmpFile, targetDir, DefaultZipMaxSize, DefaultZipMaxTotal, DefaultZipMaxFiles)
	if err == nil {
		t.Fatalf("expected error for case-sensitive path traversal; "+
			"targetDir=%s, entry=../module/evil.txt", targetDir)
	}
	if _, statErr := os.Stat(filepath.Join(parent, "module", "evil.txt")); statErr == nil {
		t.Fatal("path traversal succeeded: file written to sibling dir")
	}
}
// TestSafeExtractZipSymlinkRejected 验证 ZIP 符号链接被拒绝。
//
// v1.3（P2-NEW-09）：替换 v1.2 的 `t.Skip`，改为真实构造 symlink 条目。
func TestSafeExtractZipSymlinkRejected(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	header := &zip.FileHeader{
		Name: "evil_link",
	}
	header.SetMode(os.ModeSymlink | 0o777)
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("create symlink entry failed: %v", err)
	}
	if _, err := w.Write([]byte("/etc/passwd")); err != nil {
		t.Fatalf("write symlink target failed: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip failed: %v", err)
	}
	tmpFile := filepath.Join(t.TempDir(), "symlink.zip")
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write tmp zip failed: %v", err)
	}
	targetDir := t.TempDir()
	err = SafeExtractZip(tmpFile, targetDir, DefaultZipMaxSize, DefaultZipMaxTotal, DefaultZipMaxFiles)
	if err == nil {
		t.Fatal("expected error for symlink zip")
	}
}
// ---------------------------------------------------------------------------
// Loader 测试
// ---------------------------------------------------------------------------
// TestLoaderMarkUnmark 验证 Loader 加载标记。
func TestLoaderMarkUnmark(t *testing.T) {
	loader := NewLoader()
	if loader.IsLoaded("auth") {
		t.Error("expected auth not loaded")
	}
	loader.MarkLoaded("auth")
	if !loader.IsLoaded("auth") {
		t.Error("expected auth loaded after MarkLoaded")
	}
	loader.UnmarkLoaded("auth")
	if loader.IsLoaded("auth") {
		t.Error("expected auth not loaded after UnmarkLoaded")
	}
}
// TestServiceSetLoader 验证 Service 注入 Loader。
func TestServiceSetLoader(t *testing.T) {
	svc := &Service{}
	if svc.loader != nil {
		t.Error("expected nil loader initially")
	}
	loader := NewLoader()
	svc.SetLoader(loader)
	if svc.loader != loader {
		t.Error("expected loader injected")
	}
}
// ---------------------------------------------------------------------------
// 流程测试
// ---------------------------------------------------------------------------
// TestLoadManifestRejectInvalidPermissionPrefix 验证 manifest 权限前缀校验。
//
// v1.3（P2-NEW-09）：从 v1.2 的 `TestInstallModuleRollback` 重命名而来。
func TestLoadManifestRejectInvalidPermissionPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "test_manifest_module")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	badManifest := map[string]interface{}{
		"id":      "test_manifest_module",
		"name":    "Manifest Test",
		"version": "1.0.0",
		"permissions": []map[string]interface{}{
			{"code": "other_module:demo:view", "name": "x", "resource": "demo", "action": "view"},
		},
		"menus":           []map[string]interface{}{},
		"entry_backend":   "router:router",
		"database_tables": []string{},
	}
	data, _ := json.Marshal(badManifest)
	if err := os.WriteFile(filepath.Join(moduleDir, "manifest.json"), data, 0o644); err != nil {
		t.Fatalf("write manifest failed: %v", err)
	}
	_, err := LoadManifestFromDir(moduleDir)
	if err == nil {
		t.Fatal("expected error for invalid manifest")
	}
}
// TestInstallModuleManifestInvalidNoResidue 验证 manifest 校验失败后无 DB/FS 残留。
//
// v1.3（P2-NEW-09）：新增，使用真实 SQLite + 临时目录覆盖"回滚"语义。
func TestInstallModuleManifestInvalidNoResidue(t *testing.T) {
	db, cfg := setupTestDBAndConfig(t)
	svc := NewService(db, cfg)
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src", "invalid_module")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	badManifest := map[string]interface{}{
		"id":      "invalid_module",
		"name":    "Invalid",
		"version": "1.0.0",
		"permissions": []map[string]interface{}{
			{"code": "other:demo:view", "name": "x", "resource": "demo", "action": "view"},
		},
		"menus":           []map[string]interface{}{},
		"entry_backend":   "router:router",
		"database_tables": []string{},
	}
	manifestBytes, _ := json.Marshal(badManifest)
	if err := os.WriteFile(filepath.Join(srcDir, "manifest.json"), manifestBytes, 0o644); err != nil {
		t.Fatalf("write manifest failed: %v", err)
	}
	zipPath := filepath.Join(tmpDir, "invalid_module.zip")
	if err := createZipFromDir(srcDir, zipPath); err != nil {
		t.Fatalf("create zip failed: %v", err)
	}
	_, err := svc.InstallModule("zip", zipPath, "", 1, "admin")
	if err == nil {
		t.Fatal("expected install error for invalid manifest")
	}
	var count int64
	if err := db.Model(&models.Module{}).Where("id = ?", "invalid_module").Count(&count).Error; err != nil {
		t.Fatalf("count modules failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no module record, got %d", count)
	}
	targetDir := filepath.Join(cfg.ModulesDir, "module_invalid_module")
	if _, statErr := os.Stat(targetDir); statErr == nil {
		t.Fatalf("expected no directory residue, but %s exists", targetDir)
	}
	if _, statErr := os.Stat(targetDir + ".new"); statErr == nil {
		t.Fatalf("expected no .new residue, but %s.new exists", targetDir)
	}
	if _, statErr := os.Stat(targetDir + ".old"); statErr == nil {
		t.Fatalf("expected no .old residue, but %s.old exists", targetDir)
	}
}
// TestUpgradeModuleVersionRejected 验证升级时低版本被拒绝。
func TestUpgradeModuleVersionRejected(t *testing.T) {
	if compareVersion("1.0.0", "1.1.0") >= 0 {
		t.Error("expected 1.0.0 < 1.1.0")
	}
	if compareVersion("1.1.0", "1.0.0") <= 0 {
		t.Error("expected 1.1.0 > 1.0.0")
	}
	if compareVersion("1.0.0", "1.0.0") != 0 {
		t.Error("expected 1.0.0 == 1.0.0")
	}
}
// TestLoadActiveModulesUsesServiceLoader 验证 LoadActiveModules 使用 s.loader。
//
// v1.3（P2-NEW-08）：新增，验证 Loader 唯一真实源。
func TestLoadActiveModulesUsesServiceLoader(t *testing.T) {
	db, cfg := setupTestDBAndConfig(t)
	m := models.Module{
		ID:           "test_active_module",
		Name:         "Test Active",
		Version:      "1.0.0",
		Status:       "active",
		EntryBackend: "router:router",
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("create module failed: %v", err)
	}
	svc := NewService(db, cfg)
	loader := NewLoader()
	svc.SetLoader(loader)
	order, err := svc.LoadActiveModules(db)
	if err != nil {
		t.Fatalf("LoadActiveModules failed: %v", err)
	}
	if len(order) == 0 {
		t.Fatal("expected non-empty order")
	}
	if !loader.IsLoaded("test_active_module") {
		t.Error("expected test_active_module marked as loaded via s.loader")
	}
}
// ---------------------------------------------------------------------------
// 测试辅助函数
// ---------------------------------------------------------------------------
// setupTestDBAndConfig 创建测试用 SQLite 文件库与临时目录配置。
func setupTestDBAndConfig(t *testing.T) (*gorm.DB, *config.Config) {
	t.Helper()
	tmpRoot := t.TempDir()
	cfg := &config.Config{
		DatabaseURL:       "sqlite:///" + filepath.Join(tmpRoot, "test.db"),
		ModulesDir:        filepath.Join(tmpRoot, "modules"),
		ModuleUploadDir:   filepath.Join(tmpRoot, "uploads"),
		ModuleZipMaxSize:  DefaultZipMaxSize,
		ModuleZipMaxTotal: DefaultZipMaxTotal,
		ModuleZipMaxFiles: DefaultZipMaxFiles,
	}
	if err := os.MkdirAll(cfg.ModulesDir, 0o755); err != nil {
		t.Fatalf("mkdir modules dir failed: %v", err)
	}
	if err := os.MkdirAll(cfg.ModuleUploadDir, 0o755); err != nil {
		t.Fatalf("mkdir uploads dir failed: %v", err)
	}
	db, err := database.Init(cfg)
	if err != nil {
		t.Fatalf("init db failed: %v", err)
	}
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	if err := models.EnsureModuleTables(db); err != nil {
		t.Fatalf("ensure module tables failed: %v", err)
	}
	return db, cfg
}
// createZipFromDir 将目录打包为 ZIP。
func createZipFromDir(srcDir, zipPath string) error {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	rootName := filepath.Base(srcDir)
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		entryName := filepath.Join(rootName, rel)
		entryName = filepath.ToSlash(entryName)
		w, err := zw.Create(entryName)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})
	if err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(zipPath, buf.Bytes(), 0o644)
}
