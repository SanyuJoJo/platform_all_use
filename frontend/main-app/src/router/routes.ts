// main-app/src/router/routes.ts
import type { RouteRecordRaw } from 'vue-router';
export const staticRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/404',
    name: 'NotFound',
    component: () => import('@/views/404/index.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      // 动态子应用路由由 qiankun 接管
      // 主应用只需定义容器路由
      {
        path: '/:pathMatch(.*)*',
        name: 'SubAppContainer',
        component: () => import('@/views/Container.vue'),
        meta: { requiresAuth: true },
      },
    ],
  },
];
