package menu

// CreateMenuRequest 新增菜单请求。
type CreateMenuRequest struct {
	ParentID   uint   `json:"parentId"`
	MenuName   string `json:"menuName" binding:"required,max=50"`
	Name       string `json:"name" binding:"max=100"`
	MicroApp   string `json:"microApp" binding:"max=50"`
	OrderNum   int    `json:"orderNum"`
	Path       string `json:"path" binding:"max=200"`
	Component  string `json:"component" binding:"max=255"`
	Query      string `json:"query" binding:"max=255"`
	IsFrame    *int8  `json:"isFrame"`
	IsCache    *int8  `json:"isCache"`
	MenuType   string `json:"menuType" binding:"required,oneof=M C F"`
	Visible    *int8  `json:"visible"`
	Status     *int8  `json:"status"`
	Perms      string `json:"perms" binding:"max=100"`
	Icon       string `json:"icon" binding:"max=100"`
	ModuleID   string `json:"moduleId" binding:"max=50"`
	Remark     string `json:"remark" binding:"max=500"`
}

// UpdateMenuRequest 修改菜单请求（ID 从路径参数取）。
type UpdateMenuRequest struct {
	ParentID   uint   `json:"parentId"`
	MenuName   string `json:"menuName" binding:"required,max=50"`
	Name       string `json:"name" binding:"max=100"`
	MicroApp   string `json:"microApp" binding:"max=50"`
	OrderNum   int    `json:"orderNum"`
	Path       string `json:"path" binding:"max=200"`
	Component  string `json:"component" binding:"max=255"`
	Query      string `json:"query" binding:"max=255"`
	IsFrame    *int8  `json:"isFrame"`
	IsCache    *int8  `json:"isCache"`
	MenuType   string `json:"menuType" binding:"required,oneof=M C F"`
	Visible    *int8  `json:"visible"`
	Status     *int8  `json:"status"`
	Perms      string `json:"perms" binding:"max=100"`
	Icon       string `json:"icon" binding:"max=100"`
	ModuleID   string `json:"moduleId" binding:"max=50"`
	Remark     string `json:"remark" binding:"max=500"`
}

// QueryMenuRequest 菜单查询参数。
type QueryMenuRequest struct {
	MenuName string `form:"menuName"`
	MenuType string `form:"menuType"`
	Status   *int8  `form:"status"`
	ModuleID string `form:"moduleId"`
	MicroApp string `form:"microApp"`
}

// RouterVO 前端动态路由节点，与前端约定严格对齐。
//
// 关键字段：
//   - ID       ：数据库主键字符串，前端用作 Map key；
//   - ParentID ：顶级为 null，子节点为父 id 字符串（同时保留 Children 嵌套）；
//   - Name     ：英文路由名（PascalCase），全局唯一；
//   - Type     ：directory（目录）| menu（叶子页面）；
//   - MicroApp ：所属微应用 id，qiankun 路由分发依据；
//   - Component：directory 固定 "Layout"，menu 为 views 下相对路径；
//   - Order    ：同级排序（与 Children 数组顺序一致，兜底用）。
type RouterVO struct {
	ID        string     `json:"id"`
	ParentID  *string    `json:"parentId"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Component string     `json:"component,omitempty"`
	MicroApp  string     `json:"microApp"`
	Type      string     `json:"type"`
	Order     int        `json:"order,omitempty"`
	Meta      *MetaVO    `json:"meta,omitempty"`
	Children  []RouterVO `json:"children,omitempty"`
}

// MetaVO 路由元信息。
type MetaVO struct {
	Title      string `json:"title"`
	Icon       string `json:"icon,omitempty"`
	Permission string `json:"permission,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
	NoCache    bool   `json:"noCache,omitempty"`
	Link       string `json:"link,omitempty"`
}
