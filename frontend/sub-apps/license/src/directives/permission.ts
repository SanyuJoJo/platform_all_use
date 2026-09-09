import type { Directive, App } from 'vue';
import { usePermissionStore } from '@/store/permission';
export const permissionDirective: Directive = {
  mounted(el, binding) {
    const { value } = binding;
    if (value) {
      const permissionStore = usePermissionStore();
      if (!permissionStore.hasPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    }
  },
  updated(el, binding) {
    const { value } = binding;
    if (value) {
      const permissionStore = usePermissionStore();
      if (!permissionStore.hasPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    }
  },
};
export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective);
}
