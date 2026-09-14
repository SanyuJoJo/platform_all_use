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
    meta: { ignoreAuth: true },
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

let subRoutesAdded = false;

router.beforeEach(async (to, _from, next) => {
  if (import.meta.env.DEV) {
    console.log(
      `[Router] 目标路径: ${to.path}, 当前匹配路由名: ${String(to.name)}`
    );
  }

  const userStore = useUserStore();
  const moduleStore = useModuleStore();

  if (to.meta.ignoreAuth) {
    next();
    return;
  }

  const token = userStore.token || localStorage.getItem('token');
  if (!token) {
    next({ path: '/login', query: { redirect: to.fullPath } });
    return;
  }

  if (!moduleStore.loaded) {
    try {
      await moduleStore.fetchModules();
    } catch (e) {
      console.error('[Router] 模块加载失败', e);
    }
  }

  if (moduleStore.loaded && !subRoutesAdded) {
    const modules = moduleStore.modules;
    modules.forEach((m) => {
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

  if (to.meta.isSubApp) {
    next();
    return;
  }

  if (to.meta.permission) {
    const hasPerm =
      userStore.permissions?.includes(to.meta.permission as string) ?? false;
    if (!hasPerm) {
      next('/403');
      return;
    }
  }

  next();
});

export default router;
