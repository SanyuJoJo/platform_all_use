import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MenuOption } from 'naive-ui';
import { useUserStore } from './user';
import { message } from '@/utils/naive';
import type { Menu, Module } from '@/types/module';
import { renderIcon } from '@/utils/icon-map';
import { eventBus } from '@/micro-frontend/event-bus';

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);

  /**
   * 构建菜单树 - 必须传入 modules 数据
   */
  async function buildMenus(modules: Module[]) {
    console.log('🔍 [menu] buildMenus 开始执行');
    try {
      const userStore = useUserStore();

      if (!modules || modules.length === 0) {
        console.warn('⚠️ [menu] 传入的模块列表为空，无法构建菜单');
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
          const safePath = menu.path.startsWith('/') ? menu.path : `/${menu.path}`;
          const fullPath = `/${module.id}${safePath}`;
          allMenus.push({ ...menu, path: fullPath });
        });
      });

      console.log('🔍 [menu] allMenus 数量:', allMenus.length);

      const sorted = [...allMenus].sort((a, b) => (a.order || 0) - (b.order || 0));

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

      // 深拷贝避免 Proxy 干扰
      menuTree.value = JSON.parse(JSON.stringify(roots));
      menuLoaded.value = true;
      console.log('✅ [menu] 菜单构建完成，根菜单数:', roots.length);
    } catch (error) {
      console.error('[Menu] 构建菜单失败:', error);
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
