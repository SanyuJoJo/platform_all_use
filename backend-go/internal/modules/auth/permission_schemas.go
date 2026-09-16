package auth
// PermissionOut 权限响应对象。
type PermissionOut struct {
	ID        uint   `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ModuleID  string `json:"module_id"`
	Resource  string `json:"resource"`
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
}
// PermissionRegisterItem 权限注册项（内部调用）。
type PermissionRegisterItem struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
