import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/store/user'
import { usePermissionStore } from '@/store/permission'
import { useMenuStore } from '@/store/menu'
import { findFirstMenuPath } from '@/utils/menu'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true },
    },
    {
      path: '/403',
      name: 'Forbidden',
      component: () => import('@/views/error/403.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'Layout',
      component: () => import('@/layouts/enterprise/index.vue'),
      children: [],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/404.vue'),
    },
  ],
})

router.beforeEach(async (to) => {
  const userStore = useUserStore()
  const permissionStore = usePermissionStore()
  const menuStore = useMenuStore()
  const token = userStore.token || localStorage.getItem('token')

  if (to.meta.public) {
    if (token && to.path === '/login') return { path: '/' }
    return true
  }

  if (!token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // 刷新后恢复用户信息（用户名、权限码），供头部与 v-permission 使用。
  // 注意：菜单树由后端 /menus/routers 按权限过滤，前端不再重复做
  // 权限校验，因此这里仅做「尽力恢复」，失败不阻塞导航。
  if (!userStore.user) {
    try {
      await userStore.fetchMe()
    } catch (e) {
      console.warn('[router] 恢复用户信息失败（忽略）', e)
    }
  }

  // 首次进入（或刷新）时拉取后端菜单并注册动态路由
  if (!permissionStore.loaded) {
    try {
      const tree = await permissionStore.fetchRouters()
      permissionStore.registerRoutes(router)
      menuStore.buildMenus(tree)
    } catch (e) {
      console.error('[router] 动态路由加载失败', e)
      userStore.logout()
      return { path: '/login' }
    }
    return { path: to.fullPath, replace: true }
  }

  if (to.path === '/') {
    const first = findFirstMenuPath(permissionStore.routeTree)
    if (first) return { path: first, replace: true }
    return true
  }

  // 后端已按权限过滤菜单树，动态路由里出现的页面均视为可访问，
  // 不再依赖客户端 permissions 数组重复校验（避免因权限数组为空误跳 403）。
  return true
})

export default router
