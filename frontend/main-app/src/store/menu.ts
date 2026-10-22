import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import type { MenuOption } from 'naive-ui';
import type { VNodeChild } from 'vue';
import type { RouteNode } from '@/types/permission';
import { renderIcon } from '@/utils/icon-map';
import { inferMicroApp, isMainAppId } from '@/router/dynamic';

interface MenuNode {
  key: string;
  label: string;
  icon?: () => VNodeChild;
  children?: MenuNode[];
  path?: string;
  parentId?: string | null;
  order?: number;
}

/**
 * 读取节点 id。
 * 后端字段名未统一时逐个兼容；全部取不到时用 microApp + path 组合兜底，
 * 保证 Map 的 key 唯一，不会出现「多节点被覆盖成 1 个」。
 */
function readId(n: any): string {
  const raw = n.id ?? n.menuId ?? n.menu_id ?? n.key ?? n.menuKey ?? n.menu_key;
  if (raw !== undefined && raw !== null && String(raw).trim() !== '') {
    return String(raw);
  }

  // fallback 1：microApp + path
  const app = String(n.microApp ?? n.micro_app ?? n.microapp ?? 'main');
  const p = String(n.path ?? n.route ?? n.menuPath ?? n.menu_path ?? '').replace(/^\/+/, '');
  if (p) return `__fb:${app}::${p}`;

  // fallback 2：name + index（最后兜底，理论上不会触发）
  const name = String(n.name ?? n.menuName ?? n.menu_name ?? n.title ?? '');
  return `__fb:${app}::${name}::${Math.random().toString(36).slice(2, 8)}`;
}

function readParentId(n: any): string | null {
  const raw = n.parentId ?? n.parent_id;
  if (raw === undefined || raw === null || String(raw).trim() === '') return null;
  return String(raw);
}

function readName(n: any): string {
  return String(n.name ?? n.menuName ?? n.menu_name ?? '');
}

function readTitle(n: any): string {
  return String(n.meta?.title ?? n.title ?? n.menu_title ?? readName(n));
}

function readIcon(n: any): string | undefined {
  return (n.meta?.icon ?? n.icon) as string | undefined;
}

function safeReplacer(_key: string, val: unknown): unknown {
  return typeof val === 'function' ? '[fn]' : val;
}

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);

  watch(
    menuTree,
    (v) => {
      console.log('[DIAG] menuStore.menuTree 变化, length =', v.length);
      if (v.length > 0) {
        console.log(
          '[DIAG] menuStore.menuTree 内容 =',
          JSON.stringify(v, safeReplacer, 2)
        );
      }
    },
    { deep: false }
  );

  function buildMenus(tree: RouteNode[]) {
    console.log('[DIAG] menuStore.buildMenus 调用');
    console.log(
      '[DIAG] 入参 tree 类型 =',
      Array.isArray(tree) ? 'Array' : typeof tree,
      '| length =',
      Array.isArray(tree) ? tree.length : 'N/A'
    );

    if (!Array.isArray(tree) || tree.length === 0) {
      console.warn('[DIAG] buildMenus 入参为空，菜单置空');
      menuTree.value = [];
      menuLoaded.value = true;
      return;
    }

    console.log(
      '[DIAG] buildMenus 入参第 1 个根节点 =',
      JSON.stringify(tree[0], safeReplacer, 2)
    );

    const flat: MenuNode[] = [];

    const walk = (nodes: RouteNode[], depth = 0) => {
      nodes.forEach((n) => {
        const meta = (n as any).meta;
        if (meta?.hidden) {
          console.log(
            `[DIAG]   ${'  '.repeat(depth)}隐藏节点 ${readName(n)}，递归子节点`
          );
          if (n.children?.length) walk(n.children, depth + 1);
          return;
        }

        const microApp = inferMicroApp(n);
        const rawPath = (n as any).path;
        let fullPath: string | undefined;
        if (rawPath) {
          const safe = rawPath.startsWith('/') ? rawPath : `/${rawPath}`;
          fullPath = isMainAppId(microApp) ? safe : `/${microApp}${safe}`;
        }

        const iconName = readIcon(n);
        const id = readId(n);
        const parentId = readParentId(n);
        console.log(
          `[DIAG]   ${'  '.repeat(depth)}节点 id=${id} name=${readName(n)} title=${readTitle(n)} microApp=${microApp} path=${fullPath ?? '-'} icon=${iconName ?? '-'} parentId=${parentId ?? '-'} type=${(n as any).type}`
        );

        flat.push({
          key: id,
          label: readTitle(n),
          icon: iconName ? renderIcon(iconName) : undefined,
          path: fullPath,
          parentId,
          order: (n as any).order ?? 0,
        });

        if (n.children?.length) walk(n.children, depth + 1);
      });
    };
    walk(tree);

    console.log('[DIAG] buildMenus 展平后节点数 =', flat.length);
    console.log(
      '[DIAG] buildMenus 展平数据 =',
      JSON.stringify(flat, safeReplacer, 2)
    );

    // 检测 id 重复（真正的覆盖问题会在这里暴露）
    const idCount = new Map<string, number>();
    flat.forEach((n) => {
      idCount.set(n.key, (idCount.get(n.key) ?? 0) + 1);
    });
    const dupes = Array.from(idCount.entries()).filter(([, c]) => c > 1);
    if (dupes.length > 0) {
      console.warn(
        '[DIAG] ⚠ 检测到重复 id（会被 Map 覆盖）：',
        dupes.map(([k, c]) => `${k} × ${c}`)
      );
    }

    const map = new Map<string, MenuNode>();
    flat.forEach((n) => map.set(n.key, n));
    console.log('[DIAG] buildMenus Map 节点数 =', map.size, '（展平前 =', flat.length, '）');

    const roots: MenuNode[] = [];
    map.forEach((n) => {
      if (n.parentId && map.has(n.parentId)) {
        const p = map.get(n.parentId)!;
        if (!p.children) p.children = [];
        p.children.push(n);
      } else {
        roots.push(n);
      }
    });
    console.log('[DIAG] buildMenus 根节点数 =', roots.length);

    const process = (nodes: MenuNode[]): MenuNode[] =>
      nodes
        .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
        .map((n) => {
          if (n.children?.length) n.children = process(n.children);
          else delete n.children;
          delete n.parentId;
          delete n.order;
          return n;
        });

    menuTree.value = process(roots) as unknown as MenuOption[];
    menuLoaded.value = true;

    console.log('[DIAG] buildMenus 完成, menuTree.length =', menuTree.value.length);
    if (menuTree.value.length === 0 && flat.length > 0) {
      console.warn(
        '[DIAG] ⚠ flat 有数据但 roots 为 0，说明 parentId 字段与 id 对不上'
      );
    }
  }

  function reset() {
    console.log('[DIAG] menuStore.reset');
    menuTree.value = [];
    menuLoaded.value = false;
  }

  return { menuTree, menuLoaded, buildMenus, reset };
});
