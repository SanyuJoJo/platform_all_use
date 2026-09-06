// src/router/index.ts
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';
import { useUserStore } from '@/store/user';
import { createDiscreteApi } from 'naive-ui';

const { message } = createDiscreteApi(['message']);

const routes: RouteRecordRaw[] = [
  // 默认首页重定向到用户列表
  {
    path: '/',
    redirect: '/users'
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('@/views/users/index.vue'),
    meta: { title: '用户管理', permission: 'auth:user:view' },
  },
  {
    path: '/roles',
    name: 'Roles',
    component: () => import('@/views/roles/index.vue'),
    meta: { title: '角色管理', permission: 'auth:role:view' },
  },
  {
    path: '/permissions',
    name: 'Permissions',
    component: () => import('@/views/permissions/index.vue'),
    meta: { title: '权限管理', permission: 'auth:permission:view' },
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/403.vue'),
    meta: { title: '无权限' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/404.vue'),
  },
];

export function createAppRouter(moduleId: string, isQiankun: boolean) {
  const base = isQiankun ? `/${moduleId}` : '/';
  console.log(`[Auth Router] 创建路由，base: ${base}, isQiankun: ${isQiankun}`);
  const router = createRouter({
    history: createWebHistory(base),
    routes,
  });

  // 路由守卫：权限校验
  router.beforeEach((to, from, next) => {
    const permission = to.meta.permission as string | undefined;
    if (permission) {
      const userStore = useUserStore();
      if (!userStore.hasPermission(permission)) {
        console.warn(`[Auth] 无权限访问 ${to.path}，需要 ${permission}`);
        if (window.__POWERED_BY_QIANKUN__) {
          next('/403');
        } else {
          message.error('您没有权限访问该页面');
          next(false);
        }
        return;
      }
    }
    next();
  });

  return router;
}
