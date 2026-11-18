package media
import (
	"gorm.io/gorm"
)
// Migrate 开发环境迁移媒体表。
func Migrate(db *gorm.DB) error {
	return EnsureTable(db)
}
