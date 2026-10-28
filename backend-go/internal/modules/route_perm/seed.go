package route_perm

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/models"
)

// seedRoutePermissions 路由权限种子。
//
// 约定：
//   - /api/v1/menus/routers 故意不配置，所有登录用户可访问；
//   - pattern 必须与 Gin c.FullPath() 完全一致（含 /api/v1 前缀和 :param）；
//   - 同一 (method, pattern) 可配多个 code，语义为"或"（任一命中即放行）；
//   - 未配置映射的路由 → DynamicPermission 放行（兼容模式），
//     所以所有需要保护的写接口都必须在这里显式列出。
//
// 覆盖范围：
//   - 菜单管理、路由权限管理（动态路由自身）
//   - 用户 / 角色 / 模块 / 审计 / License
//   - SSL 代理 HTTP（ssl_proxies）
//   - SSL 代理 TCPS（tcps-proxies，4 层 TCP/TLS）  ← 本次新增
//   - Crypto 控制台（CA / Cert / CSR / CRL / Key / Task）
var seedRoutePermissions = []models.RoutePermission{
	// =====================================================================
	// 菜单管理
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/menus", PermissionCode: "auth:menu:list", ModuleID: "auth", Description: "菜单列表"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/menus/tree", PermissionCode: "auth:menu:list", ModuleID: "auth", Description: "菜单树"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/menus/micro-apps", PermissionCode: "auth:menu:list", ModuleID: "auth", Description: "已知微应用列表"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/menus/detail/:id", PermissionCode: "auth:menu:list", ModuleID: "auth", Description: "菜单详情"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/menus", PermissionCode: "auth:menu:create", ModuleID: "auth", Description: "新增菜单"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/menus/detail/:id", PermissionCode: "auth:menu:update", ModuleID: "auth", Description: "修改菜单"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/menus/detail/:id", PermissionCode: "auth:menu:delete", ModuleID: "auth", Description: "删除菜单"},

	// =====================================================================
	// 路由权限管理
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/route-permissions", PermissionCode: "auth:route_perm:list", ModuleID: "auth", Description: "路由权限列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/route-permissions", PermissionCode: "auth:route_perm:create", ModuleID: "auth", Description: "新增路由权限"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/route-permissions/:id", PermissionCode: "auth:route_perm:update", ModuleID: "auth", Description: "修改路由权限"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/route-permissions/:id", PermissionCode: "auth:route_perm:delete", ModuleID: "auth", Description: "删除路由权限"},

	// =====================================================================
	// 用户管理
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/auth/users", PermissionCode: "auth:user:view", ModuleID: "auth", Description: "用户列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/auth/users", PermissionCode: "auth:user:create", ModuleID: "auth", Description: "创建用户"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/auth/users/:user_id", PermissionCode: "auth:user:view", ModuleID: "auth", Description: "用户详情"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/auth/users/:user_id", PermissionCode: "auth:user:edit", ModuleID: "auth", Description: "更新用户"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/auth/users/:user_id", PermissionCode: "auth:user:delete", ModuleID: "auth", Description: "删除用户"},
	{HTTPMethod: "PATCH", RoutePattern: "/api/v1/auth/users/:user_id/status", PermissionCode: "auth:user:edit", ModuleID: "auth", Description: "启用/禁用用户"},
	{HTTPMethod: "PATCH", RoutePattern: "/api/v1/auth/users/:user_id/password", PermissionCode: "auth:user:edit", ModuleID: "auth", Description: "重置用户密码"},

	// =====================================================================
	// 角色管理
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/auth/roles", PermissionCode: "auth:role:list", ModuleID: "auth", Description: "角色列表"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/auth/roles/:id", PermissionCode: "auth:role:list", ModuleID: "auth", Description: "角色详情"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/auth/roles", PermissionCode: "auth:role:create", ModuleID: "auth", Description: "新增角色"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/auth/roles/:id", PermissionCode: "auth:role:edit", ModuleID: "auth", Description: "修改角色"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/auth/roles/:id", PermissionCode: "auth:role:delete", ModuleID: "auth", Description: "删除角色"},

	// =====================================================================
	// 模块管理
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/modules", PermissionCode: "module_manager:module:view", ModuleID: "module_manager", Description: "模块列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/modules", PermissionCode: "module_manager:module:create", ModuleID: "module_manager", Description: "安装模块"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/modules/upload", PermissionCode: "module_manager:module:create", ModuleID: "module_manager", Description: "上传并安装模块"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/modules/:module_id", PermissionCode: "module_manager:module:delete", ModuleID: "module_manager", Description: "卸载模块"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/modules/:module_id/enable", PermissionCode: "module_manager:module:edit", ModuleID: "module_manager", Description: "启用模块"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/modules/:module_id/disable", PermissionCode: "module_manager:module:edit", ModuleID: "module_manager", Description: "禁用模块"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/modules/:module_id/upgrade", PermissionCode: "module_manager:module:edit", ModuleID: "module_manager", Description: "升级模块"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/modules/:module_id/config", PermissionCode: "module_manager:module:view", ModuleID: "module_manager", Description: "查看模块配置"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/modules/:module_id/config", PermissionCode: "module_manager:module:edit", ModuleID: "module_manager", Description: "更新模块配置"},

	// =====================================================================
	// 日志审计
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/audit-logs", PermissionCode: "audit_log:log:view", ModuleID: "audit_log", Description: "日志列表"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/audit-logs/export", PermissionCode: "audit_log:log:export", ModuleID: "audit_log", Description: "导出日志"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/audit-logs/:log_id", PermissionCode: "audit_log:log:view", ModuleID: "audit_log", Description: "日志详情"},

	// =====================================================================
	// License 管理
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/license/status", PermissionCode: "license:license:view", ModuleID: "license", Description: "License 状态"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/license/modules", PermissionCode: "license:license:view", ModuleID: "license", Description: "授权模块列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/license/import", PermissionCode: "license:license:create", ModuleID: "license", Description: "导入 License"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/license/activate", PermissionCode: "license:license:create", ModuleID: "license", Description: "激活 License"},

	// =====================================================================
	// SSL 代理 — HTTP 管理（4 层 ssl-proxies 之外的 7 层 HTTP 反代）
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/ssl-proxies", PermissionCode: "ssl_proxy:proxy:view", ModuleID: "ssl_proxy", Description: "代理列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies", PermissionCode: "ssl_proxy:proxy:create", ModuleID: "ssl_proxy", Description: "新建代理"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/ssl-proxies/:id", PermissionCode: "ssl_proxy:proxy:view", ModuleID: "ssl_proxy", Description: "代理详情"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/ssl-proxies/:id", PermissionCode: "ssl_proxy:proxy:update", ModuleID: "ssl_proxy", Description: "编辑代理"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/ssl-proxies/:id", PermissionCode: "ssl_proxy:proxy:delete", ModuleID: "ssl_proxy", Description: "删除代理"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies/:id/start", PermissionCode: "ssl_proxy:proxy:start", ModuleID: "ssl_proxy", Description: "启动代理"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies/:id/stop", PermissionCode: "ssl_proxy:proxy:stop", ModuleID: "ssl_proxy", Description: "停止代理"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies/:id/restart", PermissionCode: "ssl_proxy:proxy:start", ModuleID: "ssl_proxy", Description: "重启代理（复用启动权限）"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/ssl-proxies/:id/config", PermissionCode: "ssl_proxy:proxy:config", ModuleID: "ssl_proxy", Description: "查看代理配置"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/ssl-proxies/:id/logs", PermissionCode: "ssl_proxy:proxy:logs", ModuleID: "ssl_proxy", Description: "查看代理日志"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies/batch/start", PermissionCode: "ssl_proxy:proxy:start", ModuleID: "ssl_proxy", Description: "批量启动"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies/batch/stop", PermissionCode: "ssl_proxy:proxy:stop", ModuleID: "ssl_proxy", Description: "批量停止"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/ssl-proxies/check-port", PermissionCode: "ssl_proxy:proxy:view", ModuleID: "ssl_proxy", Description: "端口占用检查"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/ssl-proxies/available-certs", PermissionCode: "ssl_proxy:proxy:view", ModuleID: "ssl_proxy", Description: "可绑定证书列表"},

	// =====================================================================
	// ★ SSL 代理 — TCPS 管理（4 层 TCP/TLS 代理）
	//
	// 与 HTTP 管理完全隔离：
	//   - API 前缀 /api/v1/tcps-proxies
	//   - 权限码 ssl_proxy:tcps:*
	//   - 数据表 platform_tcps_proxy
	// =====================================================================
	{HTTPMethod: "GET", RoutePattern: "/api/v1/tcps-proxies", PermissionCode: "ssl_proxy:tcps:view", ModuleID: "ssl_proxy", Description: "TCPS 列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies", PermissionCode: "ssl_proxy:tcps:create", ModuleID: "ssl_proxy", Description: "新建 TCPS"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/tcps-proxies/:id", PermissionCode: "ssl_proxy:tcps:view", ModuleID: "ssl_proxy", Description: "TCPS 详情"},
	{HTTPMethod: "PUT", RoutePattern: "/api/v1/tcps-proxies/:id", PermissionCode: "ssl_proxy:tcps:update", ModuleID: "ssl_proxy", Description: "编辑 TCPS"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/tcps-proxies/:id", PermissionCode: "ssl_proxy:tcps:delete", ModuleID: "ssl_proxy", Description: "删除 TCPS"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies/:id/start", PermissionCode: "ssl_proxy:tcps:start", ModuleID: "ssl_proxy", Description: "启动 TCPS"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies/:id/stop", PermissionCode: "ssl_proxy:tcps:stop", ModuleID: "ssl_proxy", Description: "停止 TCPS"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies/:id/restart", PermissionCode: "ssl_proxy:tcps:start", ModuleID: "ssl_proxy", Description: "重启 TCPS（复用启动权限）"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/tcps-proxies/:id/config", PermissionCode: "ssl_proxy:tcps:config", ModuleID: "ssl_proxy", Description: "TCPS 配置查看"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/tcps-proxies/:id/logs", PermissionCode: "ssl_proxy:tcps:logs", ModuleID: "ssl_proxy", Description: "TCPS 日志查看"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies/batch/start", PermissionCode: "ssl_proxy:tcps:start", ModuleID: "ssl_proxy", Description: "TCPS 批量启动"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies/batch/stop", PermissionCode: "ssl_proxy:tcps:stop", ModuleID: "ssl_proxy", Description: "TCPS 批量停止"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tcps-proxies/check-port", PermissionCode: "ssl_proxy:tcps:view", ModuleID: "ssl_proxy", Description: "TCPS 端口检测"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/tcps-proxies/available-certs", PermissionCode: "ssl_proxy:tcps:view", ModuleID: "ssl_proxy", Description: "TCPS 可绑定证书"},

	// =====================================================================
	// Crypto 控制台
	// =====================================================================

	// CA
	{HTTPMethod: "GET", RoutePattern: "/api/v1/cas", PermissionCode: "crypto_console:ca:view", ModuleID: "crypto", Description: "CA 列表"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/cas/domains", PermissionCode: "crypto_console:ca:view", ModuleID: "crypto", Description: "CA 域列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/cas/domains", PermissionCode: "crypto_console:ca:create", ModuleID: "crypto", Description: "创建 CA 域"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/cas/domains/:name", PermissionCode: "crypto_console:ca:delete", ModuleID: "crypto", Description: "删除 CA 域"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/cas/import", PermissionCode: "crypto_console:ca:import", ModuleID: "crypto", Description: "导入 CA"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/cas/:ca_id", PermissionCode: "crypto_console:ca:view", ModuleID: "crypto", Description: "CA 详情"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/cas/:ca_id/detail", PermissionCode: "crypto_console:ca:view", ModuleID: "crypto", Description: "CA 详细信息"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/cas/:ca_id/export", PermissionCode: "crypto_console:ca:export", ModuleID: "crypto", Description: "导出 CA"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/cas/:ca_id/download", PermissionCode: "crypto_console:ca:export", ModuleID: "crypto", Description: "下载 CA"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/cas/:ca_id", PermissionCode: "crypto_console:ca:delete", ModuleID: "crypto", Description: "删除 CA"},

	// 证书
	{HTTPMethod: "GET", RoutePattern: "/api/v1/certs", PermissionCode: "crypto_console:cert:view", ModuleID: "crypto", Description: "证书列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/certs/sign", PermissionCode: "crypto_console:cert:sign", ModuleID: "crypto", Description: "签发证书"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/certs/import", PermissionCode: "crypto_console:cert:import", ModuleID: "crypto", Description: "导入证书"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/certs/dual/import", PermissionCode: "crypto_console:cert:import", ModuleID: "crypto", Description: "导入双证书"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/certs/envelope/query", PermissionCode: "crypto_console:cert:view", ModuleID: "crypto", Description: "查询信封"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/certs/:cert_id", PermissionCode: "crypto_console:cert:view", ModuleID: "crypto", Description: "证书详情"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/certs/:cert_id/detail", PermissionCode: "crypto_console:cert:view", ModuleID: "crypto", Description: "证书详细信息"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/certs/:cert_id/export", PermissionCode: "crypto_console:cert:export", ModuleID: "crypto", Description: "导出证书"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/certs/:cert_id", PermissionCode: "crypto_console:cert:delete", ModuleID: "crypto", Description: "删除证书"},

	// CSR
	{HTTPMethod: "GET", RoutePattern: "/api/v1/csrs", PermissionCode: "crypto_console:csr:view", ModuleID: "crypto", Description: "CSR 列表"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/csrs/import", PermissionCode: "crypto_console:csr:create", ModuleID: "crypto", Description: "导入 CSR"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/csrs/:csr_id", PermissionCode: "crypto_console:csr:view", ModuleID: "crypto", Description: "CSR 详情"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/csrs/:csr_id/download", PermissionCode: "crypto_console:csr:view", ModuleID: "crypto", Description: "下载 CSR"},
	{HTTPMethod: "GET", RoutePattern: "/api/v1/csrs/:csr_id/key", PermissionCode: "crypto_console:csr:view", ModuleID: "crypto", Description: "下载私钥"},
	{HTTPMethod: "DELETE", RoutePattern: "/api/v1/csrs/:csr_id", PermissionCode: "crypto_console:csr:delete", ModuleID: "crypto", Description: "删除 CSR"},

	// CRL
	{HTTPMethod: "GET", RoutePattern: "/api/v1/crls", PermissionCode: "crypto_console:crl:view", ModuleID: "crypto", Description: "CRL 列表"},

	// 密钥
	{HTTPMethod: "GET", RoutePattern: "/api/v1/keys", PermissionCode: "crypto_console:key:view", ModuleID: "crypto", Description: "密钥列表"},

	// 任务
	{HTTPMethod: "GET", RoutePattern: "/api/v1/tasks/:task_id", PermissionCode: "crypto_console:task:view", ModuleID: "crypto", Description: "任务详情"},
	{HTTPMethod: "POST", RoutePattern: "/api/v1/tasks/:task_id/cancel", PermissionCode: "crypto_console:task:cancel", ModuleID: "crypto", Description: "取消任务"},

	// 密码操作（后端 API）
	{HTTPMethod: "POST", RoutePattern: "/api/v1/crypto/operations/:operation_id", PermissionCode: "crypto:operation:execute", ModuleID: "crypto", Description: "执行密码操作"},
}

// EnsureSeedData 幂等地写入路由权限种子数据。
//
// 在 SQLite 下可能与启动期其他写操作并发（如 EnsureAdminRoleMenus），
// 加锁冲突重试，避免种子静默失败导致后续 fail-open 放行。
func EnsureSeedData(db *gorm.DB) error {
	const maxRetries = 5
	now := time.Now()

	for i := range seedRoutePermissions {
		rp := seedRoutePermissions[i]

		var lastErr error
		for attempt := 0; attempt < maxRetries; attempt++ {
			var exist models.RoutePermission
			err := db.Where("http_method = ? AND route_pattern = ? AND permission_code = ?",
				rp.HTTPMethod, rp.RoutePattern, rp.PermissionCode).First(&exist).Error
			if err == nil {
				lastErr = nil
				break
			}
			if err != gorm.ErrRecordNotFound {
				lastErr = err
				if isDatabaseLocked(err) {
					time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
					continue
				}
				return err
			}

			rp.CreatedAt = now
			rp.UpdatedAt = now
			rp.Status = 0
			if err := db.Create(&rp).Error; err != nil {
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
	}
	log.Info().Msg("路由权限种子数据已就绪")
	return nil
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
