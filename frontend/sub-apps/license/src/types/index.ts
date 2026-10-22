export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
  requestId: string;
}
export interface LicenseStatus {
  is_valid: boolean;
  license_type: string;
  expires_at: string | null;
  days_remaining: number | null;
  max_users: number | null;
  current_users: number | null;
  authorized_modules: string[];
  is_expired: boolean;
  is_expiring_soon: boolean;
}
export interface ModuleAuthStatus {
  module_id: string;
  module_name: string;
  is_authorized: boolean;
  expires_at: string | null;
}

/* ========== 动态路由 ========== */
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
