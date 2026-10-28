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
	"backend-go/internal/modules/menu"
	"backend-go/internal/modules/module_manager"
	"backend-go/internal/modules/route_perm"
	"backend-go/internal/modules/ssl_proxy"
	"backend-go/internal/modules/tcps_proxy"
)
func main() {
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
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("数据库初始化失败")
	}
	// AutoMigrate
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
		// ★ 新增：TCPS 代理表迁移
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
	// 种子数据
	if err := auth.EnsureAuthSeedData(db); err != nil {
		log.Error().Err(err).Msg("认证种子数据失败")
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
	// 模块管理
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
	// 业务服务
	auditLogSvc := audit_log.NewService(db, cfg)
	licenseSvc := license.NewService(db, cfg)
	cryptoSvc := crypto.NewService(cfg, auditLogSvc, db)
	menuSvc := menu.NewService(db)
	routePermSvc := route_perm.NewService(db)
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
	authSvc := auth.NewService(db, cfg)
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterPublicRoutes(api)
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
	// ---- SSL 代理（HTTP）----
	cryptoMeta := cryptoHandler.Meta()
	certProvider := ssl_proxy.NewCryptoCertProvider(cryptoMeta.CA, cryptoMeta.Cert)
	certResolver := ssl_proxy.NewCertResolver(certProvider)
	sslProxySvc := ssl_proxy.NewService(db, certResolver)
	sslProxyHandler := ssl_proxy.NewHandler(sslProxySvc)
	sslProxyHandler.Register(protected)
	// ---- ★ TCPS 代理（4 层）----
	tcpsCertProvider := tcps_proxy.NewCryptoCertProvider(cryptoMeta.CA, cryptoMeta.Cert)
	tcpsCertResolver := tcps_proxy.NewCertResolver(tcpsCertProvider)
	tcpsProxySvc := tcps_proxy.NewService(db, tcpsCertResolver)
	tcpsProxyHandler := tcps_proxy.NewHandler(tcpsProxySvc)
	tcpsProxyHandler.Register(protected)
	// 菜单管理与路由权限
	menuHandler := menu.NewHandler(menuSvc)
	menuHandler.RegisterRoutes(protected)
	routePermHandler := route_perm.NewHandler(routePermSvc)
	routePermHandler.RegisterRoutes(protected)
	// 前端静态资源
	frontendHandler := frontend.NewHandler(cfg)
	if frontendHandler.Enabled() {
		frontendHandler.Register(r)
	} else {
		r.NoRoute(frontend.HandleDisabled)
	}
	startupResult := licenseSvc.VerifyLicenseOnStartup()
	if ok, _ := startupResult["ok"].(bool); !ok {
		log.Warn().Interface("result", startupResult).Msg("License 校验未通过")
	}
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
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("收到关闭信号")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP 服务关闭失败")
	}
	auditLogSvc.Wait()
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("数据库关闭失败")
	}
	log.Info().Msg("服务已退出")
}
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
