package auth
// RoleCreate 创建角色请求。
type RoleCreate struct {
	Name            string   `json:"name" binding:"required"`
	Code            string   `json:"code" binding:"required"`
	Description     *string  `json:"description"`
	PermissionCodes []string `json:"permission_codes"`
}
// RoleUpdate 更新角色请求（所有字段可选）。
//
// v1.3 语义：客户端传 code / is_system / 任何未定义字段 → 422 / 90004。
// Go 通过 JSON 解码阶段的 DisallowUnknownFields 实现，见 role_handler.go。
type RoleUpdate struct {
	Name            *string   `json:"name"`
	Description     *string   `json:"description"`
	PermissionCodes *[]string `json:"permission_codes"`
}
// RoleOut 角色响应对象。
type RoleOut struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	Code            string   `json:"code"`
	Description     *string  `json:"description"`
	IsSystem        int8     `json:"is_system"`
	PermissionCodes []string `json:"permission_codes"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}
// RoleListQuery 角色列表查询参数。
type RoleListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
}
