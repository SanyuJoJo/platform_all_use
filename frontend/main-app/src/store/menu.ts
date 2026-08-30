import { defineStore } from 'pinia';
import { ref } from 'vue';
import { h, type VNode } from 'vue';
import type { MenuOption } from 'naive-ui';
import { NIcon } from 'naive-ui';
import {
  PeopleOutline as PeopleIcon,
  LockClosedOutline as LockIcon,        // 修正：使用 LockClosedOutline
  ShieldOutline as ShieldIcon,
  PersonOutline as PersonIcon,
  GridOutline as GridIcon,
  ListOutline as ListIcon,
  SettingsOutline as SettingsIcon,
} from '@vicons/ionicons5';
import { getModules } from '@/api/module';
import { useUserStore } from './user';
import { message } from '@/utils/naive';

const iconMap: Record<string, any> = {
  People: PeopleIcon,
  Lock: LockIcon,
  Shield: ShieldIcon,
  Person: PersonIcon,
  Grid: GridIcon,
  List: ListIcon,
  Dashboard: GridIcon,   // Dashboard 使用 Grid 替代
  Settings: SettingsIcon,
};

function renderIcon(iconName: string): () => VNode {
  const IconComponent = iconMap[iconName];
  if (!IconComponent) {
    console.warn(`[Menu] Icon "${iconName}" not found, using fallback (Grid).`);
    const FallbackIcon = GridIcon;
    return () => h(NIcon, null, { default: () => h(FallbackIcon) });
  }
  return () => h(NIcon, null, { default: () => h(IconComponent) });
}

interface MenuItem {
  id: string;
  parent_id: string | null;
  title: string;
  icon: string;
  path: string;
  permission: string;
  order: number;
}

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);

  async function buildMenus() {
    try {
      const userStore = useUserStore();
      const permissions = userStore.permissions ?? [];

      const modules = await getModules();

      if (!Array.isArray(modules)) {
        console.error('[Menu] Invalid modules data:', modules);
        menuTree.value = [];
        menuLoaded.value = true;
        return;
      }

      const allMenus: MenuItem[] = [];
      modules.forEach((module: any) => {
        const moduleMenus = module.manifest?.menus || [];
        moduleMenus.forEach((menu: any) => {
          if (menu.permission && !permissions.includes(menu.permission)) {
            return;
          }
          allMenus.push({
            ...menu,
            path: `/${module.id}${menu.path}`,
          });
        });
      });

      const sorted = [...allMenus].sort((a, b) => (a.order || 0) - (b.order || 0));
      const map = new Map<string, MenuOption>();
      const roots: MenuOption[] = [];

      sorted.forEach(item => {
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

      menuTree.value = roots;
      menuLoaded.value = true;
    } catch (error) {
      console.error('[Menu] buildMenus error:', error);
      message.error('加载菜单失败，请刷新页面重试');
      menuLoaded.value = true;
    }
  }

  return { menuTree, menuLoaded, buildMenus };
});
