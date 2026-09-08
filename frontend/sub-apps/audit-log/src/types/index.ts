// 通用 API 响应
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
  requestId: string;
}
// 分页响应
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  pages: number;
}
// 日志对象
export interface AuditLog {
  id: number;
  user_id: number | null;
  username: string | null;
  module_id: string;
  action: string;
  resource: string | null;
  resource_id: string | null;
  detail: string | null;
  ip: string | null;
  user_agent: string | null;
  status: 'success' | 'fail';
  error_code: number | null;
  request_id: string | null;
  created_at: string;
}
// 模块对象（简略）
export interface Module {
  id: string;
  name: string;
  status: string;
}