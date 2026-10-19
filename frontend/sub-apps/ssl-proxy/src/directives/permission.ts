import type { Directive } from 'vue';
import { usePermissionStore } from '@/store/permission';
export const permissionDirective: Directive<HTMLElement, string> = {
  mounted(el, binding) {
    const permission = binding.value;
    if (!permission) return;
    const store = usePermissionStore();
    if (!store.hasPermission(permission)) {
      el.parentNode?.removeChild(el);
    }
  },
};
