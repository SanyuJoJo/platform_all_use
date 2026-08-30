import { defineStore } from 'pinia';

export interface User {
  id: number;
  username: string;
  nickname: string;
  email?: string;
  avatar?: string;
  status: number;
  roles: string[];
  permissions: string[];
}

export const useUserStore = defineStore('user', {
  state: () => {
    // 从 localStorage 恢复状态
    const token = localStorage.getItem('token') || '';
    const userStr = localStorage.getItem('user');
    let user: User | null = null;
    let permissions: string[] = [];
    if (userStr) {
      try {
        user = JSON.parse(userStr);
        permissions = user.permissions || [];
      } catch {}
    }
    return {
      token,
      user,
      permissions,
    };
  },
  actions: {
    setToken(newToken: string) {
      this.token = newToken;
      localStorage.setItem('token', newToken);
    },
    setUser(userInfo: User) {
      this.user = userInfo;
      this.permissions = userInfo.permissions || [];
      localStorage.setItem('user', JSON.stringify(userInfo));
      localStorage.setItem('permissions', JSON.stringify(userInfo.permissions || []));
    },
    logout() {
      this.token = '';
      this.user = null;
      this.permissions = [];
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('permissions');
    },
  },
  getters: {
    isLoggedIn: (state) => !!state.token,
  },
});
