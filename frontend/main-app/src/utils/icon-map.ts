// src/utils/icon-map.ts
import { h, type VNode } from 'vue';
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
  FolderOpenOutline as FolderOpenIcon,
  HomeOutline as HomeIcon,
  AppsOutline as AppsIcon,
} from '@vicons/ionicons5';

console.log('[DEBUG] utils/icon-map.ts: 图标映射初始化');

const iconMap: Record<string, any> = {
  People: PeopleIcon,
  Lock: LockIcon,
  Shield: ShieldIcon,
  Person: PersonIcon,
  Grid: GridIcon,
  List: ListIcon,
  Dashboard: GridIcon,
  FolderOpen: FolderOpenIcon,
  Home: HomeIcon,
  Apps: AppsIcon,
  Settings: SettingsIcon,
  File: FileIcon,
  Edit: EditIcon,
  Delete: DeleteIcon,
  Add: AddIcon,
  Search: SearchIcon,
};

export function renderIcon(iconName: string): () => VNode {
  const IconComponent = iconMap[iconName];
  if (!IconComponent) {
    console.warn(`[DEBUG] renderIcon: 图标 "${iconName}" 未找到，使用默认 Grid`);
    const FallbackIcon = GridIcon;
    return () => h(NIcon, null, { default: () => h(FallbackIcon) });
  }
  return () => h(NIcon, null, { default: () => h(IconComponent) });
}