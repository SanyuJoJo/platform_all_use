import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
import DefaultLayout from '@/layouts/default/index.vue';

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
    component: DefaultLayout,
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
let redirecting = false;

function isSubAppPath(path: string, modules: any[]) {
  return modules.some(m => path.startsWith(`/${m.id}`));
}

router.beforeEach(async (to, from, next) => {
  console.log(`[Router] 目标路径: ${to.path}, 当前匹配路由名: ${to.name}`);

  const userStore = useUserStore();
  const menuStore = useMenuStore();
  const moduleStore = useModuleStore();

  // 1. 对于 ignoreAuth 的路由（登录、403、404）
  if (to.meta.ignoreAuth) {
    // 如果当前匹配的是 NotFound 且路径可能为子应用路径，动态添加路由
    if (to.name === 'NotFound' && !redirecting) {
      if (!moduleStore.loaded) {
        try {
          await moduleStore.fetchModules();
          console.log('[Router] 模块加载完成 (在 NotFound 分支)');
        } catch (e) {
          console.error('[Router] 模块加载失败', e);
          next();
          return;
        }
      }
      if (moduleStore.loaded && isSubAppPath(to.path, moduleStore.modules)) {
        const matchedModule = moduleStore.modules.find(m => to.path.startsWith(`/${m.id}`));
        if (matchedModule) {
          const routeName = `subapp_${matchedModule.id}`;
          if (!router.hasRoute(routeName)) {
            console.log(`[Router] 动态添加子应用路由 (NotFound分支): /${matchedModule.id}/:pathMatch(.*)*`);
            router.addRoute('Layout', {
              path: `/${matchedModule.id}/:pathMatch(.*)*`,
              name: routeName,
              component: { render: () => null },
              meta: { ignoreAuth: true, isSubApp: true },
            });
            redirecting = true;
            next(to.path);
            return;
          }
        }
      }
    }
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

  // 4. 添加子应用动态路由（仅首次）
  if (moduleStore.loaded && !subRoutesAdded) {
    console.log('[Router] 开始添加子应用路由...');
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
        console.log(`[Router] 添加子应用路由到 Layout: /${m.id}/:pathMatch(.*)*, 成功? ${router.hasRoute(routeName)}`);
      }
    });
    subRoutesAdded = true;
    console.log(`[Router] 重定向到 ${to.path} 以匹配新路由`);
    redirecting = true;
    next(to.path);
    return;
  }

  // 5. 如果当前是子应用路由，直接放行（菜单已在 main.ts 中提前构建）
  if (to.meta.isSubApp) {
    console.log('[Router] 子应用路由，直接放行');
    redirecting = false;
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

  redirecting = false;
  next();
});

export default router;