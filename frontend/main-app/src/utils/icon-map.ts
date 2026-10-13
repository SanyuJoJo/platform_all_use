// main-app/src/utils/icon-map.ts
import { h, type VNode } from 'vue';
import { NIcon } from 'naive-ui';
import {
  // ---- 基础常用 ----
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

  // ---- 文档 / 密钥相关 ----
  DocumentTextOutline as DocumentIcon,
  KeyOutline as KeyIcon,

  // ---- 证书 / 安全相关（替代不存在的 CertificateOutline）----
  RibbonOutline as CertificateIcon,
  ShieldCheckmarkOutline as ShieldCheckIcon,

  // ---- 额外补充的常用别名 ----
  BarChartOutline as ChartIcon,
  CashOutline as CashIcon,
  CartOutline as CartIcon,
  CloudOutline as CloudIcon,
  CodeSlashOutline as CodeIcon,
  CubeOutline as CubeIcon,
  EarthOutline as GlobeIcon,
  HeartOutline as HeartIcon,
  InformationCircleOutline as InfoIcon,
  LayersOutline as LayersIcon,
  NotificationsOutline as NotificationIcon,
  PricetagOutline as TagIcon,
  PulseOutline as PulseIcon,
  RocketOutline as RocketIcon,
  ServerOutline as ServerIcon,
  SpeedometerOutline as DashboardIcon,
  StatsChartOutline as StatsIcon,
  TimeOutline as TimeIcon,
  TrendingUpOutline as TrendingUpIcon,
  WarningOutline as WarningIcon,
} from '@vicons/ionicons5';

if (import.meta.env.DEV) {
  console.log('[DEBUG] utils/icon-map.ts: 图标映射初始化');
}

/**
 * 图标名 → 组件 映射表
 *
 * 命名约定：使用「语义化名称」，与后端 menus[].icon 字段保持一致。
 * 未登记的 icon 名会触发 console.warn 并降级为 GridIcon。
 */
const iconMap: Record<string, unknown> = {
  // ============ 基础常用 ============
  Dashboard: GridIcon,
  Grid: GridIcon,
  Home: HomeIcon,
  Apps: AppsIcon,
  Settings: SettingsIcon,
  List: ListIcon,

  // ============ 用户 / 权限 ============
  People: PeopleIcon,
  Person: PersonIcon,
  Lock: LockIcon,
  Shield: ShieldIcon,
  ShieldCheck: ShieldCheckIcon,

  // ============ 文件 / 文档 ============
  File: FileIcon,
  Document: DocumentIcon,
  DocumentText: DocumentIcon,

  // ============ 密钥 / License / 证书 ============
  Key: KeyIcon,
  Certificate: CertificateIcon,  // 使用 RibbonOutline 代替
  Ribbon: CertificateIcon,

  // ============ 操作类 ============
  Edit: EditIcon,
  Delete: DeleteIcon,
  Add: AddIcon,
  Search: SearchIcon,
  FolderOpen: FolderOpenIcon,

  // ============ 业务类 ============
  Chart: ChartIcon,
  Stats: StatsIcon,
  TrendingUp: TrendingUpIcon,
  Pulse: PulseIcon,
  Cash: CashIcon,
  Cart: CartIcon,
  Cloud: CloudIcon,
  Code: CodeIcon,
  Cube: CubeIcon,
  Globe: GlobeIcon,
  Heart: HeartIcon,
  Info: InfoIcon,
  Layers: LayersIcon,
  Notification: NotificationIcon,
  Tag: TagIcon,
  Rocket: RocketIcon,
  Server: ServerIcon,
  Speedometer: DashboardIcon,
  Time: TimeIcon,
  Warning: WarningIcon,
};

export function renderIcon(iconName: string): () => VNode {
  const IconComponent = iconMap[iconName];

  if (!IconComponent) {
    console.warn(
      `[icon-map] 图标 "${iconName}" 未在 iconMap 中登记，已降级为 GridIcon。` +
        `请在 main-app/src/utils/icon-map.ts 中补充映射。`
    );
    return () => h(NIcon, null, { default: () => h(GridIcon) });
  }

  return () =>
    h(NIcon, null, {
      default: () => h(IconComponent as Parameters<typeof h>[0]),
    });
}
