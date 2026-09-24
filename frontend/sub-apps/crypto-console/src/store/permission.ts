import { defineStore } from 'pinia';
export const usePermissionStore = defineStore('permission', {
  state: () => ({
    permissions: [] as string[],
  }),
  actions: {
    setPermissions(perms: string[]) {
      this.permissions = perms;
    },
    hasPermission(code: string): boolean {
      return this.permissions.includes(code);
    },
  },
});
