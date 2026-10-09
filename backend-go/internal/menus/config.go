package menus

// FrontendConfig 前端动态配置根对象。
//
// 对应 configs/frontend.yaml。
type FrontendConfig struct {
	Version string `json:"version" yaml:"version"`

	// ModuleMenus 按模块 ID 覆盖对应模块的菜单。
	//
	// key = 模块 ID；value = 该模块的菜单列表（结构与 manifest.menus 一致）。
	//
	// 语义：
	//   - 列出：使用配置的菜单替换 DB manifest 的菜单；
	//   - 未列出：保留 DB manifest 的菜单（自动加载）。
	ModuleMenus map[string][]interface{} `json:"module_menus" yaml:"module_menus"`

	// MenusHidden 显式隐藏菜单的模块 ID 列表。
	//
	// 语义：
	//   - 列出：强制清空该模块的菜单（即使 DB manifest 中有）；
	//   - 优先级高于 ModuleMenus（同时出现时以 MenusHidden 为准）。
	//
	// 典型场景：
	//   - 通过模块管理上传了模块，但暂时不想在侧边栏展示其菜单；
	//   - 模块本身仍可加载（提供 API），仅菜单不显示。
	MenusHidden []string `json:"menus_hidden" yaml:"menus_hidden"`

	// EntryFrontendOverrides 按模块 ID 覆盖 entry_frontend。
	//
	// 语义：
	//   - 列出且值非空：使用配置的入口；
	//   - 未列出：保留 DB manifest 的 entry_frontend。
	EntryFrontendOverrides map[string]string `json:"entry_frontend_overrides" yaml:"entry_frontend_overrides"`
}
