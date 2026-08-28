// main-app/src/utils/error.ts
import { useMessage } from 'naive-ui';
const message = useMessage();
const errorCodeMap: Record<number, string> = {
  10001: '用户名或密码错误',
  10002: '账号已被禁用',
  10003: 'Token 已过期',
  // ... 参照错误码对照表
};
export function handleError(error: any): never {
  const code = error.response?.data?.code;
  const msg = errorCodeMap[code] || error.response?.data?.message || '系统错误，请稍后重试';
  
  message.error(msg);
  throw error;
}
