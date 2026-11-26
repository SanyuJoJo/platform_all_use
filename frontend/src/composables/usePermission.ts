import { useUserStore } from '@/store/user'

/** 页面内权限判断：const { hasPermission } = usePermission() */
export function usePermission() {
  const userStore = useUserStore()

  function hasPermission(code?: string): boolean {
    return userStore.hasPermission(code)
  }

  function hasAnyPermission(codes: string[]): boolean {
    return codes.some((code) => userStore.hasPermission(code))
  }

  function hasAllPermissions(codes: string[]): boolean {
    return codes.every((code) => userStore.hasPermission(code))
  }

  return { hasPermission, hasAnyPermission, hasAllPermissions }
}
