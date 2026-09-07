import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useModuleStore } from '@/store/module';
import { createDiscreteApi } from 'naive-ui';
const { message } = createDiscreteApi(['message']);
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/modules'
  },
  {
    path: '/modules',
    name: 'Modules',
    component: () => import('@/views/modules/index.vue'),
    meta: { title: '模块管理', permission: 'module_manager:module:view' },
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
      const moduleStore = useModuleStore();
      if (!moduleStore.hasPermission(permission)) {
        console.warn(`[ModuleManager] 无权限访问 ${to.path}，需要 ${permission}`);
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
