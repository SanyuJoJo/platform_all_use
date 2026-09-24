package module_manager

// CORE_MODULE_IDS 核心模块 ID，不允许停用/卸载/升级。
var CORE_MODULE_IDS = map[string]struct{}{
	"auth":           {},
	"platform":       {},
	"module_manager": {},
	"audit_log":      {},
	"license":        {},
	"crypto_console": {},
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

// CoreModuleSeed 核心模块种子数据。
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

// CoreModules 核心模块种子数据列表。
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
	// ========================================================================
	// 证书管理控制台（P0 平台后端新增）
	//
	// 菜单结构（与 auth 模块结构对齐）：
	//   证书管理（父，parent_id=null，有真实 path）
	//     ├── CA 管理
	//     ├── 证书管理
	//     ├── CSR 管理
	//     ├── CRL 管理
	//     ├── 密钥管理
	//     ├── 任务中心
	//     └── 审计日志
	//
	// ★ 关键约束（前端 menu.ts 排序+建树的实现决定）：
	//   前端把父菜单与子菜单混合后按 order 全局排序，
	//   遍历时父节点必须已经在 map 中，子节点才能正确挂载。
	//   因此：父菜单 order 必须 < 所有子菜单 order。
	//   这里：父菜单 order = 60，子菜单 order = 61 ~ 67。
	// ========================================================================
	{
		ID: "crypto_console", Name: "证书管理控制台", Version: "1.0.0",
		Description:   "CA、证书、CSR、CRL、密钥、任务、审计",
		Author:        "平台团队",
		Status:        "active",
		EntryBackend:  "router:router",
		EntryFrontend: strPtr("/sub-apps/crypto-console/"),
		Manifest: map[string]interface{}{
			"id":          "crypto_console",
			"name":        "证书管理控制台",
			"version":     "1.0.0",
			"description": "CA、证书、CSR、CRL、密钥、任务、审计",
			"dependencies": []interface{}{
				"auth",
			},
			"permissions": []interface{}{},
			"menus": []interface{}{
				// ---- 父菜单：证书管理（order=60，必须小于所有子菜单） ----
				map[string]interface{}{
					"id":         "crypto_console:dashboard",
					"parent_id":  nil,
					"title":      "证书管理",
					"icon":       "Certificate",
					"path":       "/ca",
					"component":  "views/ca/index.vue",
					"permission": "crypto_console:ca:view",
					"order":      60,
				},
				// ---- 子菜单：CA 管理 ----
				map[string]interface{}{
					"id":         "crypto_console:ca",
					"parent_id":  "crypto_console:dashboard",
					"title":      "CA 管理",
					"icon":       "Certificate",
					"path":       "/ca",
					"component":  "views/ca/index.vue",
					"permission": "crypto_console:ca:view",
					"order":      61,
				},
				// ---- 子菜单：证书管理 ----
				map[string]interface{}{
					"id":         "crypto_console:cert",
					"parent_id":  "crypto_console:dashboard",
					"title":      "证书管理",
					"icon":       "Document",
					"path":       "/certs",
					"component":  "views/certs/index.vue",
					"permission": "crypto_console:cert:view",
					"order":      62,
				},
				// ---- 子菜单：CSR 管理 ----
				map[string]interface{}{
					"id":         "crypto_console:csr",
					"parent_id":  "crypto_console:dashboard",
					"title":      "CSR 管理",
					"icon":       "DocumentText",
					"path":       "/csr",
					"component":  "views/csr/index.vue",
					"permission": "crypto_console:csr:view",
					"order":      63,
				},
				// ---- 子菜单：CRL 管理 ----
				map[string]interface{}{
					"id":         "crypto_console:crl",
					"parent_id":  "crypto_console:dashboard",
					"title":      "CRL 管理",
					"icon":       "List",
					"path":       "/crl",
					"component":  "views/crl/index.vue",
					"permission": "crypto_console:crl:view",
					"order":      64,
				},
				// ---- 子菜单：密钥管理 ----
				map[string]interface{}{
					"id":         "crypto_console:key",
					"parent_id":  "crypto_console:dashboard",
					"title":      "密钥管理",
					"icon":       "Key",
					"path":       "/keys",
					"component":  "views/keys/index.vue",
					"permission": "crypto_console:key:view",
					"order":      65,
				},
				// ---- 子菜单：任务中心 ----
				map[string]interface{}{
					"id":         "crypto_console:task",
					"parent_id":  "crypto_console:dashboard",
					"title":      "任务中心",
					"icon":       "Time",
					"path":       "/tasks",
					"component":  "views/tasks/index.vue",
					"permission": "crypto_console:task:view",
					"order":      66,
				},
				// ---- 子菜单：审计日志 ----
				map[string]interface{}{
					"id":         "crypto_console:audit",
					"parent_id":  "crypto_console:dashboard",
					"title":      "审计日志",
					"icon":       "Search",
					"path":       "/audits",
					"component":  "views/audits/index.vue",
					"permission": "crypto_console:audit:view",
					"order":      67,
				},
			},
		},
		Config: map[string]interface{}{},
	},
}

// strPtr 返回字符串指针。
func strPtr(s string) *string { return &s }
