// main-app/src/utils/icon-map.ts
import { h, type VNode, type Component } from 'vue';
import { NIcon } from 'naive-ui';
import {
  // 基础常用
  PeopleOutline,
  LockClosedOutline,
  ShieldOutline,
  PersonOutline,
  GridOutline,
  ListOutline,
  SettingsOutline,
  DocumentOutline,
  CreateOutline,
  TrashOutline,
  AddOutline,
  SearchOutline,
  FolderOpenOutline,
  HomeOutline,
  AppsOutline,

  // 文档 / 密钥 / 证书
  DocumentTextOutline,
  KeyOutline,
  RibbonOutline,
  ShieldCheckmarkOutline,

  // 业务 / 其他
  BarChartOutline,
  CashOutline,
  CartOutline,
  CloudOutline,
  CodeSlashOutline,
  CubeOutline,
  EarthOutline,
  HeartOutline,
  InformationCircleOutline,
  LayersOutline,
  NotificationsOutline,
  PricetagOutline,
  PulseOutline,
  RocketOutline,
  ServerOutline,
  SpeedometerOutline,
  StatsChartOutline,
  TimeOutline,
  TrendingUpOutline,
  WarningOutline,
} from '@vicons/ionicons5';

if (import.meta.env.DEV) {
  console.log('[DEBUG] utils/icon-map.ts: 图标映射初始化');
}

/**
 * 图标名 → 组件映射。
 *
 * key 与后端 /menus/routers 返回的 meta.icon 值保持一致。
 * 未登记的 key 会 console.warn 并降级为 GridOutline。
 */
const iconMap: Record<string, Component> = {
  // ============ 基础 ============
  Dashboard: GridOutline,
  Grid: GridOutline,
  Home: HomeOutline,
  Apps: AppsOutline,
  Setting: SettingsOutline,        // 后端用单数
  Settings: SettingsOutline,       // 兼容复数
  List: ListOutline,
  Menu: ListOutline,               // 后端「菜单管理」用 Menu

  // ============ 用户 / 权限 ============
  People: PeopleOutline,
  Person: PersonOutline,
  User: PersonOutline,             // 后端「用户管理」
  UserFilled: PeopleOutline,       // 后端「角色管理」
  Lock: LockClosedOutline,
  Shield: ShieldOutline,
  ShieldCheck: ShieldCheckmarkOutline,

  // ============ 文件 / 文档 ============
  File: DocumentOutline,
  Document: DocumentTextOutline,   // 后端「日志审计 / 审计」
  DocumentText: DocumentTextOutline,

  // ============ 密钥 / 证书 ============
  Key: KeyOutline,                 // 后端「权限管理 / License」
  Certificate: RibbonOutline,
  Ribbon: RibbonOutline,           // 后端「证书管理」

  // ============ 操作 ============
  Edit: CreateOutline,
  Delete: TrashOutline,
  Add: AddOutline,
  Search: SearchOutline,
  FolderOpen: FolderOpenOutline,

  // ============ 业务 / 状态 ============
  Chart: BarChartOutline,
  Stats: StatsChartOutline,
  TrendingUp: TrendingUpOutline,
  Pulse: PulseOutline,
  Connection: PulseOutline,        // 后端「SSL 代理」
  Odometer: SpeedometerOutline,    // 后端「仪表盘」
  Cash: CashOutline,
  Cart: CartOutline,
  Cloud: CloudOutline,
  Code: CodeSlashOutline,
  Cube: CubeOutline,
  Globe: EarthOutline,
  Heart: HeartOutline,
  Info: InformationCircleOutline,
  Layers: LayersOutline,
  Notification: NotificationsOutline,
  Tag: PricetagOutline,
  Rocket: RocketOutline,
  Server: ServerOutline,
  Speedometer: SpeedometerOutline,
  Time: TimeOutline,
  Warning: WarningOutline,
};

export function renderIcon(iconName: string): () => VNode {
  const IconComponent = iconMap[iconName];

  if (!IconComponent) {
    console.warn(
      `[icon-map] 图标 "${iconName}" 未在 iconMap 中登记，已降级为 GridOutline。` +
        `请在 main-app/src/utils/icon-map.ts 中补充映射。`
    );
    return () => h(NIcon, null, { default: () => h(GridOutline) });
  }

  return () => h(NIcon, null, { default: () => h(IconComponent) });
}
