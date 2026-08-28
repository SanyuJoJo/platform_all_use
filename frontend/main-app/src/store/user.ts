// main-app/src/store/user.ts
import { defineStore } from 'pinia';
import { login, getUserInfo } from '@/api/auth';
import type { UserInfo, LoginParams } from '@/api/types';
export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: null as UserInfo | null,
    roles: [] as string[],
  }),
  actions: {
    async login(params: LoginParams) {
      const res = await login(params);
      this.token = res.token;
      localStorage.setItem('token', res.token);
      await this.fetchUserInfo();
    },
    async fetchUserInfo() {
      const res = await getUserInfo();
      this.userInfo = res;
      this.roles = res.roles;
    },
    logout() {
      this.token = '';
      this.userInfo = null;
      localStorage.removeItem('token');
    },
  },
});
