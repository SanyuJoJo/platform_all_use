// src/store/menu.ts
import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MenuOption } from 'naive-ui';
import { useUserStore } from './user';
import { message } from '@/utils/naive';
import type { Menu, Module } from '@/types/module';
import { renderIcon } from '@/utils/icon-map';
import { eventBus } from '@/micro-frontend/event-bus';

console.log('[DEBUG] store/menu.ts: 菜单 Store 初始化');

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);

  // 监听模块刷新事件（由 moduleStore 触发）
  eventBus.on('modules:changed', async (modules: Module[]) => {
    console.log('[DEBUG] menuStore: 收到 modules:changed 事件，重建菜单');
    if (modules && modules.length > 0) {
      await buildMenus(modules);
    } else {
      console.warn('[DEBUG] menuStore: 事件传入的模块列表为空，尝试从 store 获取');
      // 如果事件未传数据，可以从 moduleStore 获取
      const { useModuleStore } = await import('./module');
      const moduleStore = useModuleStore();
      if (moduleStore.modules.value.length > 0) {
        await buildMenus(moduleStore.modules.value);
      }
    }
  });

  /**
   * 构建菜单树 - 必须传入 modules 数据
   */
  async function buildMenus(modules: Module[]) {
    console.log('[DEBUG] menuStore.buildMenus: 开始构建，模块数:', modules?.length);
    try {
      const userStore = useUserStore();

      if (!modules || modules.length === 0) {
        console.warn('[DEBUG] menuStore.buildMenus: 传入的模块列表为空，清空菜单');
        menuTree.value = [];
        menuLoaded.value = true;
        return;
      }

      const permissions = userStore.permissions ?? [];
      console.log('[DEBUG] menuStore.buildMenus: 用户权限:', permissions);

      const allMenus: Menu[] = [];

      modules.forEach((module) => {
        if (module.status !== 'active') {
          console.log(`[DEBUG] menuStore.buildMenus: 模块 ${module.id} 非激活，跳过`);
          return;
        }
        const moduleMenus = module.menus || [];
        console.log(`[DEBUG] menuStore.buildMenus: 模块 ${module.id} 菜单数:`, moduleMenus.length);
        moduleMenus.forEach((menu) => {
          if (menu.permission && !permissions.includes(menu.permission)) {
            console.log(`[DEBUG] menuStore.buildMenus: 菜单 ${menu.id} 无权限，跳过`);
            return;
          }
          const safePath = menu.path.startsWith('/') ? menu.path : `/${menu.path}`;
          const fullPath = `/${module.id}${safePath}`;
          allMenus.push({ ...menu, path: fullPath });
        });
      });

      console.log('[DEBUG] menuStore.buildMenus: 有效菜单数:', allMenus.length);

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

      menuTree.value = JSON.parse(JSON.stringify(roots)); // 深拷贝避免 Proxy 干扰
      menuLoaded.value = true;
      console.log('[DEBUG] menuStore.buildMenus: 构建完成，根菜单数:', roots.length);
    } catch (error) {
      console.error('[DEBUG] menuStore.buildMenus: 失败', error);
      message.error('加载菜单失败，请刷新页面重试');
      menuLoaded.value = true;
    }
  }

  async function refreshMenus(modules: Module[]) {
    console.log('[DEBUG] menuStore.refreshMenus: 强制刷新');
    menuLoaded.value = false;
    await buildMenus(modules);
  }

  return { menuTree, menuLoaded, buildMenus, refreshMenus };
});