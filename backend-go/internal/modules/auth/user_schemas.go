package auth
// UserCreate 创建用户请求。
//
// 必要字段增加 `binding:"required"`，缺失时在 HTTP 绑定层返回 422 / 90004。
// 长度/格式校验由 Service 层完成，统一返回业务错误码。
//
// Status 说明：
//   - 使用 *int8，省略时由 Service 层默认为 1（启用），
//     与 Python Pydantic Field(1, ge=0, le=1) 一致；
//   - 增加 binding:"omitempty,min=0,max=1"，
//     非法值（如 5 / -1）在 HTTP 绑定层返回 422 / 90004；
//   - ⚠️ P1-C 边界：JSON null 会被视为"未提供"，与 Python 行为不一致，
//     详见文档 § 4.12。
type UserCreate struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Nickname string  `json:"nickname" binding:"required"`
	Email    *string `json:"email"`
	RoleIDs  []uint  `json:"role_ids"`
	Status   *int8   `json:"status" binding:"omitempty,min=0,max=1"`
}
// UserUpdate 更新用户请求（所有字段可选）。
type UserUpdate struct {
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	RoleIDs  *[]uint `json:"role_ids"`
	Status   *int8   `json:"status" binding:"omitempty,min=0,max=1"`
}
// UserStatusUpdate 启用/禁用请求。
//
// Status 使用 *int8 + binding:"required"，
// 保证字段存在性由 HTTP 绑定层校验，handler 无需二次检查。
type UserStatusUpdate struct {
	Status *int8 `json:"status" binding:"required,min=0,max=1"`
}
// UserPasswordReset 管理员重置用户密码请求。
type UserPasswordReset struct {
	NewPassword string `json:"new_password" binding:"required"`
}
// RoleBrief 角色简要信息（用户对象内嵌）。
type RoleBrief struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
// UserOut 用户响应对象。
type UserOut struct {
	ID        uint        `json:"id"`
	Username  string      `json:"username"`
	Nickname  string      `json:"nickname"`
	Email     *string     `json:"email"`
	Avatar    *string     `json:"avatar"`
	Status    int8        `json:"status"`
	Roles     []RoleBrief `json:"roles"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}
// UserListQuery 用户列表查询参数。
type UserListQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Status   *int8  `form:"status" binding:"omitempty,min=0,max=1"`
	RoleID   *uint  `form:"role_id"`
}
