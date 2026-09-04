// src/router/index.ts
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

console.log('[SubApp Router] 路由配置加载');

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '首页' }
  },
  {
    path: '/dashboard',
    component: () => import('@/views/Dashboard.vue'),
    meta: { title: '仪表盘' }
  },
  {
    path: '/users',
    component: () => import('@/views/Users.vue'),
    meta: { title: '用户管理' }
  },
  {
    path: '/roles',
    component: () => import('@/views/Roles.vue'),
    meta: { title: '角色管理' }
  },
  // 404 重定向到首页
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
];

console.log('[SubApp Router] 路由表:', routes.map(r => r.path));

export function createAppRouter(moduleId: string, isQiankun: boolean) {
  const base = isQiankun ? `/${moduleId}` : '/';
  console.log(`[SubApp Router] 创建路由，base: ${base}, isQiankun: ${isQiankun}`);
  const router = createRouter({
    history: createWebHistory(base),
    routes,
  });

  // 添加路由守卫，打印每次跳转
  router.beforeEach((to, from, next) => {
    console.log(`[SubApp Router] 导航: ${from.path} -> ${to.path}`);
    next();
  });

  return router;
}