import type { Router, RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '../types';

/** 子应用视图映射表（相对路径 + 绝对路径双保险） */
const viewModules: Record<string, () => Promise<unknown>> = {
  ...import.meta.glob('../views/**/*.vue'),
  ...import.meta.glob('/src/views/**/*.vue'),
};

/**
 * 后端 component → 子应用 views 实际路径映射。
 *
 * 后端 /menus/routers 返回的 component 与 auth 子应用实际目录不一致：
 *   后端 system/user/index       → 实际 users/index.vue
 *   后端 system/role/index       → 实际 roles/index.vue
 *   后端 system/permission/index → 实际 permissions/index.vue
 *   后端 system/menu/index       → 实际无对应文件（降级到 404，见 resolveView）
 */
const COMPONENT_ALIAS: Record<string, string> = {
  'system/user/index': 'users/index',
  'system/role/index': 'roles/index',
  'system/permission/index': 'permissions/index',
  // system/menu/index 无对应文件，保留原值走降级路径
};

function resolveView(component: string): (() => Promise<unknown>) | null {
  if (!component) {
    console.warn('[auth/dynamic] component 为空');
    return null;
  }
  const aliased = COMPONENT_ALIAS[component] ?? component;
  const normalized = aliased.replace(/^\/+/, '').replace(/\.vue$/, '');
  const candidates = [
    `../views/${normalized}.vue`,
    `/src/views/${normalized}.vue`,
  ];
  for (const key of candidates) {
    if (viewModules[key]) {
      console.log(`[auth/dynamic] resolveView('${component}') → ${key} ✓`);
      return viewModules[key];
    }
  }
  console.warn(
    `[auth/dynamic] 未找到组件: '${component}' (alias='${aliased}')，尝试过：`,
    candidates
  );
  return null;
}

/**
 * 路径规范化：
 * - 剥离 /{moduleId} 前缀（由 createWebHistory(basename) 负责该前缀）
 * - 保证以 / 开头
 */
function normalizePath(path: string, moduleId: string): string {
  let p = path || '/';
  const prefix = `/${moduleId}`;
  if (p === prefix) p = '/';
  else if (p.startsWith(prefix + '/')) p = p.slice(prefix.length);
  if (!p.startsWith('/')) p = '/' + p;
  return p;
}

/**
 * 子应用归属判定：
 * - 自身 microApp === moduleId → 是
 * - 父节点属于本应用、自身 microApp 为 'main' 或空 → 继承父节点（兼容后端 children 未继承父目录的旧数据）
 */
function isOwnedBy(
  node: RouteNode,
  moduleId: string,
  parentOwned: boolean
): boolean {
  const app = String((node as any).microApp ?? '').trim();
  if (app === moduleId) return true;
  if (parentOwned && (app === '' || app === 'main')) return true;
  return false;
}

/**
 * 从完整路由树中筛选归属本应用的叶子节点。
 *
 * 关键设计：component 找不到时**跳过**该节点（不再降级 404），
 * 让用户点该菜单时落到 SubNotFound 或者根本不显示该菜单。
 * 这样 `system/menu/index` 缺失时不会把整个子应用搞挂。
 */
export function buildRoutes(nodes: RouteNode[], moduleId: string): RouteRecordRaw[] {
  const out: RouteRecordRaw[] = [];

  const walk = (list: RouteNode[], parentOwned: boolean) => {
    list.forEach((n) => {
      const owned = isOwnedBy(n, moduleId, parentOwned);

      if (!owned) {
        if (n.children?.length) walk(n.children, false);
        return;
      }

      if (n.meta?.hidden) {
        if (n.children?.length) walk(n.children, owned);
        return;
      }
      if (n.type === 'directory') {
        if (n.children?.length) walk(n.children, owned);
        return;
      }
      if (!n.component) {
        console.warn(`[auth/dynamic] 节点 ${n.name} 缺少 component，跳过`);
        return;
      }

      const view = resolveView(n.component);
      if (!view) {
        console.warn(
          `[auth/dynamic] 节点 ${n.name} (component=${n.component}) 无对应视图，跳过注册`
        );
        return;
      }

      out.push({
        path: normalizePath(n.path || '', moduleId),
        name: n.name,
        component: view,
        meta: (n.meta ?? { title: n.name }) as Record<PropertyKey, unknown>,
      });
    });
  };

  walk(nodes, false);
  return out;
}

/** 已注册的动态路由名，用于 reset */
let addedNames: string[] = [];

export function setupDynamicRoutes(
  router: Router,
  nodes: RouteNode[],
  moduleId: string
) {
  const routes = buildRoutes(nodes, moduleId);
  routes.forEach((r) => {
    if (r.name && router.hasRoute(r.name)) return;
    router.addRoute(r);
    if (r.name) addedNames.push(r.name as string);
  });
  console.log(
    `[auth/dynamic] 注册动态路由 ${routes.length} 条:`,
    routes.map((r) => `${String(r.name)} -> ${r.path}`)
  );

  const allNames = router.getRoutes().map((r) => String(r.name));
  console.log('[auth/dynamic] 当前路由表:', allNames);
}

export function resetDynamicRoutes(router: Router) {
  addedNames.forEach((name) => {
    if (router.hasRoute(name)) router.removeRoute(name);
  });
  addedNames = [];
}

/* ============================================================================
 * 独立运行（未嵌入 qiankun）时的默认路由树。
 * ========================================================================== */
export const defaultRouters: RouteNode[] = [
  {
    id: 'auth',
    parentId: null,
    name: 'AuthRoot',
    path: '',
    microApp: 'auth',
    type: 'directory',
    meta: { title: '认证中心', icon: 'Shield' },
    order: 1,
    children: [
      {
        id: 'auth:users',
        parentId: 'auth',
        name: 'AuthUsers',
        path: '/users',
        component: 'users/index',
        microApp: 'auth',
        type: 'menu',
        meta: { title: '用户管理', icon: 'People', permission: 'auth:user:list' },
        order: 1,
      },
      {
        id: 'auth:roles',
        parentId: 'auth',
        name: 'AuthRoles',
        path: '/roles',
        component: 'roles/index',
        microApp: 'auth',
        type: 'menu',
        meta: { title: '角色管理', icon: 'Shield', permission: 'auth:role:list' },
        order: 2,
      },
      {
        id: 'auth:permissions',
        parentId: 'auth',
        name: 'AuthPermissions',
        path: '/permissions',
        component: 'permissions/index',
        microApp: 'auth',
        type: 'menu',
        meta: { title: '权限管理', icon: 'Key', permission: 'auth:permission:list' },
        order: 3,
      },
    ],
  },
];
