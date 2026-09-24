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
	"backend-go/internal/middleware"
	"backend-go/internal/models"
	"backend-go/internal/modules/audit_log"
	"backend-go/internal/modules/auth"
	"backend-go/internal/modules/crypto"
	"backend-go/internal/modules/license"
	"backend-go/internal/modules/module_manager"
)

func main() {
	// ========================================================================
	// 1. 加载配置
	// ========================================================================
	cfg, warnings, err := config.Load()
	if err != nil {
		panic(err)
	}

	// ========================================================================
	// 2. 初始化日志
	// ========================================================================
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

	// ========================================================================
	// 3. 初始化数据库
	// ========================================================================
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("数据库初始化失败")
	}

	// ------------------------------------------------------------------------
	// AutoMigrate：仅非生产环境执行。
	//
	// 生产环境跳过 AutoMigrate，必须通过 make migrate-up（goose）执行迁移。
	// 非生产环境 AutoMigrate 失败视为致命错误（建议-6），避免表结构缺失时
	// 服务继续启动导致后续业务请求全部失败。
	// ------------------------------------------------------------------------
	if cfg.AppEnv != "production" {
		if err := models.AutoMigrate(db); err != nil {
			log.Fatal().Err(err).Msg("AutoMigrate 失败，开发环境无法继续启动（请检查数据库权限/驱动）")
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
		if err := models.EnsureTaskTable(db); err != nil {
			log.Fatal().Err(err).Msg("密码操作任务表迁移失败")
		}
		if err := models.EnsureCryptoMetaTables(db); err != nil {
			log.Fatal().Err(err).Msg("密码元数据表迁移失败")
		}
	} else {
		log.Info().Msg("生产环境跳过 AutoMigrate，请使用 make migrate-up 执行 goose 迁移")
	}

	// ========================================================================
	// 4. 种子数据初始化
	// ========================================================================
	// 认证种子数据（幂等、增量）
	if err := auth.EnsureAuthSeedData(db); err != nil {
		log.Error().Err(err).Msg("认证种子数据初始化失败")
	}
	// 模块管理种子数据
	if err := module_manager.EnsureModuleSeedData(db); err != nil {
		log.Error().Err(err).Msg("模块管理种子数据初始化失败")
	}

	// ========================================================================
	// 5. 模块管理：清理残留 + 注入 Loader + 加载 active 模块
	// ========================================================================
	moduleSvc := module_manager.NewService(db, cfg)
	moduleSvc.CleanupResidue()

	moduleLoader := module_manager.NewLoader()
	moduleSvc.SetLoader(moduleLoader)
	if order, err := moduleSvc.LoadActiveModules(db); err != nil {
		log.Warn().Err(err).Msg("加载 active 模块失败")
	} else {
		log.Info().Strs("loaded", order).Msg("已加载 active 模块")
	}

	// ========================================================================
	// 6. 各业务服务初始化
	// ========================================================================
	// 日志审计服务
	auditLogSvc := audit_log.NewService(db, cfg)

	// License 管理服务
	licenseSvc := license.NewService(db, cfg)

	// 密码操作服务（依赖 auditLogSvc 写审计）
	cryptoSvc := crypto.NewService(cfg, auditLogSvc,db)

	// 认证服务
	authSvc := auth.NewService(db, cfg)

	// 密码操作健康检查（core 是否可用）
	if err := cryptoSvc.HealthCheck(); err != nil {
		log.Warn().Err(err).Msg("core dispatch.sh 不可用，密码操作将不可用（不影响其他模块）")
	}

	// ========================================================================
	// 7. Gin 引擎初始化
	// ========================================================================
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// ★ 修复 405：默认 Gin 会把"已注册路径 + 错误方法"也当作 404，
	//   开启此选项后才会走 NoMethod（405）。
	r.HandleMethodNotAllowed = true

	// ------------------------------------------------------------------------
	// 中间件注册顺序（日志审计 v1.2 P0-1 修复）：
	//
	//   RequestID → AccessLog → CORS → AuditLog → Recovery → 路由
	//
	// 关键点：
	//   - AuditLogMiddleware 必须位于 Recovery 外层，确保路由 panic 时
	//     AuditLogMiddleware 的 defer 逻辑能读取最终响应状态码与 X-Error-Code；
	//   - Recovery 位于最内层，直接捕获路由 panic 并写入错误响应；
	//   - 结合 AuditLogMiddleware 内部的 defer 兜底，任何路径都能记录。
	// ------------------------------------------------------------------------
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(middleware.SetupCORS(cfg.CORSOrigins))
	r.Use(audit_log.AuditLogMiddleware(auditLogSvc))
	r.Use(exception.Recovery())

	// ------------------------------------------------------------------------
	// NoMethod 始终注册（开发与维护指南 v1.1 P0-1 修复）：
	//
	//   - r.HandleMethodNotAllowed = true 时，对已注册路径使用错误方法，
	//     Gin 会调用 NoMethod 处理器；
	//   - 若不注册，返回默认空响应，审计中间件拿不到 X-Error-Code；
	//   - 前端托管启用/未启用均需保持统一 JSON 405。
	//
	// NoRoute 的处理分两种情况：
	//   - 未启用前端托管：r.NoRoute(exception.NoRoute) 保持 JSON 404；
	//   - 启用前端托管：frontendHandler.Register(r) 内部覆盖 NoRoute，
	//     实现 SPA fallback + 子应用入口 + API 优先。
	// ------------------------------------------------------------------------
	r.NoMethod(exception.NoMethod)

	// ========================================================================
	// 8. 系统路由
	// ========================================================================
	r.GET("/health", health.Handler)

	// ========================================================================
	// 9. API 路由组
	// ========================================================================
	api := r.Group("/api/v1")

	// ------------------------------------------------------------------------
	// 9.1 认证模块（公开路由）
	// ------------------------------------------------------------------------
	authHandler := auth.NewHandler(authSvc)
	authHandler.RegisterPublicRoutes(api)

	// ------------------------------------------------------------------------
	// 9.2 受保护路由（需认证）
	// ------------------------------------------------------------------------
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(db, cfg.SecretKey))

	// 认证模块 - 受保护路由：/logout、/me、/me/password
	authHandler.RegisterProtectedRoutes(protected)

	// 用户管理路由
	authHandler.RegisterUserRoutes(protected)

	// 角色管理路由
	authHandler.RegisterRoleRoutes(protected)

	// 权限管理路由
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

	// 密码操作路由（/crypto/operations/:operation_id、/tasks/:task_id）
	cryptoHandler := crypto.NewHandler(cryptoSvc)
	cryptoHandler.RegisterRoutes(protected)
	metaSvc := crypto.NewMetaService(db,cfg)
	metaHandler := crypto.NewMetaHandler(metaSvc)
	metaHandler.RegisterRoutes(protected)

	// ------------------------------------------------------------------------
	// 9.3 后续业务路由扩展点（示例）
	//
	// 例如 CA/证书/CSR/CRL/密钥的列表查询接口：
	//   caHandler := crypto.NewCAHandler(...)
	//   caHandler.RegisterRoutes(protected)
	// ------------------------------------------------------------------------

	// ========================================================================
	// 10. 前端静态资源自托管（必须在所有 API 路由之后注册）
	//
	// 若未配置 FRONTEND_DEPLOY_DIR 或 main-app/index.html 不存在，
	// 保持后端默认 JSON 404 行为。
	// ========================================================================
	frontendHandler := frontend.NewHandler(cfg)
	if frontendHandler.Enabled() {
		frontendHandler.Register(r)
	} else {
		r.NoRoute(exception.NoRoute)
	}

	// ========================================================================
	// 11. License 启动校验（不阻断启动）
	// ========================================================================
	startupResult := licenseSvc.VerifyLicenseOnStartup()
	if ok, _ := startupResult["ok"].(bool); !ok {
		log.Warn().
			Interface("result", startupResult).
			Msg("License 启动校验未通过（不影响服务启动）")
	}

	// ========================================================================
	// 12. HTTP 服务启动
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
	// 13. 优雅关闭
	// ========================================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("收到关闭信号，开始优雅关闭")

	// 13.1 关闭 HTTP 服务
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP 服务关闭失败")
	}

	// 13.2 等待审计日志异步写入完成
	auditLogSvc.Wait()

	// 13.3 关闭数据库
	if err := database.Close(); err != nil {
		log.Error().Err(err).Msg("数据库关闭失败")
	}

	log.Info().Msg("服务已退出")
}

// schemeOf 提取 DATABASE_URL 的 scheme，用于启动日志。
func schemeOf(url string) string {
	if idx := strings.Index(url, "://"); idx >= 0 {
		return url[:idx]
	}
	return "unknown"
}
