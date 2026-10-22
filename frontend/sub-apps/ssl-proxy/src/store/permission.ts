import { defineStore } from 'pinia';

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    permissions: [] as string[],
  }),
  actions: {
    setPermissions(perms: string[]) {
      this.permissions = Array.isArray(perms) ? perms : [];
    },
    hasPermission(code: string): boolean {
      if (!code) return true;
      return this.permissions.includes(code);
    },
  },
});
