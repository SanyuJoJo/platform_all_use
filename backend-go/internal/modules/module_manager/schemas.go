package module_manager
import "time"
// MenuItem 菜单项。
type MenuItem struct {
	ID         string  `json:"id"`
	ParentID   *string `json:"parent_id"`
	Title      string  `json:"title"`
	Icon       *string `json:"icon"`
	Path       string  `json:"path"`
	Component  string  `json:"component"`
	Permission *string `json:"permission"`
	Order      int     `json:"order"`
}
// ModuleInstallReq 从受控目录安装模块。
type ModuleInstallReq struct {
	InstallType string `json:"install_type" binding:"omitempty,oneof=path"`
	SourcePath  string `json:"source_path" binding:"required"`
}
// ModuleUpgradeReq 升级模块。
type ModuleUpgradeReq struct {
	InstallType string  `json:"install_type" binding:"required,oneof=zip path"`
	FilePath    *string `json:"file_path"`
	SourcePath  *string `json:"source_path"`
}
// ModuleConfigUpdate 更新模块配置。
type ModuleConfigUpdate struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}
// ModuleOut 模块响应对象。
type ModuleOut struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Description   *string                `json:"description"`
	Author        *string                `json:"author"`
	Homepage      *string                `json:"homepage"`
	Status        string                 `json:"status"`
	EntryBackend  string                 `json:"entry_backend"`
	EntryFrontend *string                `json:"entry_frontend"`
	Dependencies  []string               `json:"dependencies"`
	Menus         []MenuItem             `json:"menus"`
	Config        map[string]interface{} `json:"config"`
	InstalledAt   string                 `json:"installed_at"`
	UpdatedAt     string                 `json:"updated_at"`
}
// ModuleListQuery 模块列表查询参数。
type ModuleListQuery struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status      string `form:"status" binding:"omitempty,oneof=active inactive"`
	Keyword     string `form:"keyword"`
	FilterMenus bool   `form:"filter_menus"`
}
// 时间辅助。
func formatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	t = t.UTC()
	if t.Nanosecond()/1000 == 0 {
		return t.Format("2006-01-02T15:04:05")
	}
	return t.Format("2006-01-02T15:04:05.000000")
}
