import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useModuleStore } from '@/store/module';
import { usePermissionStore } from '@/store/permission';
import { useMenuStore } from '@/store/menu';
import { findFirstMenuPath } from './dynamic';

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
    children: [],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

const MAIN_APP_IDS = ['main', 'platform', 'main-app', 'root'];
let subRoutesAdded = false;

export function setupSubAppRoutes(): boolean {
  if (subRoutesAdded) return false;
  const moduleStore = useModuleStore();
  if (!moduleStore.loaded) return false;

  let added = false;
  moduleStore.modules.forEach((m) => {
    if (MAIN_APP_IDS.includes(m.id)) return;
    const routeName = `subapp_${m.id}`;
    if (router.hasRoute(routeName)) return;
    router.addRoute('Layout', {
      path: `/${m.id}/:pathMatch(.*)*`,
      name: routeName,
      component: { render: () => null },
      meta: { ignoreAuth: true, isSubApp: true },
    });
    added = true;
    console.log(`[Router] 注册子应用 catch-all: ${routeName} -> /${m.id}/**`);
  });
  subRoutesAdded = true;
  if (added) console.log('[Router] 全部子应用 catch-all 注册完成');
  return added;
}

function isSubAppPath(path: string): boolean {
  const moduleStore = useModuleStore();
  if (!moduleStore.loaded) return false;
  return moduleStore.modules.some(
    (m) => path === `/${m.id}` || path.startsWith(`/${m.id}/`)
  );
}

router.beforeEach(async (to, _from, next) => {
  console.log(
    `[Router] 守卫: to.path=${to.path} to.name=${String(to.name)} matched=${to.matched.map((m) => String(m.name)).join('/') || '(空)'}`
  );

  const userStore = useUserStore();
  const moduleStore = useModuleStore();
  const permissionStore = usePermissionStore();
  const menuStore = useMenuStore();
  const token = userStore.token || localStorage.getItem('token');

  // ============================================================
  // 1. 登录页：已登录跳首页，未登录放行
  // ============================================================
  if (to.path === '/login') {
    if (token) {
      console.log('[Router] 已登录访问登录页，跳首页');
      next('/');
      return;
    }
    console.log('[Router] 未登录访问登录页，放行');
    next();
    return;
  }

  // ============================================================
  // 2. 403 页：直接放行（权限不足跳转目标，不需要鉴权）
  // ============================================================
  if (to.path === '/403') {
    console.log('[Router] 403 页放行');
    next();
    return;
  }

  // ============================================================
  // 3. 无 token：跳登录（携带 redirect）
  // ============================================================
  if (!token) {
    console.log('[Router] 无 token，跳登录');
    next({ path: '/login', query: { redirect: to.fullPath } });
    return;
  }

  // ============================================================
  // 4. 已登录：确保模块列表已加载（子应用注册、catch-all 依赖）
  // ============================================================
  if (!moduleStore.loaded) {
    try {
      await moduleStore.fetchModules();
    } catch (e) {
      console.error('[Router] 模块加载失败', e);
    }
  }

  // ============================================================
  // 5. 已登录：确保权限路由已加载 + 注册
  //
  // ★ 关键修复：此判断必须在 ignoreAuth 之前。
  //   首次刷新 /dashboard 时，to 因为动态路由未注册而匹配 NotFound，
  //   如果先判断 ignoreAuth 会被提前放行，导致 currentRoute 锁定 NotFound。
  // ============================================================
  if (!permissionStore.loaded) {
    try {
      await permissionStore.fetchRouters();
      permissionStore.registerRoutes(router);
      menuStore.buildMenus(permissionStore.routeTree);
    } catch (e) {
      console.error('[Router] 权限路由加载失败，强制登出', e);
      userStore.logout();
      next({ path: '/login' });
      return;
    }
    console.log(`[Router] 动态路由注册完成，重新导航: ${to.fullPath}`);
    next({ path: to.fullPath, replace: true });
    return;
  }

  // ============================================================
  // 6. 兜底：菜单未构建但路由树有数据
  // ============================================================
  if (!menuStore.menuLoaded && permissionStore.routeTree.length > 0) {
    menuStore.buildMenus(permissionStore.routeTree);
  }

  // ============================================================
  // 7. 子应用 catch-all 注册：首次访问子应用路径时补注册
  // ============================================================
  if (!subRoutesAdded && isSubAppPath(to.path)) {
    setupSubAppRoutes();
    console.log(`[Router] 子应用路径 ${to.path}，重新导航`);
    next({ path: to.fullPath, replace: true });
    return;
  }

  // ============================================================
  // 8. 根路径重定向到第一个可访问菜单
  // ============================================================
  if (to.path === '/') {
    const first = findFirstMenuPath(permissionStore.routeTree);
    if (first) {
      console.log('[Router] 根路径重定向到:', first);
      next({ path: first, replace: true });
      return;
    }
    next();
    return;
  }

  // ============================================================
  // 9. 子应用路由放行（权限由子应用内部校验）
  // ============================================================
  if (to.meta.isSubApp) {
    console.log('[Router] 子应用路由放行:', to.path);
    next();
    return;
  }

  // ============================================================
  // 10. 页面级权限
  // ============================================================
  if (to.meta.permission) {
    const hasPerm =
      userStore.permissions?.includes(to.meta.permission as string) ?? false;
    if (!hasPerm) {
      console.log('[Router] 权限不足，跳 403:', to.meta.permission);
      next('/403');
      return;
    }
  }

  // ============================================================
  // 11. 兜底：此时若仍匹配 NotFound，说明该路径确实不存在
  // ============================================================
  if (to.name === 'NotFound') {
    console.log('[Router] 目标路径不存在（动态路由已注册完毕）:', to.path);
  }

  console.log('[Router] 放行:', to.path, '| name =', String(to.name));
  next();
});

export function resetRouter() {
  usePermissionStore().reset();
  subRoutesAdded = false;
}

export default router;
