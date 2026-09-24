export const ERROR_MESSAGE_MAP: Record<string, string> = {
  CRYPTO_INVALID_PARAM: '参数不合法',
  CRYPTO_ALGORITHM_NOT_ALLOWED: '算法不在白名单',
  CRYPTO_PATH_NOT_ALLOWED: '路径不允许',
  AUTH_FORBIDDEN: '无权限',
  CRYPTO_KEY_NOT_FOUND: '密钥不存在',
  CRYPTO_CERT_NOT_FOUND: '证书不存在',
  CRYPTO_CERT_PARSE_FAILED: '证书解析失败',
  CRYPTO_CORE_FAILED: '密码服务执行失败',
  CRYPTO_CORE_TIMEOUT: '密码服务超时',
  CRYPTO_CORE_RESPONSE_INVALID: '密码服务响应异常',
  CRYPTO_VERSION_UNSUPPORTED: '铜锁版本不满足',
  TASK_NOT_FOUND: '任务不存在',
  TASK_STATE_INVALID: '任务状态不允许该操作',
  TASK_CANCELLED: '任务已取消',
  TASK_TIMEOUT: '任务超时',
  RATE_LIMITED: '请求过于频繁',
  SERVICE_UNAVAILABLE: '服务暂不可用',
  AUTH_UNAUTHORIZED: '未认证或登录已过期',
  RBAC_DENIED: '当前角色无权执行该操作',
  REQUEST_INVALID: '请求格式错误',
  RESOURCE_CONFLICT: '资源冲突',
  INTERNAL_ERROR: '系统内部错误',
};
export function errorMessage(code?: string, fallback?: string): string {
  if (!code) return fallback || '操作失败';
  return ERROR_MESSAGE_MAP[code] || fallback || code;
}
