import type { Router, RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '../types';
const MODULE_ID = 'media';
const viewModules = import.meta.glob('../views/**/*.vue');
function resolveView(component: string) {
  if (!component) return () => import('../views/404.vue');
  const normalized = component.replace(/^\/+/, '').replace(/\.vue$/, '');
  const key = `../views/${normalized}.vue`;
  if (viewModules[key]) return viewModules[key];
  console.warn(`[${MODULE_ID}/dynamic] 未找到组件: ${key}`);
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
      if (!owned) { if (n.children?.length) walk(n.children, false); return; }
      if (n.meta?.hidden) { if (n.children?.length) walk(n.children, owned); return; }
      if (n.type === 'directory') { if (n.children?.length) walk(n.children, owned); return; }
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
  buildRoutes(nodes, moduleId).forEach((r) => {
    if (r.name && router.hasRoute(r.name)) return;
    router.addRoute(r);
    if (r.name) addedNames.push(r.name as string);
  });
}
export function resetDynamicRoutes(router: Router) {
  addedNames.forEach((name) => { if (router.hasRoute(name)) router.removeRoute(name); });
  addedNames = [];
}
export const defaultRouters: RouteNode[] = [
  {
    id: MODULE_ID,
    parentId: null,
    name: 'MediaRoot',
    path: '',
    microApp: MODULE_ID,
    type: 'directory',
    meta: { title: '音视频管理', icon: 'Video' },
    children: [
      {
        id: `${MODULE_ID}:list`,
        parentId: MODULE_ID,
        name: 'MediaList',
        path: '/media',
        component: 'media/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '媒体库', icon: 'FolderOpen' },
      },
      {
        id: `${MODULE_ID}:live`,
        parentId: MODULE_ID,
        name: 'MediaLive',
        path: '/live',
        component: 'live/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '直播管理', icon: 'Pulse' },
      },
    ],
  },
];
