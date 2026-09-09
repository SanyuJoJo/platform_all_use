import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { createDiscreteApi } from 'naive-ui';
import { usePermissionStore } from '@/store/permission';
const { message } = createDiscreteApi(['message']);
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/status'
  },
  {
    path: '/status',
    name: 'LicenseStatus',
    component: () => import('@/views/license/index.vue'),
    meta: { title: 'License状态', permission: 'license:license:view' },
  },
  {
    path: '/modules',
    name: 'LicenseModules',
    component: () => import('@/views/modules/index.vue'),
    meta: { title: '模块授权', permission: 'license:license:view' },
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/403.vue'),
    meta: { title: '无权限' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
  },
];
export function createAppRouter(moduleId: string, isQiankun: boolean) {
  const base = isQiankun ? `/${moduleId}` : '/';
  const router = createRouter({
    history: createWebHistory(base),
    routes,
  });
  router.beforeEach((to, from, next) => {
    const permission = to.meta.permission as string | undefined;
    if (permission) {
      const permissionStore = usePermissionStore();
      if (!permissionStore.hasPermission(permission)) {
        console.warn(`[License] 无权限访问 ${to.path}，需要 ${permission}`);
        if (window.__POWERED_BY_QIANKUN__) {
          next('/403');
        } else {
          message.error('您没有权限访问该页面');
          next(false);
        }
        return;
      }
    }
    next();
  });
  return router;
}
