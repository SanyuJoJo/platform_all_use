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
  FolderOpenOutline as FolderOpenIcon,  // ✅ 存在
  HomeOutline as HomeIcon,              // ✅ 存在
  AppsOutline as AppsIcon,              // ✅ 存在
} from '@vicons/ionicons5';

const iconMap: Record<string, any> = {
  People: PeopleIcon,
  Lock: LockIcon,
  Shield: ShieldIcon,
  Person: PersonIcon,
  Grid: GridIcon,
  List: ListIcon,
  Dashboard: GridIcon,           // ✅ Dashboard 使用 GridIcon
  FolderOpen: FolderOpenIcon,    // ✅ 对应 FolderOpenOutline
  Home: HomeIcon,                // ✅ 对应 HomeOutline
  Apps: AppsIcon,                // ✅ 对应 AppsOutline
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
    console.warn(`[Menu] 图标 "${iconName}" 未找到，使用默认图标 Grid`);
    const FallbackIcon = GridIcon;
    return () => h(NIcon, null, { default: () => h(FallbackIcon) });
  }
  return () => h(NIcon, null, { default: () => h(IconComponent) });
}
