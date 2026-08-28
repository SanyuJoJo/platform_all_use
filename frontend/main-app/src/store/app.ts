// main-app/src/store/app.ts
import { defineStore } from 'pinia';
import { getModules } from '@/api/modules';
export const useAppStore = defineStore('app', {
  state: () => ({
    modules: [] as Module[],
    menus: [] as MenuItem[],
  }),
  actions: {
    async fetchModules() {
      const res = await getModules();
      this.modules = res.filter(m => m.enabled);
      this.menus = this.buildMenus(this.modules);
    },
    buildMenus(modules: Module[]): MenuItem[] {
      return modules.flatMap((module) =>
        module.menus.map((menu) => ({
          ...menu,
          moduleCode: module.code,
        }))
      );
    },
  },
});
