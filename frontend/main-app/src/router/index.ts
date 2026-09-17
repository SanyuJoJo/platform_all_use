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

/**
 * 注册所有子应用路由（幂等）。
 * 返回本次是否真的新增了路由。
 */
function registerSubAppRoutes(): boolean {
  const moduleStore = useModuleStore();
  if (!moduleStore.loaded) return false;

  let added = false;
  moduleStore.modules.forEach((m) => {
    const routeName = `subapp_${m.id}`;
    if (router.hasRoute(routeName)) return;

    router.addRoute('Layout', {
      path: `/${m.id}/:pathMatch(.*)*`,
      name: routeName,
      component: { render: () => null },
      meta: { ignoreAuth: true, isSubApp: true },
    });
    added = true;

    if (import.meta.env.DEV) {
      console.log(`[Router] 注册子应用路由: ${routeName}`);
    }
  });

  subRoutesAdded = true;
  return added;
}

/**
 * 判断路径是否落在某个已激活模块的 activeRule 下。
 * 注意：`/auth` 和 `/auth/...` 都算，`/auth2/...` 不算。
 */
function isSubAppPath(path: string): boolean {
  const moduleStore = useModuleStore();
  return moduleStore.modules.some(
    (m) => path === `/${m.id}` || path.startsWith(`/${m.id}/`)
  );
}

router.beforeEach(async (to, _from, next) => {
  if (import.meta.env.DEV) {
    console.log(
      `[Router] 目标路径: ${to.path}, 当前匹配路由名: ${String(to.name)}`
    );
  }

  const userStore = useUserStore();
  const moduleStore = useModuleStore();
  const token = userStore.token || localStorage.getItem('token');

  // ============================================================
  // 第一步：已登录 → 保证模块已加载
  // ============================================================
  if (token && !moduleStore.loaded) {
    try {
      await moduleStore.fetchModules();
    } catch (e) {
      console.error('[Router] 模块加载失败', e);
    }
  }

  // ============================================================
  // 第二步：注册子应用路由；如果本次新增了路由，
  //        且当前 to 还是 NotFound 但实际是子应用路径 → 强制重解析
  //
  // 这是刷新子应用路径 404 的关键修复：
  //   router.isReady() 触发的首次导航早于路由注册，
  //   只能先匹配到 NotFound；注册完必须主动再导航一次。
  // ============================================================
  if (token && moduleStore.loaded && !subRoutesAdded) {
    const added = registerSubAppRoutes();

    if (added && to.name === 'NotFound' && isSubAppPath(to.path)) {
      if (import.meta.env.DEV) {
        console.log(`[Router] 重新解析子应用路径: ${to.fullPath}`);
      }
      next({ path: to.fullPath, replace: true });
      return;
    }
  }

  // ============================================================
  // 以下为原有逻辑，仅把 ignoreAuth 的判断挪到最后
  // ============================================================

  // 白名单（含 404、子应用路由）直接放行
  if (to.meta.ignoreAuth) {
    next();
    return;
  }

  // 未登录跳登录
  if (!token) {
    next({ path: '/login', query: { redirect: to.fullPath } });
    return;
  }

  // 子应用路由放行
  if (to.meta.isSubApp) {
    next();
    return;
  }

  // 主应用页面权限检查
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
