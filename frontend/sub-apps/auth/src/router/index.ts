// sub-apps/auth/src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
const routes: RouteRecordRaw[] = [
  {
    path: '/auth/users',
    name: 'Users',
    component: () => import('@/views/users/index.vue'),
    meta: { permission: 'auth:users:view' },
  },
  {
    path: '/auth/roles',
    name: 'Roles',
    component: () => import('@/views/roles/index.vue'),
    meta: { permission: 'auth:roles:view' },
  },
];
const router = createRouter({
  history: createWebHistory(
    window.__POWERED_BY_QIANKUN__ 
      ? (window as any).__QIANKUN_BASE__ || '/' 
      : '/'
  ),
  routes,
});
export default router;
