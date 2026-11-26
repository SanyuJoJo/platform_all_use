import type { Directive, App } from 'vue'
import { useUserStore } from '@/store/user'

/**
 * v-permission="'auth:user:create'"
 * 无权限时直接移除元素。
 */
export const permissionDirective: Directive<HTMLElement, string> = {
  mounted(el, binding) {
    const code = binding.value
    if (!code) return
    const userStore = useUserStore()
    if (!userStore.hasPermission(code)) {
      el.parentNode?.removeChild(el)
    }
  },
  updated(el, binding) {
    const code = binding.value
    if (!code) return
    const userStore = useUserStore()
    if (!userStore.hasPermission(code)) {
      el.parentNode?.removeChild(el)
    }
  },
}

export function setupPermissionDirective(app: App) {
  app.directive('permission', permissionDirective)
}
