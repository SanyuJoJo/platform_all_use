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
  app.config.globalProperties.$hasPermission = hasPermission;

  app.directive('permission', {
    mounted(el, binding) {
      const permission = binding.value;
      if (permission && !hasPermission(permission)) {
        el.parentNode?.removeChild(el);
      }
    },
    updated(el, binding) {
      const permission = binding.value;
      if (permission && !hasPermission(permission)) {
        el.parentNode?.removeChild(el);
      }
    },
  });
}

// 使用 `vue` 模块增强（Vue 3.5 推荐），而不是 @vue/runtime-core
declare module 'vue' {
  interface ComponentCustomProperties {
    $hasPermission: typeof hasPermission;
  }
}

export {};
