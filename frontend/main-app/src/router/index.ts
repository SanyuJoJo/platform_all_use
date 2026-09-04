import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
import DefaultLayout from '@/layouts/default/index.vue';

const routes: RouteRecordRaw[] = [
  // 不需要布局的页面
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
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/error/404.vue'),
    meta: { ignoreAuth: true },
  },

  // 需要布局的页面，给布局路由起名 'Layout'
  {
    path: '/',
    name: 'Layout',   // ✅ 关键：命名布局路由
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
      // 其他主应用页面在此添加 children
    ],
  },

  // 通配符路由（真正的 404）
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

let subRoutesAdded = false;

router.beforeEach(async (to, from, next) => {
  console.log(`[Router] 目标路径: ${to.path}`);

  const userStore = useUserStore();
  const menuStore = useMenuStore();
  const moduleStore = useModuleStore();

  // 1. 忽略认证的页面
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
    } catch (e) {
      console.error('[Router] 模块加载失败', e);
    }
  }

  // 4. 构建菜单（如果未构建）
  if (!menuStore.menuLoaded) {
    const modules = moduleStore.modules;
    if (modules && modules.length > 0) {
      await menuStore.buildMenus(modules);
    } else {
      console.warn('[Router] 模块列表为空，无法构建菜单');
    }
  }

  // 5. 添加子应用动态路由到布局的 children 中（一次性）
  if (moduleStore.loaded && !subRoutesAdded) {
    const modules = moduleStore.modules;
    modules.forEach(m => {
      const routeName = `subapp_${m.id}`;
      if (!router.hasRoute(routeName)) {
        // ✅ 关键：添加到名为 'Layout' 的父路由下
        router.addRoute('Layout', {
          path: `/${m.id}/:pathMatch(.*)*`,
          name: routeName,
          component: { render: () => null }, // 空组件，实际由 qiankun 渲染
          meta: { ignoreAuth: true, isSubApp: true },
        });
        console.log(`[Router] 添加子应用路由到 Layout: /${m.id}/:pathMatch(.*)*`);
      }
    });
    subRoutesAdded = true;
    // 重新导航以匹配新路由
    next({ ...to, replace: true });
    return;
  }

  // 6. 如果当前路径是子应用路由，直接放行
  if (to.meta.isSubApp) {
    next();
    return;
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