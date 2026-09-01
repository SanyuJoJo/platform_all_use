import { useUserStore } from '@/store/user';
/**
 * 检查当前用户是否拥有指定权限
 * @param code 权限编码
 * @returns boolean
 */
export function hasPermission(code: string): boolean {
  try {
    const userStore = useUserStore();
    return userStore.permissions?.includes(code) ?? false;
  } catch {
    return false;
  }
}
