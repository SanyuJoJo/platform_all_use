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
	if cfg.AppEnv != "production" {
		if err := models.AutoMigrate(db); err != nil {
			log.Fatal().Err(err).Msg("AutoMigrate 失败，开发环境无法继续启动（请检查数据库权限/驱动）")
		}
	} else {
		log.Info().Msg("生产环境跳过 AutoMigrate，请使用 make migrate-up 执行 goose 迁移")
	}
	if err := auth.EnsureAuthSeedData(db); err != nil {
		log.Error().Err(err).Msg("认证种子数据初始化失败")
	}
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLog())
	r.Use(exception.Recovery())
	r.Use(middleware.SetupCORS(cfg.CORSOrigins))
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
