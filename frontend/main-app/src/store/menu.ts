import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MenuOption } from 'naive-ui';
import { useUserStore } from './user';
import { useModuleStore } from './module';
import { message } from '@/utils/naive';
import type { Menu, Module } from '@/types/module';
import { renderIcon } from '@/utils/icon-map';
import { eventBus } from '@/micro-frontend/event-bus';

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);

  /**
   * 监听模块变更事件，重建菜单。
   * 事件类型由 EventBusEvents 保证，modules 直接为 Module[]。
   */
  eventBus.on('modules:changed', async (modules) => {
    console.log('[MenuStore] 收到 modules:changed 事件，重建菜单');
    if (Array.isArray(modules) && modules.length > 0) {
      await buildMenus(modules);
    } else {
      const moduleStore = useModuleStore();
      // Pinia setup store：ref 已自动解包，直接访问 .modules
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

      modules.forEach((module) => {
        if (module.status !== 'active') return;
        const moduleMenus = module.menus || [];
        moduleMenus.forEach((menu) => {
          if (menu.permission && !permissions.includes(menu.permission)) {
            return;
          }
          const safePath = menu.path.startsWith('/')
            ? menu.path
            : `/${menu.path}`;
          const fullPath = `/${module.id}${safePath}`;
          allMenus.push({ ...menu, path: fullPath });
        });
      });

      const sorted = [...allMenus].sort(
        (a, b) => (a.order || 0) - (b.order || 0)
      );

      const map = new Map<string, MenuOption>();
      const roots: MenuOption[] = [];

      sorted.forEach((item) => {
        const node: MenuOption = {
          key: item.id,
          label: item.title,
          icon: item.icon ? renderIcon(item.icon) : undefined,
          path: item.path,
        };
        map.set(item.id, node);

        if (item.parent_id && map.has(item.parent_id)) {
          const parent = map.get(item.parent_id)!;
          if (!parent.children) parent.children = [];
          parent.children.push(node);
        } else {
          roots.push(node);
        }
      });

      menuTree.value = JSON.parse(JSON.stringify(roots));
      menuLoaded.value = true;
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
