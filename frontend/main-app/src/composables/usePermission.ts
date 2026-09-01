import type { App } from 'vue';
import { hasPermission } from '@/utils/permission';
/**
 * 权限检查组合式函数
 */
export function usePermission() {
  return { hasPermission };
}
/**
 * 注册全局权限指令和属性
 */
export function setupPermission(app: App) {
  // 全局方法（供模板使用）
  app.config.globalProperties.$hasPermission = hasPermission;
  // 自定义指令 v-permission
  app.directive('permission', {
    mounted(el, binding) {
      const permission = binding.value;
      if (permission && !hasPermission(permission)) {
        // 无权限，移除元素
        el.parentNode?.removeChild(el);
      }
    },
    // 可选：更新时重新判断
    updated(el, binding) {
      const permission = binding.value;
      if (permission && !hasPermission(permission)) {
        el.parentNode?.removeChild(el);
      }
    },
  });
}
// 类型扩展，使模板中的 $hasPermission 可识别
declare module '@vue/runtime-core' {
  interface ComponentCustomProperties {
    $hasPermission: typeof hasPermission;
  }
}
