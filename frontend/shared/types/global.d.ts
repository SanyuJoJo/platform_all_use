export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
  requestId: string;
}
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  pages: number;
}
