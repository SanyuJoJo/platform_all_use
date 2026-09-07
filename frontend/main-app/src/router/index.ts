import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
import DefaultLayout from '@/layouts/default/index.vue';

// 白名单路由（无需认证，无需 Layout）
const publicRoutes: RouteRecordRaw[] = [
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
];

// Layout 路由：匹配所有需要认证的路径
const layoutRoute: RouteRecordRaw = {
  path: '/',
  name: 'Layout',
  component: DefaultLayout,
  meta: { requiresAuth: true },
  children: [
    {
      path: '',
      redirect: '/dashboard',
    },
    {
      path: 'dashboard',
      name: 'Dashboard',
      component: () => import('@/views/dashboard/index.vue'),
      meta: { title: '仪表盘', permission: 'dashboard:view' },
    },
    // 子应用路由将在 beforeEach 中动态添加
    // 404 fallback - 必须放在最后
    {
      path: ':pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/404.vue'),
      meta: { ignoreAuth: true },
    },
  ],
};

const routes: RouteRecordRaw[] = [...publicRoutes, layoutRoute];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

let subRoutesAdded = false;

function isSubAppPath(path: string, modules: any[]) {
  return modules.some(m => path.startsWith(`/${m.id}`));
}

// 获取子应用路由名称
function getSubAppRouteName(moduleId: string): string {
  return `subapp_${moduleId}`;
}

router.beforeEach(async (to, from, next) => {
  console.log(`[Router] 目标路径: ${to.path}, 当前匹配路由名: ${to.name}`);

  const userStore = useUserStore();
  const menuStore = useMenuStore();
  const moduleStore = useModuleStore();

  // 1. 白名单直接放行
  if (to.meta.ignoreAuth) {
    next();
    return;
  }

  // 2. 检查登录
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

  // 4. 添加所有子应用路由到 Layout（仅首次，但每次检查缺失）
  if (moduleStore.loaded) {
    const modules = moduleStore.modules;
    let needRedirect = false;
    modules.forEach(m => {
      const routeName = getSubAppRouteName(m.id);
      if (!router.hasRoute(routeName)) {
        console.log(`[Router] 添加子应用路由: ${routeName}`);
        router.addRoute('Layout', {
          path: `/${m.id}/:pathMatch(.*)*`,
          name: routeName,
          component: { render: () => null },
          meta: { ignoreAuth: true, isSubApp: true },
        });
        needRedirect = true;
      }
    });
    // 如果添加了新路由，且当前路径是子应用路径，重定向以触发匹配
    if (needRedirect && isSubAppPath(to.path, modules)) {
      console.log(`[Router] 添加路由后重定向到 ${to.path}`);
      next(to.path);
      return;
    }
    subRoutesAdded = true;
  }

  // 5. 如果当前是子应用路由，直接放行
  if (to.meta.isSubApp) {
    console.log('[Router] 子应用路由，直接放行');
    next();
    return;
  }

  // 6. 如果当前是子应用路径但路由名不是子应用路由（即匹配到 NotFound），重试
  if (moduleStore.loaded && isSubAppPath(to.path, moduleStore.modules) && to.name === 'NotFound') {
    const matchedModule = moduleStore.modules.find(m => to.path.startsWith(`/${m.id}`));
    if (matchedModule) {
      const routeName = getSubAppRouteName(matchedModule.id);
      if (router.hasRoute(routeName)) {
        console.log(`[Router] 子应用路由存在但未匹配，强制重定向到 ${to.path}`);
        next(to.path);
        return;
      }
    }
  }

  // 7. 主应用页面权限检查
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