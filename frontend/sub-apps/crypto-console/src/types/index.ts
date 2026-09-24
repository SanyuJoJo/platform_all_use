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
