import type { Router, RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '../types';

const MODULE_ID = 'crypto_console';

/**
 * 后端 component 与 views 实际路径的映射。
 *
 * 后端现已返回 7 个子菜单（keys/ca/certs/csr/crl/audits/tasks），
 * component 与 views 一一对应，无需 alias。
 * 如未来后端改回 crypto/index 或字段变化，在此追加。
 */
const COMPONENT_ALIAS: Record<string, string> = {};

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

/* 独立运行时的默认路由（qiankun 下由 props.routers 覆盖） */
export const defaultRouters: RouteNode[] = [
  {
    id: MODULE_ID,
    parentId: null,
    name: 'CryptoConsoleRoot',
    path: '',
    microApp: MODULE_ID,
    type: 'directory',
    meta: { title: '密码服务', icon: 'Lock' },
    children: [
      {
        id: `${MODULE_ID}:keys`,
        parentId: MODULE_ID,
        name: 'CryptoKeys',
        path: '/crypto/keys',
        component: 'keys/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '密钥管理', icon: 'Key' },
      },
      {
        id: `${MODULE_ID}:ca`,
        parentId: MODULE_ID,
        name: 'CryptoCA',
        path: '/crypto/ca',
        component: 'ca/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: 'CA 管理', icon: 'Shield' },
      },
      {
        id: `${MODULE_ID}:certs`,
        parentId: MODULE_ID,
        name: 'CryptoCerts',
        path: '/crypto/certs',
        component: 'certs/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '证书管理', icon: 'Ribbon' },
      },
      {
        id: `${MODULE_ID}:csr`,
        parentId: MODULE_ID,
        name: 'CryptoCSR',
        path: '/crypto/csr',
        component: 'csr/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: 'CSR 管理', icon: 'Document' },
      },
      {
        id: `${MODULE_ID}:crl`,
        parentId: MODULE_ID,
        name: 'CryptoCRL',
        path: '/crypto/crl',
        component: 'crl/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: 'CRL 管理', icon: 'Document' },
      },
      {
        id: `${MODULE_ID}:audits`,
        parentId: MODULE_ID,
        name: 'CryptoAudits',
        path: '/crypto/audits',
        component: 'audits/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '操作审计', icon: 'Document' },
      },
      {
        id: `${MODULE_ID}:tasks`,
        parentId: MODULE_ID,
        name: 'CryptoTasks',
        path: '/crypto/tasks',
        component: 'tasks/index',
        microApp: MODULE_ID,
        type: 'menu',
        meta: { title: '任务管理', icon: 'List' },
      },
    ],
  },
];
