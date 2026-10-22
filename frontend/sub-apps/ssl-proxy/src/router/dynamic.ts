import type { Router, RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '../types';

const MODULE_ID = 'ssl_proxy';
const COMPONENT_ALIAS: Record<string, string> = {
  // 后端返回 ssl-proxy/index，子应用实际是 list/index.vue
  'ssl-proxy/index': 'list/index',
};

const viewModules = import.meta.glob('../views/**/*.vue');

function resolveView(component: string) {
  if (!component) {
    console.warn(`[${MODULE_ID}/dynamic] component 为空，降级 404`);
    return () => import('../views/404.vue');
  }
  const aliased = COMPONENT_ALIAS[component] ?? component;
  const normalized = aliased.replace(/^\/+/, '').replace(/\.vue$/, '');
  const key = `../views/${normalized}.vue`;
  if (viewModules[key]) return viewModules[key];
  console.warn(`[${MODULE_ID}/dynamic] 未找到组件: ${key}，降级 404`);
  return () => import('../views/404.vue');
}

function normalizePath(path: string, moduleId: string): string {
  let p = path || '/';
  const prefix = `/${moduleId}`;
  if (p === prefix) p = '/';
  else if (p.startsWith(prefix + '/')) p = p.slice(prefix.length);
  if (!p.startsWith('/')) p = '/' + p;
  return p;
}

function isOwnedBy(node: RouteNode, moduleId: string, parentOwned: boolean): boolean {
  const app = String((node as any).microApp ?? '').trim();
  if (app === moduleId) return true;
  if (parentOwned && (app === '' || app === 'main')) return true;
  return false;
}

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
      if (!n.component) return;
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

let addedNames: string[] = [];

export function setupDynamicRoutes(router: Router, nodes: RouteNode[], moduleId: string) {
  const routes = buildRoutes(nodes, moduleId);
  routes.forEach((r) => {
    if (r.name && router.hasRoute(r.name)) return;
    router.addRoute(r);
    if (r.name) addedNames.push(r.name as string);
  });
  console.warn(
    `[${MODULE_ID}/dynamic] 注册动态路由 ${routes.length} 条:`,
    routes.map((r) => `${String(r.name)} -> ${r.path}`)
  );
}

export function resetDynamicRoutes(router: Router) {
  addedNames.forEach((name) => {
    if (router.hasRoute(name)) router.removeRoute(name);
  });
  addedNames = [];
}

/* ============================================================================
 * 独立运行 / props.routers 缺失时的兜底路由树。
 *
 * ★ path 必须与后端 /menus/routers 返回的「模块内相对路径」一致：
 *   后端 SslProxyList.path = "/ssl_proxy/list"
 *   normalizePath 剥离 /ssl_proxy 前缀后 → "/list"
 *   所以 defaultRouters 的 path 也必须是 "/list"
 *
 *   之前写 "/ssl-proxy"（连字符）会与 base /ssl_proxy（下划线）不匹配，
 *   导致 normalizePath 不剥离，注册 path 变成 "/ssl-proxy"，
 *   与 URL /ssl_proxy/list 剥离后的 "/list" 不一致 → 404。
 * ========================================================================== */
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
        path: '/list',           // ★ 从 '/ssl-proxy' 改成 '/list'
        component: 'ssl-proxy/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: 'HTTP 管理', icon: 'Connection' },
      },
    ],
  },
];
