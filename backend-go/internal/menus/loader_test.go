package menus
import "testing"
func TestValidateOK(t *testing.T) {
	cfg := &FrontendConfig{
		PlatformMenus: []interface{}{
			map[string]interface{}{
				"id": "platform:dashboard", "title": "仪表盘", "path": "/dashboard",
				"permission": "platform:dashboard:view",
				"children": []interface{}{
					map[string]interface{}{
						"id": "platform:sub", "title": "子菜单", "path": "/dashboard/sub",
					},
				},
			},
		},
		EntryFrontendOverrides: map[string]string{"auth": "/sub-apps/auth/"},
	}
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}
func TestValidateDuplicateMenuID(t *testing.T) {
	cfg := &FrontendConfig{
		PlatformMenus: []interface{}{
			map[string]interface{}{"id": "a", "title": "A", "path": "/a"},
			map[string]interface{}{"id": "a", "title": "A2", "path": "/a2"},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Fatal("expected duplicate id error")
	}
}
func TestValidateDuplicatePath(t *testing.T) {
	cfg := &FrontendConfig{
		PlatformMenus: []interface{}{
			map[string]interface{}{"id": "a", "title": "A", "path": "/dup"},
			map[string]interface{}{"id": "b", "title": "B", "path": "/dup"},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Fatal("expected duplicate path error")
	}
}
func TestValidateInvalidPermission(t *testing.T) {
	cfg := &FrontendConfig{
		PlatformMenus: []interface{}{
			map[string]interface{}{
				"id": "a", "title": "A", "path": "/a", "permission": "INVALID",
			},
		},
	}
	if err := Validate(cfg); err == nil {
		t.Fatal("expected invalid permission error")
	}
}
func TestValidateEmptyOverrideValue(t *testing.T) {
	cfg := &FrontendConfig{
		EntryFrontendOverrides: map[string]string{"auth": ""},
	}
	if err := Validate(cfg); err == nil {
		t.Fatal("expected empty override value error")
	}
}
