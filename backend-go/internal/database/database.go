package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	// 纯 Go SQLite 驱动，CGO_ENABLED=0 可用。
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend-go/internal/config"
)

var DB *gorm.DB

// Init 初始化数据库连接。
//
// SQLite 特殊处理：
//   - 强制 MaxOpenConns=1：SQLite 写是文件级排他锁，多连接并发写会
//     "database is locked"；即使有 busy_timeout，多连接在写密集时依然
//     可能耗尽重试。管理后台并发量低，单连接完全够用。
//   - DSN 追加 busy_timeout(5000) + journal_mode(WAL)：即使外部工具
//     （如 sqlite3 CLI、测试脚本）同时访问，也能等待锁释放而不是直接报错。
//
// PostgreSQL 保持原有连接池配置。
func Init(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	dialector, isSQLite, err := buildDialector(cfg.DatabaseURL)
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

	if isSQLite {
		if cfg.DBMaxOpenConns > 1 {
			log.Warn().
				Int("configured", cfg.DBMaxOpenConns).
				Msg("SQLite 不支持并发写，强制 MaxOpenConns=1（避免 database is locked）")
		}
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	} else {
		sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	}
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	DB = db
	return db, nil
}

// buildDialector 根据 DSN 前缀构造 dialector。
//
// 第二个返回值表示是否 SQLite（供 Init 决定连接池策略）。
func buildDialector(databaseURL string) (gorm.Dialector, bool, error) {
	dsn := strings.TrimSpace(databaseURL)
	switch {
	case strings.HasPrefix(dsn, "sqlite+aiosqlite:///"):
		path := strings.TrimPrefix(dsn, "sqlite+aiosqlite:///")
		return sqlite.Open(buildSQLiteDSN(path)), true, nil
	case strings.HasPrefix(dsn, "sqlite:///"):
		path := strings.TrimPrefix(dsn, "sqlite:///")
		return sqlite.Open(buildSQLiteDSN(path)), true, nil
	case strings.HasPrefix(dsn, "postgresql+asyncpg://"):
		dsn = strings.Replace(dsn, "postgresql+asyncpg://", "postgresql://", 1)
		return postgres.Open(dsn), false, nil
	case strings.HasPrefix(dsn, "postgres://"), strings.HasPrefix(dsn, "postgresql://"):
		return postgres.Open(dsn), false, nil
	default:
		return nil, false, fmt.Errorf("不支持的 DATABASE_URL: %s", databaseURL)
	}
}

// buildSQLiteDSN 为 glebarez/sqlite（纯 Go）追加 pragma：
//   - foreign_keys=1     ：与 Python 版 PRAGMA foreign_keys=ON 行为一致；
//   - busy_timeout(5000) ：锁冲突时等待 5s 再报错，避免直接 "database is locked"；
//   - journal_mode(WAL)  ：读写不互斥，减少锁冲突；
//   - synchronous(NORMAL)：WAL 下兼顾性能与安全。
func buildSQLiteDSN(path string) string {
	if path == "" {
		path = "./app.db"
	}
	pragmas := []string{
		"_pragma=foreign_keys(1)",
		"_pragma=busy_timeout(5000)",
		"_pragma=journal_mode(WAL)",
		"_pragma=synchronous(NORMAL)",
	}
	joined := strings.Join(pragmas, "&")
	if strings.Contains(path, "?") {
		return path + "&" + joined
	}
	return path + "?" + joined
}

// HealthCheck 数据库健康检查。
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

// Close 关闭数据库连接。
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
