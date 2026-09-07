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
// 模块对象（列表接口返回）
export interface Module {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  status: 'active' | 'inactive';
  dependencies: string[];
  installed_at: string;
  updated_at: string;
}
// 模块配置（键值对）
export interface Config {
  [key: string]: any;
}
