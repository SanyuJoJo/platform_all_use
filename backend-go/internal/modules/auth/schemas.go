package auth
// LoginReq 登录请求。
type LoginReq struct {
	Username string `json:"username" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=1"`
}
// RefreshReq 刷新请求。
type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}
// ChangePasswordReq 修改密码请求。
//
// v1.1（P2-01）：使用指针类型精确对齐 Python Pydantic 语义：
//   - Python `Field(..., min_length=0)` 语义为"字段必填但允许空字符串"；
//   - Go 通过 `*string` + `binding:"required"` 实现：
//       · 字段缺失   → 指针为 nil → binding 触发 → 422 / 90004
//       · 字段存在但空 → 指针非 nil，值为 "" → binding 通过 → 进入 Service
//     （旧密码为空 → 10003，与 Python 一致）
type ChangePasswordReq struct {
	OldPassword     *string `json:"old_password" binding:"required"`
	NewPassword     *string `json:"new_password" binding:"required"`
	ConfirmPassword *string `json:"confirm_password" binding:"required"`
}
// UserInfo 当前用户信息。
type UserInfo struct {
	ID          uint     `json:"id"`
	Username    string   `json:"username"`
	Nickname    string   `json:"nickname"`
	Email       *string  `json:"email"`
	Avatar      *string  `json:"avatar"`
	Status      int8     `json:"status"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}
// TokenResp 登录/刷新响应。
type TokenResp struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	User         *UserInfo `json:"user,omitempty"`
}
// RefreshResp 刷新响应。
type RefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
