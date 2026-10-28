// backend-go/internal/modules/auth/seed.go
package auth

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/models"
	"backend-go/internal/security"
)

// DefaultPermission 默认权限种子。
type DefaultPermission struct {
	Code     string
	Name     string
	ModuleID string
	Resource string
	Action   string
}

// DefaultPermissions 核心权限列表。
//
// 说明：
//   - 该列表是 auth_permission 表（旧表）的种子，供 RBAC 使用；
//   - 方案 A 上线后，角色权限通过 auth_role_menu 关联 platform_menu，
//     但 auth_permission 表仍保留作为"权限码字典"，供审计/兼容使用；
//   - 新增权限码时，需要在 platform_menu 里同时补一条 F 类型菜单（perms 一致），
//     否则管理员在角色菜单界面无法勾选到该按钮。
//
// 当前已涵盖的模块：
//   - platform        ：仪表盘
//   - auth            ：用户 / 角色 / 权限管理
//   - module_manager  ：模块管理
//   - audit_log       ：日志审计
//   - license         ：License 管理
//   - crypto_console  ：证书控制台（CA / Cert / CSR / CRL / Key / Task / Audit）
//   - ssl_proxy       ：SSL 代理 HTTP 管理（proxy）+ TCPS 4 层代理（tcps）
//   - crypto          ：后端密码操作 API
var DefaultPermissions = []DefaultPermission{
	// ============================================================
	// platform
	// ============================================================
	{"platform:dashboard:view", "查看仪表盘", "platform", "dashboard", "view"},

	// ============================================================
	// auth
	// ============================================================
	{"auth:user:view", "查看用户", "auth", "user", "view"},
	{"auth:user:create", "创建用户", "auth", "user", "create"},
	{"auth:user:edit", "编辑用户", "auth", "user", "edit"},
	{"auth:user:delete", "删除用户", "auth", "user", "delete"},
	{"auth:role:view", "查看角色", "auth", "role", "view"},
	{"auth:role:create", "创建角色", "auth", "role", "create"},
	{"auth:role:edit", "编辑角色", "auth", "role", "edit"},
	{"auth:role:delete", "删除角色", "auth", "role", "delete"},
	{"auth:permission:view", "查看权限", "auth", "permission", "view"},

	// ============================================================
	// module_manager
	// ============================================================
	{"module_manager:module:view", "查看模块", "module_manager", "module", "view"},
	{"module_manager:module:create", "安装模块", "module_manager", "module", "create"},
	{"module_manager:module:edit", "编辑模块", "module_manager", "module", "edit"},
	{"module_manager:module:delete", "卸载模块", "module_manager", "module", "delete"},

	// ============================================================
	// audit_log
	// ============================================================
	{"audit_log:log:view", "查看日志", "audit_log", "log", "view"},
	{"audit_log:log:export", "导出日志", "audit_log", "log", "export"},

	// ============================================================
	// license
	// ============================================================
	{"license:license:view", "查看License", "license", "license", "view"},
	{"license:license:create", "管理License", "license", "license", "create"},

	// ============================================================
	// crypto_console（证书管理控制台前端菜单）
	// ============================================================

	// CA
	{"crypto_console:ca:view", "查看 CA", "crypto_console", "ca", "view"},
	{"crypto_console:ca:create", "创建 CA", "crypto_console", "ca", "create"},
	{"crypto_console:ca:import", "导入 CA", "crypto_console", "ca", "import"},
	{"crypto_console:ca:export", "导出 CA", "crypto_console", "ca", "export"},
	{"crypto_console:ca:delete", "删除 CA", "crypto_console", "ca", "delete"},

	// 证书
	{"crypto_console:cert:view", "查看证书", "crypto_console", "cert", "view"},
	{"crypto_console:cert:sign", "签发证书", "crypto_console", "cert", "sign"},
	{"crypto_console:cert:import", "导入证书", "crypto_console", "cert", "import"},
	{"crypto_console:cert:export", "导出证书", "crypto_console", "cert", "export"},
	{"crypto_console:cert:delete", "删除证书", "crypto_console", "cert", "delete"},

	// P10（CSR）
	{"crypto_console:csr:view", "查看 P10", "crypto_console", "csr", "view"},
	{"crypto_console:csr:create", "创建/导入 P10", "crypto_console", "csr", "create"},
	{"crypto_console:csr:delete", "删除 P10", "crypto_console", "csr", "delete"},

	// CRL
	{"crypto_console:crl:view", "查看 CRL", "crypto_console", "crl", "view"},
	{"crypto_console:crl:create", "创建 CRL", "crypto_console", "crl", "create"},

	// 密钥
	{"crypto_console:key:view", "查看密钥元数据", "crypto_console", "key", "view"},
	{"crypto_console:key:manage", "管理密钥", "crypto_console", "key", "manage"},

	// 任务 / 审计
	{"crypto_console:task:view", "查看任务", "crypto_console", "task", "view"},
	{"crypto_console:task:cancel", "取消任务", "crypto_console", "task", "cancel"},
	{"crypto_console:audit:view", "查看审计", "crypto_console", "audit", "view"},

	// ============================================================
	// ssl_proxy（SSL 代理）— HTTP 管理
	// ============================================================
	{"ssl_proxy:proxy:view", "查看 SSL 代理", "ssl_proxy", "proxy", "view"},
	{"ssl_proxy:proxy:create", "新建 SSL 代理", "ssl_proxy", "proxy", "create"},
	{"ssl_proxy:proxy:update", "编辑 SSL 代理", "ssl_proxy", "proxy", "update"},
	{"ssl_proxy:proxy:delete", "删除 SSL 代理", "ssl_proxy", "proxy", "delete"},
	{"ssl_proxy:proxy:start", "启动 SSL 代理", "ssl_proxy", "proxy", "start"},
	{"ssl_proxy:proxy:stop", "停止 SSL 代理", "ssl_proxy", "proxy", "stop"},
	{"ssl_proxy:proxy:config", "查看 SSL 代理配置", "ssl_proxy", "proxy", "config"},
	{"ssl_proxy:proxy:logs", "查看 SSL 代理日志", "ssl_proxy", "proxy", "logs"},

	// ============================================================
	// ssl_proxy（SSL 代理）— TCPS 4 层 TCP/TLS 管理
	//
	// ★ 与 HTTP 管理使用相同的模块 ID（ssl_proxy），
	//   但 resource 切为 "tcps"，权限码与 ssl_proxy:proxy:* 完全隔离。
	//   前端 tcps 页面、后端 /api/v1/tcps-proxies 接口都校验这套权限码。
	// ============================================================
	{"ssl_proxy:tcps:view", "查看 TCPS 代理", "ssl_proxy", "tcps", "view"},
	{"ssl_proxy:tcps:create", "新建 TCPS 代理", "ssl_proxy", "tcps", "create"},
	{"ssl_proxy:tcps:update", "编辑 TCPS 代理", "ssl_proxy", "tcps", "update"},
	{"ssl_proxy:tcps:delete", "删除 TCPS 代理", "ssl_proxy", "tcps", "delete"},
	{"ssl_proxy:tcps:start", "启动 TCPS 代理", "ssl_proxy", "tcps", "start"},
	{"ssl_proxy:tcps:stop", "停止 TCPS 代理", "ssl_proxy", "tcps", "stop"},
	{"ssl_proxy:tcps:config", "查看 TCPS 代理配置", "ssl_proxy", "tcps", "config"},
	{"ssl_proxy:tcps:logs", "查看 TCPS 代理日志", "ssl_proxy", "tcps", "logs"},

	// ============================================================
	// crypto（后端密码操作 API）
	// ============================================================
	{"crypto:operation:execute", "执行密码操作", "crypto", "operation", "execute"},
	{"crypto:task:view", "查看密码任务", "crypto", "task", "view"},
	{"crypto:task:cancel", "取消密码任务", "crypto", "task", "cancel"},
}

// guestPermissionCodes guest 角色确定性初始权限集合。
//
// 注意：仅为 guest 角色的初始权限。管理员可以后续在角色界面调整。
var guestPermissionCodes = map[string]struct{}{
	"platform:dashboard:view":    {},
	"auth:user:view":             {},
	"module_manager:module:view": {},
}

// EnsureAuthSeedData 初始化认证模块种子数据（幂等、增量）。
//
// 执行顺序建议：
//  1. 本函数（写 auth_permission / auth_role / auth_user / auth_user_role）；
//  2. module_manager.EnsureModuleSeedData（写 module_manager_module）；
//  3. menu.EnsureMenuSeedData（写 platform_menu）；
//  4. EnsureAdminRoleMenus（写 auth_role_menu，依赖 platform_menu 已有数据）；
//  5. route_perm.EnsureSeedData（写 platform_route_permission）。
//
// 注意（方案 A）：
//   本函数只负责 auth_permission / auth_role / auth_user 种子。
//   admin 角色的 auth_role_menu 关联由 EnsureAdminRoleMenus 负责，
//   必须在 menu.EnsureMenuSeedData 之后调用，以保证 platform_menu 已有数据。
//
// 幂等语义：
//   - 权限码：按 code 判重，已存在跳过；
//   - 角色：按 code 判重，已存在则补齐缺失权限（老角色升级）；
//   - 用户：按 username 判重，已存在跳过；
//   - 用户-角色：按 (user_id, role_id) 判重。
func EnsureAuthSeedData(db *gorm.DB) error {
	permissionsByCode := make(map[string]*models.Permission)

	// ---- 1. 写权限码（幂等、增量）----
	for _, item := range DefaultPermissions {
		var perm models.Permission
		err := db.Where("code = ?", item.Code).First(&perm).Error
		if err == gorm.ErrRecordNotFound {
			perm = models.Permission{
				Code:     item.Code,
				Name:     item.Name,
				ModuleID: item.ModuleID,
				Resource: item.Resource,
				Action:   item.Action,
			}
			if err := db.Create(&perm).Error; err != nil {
				return err
			}
			log.Info().
				Str("code", item.Code).
				Str("module", item.ModuleID).
				Msg("权限码已新增")
		} else if err != nil {
			return err
		}
		permissionsByCode[item.Code] = &perm
	}

	// ---- 2. 写 admin 角色（system role，拥有全部权限）----
	adminRole, err := upsertRole(
		db, "admin", "管理员", "系统内置管理员，拥有全部权限",
		true, allPermissions(permissionsByCode),
	)
	if err != nil {
		return err
	}

	// ---- 3. 写 guest 角色（system role，只读权限）----
	guestPerms := filterPermissions(permissionsByCode, guestPermissionCodes)
	guestRole, err := upsertRole(
		db, "guest", "访客", "系统内置访客，只读权限",
		true, guestPerms,
	)
	if err != nil {
		return err
	}

	// ---- 4. 写 admin / guest 用户（默认密码 123456）----
	adminUser, err := upsertUser(db, "admin", "管理员", "admin@example.com")
	if err != nil {
		return err
	}
	guestUser, err := upsertUser(db, "guest", "访客", "guest@example.com")
	if err != nil {
		return err
	}

	// ---- 5. 建立用户-角色关联 ----
	if err := ensureUserRole(db, adminUser.ID, adminRole.ID); err != nil {
		return err
	}
	if err := ensureUserRole(db, guestUser.ID, guestRole.ID); err != nil {
		return err
	}

	log.Info().
		Int("total_permissions", len(DefaultPermissions)).
		Msg("认证种子数据初始化完成（幂等、增量）")
	return nil
}

// EnsureAdminRoleMenus 确保 admin 角色的 auth_role_menu 覆盖所有菜单。
//
// 调用时机：必须在 menu.EnsureMenuSeedData 之后（依赖 platform_menu 已有数据）。
//
// 语义（v3）：
//   - admin 是超管，拥有所有菜单（M/C/F，且**不过滤 status**，
//     与 rbac.LoadUserPermissions 的 admin 分支保持一致）；
//   - 每次启动幂等同步：新增菜单自动追加，已存在的不重复插入；
//   - 只追加不删除，避免误删管理员手动调整的行。
//
// 注意：
//   本函数**不是关键路径**——admin 的权限由 rbac.LoadUserPermissions 的
//   admin 分支直接返回全量 perms，不依赖 auth_role_menu。保留本函数有两个作用：
//     1. 兜底：若以后调整 admin 分支逻辑，auth_role_menu 里已有完整数据；
//     2. 可视化：管理后台的"角色菜单"界面能正确显示 admin 勾选的菜单。
func EnsureAdminRoleMenus(db *gorm.DB) error {
	const maxRetries = 5

	var adminRole models.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// admin 角色还没创建（异常情况），交给 EnsureAuthSeedData 处理
			return nil
		}
		return err
	}

	// 1. 所有菜单的 id（M/C/F 全部，★ 不再过滤 status）
	var menuIDs []uint
	if err := db.Model(&models.Menu{}).
		Pluck("id", &menuIDs).Error; err != nil {
		return err
	}
	if len(menuIDs) == 0 {
		// platform_menu 还未 seed，跳过（本次启动 menu seed 会在之后调用）
		return nil
	}

	// 2. admin 角色已有的关联
	var existing []uint
	if err := db.Model(&models.AuthRoleMenu{}).
		Where("role_id = ?", adminRole.ID).
		Pluck("menu_id", &existing).Error; err != nil {
		return err
	}
	existSet := make(map[uint]struct{}, len(existing))
	for _, id := range existing {
		existSet[id] = struct{}{}
	}

	// 3. 只插入缺失的
	toAdd := make([]models.AuthRoleMenu, 0, len(menuIDs))
	for _, mid := range menuIDs {
		if _, ok := existSet[mid]; ok {
			continue
		}
		toAdd = append(toAdd, models.AuthRoleMenu{RoleID: adminRole.ID, MenuID: mid})
	}
	if len(toAdd) == 0 {
		return nil
	}

	// 4. 批量插入（SQLite 下可能与其它写操作并发，加锁冲突重试）
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := db.CreateInBatches(toAdd, 100).Error; err != nil {
			lastErr = err
			if isDatabaseLocked(err) {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
				continue
			}
			return err
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return lastErr
	}

	log.Info().
		Uint("role_id", adminRole.ID).
		Int("added", len(toAdd)).
		Int("total_menus", len(menuIDs)).
		Msg("admin 角色菜单关联已同步")
	return nil
}

// ============================================================================
// 内部工具
// ============================================================================

// upsertRole 幂等地创建/更新角色。
//
//   - 角色不存在 → 创建，并关联全部权限（admin）或指定权限（guest）；
//   - 角色已存在 → 补齐缺失权限（老角色新增权限码后自动升级）。
func upsertRole(
	db *gorm.DB, code, name, description string,
	isSystem bool, perms []*models.Permission,
) (*models.Role, error) {
	var role models.Role
	err := db.Preload("Permissions").Where("code = ?", code).First(&role).Error

	if err == gorm.ErrRecordNotFound {
		role = models.Role{
			Name:        name,
			Code:        code,
			Description: &description,
			IsSystem:    boolToInt8(isSystem),
		}
		if err := db.Create(&role).Error; err != nil {
			return nil, err
		}
		if len(perms) > 0 {
			if err := db.Model(&role).Association("Permissions").Append(perms); err != nil {
				return nil, err
			}
		}
		log.Info().
			Str("role", code).
			Int("perms", len(perms)).
			Msg("角色已创建")
		return &role, nil
	} else if err != nil {
		return nil, err
	}

	// 已存在：补齐缺失权限（增量升级）
	existing := make(map[string]struct{}, len(role.Permissions))
	for _, p := range role.Permissions {
		existing[p.Code] = struct{}{}
	}
	var toAdd []*models.Permission
	for _, p := range perms {
		if _, ok := existing[p.Code]; !ok {
			toAdd = append(toAdd, p)
		}
	}
	if len(toAdd) > 0 {
		if err := db.Model(&role).Association("Permissions").Append(toAdd); err != nil {
			return nil, err
		}
		log.Info().
			Str("role", code).
			Int("added", len(toAdd)).
			Msg("角色补充新权限")
	}
	return &role, nil
}

// upsertUser 幂等地创建用户（默认密码 123456）。
func upsertUser(db *gorm.DB, username, nickname, email string) (*models.User, error) {
	var user models.User
	err := db.Where("username = ?", username).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		hashed, err := security.HashPassword("123456")
		if err != nil {
			return nil, err
		}
		user = models.User{
			Username:     username,
			PasswordHash: hashed,
			Nickname:     nickname,
			Email:        &email,
			Status:       1,
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
		log.Info().
			Str("username", username).
			Msg("初始用户已创建（默认密码 123456，请尽快修改）")
		return &user, nil
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

// ensureUserRole 幂等地建立用户-角色关联。
func ensureUserRole(db *gorm.DB, userID, roleID uint) error {
	var count int64
	db.Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count)
	if count == 0 {
		return db.Create(&models.UserRole{UserID: userID, RoleID: roleID}).Error
	}
	return nil
}

// allPermissions 返回 map 中全部权限（用于 admin 角色）。
func allPermissions(m map[string]*models.Permission) []*models.Permission {
	result := make([]*models.Permission, 0, len(m))
	for _, p := range m {
		result = append(result, p)
	}
	return result
}

// filterPermissions 按 codes 集合过滤权限（用于 guest 角色）。
func filterPermissions(m map[string]*models.Permission, codes map[string]struct{}) []*models.Permission {
	result := make([]*models.Permission, 0)
	for code := range codes {
		if p, ok := m[code]; ok {
			result = append(result, p)
		}
	}
	return result
}

// boolToInt8 bool → int8（GORM 兼容）。
func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// isDatabaseLocked 判断错误是否为 SQLite 锁冲突。
func isDatabaseLocked(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database table is locked")
}
