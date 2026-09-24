import {
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
} from 'vue-router';
import { createDiscreteApi } from 'naive-ui';
import { usePermissionStore } from '@/store/permission';
const { message } = createDiscreteApi(['message']);
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/ca',
  },
  {
    path: '/ca',
    name: 'CaManager',
    component: () => import('@/views/ca/index.vue'),
    meta: { title: 'CA 管理', permission: 'crypto_console:ca:view' },
  },
  {
    path: '/certs',
    name: 'CertManager',
    component: () => import('@/views/certs/index.vue'),
    meta: { title: '证书管理', permission: 'crypto_console:cert:view' },
  },
  {
    path: '/csr',
    name: 'CsrManager',
    component: () => import('@/views/csr/index.vue'),
    meta: { title: 'CSR 管理', permission: 'crypto_console:csr:view' },
  },
  {
    path: '/crl',
    name: 'CrlManager',
    component: () => import('@/views/crl/index.vue'),
    meta: { title: 'CRL 管理', permission: 'crypto_console:crl:view' },
  },
  {
    path: '/keys',
    name: 'KeyManager',
    component: () => import('@/views/keys/index.vue'),
    meta: { title: '密钥管理', permission: 'crypto_console:key:view' },
  },
  {
    path: '/tasks',
    name: 'TaskCenter',
    component: () => import('@/views/tasks/index.vue'),
    meta: { title: '任务中心', permission: 'crypto_console:task:view' },
  },
  {
    path: '/audits',
    name: 'AuditLog',
    component: () => import('@/views/audits/index.vue'),
    meta: { title: '审计日志', permission: 'crypto_console:audit:view' },
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/403.vue'),
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
        console.warn(`[CryptoConsole] 无权限访问 ${to.path}，需要 ${permission}`);
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
