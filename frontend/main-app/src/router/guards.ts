// main-app/src/router/guards.ts
import router from './index';
import { useUserStore } from '@/store/user';
import { usePermissionStore } from '@/store/permission';
import { useAppStore } from '@/store/app';
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  const permissionStore = usePermissionStore();
  const appStore = useAppStore();
  // 1. 登录校验
  if (to.meta.requiresAuth !== false && !userStore.token) {
    return next({ path: '/login', query: { redirect: to.fullPath } });
  }
  // 2. 已登录但无用户信息 → 获取用户信息
  if (userStore.token && !userStore.userInfo) {
    await userStore.fetchUserInfo();
    await permissionStore.fetchPermissions();
    await appStore.fetchModules();
  }
  // 3. 权限校验
  if (to.meta.permission) {
    const hasPermission = permissionStore.hasPermission(to.meta.permission);
    if (!hasPermission) {
      return next({ path: '/403' });
    }
  }
  next();
});
