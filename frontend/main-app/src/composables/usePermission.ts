import type { App } from 'vue';
import { hasPermission } from '@/store/permission';
export function usePermission() {
  return { hasPermission };
}
export function setupPermission(app: App) {
  app.config.globalProperties.$hasPermission = hasPermission;
}
declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $hasPermission: typeof hasPermission;
  }
}
