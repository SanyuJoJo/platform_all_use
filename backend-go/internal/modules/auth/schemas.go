package auth
import "time"
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
// formatDateTime 将时间格式化为 Python datetime.isoformat() 兼容的字符串。
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
// intPtr 返回 int 指针。
func intPtr(v int) *int {
	return &v
}
