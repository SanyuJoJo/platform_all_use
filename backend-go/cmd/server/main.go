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
	"backend-go/internal/modules/auth"
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
	// ---------------------------------------------------------------------
	// AutoMigrate：仅非生产环境执行。
	//
	// v1.1（P1-03）：生产环境跳过 AutoMigrate，避免修改生产表结构。
	// v1.2（建议-6）：非生产环境 AutoMigrate 失败视为致命错误，
	//   避免表结构缺失时服务继续启动导致后续业务请求全部失败。
	// ---------------------------------------------------------------------
	if cfg.AppEnv != "production" {
		if err := models.AutoMigrate(db); err != nil {
			log.Fatal().Err(err).Msg("AutoMigrate 失败，开发环境无法继续启动（请检查数据库权限/驱动）")
		}
	} else {
		log.Info().Msg("生产环境跳过 AutoMigrate，请使用 make migrate-up 执行 goose 迁移")
	}
	// 认证种子数据（幂等、增量）
	if err := auth.EnsureAuthSeedData(db); err != nil {
		log.Error().Err(err).Msg("认证种子数据初始化失败")
	}
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	// 中间件顺序（与 Python 版语义对齐）：
	//   RequestID → AccessLog → Recovery → CORS → 业务路由
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(exception.Recovery())
	r.Use(middleware.SetupCORS(cfg.CORSOrigins))
	r.NoRoute(exception.NoRoute)
	r.NoMethod(exception.NoMethod)
	r.GET("/health", health.Handler)
	// ---------------------------------------------------------------------
	// 认证模块路由注册（v1.1 P0-01：拆分公开/受保护分组）
	//
	// 路径拼接说明：
	//   - `api := r.Group("/api/v1")`，路由注册时为 `/api/v1` 前缀；
	//   - `authHandler.RegisterPublicRoutes(api)` 内部使用 `api.Group("/auth")`，
	//     拼出 `/api/v1/auth/...`；
	//   - `protected := api.Group("")` 不改变前缀，仍为 `/api/v1`；
	//     内部 `protected.Group("/auth")` 同样拼出 `/api/v1/auth/...`；
	//   - 因此**不需要**在受保护分组中重复书写 `/auth`。
	// ---------------------------------------------------------------------
	api := r.Group("/api/v1")
	authSvc := auth.NewService(db, cfg)
	authHandler := auth.NewHandler(authSvc)
	// 公开路由（无需认证）：/login、/refresh
	authHandler.RegisterPublicRoutes(api)
	// 受保护路由（需认证）：/logout、/me、/me/password
	// 必须先挂载 AuthMiddleware，否则 /me、/me/password 会因为
	// 缺少 `user_info`/`user_id` 而返回 401。
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(db, cfg.SecretKey))
	authHandler.RegisterProtectedRoutes(protected)
	// ---------------------------------------------------------------------
	// 后续业务路由示例（P2/P3 阶段补充）：
	//
	// users := protected.Group("/auth/users")
	// users.Use(middleware.RequirePermission("auth:user:view"))
	// users.GET("", userHandler.List)
	// ...
	// ---------------------------------------------------------------------
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
