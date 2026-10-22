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

// ========== 角色（方案 A：permission_codes → menuIds） ==========
export interface Role {
  id: number;
  name: string;
  code: string;
  description?: string;
  is_system: boolean;
  /** 已勾选的菜单 id 列表（含 M/C/F） */
  menuIds?: number[];
  /** 菜单数量（列表接口可选返回的轻量统计） */
  menuCount?: number;
  created_at: string;
  updated_at: string;
}
export interface RoleCreate {
  name: string;
  code: string;
  description?: string;
  /** 菜单 id 列表 */
  menuIds?: number[];
}
export interface RoleUpdate {
  name?: string;
  description?: string;
  /** 菜单 id 列表 */
  menuIds?: number[];
}

// ========== 权限（方案 A 后逐步废弃，暂保留以兼容 PermissionTree） ==========
export interface Permission {
  id: number;
  code: string;
  name: string;
  module_id: string;
  resource: string;
  action: string;
  created_at: string;
}

/* ============================================================================
 * 动态路由（子应用消费主应用下发路由树）
 * ========================================================================== */

export interface RouteMeta {
  title: string;
  icon?: string;
  permission?: string;
  hidden?: boolean;
  keepAlive?: boolean;
  [key: string]: unknown;
  [key: symbol]: unknown;
}

export interface RouteNode {
  id: string;
  parentId: string | null;
  name: string;
  path?: string;
  component?: string;
  microApp: string;
  type: 'directory' | 'menu';
  meta: RouteMeta;
  order?: number;
  children?: RouteNode[];
}
