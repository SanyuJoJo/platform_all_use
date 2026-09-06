import { defineStore } from 'pinia';
import { ref } from 'vue';
// 从 localStorage 读取权限列表
function getInitialPermissions(): string[] {
  try {
    return JSON.parse(localStorage.getItem('permissions') || '[]');
  } catch {
    return [];
  }
}
export const useUserStore = defineStore('user', () => {
  const token = ref<string>(localStorage.getItem('token') || '');
  const permissions = ref<string[]>(getInitialPermissions());
  const userInfo = ref<any>(null);
  const setToken = (newToken: string) => {
    token.value = newToken;
    localStorage.setItem('token', newToken);
  };
  const setPermissions = (perms: string[]) => {
    permissions.value = perms;
    localStorage.setItem('permissions', JSON.stringify(perms));
  };
  const setUser = (user: any) => {
    userInfo.value = user;
    localStorage.setItem('user', JSON.stringify(user));
  };
  const logout = () => {
    token.value = '';
    permissions.value = [];
    userInfo.value = null;
    localStorage.removeItem('token');
    localStorage.removeItem('permissions');
    localStorage.removeItem('user');
  };
  const hasPermission = (code: string) => permissions.value.includes(code);
  return {
    token,
    permissions,
    userInfo,
    setToken,
    setPermissions,
    setUser,
    logout,
    hasPermission,
  };
});
