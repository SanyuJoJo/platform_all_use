package auth

// CreateRoleRequest 新增角色请求。
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Code        string `json:"code" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	MenuIDs     []uint `json:"menuIds"`
}

// UpdateRoleRequest 修改角色请求。
//
// menuIds 语义：
//   - 字段缺失 / null → 不更新菜单（保留原有）
//   - []              → 清空所有菜单
//   - [1,2]           → 全量替换
type UpdateRoleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Description string  `json:"description" binding:"max=255"`
	MenuIDs     *[]uint `json:"menuIds"`
}

// RoleListItem 角色列表项。
type RoleListItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	MenuCount   int64  `json:"menuCount"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// RoleDetail 角色详情。
type RoleDetail struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	MenuIDs     []uint `json:"menuIds"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// RoleListData 角色列表响应数据。
type RoleListData struct {
	Items    []RoleListItem `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Pages    int            `json:"pages"`
}
