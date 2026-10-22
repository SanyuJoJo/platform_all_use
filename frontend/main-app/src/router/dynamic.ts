import type { RouteRecordRaw } from 'vue-router';
import type { RouteNode } from '@/types/permission';
import { MAIN_APP_IDS } from '@/types/permission';

/**
 * 双 glob：同时用相对路径和绝对路径，任一生效即可。
 */
const viewModules: Record<string, () => Promise<unknown>> = {
  ...import.meta.glob('../views/**/*.vue'),
  ...import.meta.glob('/src/views/**/*.vue'),
};

console.log('[dynamic] glob 扫描到', Object.keys(viewModules).length, '个视图文件');
if (Object.keys(viewModules).length > 0) {
  console.log('[dynamic] 前 10 个 key:', Object.keys(viewModules).slice(0, 10));
}

export function isMainAppId(id: string): boolean {
  return MAIN_APP_IDS.includes(id);
}

/**
 * component 字符串 → 动态 import。
 * 依次尝试相对/绝对两种 key，都找不到返回 null 并打日志。
 */
function resolveView(component: string): (() => Promise<unknown>) | null {
  if (!component) return null;
  const normalized = component.replace(/^\/+/, '').replace(/\.vue$/, '');
  const candidates = [
    `../views/${normalized}.vue`,
    `/src/views/${normalized}.vue`,
  ];
  for (const key of candidates) {
    if (viewModules[key]) {
      console.log(`[dynamic] resolveView('${component}') → ${key} ✓`);
      return viewModules[key];
    }
  }
  console.warn(`[dynamic] resolveView('${component}') 未找到，尝试过：`, candidates);
  return null;
}

function normalizeAbsolutePath(p: string): string {
  if (!p) return '/';
  return p.startsWith('/') ? p : `/${p}`;
}

/**
 * 节点归属判定：
 * - 自身 microApp 非空且非 main → 用自身
 * - 自身 microApp === 'main' 或空 → 继承父级 owner
 */
function resolveOwner(node: RouteNode, parentOwner: string): string {
  const app = String((node as any).microApp ?? '').trim();
  if (!app) return parentOwner;
  if (app === 'main') return parentOwner;
  return app;
}

/**
 * 兼容 menu.ts 的调用：单参时 parentOwner 默认 'main'。
 */
export function inferMicroApp(node: RouteNode, parentOwner = 'main'): string {
  return resolveOwner(node, parentOwner);
}

/**
 * 推断节点类型（兼容 menu.ts）。
 * - 有 type 字段直接用
 * - component === 'Layout' 或 children 非空 → directory
 * - 其余 → menu
 */
export function inferType(node: RouteNode): 'directory' | 'menu' {
  const declared = (node as any).type as string | undefined;
  if (declared === 'directory' || declared === 'menu') return declared;
  const comp = String((node as any).component ?? '');
  const hasChildren = Array.isArray(node.children) && node.children.length > 0;
  return comp === 'Layout' || hasChildren ? 'directory' : 'menu';
}

/**
 * 构建主应用动态路由。
 */
export function buildMainAppRoutes(tree: RouteNode[]): RouteRecordRaw[] {
  const result: RouteRecordRaw[] = [];

  const walk = (nodes: RouteNode[], parentOwner: string) => {
    nodes.forEach((node) => {
      if ((node as any).meta?.hidden) {
        console.log(`[dynamic]   ${node.name} hidden，跳过`);
        return;
      }

      const owner = resolveOwner(node, parentOwner);
      console.log(
        `[dynamic] 节点 ${node.name} | microApp=${(node as any).microApp} | owner=${owner} | type=${node.type}`
      );

      if (!isMainAppId(owner)) {
        if (node.children?.length) walk(node.children, owner);
        return;
      }

      if (node.type === 'directory') {
        if (node.children?.length) walk(node.children, owner);
        return;
      }

      const comp = String((node as any).component ?? '');
      if (!comp) {
        console.warn(`[dynamic] 主应用菜单 ${node.name} 缺少 component，已跳过`);
        return;
      }

      const view = resolveView(comp);
      if (!view) {
        console.warn(
          `[dynamic] 主应用菜单 ${node.name} 的 component=${comp} 在主应用 views 中不存在，已跳过`
        );
        return;
      }

      result.push({
        path: normalizeAbsolutePath(node.path || ''),
        name: node.name,
        component: view,
        meta: { ...(node.meta || {}), title: node.meta?.title ?? node.name },
      });
      console.log(
        `[dynamic]   + 产出路由 ${node.name} -> ${normalizeAbsolutePath(node.path || '')}`
      );
    });
  };

  walk(tree, 'main');

  console.log(
    `[dynamic] buildMainAppRoutes 完成，产出 ${result.length} 条:`,
    result.map((r) => `${String(r.name)} -> ${r.path}`)
  );
  return result;
}

export function findFirstMenuPath(tree: RouteNode[]): string | null {
  let found: string | null = null;
  const walk = (nodes: RouteNode[], parentOwner: string) => {
    for (const node of nodes) {
      if (found) return;
      if (node.meta?.hidden) continue;
      const owner = resolveOwner(node, parentOwner);
      if (node.type === 'menu' && node.path) {
        const safe = normalizeAbsolutePath(node.path);
        found = isMainAppId(owner) ? safe : `/${owner}${safe}`;
        return;
      }
      if (node.children?.length) walk(node.children, owner);
    }
  };
  walk(tree, 'main');
  return found;
}
