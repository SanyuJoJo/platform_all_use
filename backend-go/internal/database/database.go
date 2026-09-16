package database
import (
	"context"
	"fmt"
	"strings"
	"time"
	// 纯 Go SQLite 驱动，CGO_ENABLED=0 可用。
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"backend-go/internal/config"
)
var DB *gorm.DB
func Init(cfg *config.Config) (*gorm.DB, error) {
	// GORM 日志设为 Silent，统一由 zerolog 输出
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	dialector, err := buildDialector(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 sql.DB 失败: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	DB = db
	return db, nil
}
func buildDialector(databaseURL string) (gorm.Dialector, error) {
	dsn := strings.TrimSpace(databaseURL)
	switch {
	case strings.HasPrefix(dsn, "sqlite+aiosqlite:///"):
		path := strings.TrimPrefix(dsn, "sqlite+aiosqlite:///")
		return sqlite.Open(buildSQLiteDSN(path)), nil
	case strings.HasPrefix(dsn, "sqlite:///"):
		path := strings.TrimPrefix(dsn, "sqlite:///")
		return sqlite.Open(buildSQLiteDSN(path)), nil
	case strings.HasPrefix(dsn, "postgresql+asyncpg://"):
		dsn = strings.Replace(dsn, "postgresql+asyncpg://", "postgresql://", 1)
		return postgres.Open(dsn), nil
	case strings.HasPrefix(dsn, "postgres://"), strings.HasPrefix(dsn, "postgresql://"):
		return postgres.Open(dsn), nil
	default:
		return nil, fmt.Errorf("不支持的 DATABASE_URL: %s", databaseURL)
	}
}
// buildSQLiteDSN 为 glebarez/sqlite（纯 Go）追加外键 pragma，
// 保证与 Python 版 PRAGMA foreign_keys=ON 行为一致。
func buildSQLiteDSN(path string) string {
	if path == "" {
		path = "./app.db"
	}
	if strings.Contains(path, "?") {
		return path + "&_pragma=foreign_keys(1)"
	}
	return path + "?_pragma=foreign_keys(1)"
}
func HealthCheck() bool {
	if DB == nil {
		return false
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx) == nil
}
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
