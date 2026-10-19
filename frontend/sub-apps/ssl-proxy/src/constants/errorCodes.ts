export const ERROR_MESSAGE_MAP: Record<string | number, string> = {
  0: '成功',
  40001: '参数错误',
  40401: '实例不存在',
  40901: '端口冲突',
  40902: '服务运行中不可修改/删除',
  50001: '配置生成失败',
  50002: '证书解析失败',
  50003: '启动失败',
  50004: '停止失败',
  50005: 'Tengine 不可用',
  50006: '协议不支持',
  50007: '权限错误',
  50008: '批量部分失败',
  AUTH_UNAUTHORIZED: '未认证或登录已过期',
  RBAC_DENIED: '当前角色无权执行该操作',
  NETWORK_ERROR: '网络异常',
  INTERNAL_ERROR: '系统内部错误',
  HTTP_400: '请求参数错误',
  HTTP_403: '无权限',
  HTTP_404: '资源不存在',
  HTTP_500: '服务内部错误',
};
export function errorMessage(
  code?: string | number,
  fallback?: string
): string {
  if (code === undefined || code === null || code === '') {
    return fallback || '操作失败';
  }
  return ERROR_MESSAGE_MAP[code] || fallback || String(code);
}
