import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
import { 
  loadSubApp, 
  unloadCurrentApp, 
  getCurrentModuleId, 
  isSubAppLoaded 
} from '@/micro-frontend/registry';
import { message } from '@/utils/naive';

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

const isDev = import.meta.env.MODE === 'development';

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  const menuStore = useMenuStore();
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

  if (!userStore.token && token) {
    userStore.setToken(token);
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        userStore.setUser(user);
      } catch (e) { /* ignore */ }
    }
  }

  // 确保模块列表已加载
  if (!moduleStore.loaded.value) {
    await moduleStore.fetchModules().catch(() => {});
  }

  const activeIds = moduleStore.getActiveModuleIds();
  const matchedModuleId = activeIds.find(id => {
    const path = to.path;
    return path === '/' + id || path.startsWith('/' + id + '/');
  });
  const isSubApp = !!matchedModuleId;
  to.meta.isMicroApp = isSubApp;

  if (!isSubApp && to.meta.permission) {
    const hasPerm = userStore.permissions?.includes(to.meta.permission as string) ?? false;
    if (!hasPerm) {
      next('/403');
      return;
    }
  }

  // ✅ 只在菜单未加载时构建，且传入数据
  if (!menuStore.menuLoaded.value) {
    const modules = moduleStore.modules.value;
    if (modules && modules.length > 0) {
      await menuStore.buildMenus(modules);
    } else {
      // 若模块数据为空，尝试重新获取
      await moduleStore.fetchModules();
      const newModules = moduleStore.modules.value;
      if (newModules && newModules.length > 0) {
        await menuStore.buildMenus(newModules);
      }
    }
  }

  // 子应用加载/卸载逻辑
  if (isSubApp && matchedModuleId) {
    const currentId = getCurrentModuleId();
    if (!isSubAppLoaded() || currentId !== matchedModuleId) {
      const module = moduleStore.modules.value.find(m => m.id === matchedModuleId);
      if (module) {
        try {
          await loadSubApp(module, router);
          if (isDev) {
            console.log(`[Router] 子应用 ${matchedModuleId} 加载成功`);
          }
        } catch (error) {
          console.error('[Router] 加载子应用失败:', error);
          message.error(`加载模块 ${module.name || matchedModuleId} 失败`);
          next('/404');
          return;
        }
      } else {
        next('/404');
        return;
      }
    }
  } else {
    if (isSubAppLoaded()) {
      await unloadCurrentApp();
      if (isDev) {
        console.log('[Router] 已卸载子应用（切换到非子应用路由）');
      }
    }
  }

  next();
});

export default router;
