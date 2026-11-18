package media
import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"backend-go/internal/models"
)
// EnsureSeedData 在开发/首次部署时补 media 模块的权限码与菜单。
//
// 生产环境应通过 goose 迁移 + 手工种子完成，此函数只做"存在即跳过"。
func EnsureSeedData(db *gorm.DB) error {
	perms := []models.Permission{
		{Code: "media:media:view", Name: "查看媒体库", ModuleID: "media", Resource: "media", Action: "view"},
		{Code: "media:media:upload", Name: "上传媒体", ModuleID: "media", Resource: "media", Action: "upload"},
		{Code: "media:media:download", Name: "下载媒体", ModuleID: "media", Resource: "media", Action: "download"},
		{Code: "media:media:delete", Name: "删除媒体", ModuleID: "media", Resource: "media", Action: "delete"},
		{Code: "media:live:push", Name: "创建推流", ModuleID: "media", Resource: "live", Action: "push"},
		{Code: "media:live:stop", Name: "停止推流", ModuleID: "media", Resource: "live", Action: "stop"},
		{Code: "media:live:play", Name: "拉流播放", ModuleID: "media", Resource: "live", Action: "play"},
	}
	for _, p := range perms {
		var count int64
		if err := db.Model(&models.Permission{}).Where("code = ?", p.Code).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&p).Error; err != nil {
				log.Warn().Err(err).Str("code", p.Code).Msg("媒体权限种子写入失败")
			}
		}
	}
	return nil
}
