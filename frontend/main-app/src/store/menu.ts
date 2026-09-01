import { defineStore } from 'pinia';
import { ref, h, type VNode } from 'vue';
import type { MenuOption } from 'naive-ui';
import { NIcon } from 'naive-ui';
import {
  PeopleOutline as PeopleIcon,
  LockClosedOutline as LockIcon,
  ShieldOutline as ShieldIcon,
  PersonOutline as PersonIcon,
  GridOutline as GridIcon,
  ListOutline as ListIcon,
  SettingsOutline as SettingsIcon,
  DocumentOutline as FileIcon,
  CreateOutline as EditIcon,
  TrashOutline as DeleteIcon,
  AddOutline as AddIcon,
  SearchOutline as SearchIcon,
} from '@vicons/ionicons5';
import { useModuleStore } from './module';
import { useUserStore } from './user';
import { message } from '@/utils/naive';
import type { Menu } from '@/types/module';
// 图标映射表（扩充常用图标）
const iconMap: Record<string, any> = {
  People: PeopleIcon,
  Lock: LockIcon,
  Shield: ShieldIcon,
  Person: PersonIcon,
  Grid: GridIcon,
  List: ListIcon,
  Dashboard: GridIcon,
  Settings: SettingsIcon,
  File: FileIcon,
  Edit: EditIcon,
  Delete: DeleteIcon,
  Add: AddIcon,
  Search: SearchIcon,
  // 可继续扩充
};
function renderIcon(iconName: string): () => VNode {
  const IconComponent = iconMap[iconName];
  if (!IconComponent) {
    console.warn(`[Menu] 图标 "${iconName}" 未找到，使用默认图标 Grid`);
    const FallbackIcon = GridIcon;
    return () => h(NIcon, null, { default: () => h(FallbackIcon) });
  }
  return () => h(NIcon, null, { default: () => h(IconComponent) });
}
export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuOption[]>([]);
  const menuLoaded = ref(false);
  /**
   * 构建菜单树
   * 从 moduleStore 获取模块列表，提取 menus，权限过滤，构建树形结构
   */
  async function buildMenus() {
    try {
      const moduleStore = useModuleStore();
      const userStore = useUserStore();
      // 确保模块列表已加载
      if (!moduleStore.loaded) {
        await moduleStore.fetchModules();
      }
      const modules = moduleStore.modules;
      const permissions = userStore.permissions ?? [];
      if (!modules || modules.length === 0) {
        menuTree.value = [];
        menuLoaded.value = true;
        return;
      }
      const allMenus: Menu[] = [];
      modules.forEach((module) => {
        // 只处理活跃模块
        if (module.status !== 'active') return;
        const moduleMenus = module.menus || [];
        moduleMenus.forEach((menu) => {
          // 权限过滤（双重保险）
          if (menu.permission && !permissions.includes(menu.permission)) {
            return;
          }
          // 防御性拼接路径：确保 path 以 '/' 开头
          const safePath = menu.path.startsWith('/') ? menu.path : `/${menu.path}`;
          allMenus.push({
            ...menu,
            path: `/${module.id}${safePath}`,
          });
        });
      });
      // 按 order 排序
      const sorted = [...allMenus].sort((a, b) => (a.order || 0) - (b.order || 0));
      // 构建树形结构
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
      menuTree.value = roots;
      menuLoaded.value = true;
    } catch (error) {
      console.error('[Menu] 构建菜单失败:', error);
      message.error('加载菜单失败，请刷新页面重试');
      menuLoaded.value = true;
    }
  }
  return { menuTree, menuLoaded, buildMenus };
});
