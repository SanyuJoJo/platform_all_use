// backend-go/cmd/server/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/exception"
	"backend-go/internal/frontend"
	"backend-go/internal/health"
	"backend-go/internal/logger"
	"backend-go/internal/menus"
	"backend-go/internal/middleware"
	"backend-go/internal/models"
	"backend-go/internal/modules/audit_log"
	"backend-go/internal/modules/auth"
	"backend-go/internal/modules/crypto"
	"backend-go/internal/modules/license"
	"backend-go/internal/modules/live"
	"backend-go/internal/modules/media"
	"backend-go/internal/modules/menu"
	"backend-go/internal/modules/module_manager"
	"backend-go/internal/modules/route_perm"
	"backend-go/internal/modules/ssl_proxy"
	"backend-go/internal/modules/tcps_proxy"
)

func main() {
	// ========================================================================
	// 1. 加载配置 + 初始化日志
	// ========================================================================
	cfg, warnings, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger.Setup(cfg.LogLevel)
	for _, w := range warnings {
		if w.Level == "error" {
			log.Error().Msg(w.Message)
		} else {
			log.Warn().Msg(w.Message)
		}
	}

	// ========================================================================
	// 2. 初始化数据库
	// ========================================================================
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("数据库初始化失败")
	}

	// ========================================================================
	// 3. AutoMigrate（生产环境跳过，走 goose）
	// ========================================================================
	if cfg.AppEnv != "production" {
		if err := models.AutoMigrate(db); err != nil {
			log.Fatal().Err(err).Msg("AutoMigrate 失败")
		}
		if err := models.EnsureModuleTables(db); err != nil {
			log.Fatal().Err(err).Msg("模块管理表迁移失败")
		}
		if err := models.EnsureAuditLogTable(db); err != nil {
			log.Fatal().Err(err).Msg("日志审计表迁移失败")
		}
		if err := models.EnsureLicenseTable(db); err != nil {
			log.Fatal().Err(err).Msg("License 表迁移失败")
		}
		if err := models.EnsureAuthRoleMenuTable(db); err != nil {
			log.Fatal().Err(err).Msg("auth_role_menu 表迁移失败")
		}
		if err := ssl_proxy.Migrate(db); err != nil {
			log.Fatal().Err(err).Msg("SSL 代理表迁移失败")
		}
		if err := tcps_proxy.Migrate(db); err != nil {
			log.Fatal().Err(err).Msg("TCPS 代理表迁移失败")
		}
		if err := models.EnsureMenuSchema(db); err != nil {
			log.Fatal().Err(err).Msg("菜单表迁移失败")
		}
		if err := models.EnsureRoutePermissionSchema(db); err != nil {
			log.Fatal().Err(err).Msg("路由权限表迁移失败")
		}
	} else {
		log.Info().Msg("生产环境跳过 AutoMigrate，请使用 make migrate-up")
	}

	// ========================================================================
	// 3.1 生产环境也执行的表结构兜底（纯增量 AutoMigrate）
	//
	// 背景：生产环境走 goose 建表并跳过 AutoMigrate。若某张表只由
	// GORM AutoMigrate 创建（例如 platform_media / media_live_room /
	// platform_ssl_proxy / platform_tcps_proxy / 证书元数据各表），
	// 或 goose 迁移的 DDL 与 model 不一致（例如缺 name / micro_app 列），
	// 种子会失败 / 接口会报 "no such table"。AutoMigrate 对已存在的列
	// 与索引只跳过、不删除，属于纯增量操作，因此这里在两种环境下都执行。
	// ========================================================================
	if err := models.EnsureMenuSchema(db); err != nil {
		log.Fatal().Err(err).Msg("菜单表结构校验失败")
	}
	if err := models.EnsureCryptoMetaTables(db); err != nil {
		log.Fatal().Err(err).Msg("证书元数据表迁移失败")
	}
	if err := ssl_proxy.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("SSL 代理表迁移失败")
	}
	if err := tcps_proxy.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("TCPS 代理表迁移失败")
	}
	// 音视频：媒体库表迁移
	if err := media.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("媒体表迁移失败")
	}
	// 音视频：直播间表迁移
	if err := live.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("直播间表迁移失败")
	}

	// ========================================================================
	// 4. 种子数据
	//
	// 顺序约束（自上而下）：
	//   1. auth         ：用户/角色/权限基础种子
	//   2. media        ：音视频权限码（media:media:* / media:live:*），
	//                     必须先于 menu seed，保证 media 菜单的 perms 有归属
	//   3. module_manager：模块表种子（含 media 模块注册，
	//                     提供 entry_frontend=/sub-apps/media/，供 qiankun 加载）
	//   4. menu         ：platform_menu 菜单树（含音视频管理节点）
	//   5. auth         ：admin 角色菜单授权同步
	//   6. route_perm   ：路由权限映射种子
	// ========================================================================
	if err := auth.EnsureAuthSeedData(db); err != nil {
		log.Error().Err(err).Msg("认证种子数据失败")
	}
	if err := media.EnsureSeedData(db); err != nil {
		log.Error().Err(err).Msg("媒体模块种子失败")
	}
	if err := module_manager.EnsureModuleSeedData(db); err != nil {
		log.Error().Err(err).Msg("模块管理种子失败")
	}
	if err := menu.EnsureMenuSeedData(db); err != nil {
		log.Error().Err(err).Msg("菜单种子失败")
	}
	if err := auth.EnsureAdminRoleMenus(db); err != nil {
		log.Error().Err(err).Msg("admin 角色菜单同步失败")
	}
	if err := route_perm.EnsureSeedData(db); err != nil {
		log.Error().Err(err).Msg("路由权限种子失败")
	}

	// ========================================================================
	// 5. 模块管理服务
	// ========================================================================
	moduleSvc := module_manager.NewService(db, cfg)
	moduleSvc.CleanupResidue()
	moduleLoader := module_manager.NewLoader()
	moduleSvc.SetLoader(moduleLoader)
	frontendCfg := loadFrontendConfigOrNil(cfg.FrontendConfigPath)
	moduleSvc.SetFrontendConfig(frontendCfg)
	if order, err := moduleSvc.LoadActiveModules(db); err != nil {
		log.Warn().Err(err).Msg("加载 active 模块失败")
	} else {
		log.Info().Strs("loaded", order).Msg("已加载 active 模块")
	}

	// ========================================================================
	// 6. 业务服务（非路由依赖）
	// ========================================================================
	auditLogSvc := audit_log.NewService(db, cfg)
	licenseSvc := license.NewService(db, cfg)
	cryptoSvc := crypto.NewService(cfg, auditLogSvc, db)
	menuSvc := menu.NewService(db)
	routePermSvc := route_perm.NewService(db)

	// ========================================================================
	// 7. Gin 引擎与全局中间件
	// ========================================================================
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.SetupCORS(cfg.CORSOrigins))
	r.Use(audit_log.AuditLogMiddleware(auditLogSvc))
	r.Use(exception.Recovery())
	r.NoMethod(exception.NoMethod)

	r.GET("/health", health.Handler)

	api := r.Group("/api/v1")

	// ---- 认证（公开路由） ----
	authSvc := auth.NewService(db, cfg)
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterPublicRoutes(api)

	// ---- 受保护路由组 ----
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(db, cfg.SecretKey))
	protected.Use(middleware.DynamicPermission(routePermSvc))

	authHandler.RegisterProtectedRoutes(protected)
	authHandler.RegisterUserRoutes(protected)
	authHandler.RegisterRoleRoutes(protected)
	authHandler.RegisterPermissionRoutes(protected)

	moduleHandler := module_manager.NewHandler(moduleSvc)
	moduleHandler.RegisterRoutes(protected)

	auditLogHandler := audit_log.NewHandler(auditLogSvc)
	auditLogHandler.RegisterRoutes(protected)

	licenseHandler := license.NewHandler(licenseSvc)
	licenseHandler.RegisterRoutes(protected)

	cryptoHandler := crypto.NewHandler(cryptoSvc)
	cryptoHandler.RegisterRoutes(protected)

	// ---- SSL 代理（HTTP） ----
	cryptoMeta := cryptoHandler.Meta()
	certProvider := ssl_proxy.NewCryptoCertProvider(cryptoMeta.CA, cryptoMeta.Cert)
	certResolver := ssl_proxy.NewCertResolver(certProvider)
	sslProxySvc := ssl_proxy.NewService(db, certResolver)
	sslProxyHandler := ssl_proxy.NewHandler(sslProxySvc)
	sslProxyHandler.Register(protected)

	// ---- TCPS 代理（4 层） ----
	tcpsCertProvider := tcps_proxy.NewCryptoCertProvider(cryptoMeta.CA, cryptoMeta.Cert)
	tcpsCertResolver := tcps_proxy.NewCertResolver(tcpsCertProvider)
	tcpsProxySvc := tcps_proxy.NewService(db, tcpsCertResolver)
	tcpsProxyHandler := tcps_proxy.NewHandler(tcpsProxySvc)
	tcpsProxyHandler.Register(protected)

	// ========================================================================
	// 8. 音视频：媒体库 + 直播
	//
	// 说明：
	//   - mediaHandler.RegisterRoutes 需要同时拿到 protected 组（正常业务）
	//     与 api 组（download / raw 直链接口，内部手工校验 Bearer 或 ?token=）。
	//   - live 依赖 ZLM 引擎；EngineManager 统一封装：
	//       · 启动时 EnsureRunning：探活失败 + ZLM_AUTO_START=true → 托管拉起
	//       · 退出时 Shutdown：按 ZLM_ENGINE_SHUTDOWN_ON_EXIT 决定是否 kill
	//       · 运行时通过 /live/engine/{status,start,stop} 提供状态与手动启停
	//   - LiveService 增加 db 参数，用于直播间（media_live_room）持久化：
	//       · POST   /live/rooms    创建直播间（同步触发 proxy 拉流）
	//       · GET    /live/rooms    列表（DB + ZLM 实时状态合并）
	//       · DELETE /live/rooms/:id 删除（踢流 + 清理代理）
	// ========================================================================
	mediaSvc := media.NewService(db, cfg)
	mediaHandler := media.NewHandler(mediaSvc, cfg)
	mediaHandler.RegisterRoutes(protected, api)

	liveEngine := live.NewEngineManager(cfg)
	liveSvc := live.NewService(db, cfg, liveEngine)
	liveHandler := live.NewHandler(liveSvc)
	liveHandler.RegisterRoutes(protected)

	// 启动时探活：未运行 + ZLM_AUTO_START=true → 托管拉起（异步探活，不阻塞启动）
	liveEngine.EnsureRunning(context.Background())

	// ========================================================================
	// 9. 菜单管理与路由权限
	// ========================================================================
	menuHandler := menu.NewHandler(menuSvc)
	menuHandler.RegisterRoutes(protected)

	routePermHandler := route_perm.NewHandler(routePermSvc)
	routePermHandler.RegisterRoutes(protected)

	// ========================================================================
	// 10. 前端静态资源
	// ========================================================================
	frontendHandler := frontend.NewHandler(cfg)
	if frontendHandler.Enabled() {
		frontendHandler.Register(r)
	} else {
		r.NoRoute(frontend.HandleDisabled)
	}

	// ========================================================================
	// 11. License 启动校验（不阻断启动）
	// ========================================================================
	startupResult := licenseSvc.VerifyLicenseOnStartup()
	if ok, _ := startupResult["ok"].(bool); !ok {
		log.Warn().Interface("result", startupResult).Msg("License 校验未通过")
	}

	// ========================================================================
	// 12. HTTP 服务
	// ========================================================================
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("HTTP 服务已启动")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP 服务启动失败")
		}
	}()

	// ========================================================================
	// 13. 优雅退出
	// ========================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("收到关闭信号")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP 服务关闭失败")
	}

	// 按 ZLM_ENGINE_SHUTDOWN_ON_EXIT 决定是否 kill 托管的 ZLM
	// （未托管或已由 systemd 管理时不动作，避免误杀）
	liveEngine.Shutdown()

	auditLogSvc.Wait()
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("数据库关闭失败")
	}
	log.Info().Msg("服务已退出")
}

// loadFrontendConfigOrNil 加载前端动态菜单配置；不存在或为空时返回 nil。
//
// 返回 nil 时 module_manager 保留 DB manifest 里的菜单（向后兼容）。
func loadFrontendConfigOrNil(path string) *menus.FrontendConfig {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	cfg, err := menus.Load(path)
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("前端配置加载失败")
	}
	return cfg
}
