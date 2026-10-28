// src/router/dynamic.ts
//
// ssl-proxy 子应用动态路由构建。
//
// 设计要点：
//   1. 后端 /menus/routers 下发路由树，主应用通过 props.routers 传给子应用；
//   2. 子应用 basename = /ssl_proxy，URL /ssl_proxy/tcps/list 剥离后为 /tcps/list；
//      因此后端菜单 path 必须是「模块内相对路径」（/tcps/list），
//      若写成 /ssl_proxy/tcps/list 会被 normalizePath 剥离 /ssl_proxy 后仍正确，
//      但推荐统一写相对路径，减少歧义；
//   3. TCPS 与 HTTP 是两个独立菜单节点，component 分别指向
//      ssl-proxy/index（HTTP）与 tcps/list/index（TCPS）；
//   4. TCPS 页面与 HTTP 页面代码完全隔离，后续差异化不会互相牵连。
import type { Router, RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '../types';

const MODULE_ID = 'ssl_proxy';

/**
 * 后端 component → 子应用 views 实际路径映射。
 *
 * 目前两者一致，保留 alias 表是为了：
 *   - 后端若沿用旧字段（ssl-proxy/index）时仍能解析；
 *   - 未来 TCPS 若拆分多个页面（如 tcps/list/index、tcps/detail/index），
 *     可在此显式登记，避免 resolveView 依赖隐式路径约定。
 */
const COMPONENT_ALIAS: Record<string, string> = {
  // HTTP 管理（历史兼容）
  'ssl-proxy/index': 'list/index',
  'list/index': 'list/index',
  // TCPS 管理（独立目录）
  'tcps/list/index': 'tcps/list/index',
  'tcps/detail/index': 'tcps/detail/index',
};

const viewModules = import.meta.glob('../views/**/*.vue');

/**
 * component 字符串 → 动态 import。
 * 依次尝试 alias / 相对 / 绝对三种 key，都找不到降级 404。
 */
function resolveView(component: string) {
  if (!component) {
    console.warn(`[${MODULE_ID}/dynamic] component 为空，降级 404`);
    return () => import('../views/404.vue');
  }
  const aliased = COMPONENT_ALIAS[component] ?? component;
  const normalized = aliased.replace(/^\/+/, '').replace(/\.vue$/, '');
  const key = `../views/${normalized}.vue`;
  if (viewModules[key]) return viewModules[key];

  // 兜底：绝对路径 key（某些构建场景下相对 key 不匹配）
  const absKey = `/src/views/${normalized}.vue`;
  if (viewModules[absKey]) return viewModules[absKey];

  console.warn(
    `[${MODULE_ID}/dynamic] 未找到组件: '${component}' (alias='${aliased}')，尝试过：`,
    [key, absKey],
  );
  return () => import('../views/404.vue');
}

/**
 * 路径规范化：
 *   - 剥离 /{moduleId} 前缀（base=/ssl_proxy 负责该前缀）；
 *   - 保证以 / 开头。
 *
 * 例：'/ssl_proxy/tcps/list' → '/tcps/list'
 *     '/tcps/list'           → '/tcps/list'
 *     ''                     → '/'
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
 *   - 自身 microApp === moduleId → 是；
 *   - 父节点属于本应用、自身 microApp 为 'main' 或空 → 继承父节点；
 *     （兼容后端 children 未继承父目录 microApp 的旧数据）
 */
function isOwnedBy(
  node: RouteNode,
  moduleId: string,
  parentOwned: boolean,
): boolean {
  const app = String((node as any).microApp ?? '').trim();
  if (app === moduleId) return true;
  if (parentOwned && (app === '' || app === 'main')) return true;
  return false;
}

/**
 * 从完整路由树中筛选归属本应用的叶子节点，构建 vue-router 记录。
 *
 * 跳过规则：
 *   - 非本应用节点：若还有 children，继续递归查找；
 *   - meta.hidden：不注册路由（也不进菜单）；
 *   - type === 'directory'：不注册自身，只递归 children；
 *   - 缺 component 或 resolveView 失败：跳过（不降级 404，避免污染路由表）。
 */
export function buildRoutes(
  nodes: RouteNode[],
  moduleId: string,
): RouteRecordRaw[] {
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
        console.warn(
          `[${MODULE_ID}/dynamic] 节点 ${n.name} 缺少 component，跳过`,
        );
        return;
      }

      out.push({
        path: normalizePath(n.path || '', moduleId),
        name: n.name,
        component: resolveView(n.component),
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
  moduleId: string,
) {
  const routes = buildRoutes(nodes, moduleId);
  routes.forEach((r) => {
    if (r.name && router.hasRoute(r.name)) return;
    router.addRoute(r);
    if (r.name) addedNames.push(r.name as string);
  });
  console.warn(
    `[${MODULE_ID}/dynamic] 注册动态路由 ${routes.length} 条:`,
    routes.map((r) => `${String(r.name)} -> ${r.path}`),
  );
}

export function resetDynamicRoutes(router: Router) {
  addedNames.forEach((name) => {
    if (router.hasRoute(name)) router.removeRoute(name);
  });
  addedNames = [];
}

/**
 * 独立运行 / props.routers 缺失时的兜底路由树。
 *
 * ★ path 必须与后端 /menus/routers 返回的「模块内相对路径」一致：
 *   - HTTP 菜单：后端 path 推荐 "/list"
 *   - TCPS 菜单：后端 path 推荐 "/tcps/list"
 *
 *   子应用 basename = /ssl_proxy：
 *     URL /ssl_proxy/list       → 剥离后 /list       → 命中 SslProxyList
 *     URL /ssl_proxy/tcps/list  → 剥离后 /tcps/list  → 命中 TcpsProxyList
 *
 *   历史坑：曾把 HTTP 写成 "/ssl-proxy"（连字符），与 base 下划线不匹配，
 *   导致 normalizePath 不剥离、注册路径错位 → 404。这里统一使用下划线模块名
 *   与相对路径，不再复现。
 */
export const defaultRouters: RouteNode[] = [
  {
    id: MODULE_ID,
    parentId: null,
    name: 'SslProxyRoot',
    path: '',
    microApp: MODULE_ID,
    type: 'directory',
    meta: { title: 'SSL 代理', icon: 'Connection' },
    children: [
      {
        id: `${MODULE_ID}:list`,
        parentId: MODULE_ID,
        name: 'SslProxyList',
        path: '/list',
        component: 'ssl-proxy/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: 'HTTP 管理', icon: 'Connection' },
      },
      {
        id: `${MODULE_ID}:tcps-list`,
        parentId: MODULE_ID,
        name: 'TcpsProxyList',
        path: '/tcps/list',
        component: 'tcps/list/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: 'TCPS 管理', icon: 'Connection' },
      },
    ],
  },
];
