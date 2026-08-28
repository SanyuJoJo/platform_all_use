// main-app/src/store/permission.ts
import { defineStore } from 'pinia';
export const usePermissionStore = defineStore('permission', {
  state: () => ({
    permissions: [] as string[],
  }),
  getters: {
    hasPermission: (state) => (permission: string) => {
      return state.permissions.includes(permission) || state.permissions.includes('*');
    },
  },
  actions: {
    async fetchPermissions() {
      const res = await getPermissions();
      this.permissions = res;
    },
  },
});
