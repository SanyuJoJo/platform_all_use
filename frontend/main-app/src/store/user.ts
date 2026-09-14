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
    const token = localStorage.getItem('token') || '';
    const userStr = localStorage.getItem('user');
    let user: User | null = null;
    let permissions: string[] = [];

    if (userStr) {
      try {
        const parsed = JSON.parse(userStr) as User | null;
        if (parsed) {
          user = parsed;
          permissions = Array.isArray(parsed.permissions)
            ? parsed.permissions
            : [];
        }
      } catch {
        // 忽略解析失败
      }
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
      this.permissions = Array.isArray(userInfo.permissions)
        ? userInfo.permissions
        : [];
      localStorage.setItem('user', JSON.stringify(userInfo));
      localStorage.setItem(
        'permissions',
        JSON.stringify(this.permissions)
      );
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
