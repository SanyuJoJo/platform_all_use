import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Router } from 'vue-router';
import type { RouteNode } from '@/types/permission';
import { getRouters } from '@/api/permission';
import { buildMainAppRoutes } from '@/router/dynamic';

export const usePermissionStore = defineStore('permission', () => {
  const routeTree = ref<RouteNode[]>([]);
  const loaded = ref(false);
  const loading = ref(false);
  let removeRouteFns: Array<() => void> = [];

  async function fetchRouters(force = false): Promise<RouteNode[]> {
    if (loaded.value && !force && routeTree.value.length > 0) {
      return routeTree.value;
    }
    if (loading.value) {
      while (loading.value) await new Promise((r) => setTimeout(r, 50));
      return routeTree.value;
    }
    loading.value = true;
    try {
      const tree = await getRouters();
      routeTree.value = tree;
      loaded.value = true;
      console.log(`[PermissionStore] fetchRouters 成功, 根节点数 = ${tree.length}`);
      return tree;
    } finally {
      loading.value = false;
    }
  }

  /**
   * 注册主应用动态路由。
   * 每一步都打日志，方便定位 addRoute 是否成功。
   */
  function registerRoutes(router: Router) {
    console.log('[PermissionStore] ===== registerRoutes 开始 =====');

    // 清理旧的动态路由
    if (removeRouteFns.length > 0) {
      console.log(`[PermissionStore] 清理旧动态路由 ${removeRouteFns.length} 条`);
    }
    removeRouteFns.forEach((fn) => fn());
    removeRouteFns = [];

    // 构建路由
    const routes = buildMainAppRoutes(routeTree.value);
    console.log(`[PermissionStore] buildMainAppRoutes 返回 ${routes.length} 条路由`);

    if (routes.length === 0) {
      console.error(
        '[PermissionStore] ⚠️ buildMainAppRoutes 返回空！请检查上方 [dynamic] 的日志'
      );
      return;
    }

    // 检查 Layout 是否存在（addRoute 的 parentName 必须存在）
    const layoutExists = router.hasRoute('Layout');
    console.log(`[PermissionStore] router.hasRoute('Layout') = ${layoutExists}`);
    if (!layoutExists) {
      console.error(
        '[PermissionStore] ⚠️ Layout 路由不存在，无法 addRoute。请检查 router/index.ts 静态路由定义'
      );
      return;
    }

    // 逐条注册
    routes.forEach((r) => {
      try {
        const remove = router.addRoute('Layout', r);
        removeRouteFns.push(remove);
        console.log(
          `[PermissionStore]   + addRoute OK: ${String(r.name)} -> ${r.path}`
        );
      } catch (e) {
        console.error(
          `[PermissionStore]   ✗ addRoute 失败: ${String(r.name)} -> ${r.path}`,
          e
        );
      }
    });

    // 验证注册结果
    const allRoutes = router.getRoutes();
    console.log(`[PermissionStore] addRoute 后 router 总路由数 = ${allRoutes.length}`);
    console.log('[PermissionStore] 当前完整路由表:');
    allRoutes.forEach((r) => {
      console.log(`[PermissionStore]   [${String(r.name).padEnd(28)}] ${r.path}`);
    });

    // 关键断言：dashboard 必须存在
    const hasDashboard = router.hasRoute('Dashboard');
    console.log(`[PermissionStore] 断言 router.hasRoute('Dashboard') = ${hasDashboard}`);
    if (!hasDashboard) {
      console.error(
        '[PermissionStore] ⚠️ /dashboard 未注册！这就是刷新后 404 的根本原因。'
      );
      console.error(
        '[PermissionStore] 请检查：[1] import.meta.glob 是否扫描到 dashboard/index.vue ' +
          '[2] views/dashboard/index.vue 文件是否存在'
      );
    }

    console.log('[PermissionStore] ===== registerRoutes 结束 =====');
  }

  function reset() {
    removeRouteFns.forEach((fn) => fn());
    removeRouteFns = [];
    routeTree.value = [];
    loaded.value = false;
  }

  return { routeTree, loaded, loading, fetchRouters, registerRoutes, reset };
});
