// backend-go/internal/modules/ssl_proxy/migration.go
package ssl_proxy

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Migrate 自动建表与字段迁移（幂等）。
//
// ★ 只做表结构（字段增删改），不处理唯一索引。
//   唯一索引走 EnsureIndexes，避免历史数据重名阻断启动。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&PlatformSSLProxy{}); err != nil {
		return fmt.Errorf("AutoMigrate 失败: %w", err)
	}
	return nil
}

// EnsureIndexes 显式维护唯一索引（幂等）。
//
// ★ 历史索引变迁：
//   - 旧：platform_ssl_proxy(service_name) UNIQUE
//     —— 全表唯一，client / server 无法同名
//   - 新：platform_ssl_proxy(mode, service_name) UNIQUE
//     —— client / server 各自独立命名空间，允许同名
//
// 处理步骤：
//  1. 删除旧单字段唯一索引（若存在）；
//  2. 尝试创建新复合唯一索引；
//  3. 若失败（历史数据在 (mode, service_name) 上仍重名），仅告警，
//     不阻断启动，由调用方决定是否人工清理。
func EnsureIndexes(db *gorm.DB) error {
	// 1. 删除旧单字段唯一索引（忽略不存在 / 不支持的错误）
	if err := db.Exec(`DROP INDEX IF EXISTS idx_ssl_proxy_service_name`).Error; err != nil {
		log.Warn().Err(err).
			Msg("ssl_proxy: 删除旧索引 idx_ssl_proxy_service_name 失败（可能不存在）")
	}

	// 2. 创建新的复合唯一索引（mode, service_name）
	if err := db.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_ssl_proxy_mode_name
		 ON platform_ssl_proxy(mode, service_name)`,
	).Error; err != nil {
		log.Warn().Err(err).
			Msg("ssl_proxy: 创建 (mode, service_name) 复合唯一索引失败（可能存在历史重名数据）")
		return err
	}
	log.Info().Msg("ssl_proxy: (mode, service_name) 复合唯一索引已就绪")
	return nil
}
