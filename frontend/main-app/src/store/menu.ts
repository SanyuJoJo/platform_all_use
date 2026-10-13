// main-app/src/store/menu.ts
import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MenuOption } from 'naive-ui';
import type { VNodeChild } from 'vue';
import { useUserStore } from './user';
import { useModuleStore } from './module';
import { message } from '@/utils/naive';
import type { Menu, Module } from '@/types/module';
import { renderIcon } from '@/utils/icon-map';
import { eventBus } from '@/micro-frontend/event-bus';

/**
 * 保留给主应用自身的模块 ID。
 *
 * 语义：
 * - 这些模块不会注册为 qiankun 子应用；
 * - 它们的 menus[].path 已经是主应用完整的路由路径（如 /dashboard），
 *   构建菜单时不能再拼 /{module.id} 前缀。
 */
const RESERVED_MAIN_APP_IDS: readonly string[] = [
  'platform',
  'main-app',
  'main',
  'root',
];

function isMainAppModule(moduleId: string): boolean {
  return RESERVED_MAIN_APP_IDS.includes(moduleId);
}

/**
 * 内部菜单节点结构。
 *
 * 注意：Naive UI 的 MenuOption 是联合类型（MenuOption | MenuGroupOption |
 * MenuDividerOption），不能直接 extends，因此这里定义独立接口，
 * 最终在赋值给 menuTree 时 cast 为 MenuOption[]。
 *
 * 字段说明：
 * - key/label/icon/children：Naive UI MenuOption 需要的渲染字段
 * - parent_id/order/raw：内部建树用的临时字段，处理完会删除
 * - path：附加到 MenuOption 上，供点击跳转使用
 */
interface MenuNode {
  key: string;
  label: string;
  icon?: () => VNodeChild;
  children?: MenuNode[];
  path?: string;
  parent_id?: string | null;
  order?: number;
  raw?: Menu;
}

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);

  eventBus.on('modules:changed', async (modules) => {
    console.log('[MenuStore] 收到 modules:changed 事件，重建菜单');
    if (Array.isArray(modules) && modules.length > 0) {
      await buildMenus(modules);
    } else {
      const moduleStore = useModuleStore();
      if (moduleStore.modules.length > 0) {
        await buildMenus(moduleStore.modules);
      }
    }
  });

  async function buildMenus(modules: Module[]) {
    try {
      const userStore = useUserStore();

      if (!modules || modules.length === 0) {
        menuTree.value = [];
        menuLoaded.value = true;
        return;
      }

      const permissions = userStore.permissions ?? [];
      const allMenus: Menu[] = [];

      // ============================================================
      // 步骤 1：收集所有菜单，拼接完整路径
      // ============================================================
      modules.forEach((module) => {
        if (module.status !== 'active') return;

        const moduleMenus = module.menus || [];
        const mainAppModule = isMainAppModule(module.id);

        moduleMenus.forEach((menu) => {
          // 权限过滤（叶子节点）
          if (menu.permission && !permissions.includes(menu.permission)) {
            return;
          }

          // 规范化 path
          let fullPath: string | undefined;
          if (typeof menu.path === 'string' && menu.path.length > 0) {
            const safePath = menu.path.startsWith('/')
              ? menu.path
              : `/${menu.path}`;
            fullPath = mainAppModule
              ? safePath
              : `/${module.id}${safePath}`;
          } else if (menu.type !== 'directory') {
            console.warn(
              `[MenuStore] 模块 ${module.id} 的菜单 ${menu.id} 缺少 path，已跳过`
            );
            return;
          }

          allMenus.push({ ...menu, path: fullPath });
        });
      });

      // ============================================================
      // 步骤 2：第一阶段建树 —— 创建所有节点并放入 Map（不排序）
      // ============================================================
      const nodeMap = new Map<string, MenuNode>();
      allMenus.forEach((item) => {
        const node: MenuNode = {
          key: item.id,
          label: item.title,
          icon: item.icon ? renderIcon(item.icon) : undefined,
          path: item.path,
          parent_id: item.parent_id,
          order: item.order ?? 0,
          raw: item,
        };
        nodeMap.set(item.id, node);
      });

      // ============================================================
      // 步骤 3：第二阶段建树 —— 根据 parent_id 挂载父子关系
      // ============================================================
      const roots: MenuNode[] = [];
      nodeMap.forEach((node) => {
        if (node.parent_id && nodeMap.has(node.parent_id)) {
          const parent = nodeMap.get(node.parent_id)!;
          if (!parent.children) parent.children = [];
          parent.children.push(node);
        } else {
          roots.push(node);
        }
      });

      // ============================================================
      // 步骤 4：递归处理 —— 同级排序、清理临时字段
      // ============================================================
      function process(nodes: MenuNode[]): MenuNode[] {
        return nodes
          .sort((a, b) => (a.order ?? 0) - (b.order ?? 0))
          .map((node) => {
            if (node.children && node.children.length > 0) {
              node.children = process(node.children);
            }

            // 空目录清理（默认保留；如需隐藏空目录，取消下面注释）
            // const isDirectory =
            //   node.raw?.type === 'directory' ||
            //   (!node.raw?.component && !node.raw?.path);
            // if (isDirectory && (!node.children || node.children.length === 0)) {
            //   return null;
            // }

            if (node.children && node.children.length === 0) {
              delete node.children;
            }

            // 移除临时字段
            delete node.parent_id;
            delete node.order;
            delete node.raw;

            return node;
          })
          .filter(Boolean) as MenuNode[];
      }

      // cast 为 MenuOption[]，因为 MenuNode 结构与 MenuOption 兼容
      menuTree.value = process(roots) as unknown as MenuOption[];
      menuLoaded.value = true;

      if (import.meta.env.DEV) {
        console.log('[MenuStore] 菜单树构建完成', menuTree.value);
      }
    } catch (error) {
      console.error('[MenuStore] buildMenus 失败', error);
      message.error('加载菜单失败，请刷新页面重试');
      menuLoaded.value = true;
    }
  }

  async function refreshMenus(modules: Module[]) {
    menuLoaded.value = false;
    await buildMenus(modules);
  }

  return { menuTree, menuLoaded, buildMenus, refreshMenus };
});
