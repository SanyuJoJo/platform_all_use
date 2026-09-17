import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MenuOption } from 'naive-ui';
import { useUserStore } from './user';
import { useModuleStore } from './module';
import { message } from '@/utils/naive';
import type { Menu, Module } from '@/types/module';
import { renderIcon } from '@/utils/icon-map';
import { eventBus } from '@/micro-frontend/event-bus';

/**
 * 保留给主应用自身的模块 ID（与 registry.ts 中同名集合保持一致）。
 *
 * 语义：
 * - 这些模块不会注册为 qiankun 子应用；
 * - 它们的 menus[].path 已经是主应用完整的路由路径（如 /dashboard），
 *   构建菜单时**不能再拼 /{module.id} 前缀**，否则会得到 /platform/dashboard 这类无效路径。
 *
 * 典型场景：
 * - platform 模块提供主应用内置的「仪表盘」菜单，path = /dashboard；
 * - 主应用路由在 router/index.ts 中直接注册 /dashboard。
 *
 * 后续如果抽到独立文件 @/constants/module.ts，把本常量改为
 *   import { RESERVED_MAIN_APP_IDS } from '@/constants/module';
 * 即可，无需改动本文件其它逻辑。
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
        // ★ 关键：主应用自身模块的 path 已是完整路由，不再拼 /{module.id}
        const mainAppModule = isMainAppModule(module.id);

        moduleMenus.forEach((menu) => {
          // 权限过滤
          if (menu.permission && !permissions.includes(menu.permission)) {
            return;
          }

          // 防御：menu.path 缺失或非字符串时跳过，避免 startsWith 抛错
          if (typeof menu.path !== 'string' || menu.path.length === 0) {
            console.warn(
              `[MenuStore] 模块 ${module.id} 的菜单 ${menu.id} 缺少 path，已跳过`
            );
            return;
          }

          // 规范化 path（确保以 / 开头）
          const safePath = menu.path.startsWith('/')
            ? menu.path
            : `/${menu.path}`;

          // ★ 关键修复：
          //   - 主应用自身模块（platform 等）→ 直接使用 safePath（如 /dashboard）
          //   - 其它业务模块（auth / module_manager / audit_log / license）→ 拼 /{module.id}
          const fullPath = mainAppModule
            ? safePath
            : `/${module.id}${safePath}`;

          allMenus.push({ ...menu, path: fullPath });
        });
      });

      // 按 order 排序（未填 order 视为 0，越靠前越优先）
      const sorted = [...allMenus].sort(
        (a, b) => (a.order || 0) - (b.order || 0)
      );

      // 构建树：先建 Map，再按 parent_id 挂载
      const map = new Map<string, MenuOption>();
      const roots: MenuOption[] = [];

      sorted.forEach((item) => {
        const node: MenuOption = {
          // 保持 key = menu.id（全局唯一，且不受 path 冲突影响）
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

      // 深拷贝，避免外部直接改写 store 内部引用
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
