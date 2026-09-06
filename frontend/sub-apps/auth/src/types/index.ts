// ========== 通用 API 响应 ==========
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
  requestId: string;
}
// ========== 通用分页响应（嵌套在 data 中） ==========
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  pages: number;
}
// ========== 用户 ==========
export interface User {
  id: number;
  username: string;
  nickname: string;
  email?: string;
  avatar?: string;
  status: number;
  roles: { id: number; name: string; code: string }[];
  created_at: string;
  updated_at: string;
}
export interface UserCreate {
  username: string;
  password: string;
  nickname: string;
  email?: string;
  role_ids?: number[];
  status?: number;
}
export interface UserUpdate {
  nickname?: string;
  email?: string;
  role_ids?: number[];
  status?: number;
}
// ========== 角色 ==========
export interface Role {
  id: number;
  name: string;
  code: string;
  description?: string;
  is_system: boolean;
  permission_codes: string[];
  created_at: string;
  updated_at: string;
}
export interface RoleCreate {
  name: string;
  code: string;
  description?: string;
  permission_codes?: string[];
}
export interface RoleUpdate {
  name?: string;
  description?: string;
  permission_codes?: string[];
}
// ========== 权限 ==========
export interface Permission {
  id: number;
  code: string;
  name: string;
  module_id: string;
  resource: string;
  action: string;
  created_at: string;
}
