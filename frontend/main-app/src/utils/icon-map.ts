// src/utils/icon-map.ts
import { h, type VNode } from 'vue';
import { NIcon } from 'naive-ui';
import {
  // ---- 原有图标（保持不变） ----
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

  // ---- 本次新增（补齐后端种子数据引用的图标） ----
  // audit_log 模块的菜单 icon = "Document"
  DocumentTextOutline as DocumentIcon,
  // license 模块的菜单 icon = "Key"
  KeyOutline as KeyIcon,

  // ---- 额外补充的常用别名（可选，用于兼容其他模块） ----
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

// 调试日志仅在开发环境输出，避免污染生产控制台
if (import.meta.env.DEV) {
  console.log('[DEBUG] utils/icon-map.ts: 图标映射初始化');
}

/**
 * 图标名 → 组件 映射表
 *
 * 命名约定：使用「语义化名称」（如 `Document`、`Key`），
 * 与后端模块 manifest.json 中 menus[].icon 字段保持一致。
 *
 * 维护指引：
 * - 后端新增菜单 icon 名时，请在此表中补充对应映射；
 * - 未登记的 icon 名会触发 console.warn 并降级为 GridIcon；
 * - 允许一个语义名映射到同一个组件（如 Dashboard / Grid 都用 GridIcon）。
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

  // ============ 文件 / 文档（含本次新增 Document） ============
  File: FileIcon,
  Document: DocumentIcon,   // ← 新增：audit_log 使用

  // ============ 密钥 / License（含本次新增 Key） ============
  Key: KeyIcon,             // ← 新增：license 使用

  // ============ 操作类 ============
  Edit: EditIcon,
  Delete: DeleteIcon,
  Add: AddIcon,
  Search: SearchIcon,
  FolderOpen: FolderOpenIcon,

  // ============ 业务类（本次补充的别名） ============
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

/**
 * 生成 Naive UI 可用的菜单图标渲染函数。
 *
 * @param iconName 语义化图标名（对应后端 manifest.json 中 menus[].icon）
 * @returns 返回一个生成 VNode 的函数，可直接用于 MenuOption.icon
 *
 * 未找到的图标会：
 * 1. 在控制台输出 warn（便于快速定位遗漏）；
 * 2. 降级为 GridIcon，保证菜单结构完整不报错。
 */
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
