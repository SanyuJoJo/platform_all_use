package module_manager
import (
	"time"
	"gorm.io/gorm"
	"backend-go/internal/models"
)
// EnsureModuleSeedData 初始化核心模块种子数据（幂等、增量）。
//
// v1.1（P1-08）：使用 strPtrOrNil 替代 strPtr，避免空字符串入库。
// v1.1（P2-06）：extractDependencies 兼容 []string / []interface{}。
// v1.3（P2-NEW-07）：extractDependencies 委托到 extractStringList，消除重复。
func EnsureModuleSeedData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, seed := range CoreModules {
			var existing models.Module
			err := tx.First(&existing, "id = ?", seed.ID).Error
			if err == nil {
				updates := map[string]interface{}{
					"name":           seed.Name,
					"version":        seed.Version,
					"description":    strPtrOrNil(seed.Description),
					"author":         strPtrOrNil(seed.Author),
					"entry_backend":  seed.EntryBackend,
					"entry_frontend": seed.EntryFrontend,
					"manifest":       models.JSONMap(seed.Manifest),
					"updated_at":     time.Now().UTC(),
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
				if err := tx.Where("module_id = ?", seed.ID).
					Delete(&models.ModuleDependency{}).Error; err != nil {
					return err
				}
			} else if err == gorm.ErrRecordNotFound {
				m := models.Module{
					ID:            seed.ID,
					Name:          seed.Name,
					Version:       seed.Version,
					Description:   strPtrOrNil(seed.Description),
					Author:        strPtrOrNil(seed.Author),
					Status:        seed.Status,
					EntryBackend:  seed.EntryBackend,
					EntryFrontend: seed.EntryFrontend,
					Config:        models.JSONMap(seed.Config),
					Manifest:      models.JSONMap(seed.Manifest),
				}
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
			} else {
				return err
			}
			deps := extractDependencies(seed.Manifest)
			for _, dep := range deps {
				if err := tx.Create(&models.ModuleDependency{
					ModuleID:     seed.ID,
					DependencyID: dep,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
// extractDependencies 从 manifest 中提取 dependencies 列表。
//
// v1.3（P2-NEW-07）：委托到 extractStringList，消除与 service.go 的重复逻辑。
func extractDependencies(manifest map[string]interface{}) []string {
	if manifest == nil {
		return nil
	}
	return extractStringList(manifest["dependencies"])
}
