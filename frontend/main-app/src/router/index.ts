import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useModuleStore } from '@/store/module';

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { ignoreAuth: true },
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { ignoreAuth: true },
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layouts/default/index.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', permission: 'platform:dashboard:view' },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    // ★ 不要再加 ignoreAuth：否则刷新时命中 NotFound 会直接放行，
    //   导致模块未加载、subapp_* 路由未注册，最终停在这里。
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

let subRoutesAdded = false;
let addingSubRoutes = false;

router.beforeEach(async (to, from, next) => {
  console.log(`[Router] 目标路径: ${to.path}, 当前匹配路由名: ${to.name}`);

  const userStore = useUserStore();
  const moduleStore = useModuleStore();

  // 1. 白名单路由（登录页 / 403）直接放行
  if (to.meta.ignoreAuth) {
    next();
    return;
  }

  // 2. 未登录跳转登录页
  const token = userStore.token || localStorage.getItem('token');
  if (!token) {
    next({ path: '/login', query: { redirect: to.fullPath } });
    return;
  }

  // 3. 加载模块（幂等，store 内部会缓存）
  if (!moduleStore.loaded) {
    try {
      await moduleStore.fetchModules();
      console.log('[Router] 模块加载完成');
    } catch (e) {
      console.error('[Router] 模块加载失败', e);
    }
  }

  // 4. 首次注册子应用路由 → 用 to.fullPath 重新导航
  if (moduleStore.loaded && !subRoutesAdded && !addingSubRoutes) {
    addingSubRoutes = true;
    moduleStore.modules.forEach(m => {
      const routeName = `subapp_${m.id}`;
      if (!router.hasRoute(routeName)) {
        router.addRoute('Layout', {
          path: `/${m.id}/:pathMatch(.*)*`,
          name: routeName,
          component: { render: () => null },
          meta: { ignoreAuth: true, isSubApp: true },
        });
        console.log(`[Router] 添加子应用路由: ${routeName}`);
      }
    });
    subRoutesAdded = true;
    addingSubRoutes = false;

    // ★ 关键：重新匹配当前路径，避免本次导航落在 NotFound
    next({ path: to.fullPath, replace: true });
    return;
  }

  // 5. 子应用路由直接放行（子应用内部自己有路由/守卫）
  if (to.meta.isSubApp) {
    next();
    return;
  }

  // 6. 主应用页面权限检查
  if (to.meta.permission) {
    const hasPerm = userStore.permissions?.includes(to.meta.permission as string) ?? false;
    if (!hasPerm) {
      next('/403');
      return;
    }
  }

  next();
});

export default router;
