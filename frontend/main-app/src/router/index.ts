import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { ignoreAuth: true },
  },
  {
  path: '/dashboard',
  name: 'Dashboard',
  component: () => import('@/views/dashboard/index.vue'),
  meta: { title: '仪表盘', permission: 'dashboard:view' }
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
        meta: { title: '仪表盘', icon: 'Grid', permission: 'dashboard:view' },
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
  const moduleStore = useModuleStore();
  // 1. 忽略认证的路由直接放行
  if (to.meta.ignoreAuth) {
    next();
    return;
  }
  // 2. 检查登录状态
  const token = userStore.token || localStorage.getItem('token');
  if (!token) {
    next({ path: '/login', query: { redirect: to.fullPath } });
    return;
  }
  // 确保 store 中的 token 被同步
  if (!userStore.token && token) {
    userStore.setToken(token);
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        userStore.setUser(user);
      } catch (e) {
        // 忽略
      }
    }
  }
  // 3. 动态判断当前路由是否为子应用路径
  // 先确保模块列表已加载
  if (!moduleStore.loaded) {
    await moduleStore.fetchModules().catch(() => {});
  }
  const activeIds = moduleStore.getActiveModuleIds();
  const isSubApp = activeIds.some((id) => to.path.startsWith(`/${id}`));
  to.meta.isMicroApp = isSubApp;
  // 4. 对于非子应用路由，校验权限
  if (!isSubApp && to.meta.permission) {
    const hasPerm = userStore.permissions?.includes(to.meta.permission as string) ?? false;
    if (!hasPerm) {
      next('/403');
      return;
    }
  }
  // 5. 首次加载或菜单未构建时，构建菜单
  if (!menuStore.menuLoaded) {
    try {
      await menuStore.buildMenus();
    } catch (error) {
      console.error('[Router] 构建菜单失败:', error);
    }
  }
  next();
});
export default router;
