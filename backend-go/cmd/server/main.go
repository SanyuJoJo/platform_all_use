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
	"backend-go/internal/health"
	"backend-go/internal/logger"
	"backend-go/internal/middleware"
	"backend-go/internal/models"
	"backend-go/internal/modules/audit_log"
	"backend-go/internal/modules/auth"
	"backend-go/internal/modules/license"
	"backend-go/internal/modules/module_manager"
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
	log.Info().
		Str("app", cfg.AppName).
		Str("env", cfg.AppEnv).
		Str("db_scheme", schemeOf(cfg.DatabaseURL)).
		Msg("后端服务启动中")
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("数据库初始化失败")
	}
	// AutoMigrate：仅非生产环境执行。
	if cfg.AppEnv != "production" {
		if err := models.AutoMigrate(db); err != nil {
			log.Fatal().Err(err).Msg("AutoMigrate 失败，开发环境无法继续启动")
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
	} else {
		log.Info().Msg("生产环境跳过 AutoMigrate，请使用 make migrate-up 执行 goose 迁移")
	}
	// 认证种子数据
	if err := auth.EnsureAuthSeedData(db); err != nil {
		log.Error().Err(err).Msg("认证种子数据初始化失败")
	}
	// 模块管理种子数据
	if err := module_manager.EnsureModuleSeedData(db); err != nil {
		log.Error().Err(err).Msg("模块管理种子数据初始化失败")
	}
	// 清理模块残留
	moduleSvc := module_manager.NewService(db, cfg)
	moduleSvc.CleanupResidue()
	// Loader 注入
	moduleLoader := module_manager.NewLoader()
	moduleSvc.SetLoader(moduleLoader)
	if order, err := moduleSvc.LoadActiveModules(db); err != nil {
		log.Warn().Err(err).Msg("加载 active 模块失败")
	} else {
		log.Info().Strs("loaded", order).Msg("已加载 active 模块")
	}
	// 日志审计服务
	auditLogSvc := audit_log.NewService(db, cfg)
	// License 管理服务
	licenseSvc := license.NewService(db, cfg)
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
	r.NoRoute(exception.NoRoute)
	r.NoMethod(exception.NoMethod)
	r.GET("/health", health.Handler)
	api := r.Group("/api/v1")
	authSvc := auth.NewService(db, cfg)
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterPublicRoutes(api)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(db, cfg.SecretKey))
	authHandler.RegisterProtectedRoutes(protected)
	authHandler.RegisterUserRoutes(protected)
	authHandler.RegisterRoleRoutes(protected)
	authHandler.RegisterPermissionRoutes(protected)
	// 模块管理路由
	moduleHandler := module_manager.NewHandler(moduleSvc)
	moduleHandler.RegisterRoutes(protected)
	// 日志审计路由
	auditLogHandler := audit_log.NewHandler(auditLogSvc)
	auditLogHandler.RegisterRoutes(protected)
	// License 管理路由
	licenseHandler := license.NewHandler(licenseSvc)
	licenseHandler.RegisterRoutes(protected)
	// License 启动校验（不阻断启动）
	startupResult := licenseSvc.VerifyLicenseOnStartup()
	if ok, _ := startupResult["ok"].(bool); !ok {
		log.Warn().Interface("result", startupResult).Msg("License 启动校验未通过（不影响服务启动）")
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
	log.Info().Msg("收到关闭信号，开始优雅关闭")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP 服务关闭失败")
	}
	// 等待审计日志异步写入完成
	auditLogSvc.Wait()
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("数据库关闭失败")
	}
	log.Info().Msg("服务已退出")
}
func schemeOf(url string) string {
	if idx := strings.Index(url, "://"); idx >= 0 {
		return url[:idx]
	}
	return "unknown"
}
