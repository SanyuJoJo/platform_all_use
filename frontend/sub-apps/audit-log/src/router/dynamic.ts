import type { Router, RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '../types';

const MODULE_ID = 'audit_log';
const COMPONENT_ALIAS: Record<string, string> = {
  // 后端返回 audit/index，子应用实际是 logs/index.vue
  'audit/index': 'logs/index',
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
  console.log(
    `[${MODULE_ID}/dynamic] 注册动态路由 ${routes.length} 条`,
    routes.map((r) => `${String(r.name)} -> ${r.path}`)
  );
}

export function resetDynamicRoutes(router: Router) {
  addedNames.forEach((name) => {
    if (router.hasRoute(name)) router.removeRoute(name);
  });
  addedNames = [];
}

export const defaultRouters: RouteNode[] = [
  {
    id: MODULE_ID,
    parentId: null,
    name: 'AuditLogRoot',
    path: '',
    microApp: MODULE_ID,
    type: 'directory',
    meta: { title: '日志审计', icon: 'Document' },
    children: [
      {
        id: `${MODULE_ID}:logs`,
        parentId: MODULE_ID,
        name: 'AuditLogs',
        path: '/audit-logs',
        component: 'audit/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '日志审计', icon: 'Document' },
      },
    ],
  },
];
