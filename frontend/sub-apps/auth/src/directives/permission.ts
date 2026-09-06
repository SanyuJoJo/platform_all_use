import type { Directive, App } from 'vue';
import { useUserStore } from '@/store/user';
export const permissionDirective: Directive = {
  mounted(el, binding) {
    const { value } = binding;
    if (value) {
      const userStore = useUserStore();
      if (!userStore.hasPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    }
  },
  updated(el, binding) {
    const { value } = binding;
    if (value) {
      const userStore = useUserStore();
      if (!userStore.hasPermission(value)) {
        el.parentNode?.removeChild(el);
      }
    }
  },
};
export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective);
}
