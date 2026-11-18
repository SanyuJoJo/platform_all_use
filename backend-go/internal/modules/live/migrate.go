package live

import "gorm.io/gorm"

// Migrate 开发环境创建 / 更新直播间表（生产环境走 goose）。
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&LiveRoom{})
}
