import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { SUB_APP_PREFIXES } from '@/micro-frontend/constants';

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { ignoreAuth: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/default/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Dashboard', permission: 'dashboard:view', isMicroApp: false },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { ignoreAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { ignoreAuth: true },
  },
];
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  const menuStore = useMenuStore();

  // 直接从 localStorage 获取 token，作为后备
  const token = localStorage.getItem('token') || userStore.token || '';

  console.log('🛡️ 路由守卫检查:', to.path, '| token:', token);
  console.log('🛡️ userStore.token:', userStore.token);
  console.log('🛡️ localStorage token:', localStorage.getItem('token'));
  if (to.meta.ignoreAuth) {
    console.log('🛡️ 白名单，放行');
    next();
    return;
  }
  if (!userStore.token) {
    console.log('🛡️ 无 token，跳转登录');
    next({ path: '/login', query: { redirect: to.fullPath } });
    return;
  }
  const isSubApp = SUB_APP_PREFIXES.some(prefix => to.path.startsWith(prefix));
  to.meta.isMicroApp = isSubApp;
  if (!isSubApp && to.meta.permission && !userStore.permissions.includes(to.meta.permission as string)) {
    next('/403');
    return;
  }
  if (!menuStore.menuLoaded && userStore.token) {
    try {
      await menuStore.buildMenus();
    } catch (error) {
      console.error('[Menu] Build menus failed:', error);
    }
  }
  console.log('🛡️ 已认证，放行');
  next();
});
export default router;
