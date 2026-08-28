// main-app/src/components/Permission/index.ts
import type { App, Directive } from 'vue';
import { usePermissionStore } from '@/store/permission';
export const permissionDirective: Directive = {
  mounted(el, binding) {
    const { value } = binding;
    if (!value) return;
    
    const permissionStore = usePermissionStore();
    const hasPermission = permissionStore.hasPermission(value);
    
    if (!hasPermission) {
      el.parentNode?.removeChild(el);
    }
  },
};
export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective);
}
