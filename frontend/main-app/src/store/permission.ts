import { useUserStore } from './user';
export function hasPermission(code: string): boolean {
  try {
    const userStore = useUserStore();
    return userStore.permissions?.includes(code) ?? false;
  } catch {
    return false;
  }
}
