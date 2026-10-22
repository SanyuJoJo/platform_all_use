export interface ApiResponse<T = unknown> {
  schema_version?: string;
  code: string | number;
  message: string;
  data: T;
  error?: {
    code: string;
    message: string;
    detail?: Record<string, unknown>;
    retryable?: boolean;
  } | null;
  audit?: Record<string, unknown> | null;
  request_id?: string;
  operation_id?: string;
  task_id?: string | null;
}
export interface OperationResponse<T = unknown> {
  task_id?: string | null;
  [key: string]: unknown;
}
export interface PageQuery {
  page?: number;
  page_size?: number;
  keyword?: string;
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
