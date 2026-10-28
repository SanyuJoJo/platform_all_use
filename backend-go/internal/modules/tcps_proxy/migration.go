// backend-go/internal/modules/tcps_proxy/migration.go
package tcps_proxy
import (
	"fmt"
	"gorm.io/gorm"
)
// Migrate 自动建表与字段迁移（幂等）。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&PlatformTCPSProxy{}); err != nil {
		return fmt.Errorf("tcps_proxy AutoMigrate 失败: %w", err)
	}
	if err := db.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tcps_proxy_mode_name
		 ON platform_tcps_proxy(mode, service_name)`,
	).Error; err != nil {
		return fmt.Errorf("tcps_proxy 创建唯一索引失败: %w", err)
	}
	return nil
}
