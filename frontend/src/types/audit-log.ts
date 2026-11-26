/* ============================================================================
 * 日志审计类型
 * ==========================================================================*/

/** 日志对象 */
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

/** 模块对象（简略，筛选下拉用） */
export interface ModuleOption {
  id: string;
  name: string;
  status: string;
}