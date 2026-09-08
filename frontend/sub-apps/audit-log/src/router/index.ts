import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { createDiscreteApi } from 'naive-ui';
import { usePermissionStore } from '@/store/permission';
const { message } = createDiscreteApi(['message']);
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/logs'
  },
  {
    path: '/logs',
    name: 'AuditLogs',
    component: () => import('@/views/logs/index.vue'),
    meta: { title: '日志审计', permission: 'audit_log:log:view' },
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
        console.warn(`[AuditLog] 无权限访问 ${to.path}，需要 ${permission}`);
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