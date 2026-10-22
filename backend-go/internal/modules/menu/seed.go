package menu

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/models"
)

// seedMenuEntry 菜单种子条目。
//
// ParentChain 用菜单名链定位父节点，例如：
//   - ""                    → 顶级（parent_id = 0）
//   - "系统管理"             → 系统管理的一级子菜单
//   - "系统管理/用户管理"     → 用户管理 下的按钮菜单
//
// 用链式定位而不是 ParentID 硬编码，是为了让 seed 数据与数据库的实际 ID
// 解耦。顺序要求：父节点必须先于其子节点出现在 seedMenuEntries 中。
type seedMenuEntry struct {
	ParentChain string
	Menu        models.Menu
}

// seedMenuEntries 全量菜单种子。
//
// 结构：
//
//	仪表盘              (C, main, /dashboard)
//	系统管理            (M, auth, /system)
//	├── 用户管理        (C, /system/users)
//	│   ├── 新增用户    (F, auth:user:create)
//	│   ├── 编辑用户    (F, auth:user:edit)
//	│   └── 删除用户    (F, auth:user:delete)
//	├── 角色管理        (C, /system/roles)
//	│   ├── 新增角色    (F, auth:role:create)
//	│   ├── 编辑角色    (F, auth:role:edit)
//	│   └── 删除角色    (F, auth:role:delete)
//	├── 权限管理        (C, /system/permissions)  ← 只读，无 F
//	└── 菜单管理        (C, /system/menus)
//	    ├── 新增菜单    (F, auth:menu:create)
//	    ├── 修改菜单    (F, auth:menu:update)
//	    └── 删除菜单    (F, auth:menu:delete)
//	模块管理            (C, module_manager, /modules)
//	├── 安装模块        (F, module_manager:module:create)
//	├── 编辑模块        (F, module_manager:module:edit)
//	└── 卸载模块        (F, module_manager:module:delete)
//	日志审计            (C, audit_log, /audit-logs)
//	└── 导出日志        (F, audit_log:log:export)
//	License 管理        (C, license, /licenses)
//	└── 管理 License    (F, license:license:create)
//	安全管理            (M, crypto_console, /security)
//	├── 证书运维        (M, /security/cert-ops)
//	│   ├── CA 管理     (C, /ca)
//	│   │   ├── 新增 CA (F, crypto_console:ca:create)
//	│   │   ├── 导入 CA (F, crypto_console:ca:import)
//	│   │   ├── 导出 CA (F, crypto_console:ca:export)
//	│   │   └── 删除 CA (F, crypto_console:ca:delete)
//	│   ├── 证书管理    (C, /certs)
//	│   │   ├── 签发证书 (F, crypto_console:cert:sign)
//	│   │   ├── 导入证书 (F, crypto_console:cert:import)
//	│   │   ├── 导出证书 (F, crypto_console:cert:export)
//	│   │   └── 删除证书 (F, crypto_console:cert:delete)
//	│   ├── CSR 管理    (C, /csr)
//	│   │   ├── 创建 P10 (F, crypto_console:csr:create)
//	│   │   └── 删除 P10 (F, crypto_console:csr:delete)
//	│   ├── CRL 管理    (C, /crl)
//	│   │   └── 创建 CRL (F, crypto_console:crl:create)
//	│   ├── 密钥管理    (C, /keys)
//	│   │   └── 管理密钥 (F, crypto_console:key:manage)
//	│   ├── 任务中心    (C, /tasks)
//	│   │   └── 取消任务 (F, crypto_console:task:cancel)
//	│   └── 审计日志    (C, /audits)
//	└── 业务管理        (M, /security/business)
//	    └── HTTP 管理   (C, /list)
//	        ├── 新建代理 (F, ssl_proxy:proxy:create)
//	        ├── 编辑代理 (F, ssl_proxy:proxy:update)
//	        ├── 删除代理 (F, ssl_proxy:proxy:delete)
//	        ├── 启动代理 (F, ssl_proxy:proxy:start)
//	        ├── 停止代理 (F, ssl_proxy:proxy:stop)
//	        ├── 查看配置 (F, ssl_proxy:proxy:config)
//	        └── 查看日志 (F, ssl_proxy:proxy:logs)
//
// 约定：
//   - M 目录：path 用累积路径（含父级），如 /security/cert-ops；
//   - C 菜单：path 用相对路径（如 /ca），前端按 `/{microApp}/{path}` 拼接；
//   - F 按钮：path 留空（不参与路由），perms 为按钮权限码；
//   - 每个节点用「自身所属模块的 microApp」，不是父节点的。
var seedMenuEntries = []seedMenuEntry{
	// ---- 顶级 ----
	{ParentChain: "", Menu: models.Menu{MenuName: "仪表盘", Name: "Dashboard", MicroApp: "main", OrderNum: 1, Path: "/dashboard", Component: "dashboard/index", MenuType: "C", Icon: "Odometer", ModuleID: "platform"}},
	{ParentChain: "", Menu: models.Menu{MenuName: "系统管理", Name: "System", MicroApp: "auth", OrderNum: 2, Path: "/system", MenuType: "M", Icon: "Setting", ModuleID: "auth"}},
	{ParentChain: "", Menu: models.Menu{MenuName: "模块管理", Name: "ModuleManager", MicroApp: "module_manager", OrderNum: 3, Path: "/modules", Component: "modules/index", MenuType: "C", Icon: "Grid", ModuleID: "module_manager"}},
	{ParentChain: "", Menu: models.Menu{MenuName: "日志审计", Name: "AuditLog", MicroApp: "audit_log", OrderNum: 4, Path: "/audit-logs", Component: "logs/index", MenuType: "C", Icon: "Document", ModuleID: "audit_log"}},
	{ParentChain: "", Menu: models.Menu{MenuName: "License 管理", Name: "License", MicroApp: "license", OrderNum: 5, Path: "/licenses", Component: "license/index", MenuType: "C", Icon: "Key", ModuleID: "license"}},
	{ParentChain: "", Menu: models.Menu{MenuName: "安全管理", Name: "Security", MicroApp: "crypto_console", OrderNum: 6, Path: "/security", MenuType: "M", Icon: "Shield", ModuleID: "crypto"}},

	// ---- 系统管理 子菜单 ----
	{ParentChain: "系统管理", Menu: models.Menu{MenuName: "用户管理", Name: "SystemUsers", MicroApp: "auth", OrderNum: 1, Path: "/system/users", Component: "system/user/index", MenuType: "C", Icon: "User", ModuleID: "auth", Perms: "auth:user:list"}},
	{ParentChain: "系统管理", Menu: models.Menu{MenuName: "角色管理", Name: "SystemRoles", MicroApp: "auth", OrderNum: 2, Path: "/system/roles", Component: "system/role/index", MenuType: "C", Icon: "UserFilled", ModuleID: "auth", Perms: "auth:role:list"}},
	{ParentChain: "系统管理", Menu: models.Menu{MenuName: "权限管理", Name: "SystemPermissions", MicroApp: "auth", OrderNum: 3, Path: "/system/permissions", Component: "system/permission/index", MenuType: "C", Icon: "Key", ModuleID: "auth", Perms: "auth:permission:list"}},
	{ParentChain: "系统管理", Menu: models.Menu{MenuName: "菜单管理", Name: "SystemMenus", MicroApp: "auth", OrderNum: 4, Path: "/system/menus", Component: "system/menu/index", MenuType: "C", Icon: "Menu", ModuleID: "auth", Perms: "auth:menu:list"}},

	// ---- 用户管理 按钮 ----
	{ParentChain: "系统管理/用户管理", Menu: models.Menu{MenuName: "新增用户", Name: "UserCreate", MicroApp: "auth", OrderNum: 1, MenuType: "F", ModuleID: "auth", Perms: "auth:user:create"}},
	{ParentChain: "系统管理/用户管理", Menu: models.Menu{MenuName: "编辑用户", Name: "UserEdit", MicroApp: "auth", OrderNum: 2, MenuType: "F", ModuleID: "auth", Perms: "auth:user:edit"}},
	{ParentChain: "系统管理/用户管理", Menu: models.Menu{MenuName: "删除用户", Name: "UserDelete", MicroApp: "auth", OrderNum: 3, MenuType: "F", ModuleID: "auth", Perms: "auth:user:delete"}},

	// ---- 角色管理 按钮 ----
	{ParentChain: "系统管理/角色管理", Menu: models.Menu{MenuName: "新增角色", Name: "RoleCreate", MicroApp: "auth", OrderNum: 1, MenuType: "F", ModuleID: "auth", Perms: "auth:role:create"}},
	{ParentChain: "系统管理/角色管理", Menu: models.Menu{MenuName: "编辑角色", Name: "RoleEdit", MicroApp: "auth", OrderNum: 2, MenuType: "F", ModuleID: "auth", Perms: "auth:role:edit"}},
	{ParentChain: "系统管理/角色管理", Menu: models.Menu{MenuName: "删除角色", Name: "RoleDelete", MicroApp: "auth", OrderNum: 3, MenuType: "F", ModuleID: "auth", Perms: "auth:role:delete"}},

	// ---- 菜单管理 按钮 ----
	{ParentChain: "系统管理/菜单管理", Menu: models.Menu{MenuName: "新增菜单", Name: "MenuCreate", MicroApp: "auth", OrderNum: 1, MenuType: "F", ModuleID: "auth", Perms: "auth:menu:create"}},
	{ParentChain: "系统管理/菜单管理", Menu: models.Menu{MenuName: "修改菜单", Name: "MenuUpdate", MicroApp: "auth", OrderNum: 2, MenuType: "F", ModuleID: "auth", Perms: "auth:menu:update"}},
	{ParentChain: "系统管理/菜单管理", Menu: models.Menu{MenuName: "删除菜单", Name: "MenuDelete", MicroApp: "auth", OrderNum: 3, MenuType: "F", ModuleID: "auth", Perms: "auth:menu:delete"}},

	// ---- 模块管理 按钮 ----
	{ParentChain: "模块管理", Menu: models.Menu{MenuName: "安装模块", Name: "ModuleInstall", MicroApp: "module_manager", OrderNum: 1, MenuType: "F", ModuleID: "module_manager", Perms: "module_manager:module:create"}},
	{ParentChain: "模块管理", Menu: models.Menu{MenuName: "编辑模块", Name: "ModuleEdit", MicroApp: "module_manager", OrderNum: 2, MenuType: "F", ModuleID: "module_manager", Perms: "module_manager:module:edit"}},
	{ParentChain: "模块管理", Menu: models.Menu{MenuName: "卸载模块", Name: "ModuleUninstall", MicroApp: "module_manager", OrderNum: 3, MenuType: "F", ModuleID: "module_manager", Perms: "module_manager:module:delete"}},

	// ---- 日志审计 按钮 ----
	{ParentChain: "日志审计", Menu: models.Menu{MenuName: "导出日志", Name: "AuditLogExport", MicroApp: "audit_log", OrderNum: 1, MenuType: "F", ModuleID: "audit_log", Perms: "audit_log:log:export"}},

	// ---- License 管理 按钮 ----
	{ParentChain: "License 管理", Menu: models.Menu{MenuName: "管理 License", Name: "LicenseManage", MicroApp: "license", OrderNum: 1, MenuType: "F", ModuleID: "license", Perms: "license:license:create"}},

	// ---- 安全管理 子目录 ----
	{ParentChain: "安全管理", Menu: models.Menu{MenuName: "证书运维", Name: "CertOps", MicroApp: "crypto_console", OrderNum: 10, Path: "/security/cert-ops", MenuType: "M", Icon: "Certificate", ModuleID: "crypto"}},
	{ParentChain: "安全管理", Menu: models.Menu{MenuName: "业务管理", Name: "Business", MicroApp: "crypto_console", OrderNum: 20, Path: "/security/business", MenuType: "M", Icon: "Apps", ModuleID: "crypto"}},

	// ---- 证书运维 子菜单 ----
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "CA 管理", Name: "CryptoCA", MicroApp: "crypto_console", OrderNum: 10, Path: "/ca", Component: "ca/index", MenuType: "C", Icon: "Certificate", ModuleID: "crypto", Perms: "crypto_console:ca:view"}},
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "证书管理", Name: "CryptoCert", MicroApp: "crypto_console", OrderNum: 20, Path: "/certs", Component: "certs/index", MenuType: "C", Icon: "Document", ModuleID: "crypto", Perms: "crypto_console:cert:view"}},
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "CSR 管理", Name: "CryptoCSR", MicroApp: "crypto_console", OrderNum: 30, Path: "/csr", Component: "csr/index", MenuType: "C", Icon: "DocumentText", ModuleID: "crypto", Perms: "crypto_console:csr:view"}},
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "CRL 管理", Name: "CryptoCRL", MicroApp: "crypto_console", OrderNum: 40, Path: "/crl", Component: "crl/index", MenuType: "C", Icon: "List", ModuleID: "crypto", Perms: "crypto_console:crl:view"}},
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "密钥管理", Name: "CryptoKey", MicroApp: "crypto_console", OrderNum: 50, Path: "/keys", Component: "keys/index", MenuType: "C", Icon: "Key", ModuleID: "crypto", Perms: "crypto_console:key:view"}},
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "任务中心", Name: "CryptoTask", MicroApp: "crypto_console", OrderNum: 60, Path: "/tasks", Component: "tasks/index", MenuType: "C", Icon: "Time", ModuleID: "crypto", Perms: "crypto_console:task:view"}},
	{ParentChain: "安全管理/证书运维", Menu: models.Menu{MenuName: "审计日志", Name: "CryptoAudit", MicroApp: "crypto_console", OrderNum: 70, Path: "/audits", Component: "audits/index", MenuType: "C", Icon: "Search", ModuleID: "crypto", Perms: "crypto_console:audit:view"}},

	// ---- CA 管理 按钮 ----
	{ParentChain: "安全管理/证书运维/CA 管理", Menu: models.Menu{MenuName: "新增 CA", Name: "CryptoCACreate", MicroApp: "crypto_console", OrderNum: 1, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:ca:create"}},
	{ParentChain: "安全管理/证书运维/CA 管理", Menu: models.Menu{MenuName: "导入 CA", Name: "CryptoCAImport", MicroApp: "crypto_console", OrderNum: 2, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:ca:import"}},
	{ParentChain: "安全管理/证书运维/CA 管理", Menu: models.Menu{MenuName: "导出 CA", Name: "CryptoCAExport", MicroApp: "crypto_console", OrderNum: 3, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:ca:export"}},
	{ParentChain: "安全管理/证书运维/CA 管理", Menu: models.Menu{MenuName: "删除 CA", Name: "CryptoCADelete", MicroApp: "crypto_console", OrderNum: 4, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:ca:delete"}},

	// ---- 证书管理 按钮 ----
	{ParentChain: "安全管理/证书运维/证书管理", Menu: models.Menu{MenuName: "签发证书", Name: "CryptoCertSign", MicroApp: "crypto_console", OrderNum: 1, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:cert:sign"}},
	{ParentChain: "安全管理/证书运维/证书管理", Menu: models.Menu{MenuName: "导入证书", Name: "CryptoCertImport", MicroApp: "crypto_console", OrderNum: 2, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:cert:import"}},
	{ParentChain: "安全管理/证书运维/证书管理", Menu: models.Menu{MenuName: "导出证书", Name: "CryptoCertExport", MicroApp: "crypto_console", OrderNum: 3, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:cert:export"}},
	{ParentChain: "安全管理/证书运维/证书管理", Menu: models.Menu{MenuName: "删除证书", Name: "CryptoCertDelete", MicroApp: "crypto_console", OrderNum: 4, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:cert:delete"}},

	// ---- CSR 管理 按钮 ----
	{ParentChain: "安全管理/证书运维/CSR 管理", Menu: models.Menu{MenuName: "创建 P10", Name: "CryptoCSRCreate", MicroApp: "crypto_console", OrderNum: 1, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:csr:create"}},
	{ParentChain: "安全管理/证书运维/CSR 管理", Menu: models.Menu{MenuName: "删除 P10", Name: "CryptoCSRDelete", MicroApp: "crypto_console", OrderNum: 2, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:csr:delete"}},

	// ---- CRL 管理 按钮 ----
	{ParentChain: "安全管理/证书运维/CRL 管理", Menu: models.Menu{MenuName: "创建 CRL", Name: "CryptoCRLCreate", MicroApp: "crypto_console", OrderNum: 1, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:crl:create"}},

	// ---- 密钥管理 按钮 ----
	{ParentChain: "安全管理/证书运维/密钥管理", Menu: models.Menu{MenuName: "管理密钥", Name: "CryptoKeyManage", MicroApp: "crypto_console", OrderNum: 1, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:key:manage"}},

	// ---- 任务中心 按钮 ----
	{ParentChain: "安全管理/证书运维/任务中心", Menu: models.Menu{MenuName: "取消任务", Name: "CryptoTaskCancel", MicroApp: "crypto_console", OrderNum: 1, MenuType: "F", ModuleID: "crypto", Perms: "crypto_console:task:cancel"}},

	// ---- 业务管理 子菜单（跨模块挂载）----
	{ParentChain: "安全管理/业务管理", Menu: models.Menu{MenuName: "HTTP 管理", Name: "SslProxyList", MicroApp: "ssl_proxy", OrderNum: 21, Path: "/list", Component: "list/index", MenuType: "C", Icon: "Connection", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:view"}},

	// ---- HTTP 管理 按钮 ----
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "新建代理", Name: "SslProxyCreate", MicroApp: "ssl_proxy", OrderNum: 1, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:create"}},
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "编辑代理", Name: "SslProxyUpdate", MicroApp: "ssl_proxy", OrderNum: 2, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:update"}},
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "删除代理", Name: "SslProxyDelete", MicroApp: "ssl_proxy", OrderNum: 3, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:delete"}},
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "启动代理", Name: "SslProxyStart", MicroApp: "ssl_proxy", OrderNum: 4, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:start"}},
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "停止代理", Name: "SslProxyStop", MicroApp: "ssl_proxy", OrderNum: 5, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:stop"}},
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "查看配置", Name: "SslProxyConfig", MicroApp: "ssl_proxy", OrderNum: 6, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:config"}},
	{ParentChain: "安全管理/业务管理/HTTP 管理", Menu: models.Menu{MenuName: "查看日志", Name: "SslProxyLogs", MicroApp: "ssl_proxy", OrderNum: 7, MenuType: "F", ModuleID: "ssl_proxy", Perms: "ssl_proxy:proxy:logs"}},
}

// seedPermissions 菜单管理相关权限码，需在 auth.EnsureAuthSeedData 之后调用。
var seedPermissions = []models.Permission{
	{Code: "auth:menu:list", Name: "菜单查询", ModuleID: "auth", Resource: "menu", Action: "list"},
	{Code: "auth:menu:create", Name: "菜单新增", ModuleID: "auth", Resource: "menu", Action: "create"},
	{Code: "auth:menu:update", Name: "菜单修改", ModuleID: "auth", Resource: "menu", Action: "update"},
	{Code: "auth:menu:delete", Name: "菜单删除", ModuleID: "auth", Resource: "menu", Action: "delete"},
	{Code: "auth:route_perm:list", Name: "路由权限查询", ModuleID: "auth", Resource: "route_perm", Action: "list"},
	{Code: "auth:route_perm:create", Name: "路由权限新增", ModuleID: "auth", Resource: "route_perm", Action: "create"},
	{Code: "auth:route_perm:update", Name: "路由权限修改", ModuleID: "auth", Resource: "route_perm", Action: "update"},
	{Code: "auth:route_perm:delete", Name: "路由权限删除", ModuleID: "auth", Resource: "route_perm", Action: "delete"},
}

// EnsureMenuSeedData 幂等地写入权限码与菜单种子数据。
//
// 顺序约束：seedMenuEntries 中父节点必须先于子节点，
// 这样 resolveParentChain 解析链时父节点已存在于数据库。
//
// 幂等：按 (parent_id, menu_name) 判重，已存在则跳过。
// 所以新增 F 节点在下次启动时会自动插入，无需清表。
func EnsureMenuSeedData(db *gorm.DB) error {
	if err := ensurePermissions(db); err != nil {
		return err
	}
	now := time.Now()

	for _, entry := range seedMenuEntries {
		parentID, ok := resolveParentChain(db, entry.ParentChain)
		if !ok {
			log.Warn().
				Str("parent_chain", entry.ParentChain).
				Str("menu_name", entry.Menu.MenuName).
				Msg("父菜单链不存在，跳过该菜单种子")
			continue
		}
		m := entry.Menu
		m.ParentID = parentID
		if err := ensureOneMenu(db, &m, now); err != nil {
			return err
		}
	}

	log.Info().Msg("菜单种子数据已就绪")
	return nil
}

// resolveParentChain 按菜单名链解析父节点 ID。
func resolveParentChain(db *gorm.DB, chain string) (uint, bool) {
	chain = strings.TrimSpace(chain)
	if chain == "" {
		return 0, true
	}
	parts := strings.Split(chain, "/")
	var cur uint
	for _, name := range parts {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var m models.Menu
		if err := db.Where("parent_id = ? AND menu_name = ?", cur, name).
			First(&m).Error; err != nil {
			return 0, false
		}
		cur = m.ID
	}
	return cur, true
}

func ensurePermissions(db *gorm.DB) error {
	for _, p := range seedPermissions {
		var exist models.Permission
		err := db.Where("code = ?", p.Code).First(&exist).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

// ensureOneMenu 按 (parent_id, menu_name) 判重插入。
//
// 已存在则跳过（不更新任何字段），因此结构性变更需清表重建。
// F 类型节点的 path 留空，perms 为按钮权限码。
func ensureOneMenu(db *gorm.DB, m *models.Menu, now time.Time) error {
	var exist models.Menu
	err := db.Where("parent_id = ? AND menu_name = ?", m.ParentID, m.MenuName).
		First(&exist).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	if m.Icon == "" {
		m.Icon = "#"
	}
	if m.Name == "" {
		m.Name = deriveRouteName(m.Path, 0)
	}
	if m.MicroApp == "" {
		m.MicroApp = "main"
	}
	m.CreateBy = "system"
	m.CreateTime = now
	m.UpdateBy = "system"
	m.UpdateTime = now
	return db.Create(m).Error
}
