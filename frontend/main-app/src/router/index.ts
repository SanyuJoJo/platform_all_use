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
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/error/403.vue'),
    meta: { ignoreAuth: true },
  },
  {
    path: '/',
    // ★ 关键：加回 name: 'Layout'，main.ts 中的 router.addRoute('Layout', ...) 依赖它
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
    meta: { ignoreAuth: true },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

let subRoutesAdded = false;

function isSubAppPath(path: string, modules: any[]) {
  return modules.some(m => path.startsWith(`/${m.id}`));
}

router.beforeEach(async (to, from, next) => {
  console.log(`[Router] 目标路径: ${to.path}, 当前匹配路由名: ${to.name}`);

  const userStore = useUserStore();
  const menuStore = useMenuStore();
  const moduleStore = useModuleStore();

  // 1. 白名单路由直接放行
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

  // 3. 加载模块（如果未加载）
  if (!moduleStore.loaded) {
    try {
      await moduleStore.fetchModules();
      console.log('[Router] 模块加载完成');
    } catch (e) {
      console.error('[Router] 模块加载失败', e);
    }
  }

  // 4. 添加所有子应用路由（仅首次）
  if (moduleStore.loaded && !subRoutesAdded) {
    const modules = moduleStore.modules;
    modules.forEach(m => {
      const routeName = `subapp_${m.id}`;
      if (!router.hasRoute(routeName)) {
        router.addRoute('Layout', {
          path: `/${m.id}/:pathMatch(.*)*`,
          name: routeName,
          component: { render: () => null },
          meta: { ignoreAuth: true, isSubApp: true },
        });
      }
    });
    subRoutesAdded = true;
  }

  // 5. 子应用路由放行
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
