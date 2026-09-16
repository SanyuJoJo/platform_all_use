package module_manager
// CORE_MODULE_IDS 核心模块 ID，不允许停用/卸载/升级。
//
// P2-01：删除原 CORE_MODULE_PROTECTED_ACTIONS（未使用），
// 保护逻辑统一通过 s.isCoreModule(moduleID) 判定。
var CORE_MODULE_IDS = map[string]struct{}{
	"auth":           {},
	"platform":       {},
	"module_manager": {},
	"audit_log":      {},
	"license":        {},
}
// ZIP 安全默认值（可被 settings 覆盖）。
const (
	DefaultZipMaxSize  = 50 * 1024 * 1024
	DefaultZipMaxTotal = 200 * 1024 * 1024
	DefaultZipMaxFiles = 2000
)
// ZipMaxCompressRatio ZIP 压缩比上限。
const ZipMaxCompressRatio = 100
// CopyIgnorePatterns 复制源码时忽略的模式。
var CopyIgnorePatterns = []string{
	".git",
	".gitignore",
	"__pycache__",
	"*.pyc",
	"*.pyo",
	".mypy_cache",
	".ruff_cache",
	".pytest_cache",
	"node_modules",
	".venv",
	"venv",
	"*.log",
	"*.db",
}
// CoreModuleSeed 核心模块种子数据（与 Python constants.py 完全一致）。
type CoreModuleSeed struct {
	ID            string
	Name          string
	Version       string
	Description   string
	Author        string
	Status        string
	EntryBackend  string
	EntryFrontend *string
	Manifest      map[string]interface{}
	Config        map[string]interface{}
}
// CoreModules 核心模块种子数据列表（与 Python constants.py 完全一致）。
var CoreModules = []CoreModuleSeed{
	{
		ID: "platform", Name: "平台基础", Version: "1.0.0",
		Description: "平台核心功能", Author: "平台团队",
		Status: "active", EntryBackend: "router:router",
		EntryFrontend: nil,
		Manifest: map[string]interface{}{
			"id": "platform", "name": "平台基础", "version": "1.0.0",
			"description": "平台核心功能",
			"dependencies": []interface{}{},
			"permissions":  []interface{}{},
			"menus": []interface{}{
				map[string]interface{}{
					"id": "platform:dashboard", "parent_id": nil, "title": "仪表盘",
					"icon": "Dashboard", "path": "/dashboard",
					"component":  "views/dashboard/index.vue",
					"permission": "platform:dashboard:view", "order": 0,
				},
			},
		},
		Config: map[string]interface{}{},
	},
	{
		ID: "auth", Name: "认证授权", Version: "1.0.0",
		Description: "用户管理、角色管理、权限管理", Author: "平台团队",
		Status: "active", EntryBackend: "router:router",
		EntryFrontend: strPtr("/sub-apps/auth/"),
		Manifest: map[string]interface{}{
			"id": "auth", "name": "认证授权", "version": "1.0.0",
			"description": "用户管理、角色管理、权限管理",
			"dependencies": []interface{}{},
			"permissions":  []interface{}{},
			"menus": []interface{}{
				map[string]interface{}{
					"id": "auth:dashboard", "parent_id": nil, "title": "认证授权",
					"icon": "Lock", "path": "/dashboard",
					"component":  "views/dashboard/index.vue",
					"permission": "auth:user:view", "order": 10,
				},
				map[string]interface{}{
					"id": "auth:users", "parent_id": "auth:dashboard", "title": "用户管理",
					"icon": "Person", "path": "/users",
					"component":  "views/users/index.vue",
					"permission": "auth:user:view", "order": 10,
				},
				map[string]interface{}{
					"id": "auth:roles", "parent_id": "auth:dashboard", "title": "角色管理",
					"icon": "Shield", "path": "/roles",
					"component":  "views/roles/index.vue",
					"permission": "auth:role:view", "order": 20,
				},
				map[string]interface{}{
					"id": "auth:permissions", "parent_id": "auth:dashboard", "title": "权限管理",
					"icon": "List", "path": "/permissions",
					"component":  "views/permissions/index.vue",
					"permission": "auth:permission:view", "order": 30,
				},
			},
		},
		Config: map[string]interface{}{},
	},
	{
		ID: "module_manager", Name: "模块管理", Version: "1.0.0",
		Description: "管理已安装模块的生命周期", Author: "平台团队",
		Status: "active", EntryBackend: "router:router",
		EntryFrontend: strPtr("/sub-apps/module-manager/"),
		Manifest: map[string]interface{}{
			"id": "module_manager", "name": "模块管理", "version": "1.0.0",
			"description": "管理已安装模块的生命周期",
			"dependencies": []interface{}{},
			"permissions":  []interface{}{},
			"menus": []interface{}{
				map[string]interface{}{
					"id": "module_manager:dashboard", "parent_id": nil, "title": "模块管理",
					"icon": "Apps", "path": "/modules",
					"component":  "views/modules/index.vue",
					"permission": "module_manager:module:view", "order": 30,
				},
			},
		},
		Config: map[string]interface{}{},
	},
	{
		ID: "audit_log", Name: "日志审计", Version: "1.0.0",
		Description: "记录和查看系统操作日志", Author: "平台团队",
		Status: "active", EntryBackend: "router:router",
		EntryFrontend: strPtr("/sub-apps/audit-log/"),
		Manifest: map[string]interface{}{
			"id": "audit_log", "name": "日志审计", "version": "1.0.0",
			"description": "记录和查看系统操作日志",
			"dependencies": []interface{}{"auth"},
			"permissions":  []interface{}{},
			"menus": []interface{}{
				map[string]interface{}{
					"id": "audit_log:dashboard", "parent_id": nil, "title": "日志审计",
					"icon": "Document", "path": "/logs",
					"component":  "views/logs/index.vue",
					"permission": "audit_log:log:view", "order": 40,
				},
			},
		},
		Config: map[string]interface{}{},
	},
	{
		ID: "license", Name: "License 管理", Version: "1.0.0",
		Description: "License 导入、激活、状态查询", Author: "平台团队",
		Status: "active", EntryBackend: "router:router",
		EntryFrontend: strPtr("/sub-apps/license/"),
		Manifest: map[string]interface{}{
			"id": "license", "name": "License 管理", "version": "1.0.0",
			"description": "License 导入、激活、状态查询",
			"dependencies": []interface{}{"auth"},
			"permissions":  []interface{}{},
			"menus": []interface{}{
				map[string]interface{}{
					"id": "license:dashboard", "parent_id": nil, "title": "License 管理",
					"icon": "Key", "path": "/status",
					"component":  "views/license/index.vue",
					"permission": "license:license:view", "order": 50,
				},
			},
		},
		Config: map[string]interface{}{},
	},
}
func strPtr(s string) *string { return &s }
