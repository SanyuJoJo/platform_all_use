/**
 * 后端菜单图标名（Element Plus 语义）→ TDesign 图标名。
 */
const ICON_MAP: Record<string, string> = {
  // 基础
  Dashboard: 'dashboard',
  Odometer: 'dashboard',
  Speedometer: 'dashboard',
  Grid: 'grid-view',
  Apps: 'app',
  App: 'app',
  Home: 'home',
  Setting: 'setting',
  Settings: 'setting',
  List: 'view-list',
  Menu: 'menu',
  FolderOpen: 'folder',
  Folder: 'folder',

  // 用户 / 权限
  People: 'usergroup',
  Person: 'user',
  User: 'user',
  UserFilled: 'usergroup',
  Lock: 'lock-on',
  Shield: 'secured',
  ShieldCheck: 'secured',
  Key: 'key',

  // 文件 / 文档
  File: 'file',
  Document: 'file',
  DocumentText: 'file',
  Certificate: 'certificate',
  Ribbon: 'certificate',

  // 业务 / 状态
  Connection: 'link',
  Link: 'link',
  Search: 'search',
  Time: 'time',
  Chart: 'chart',
  Stats: 'chart',
  BarChart: 'chart',
  TrendingUp: 'chart',
  Pulse: 'chart',
  Cash: 'cart',
  Cart: 'cart',
  Cloud: 'cloud',
  Globe: 'internet',
  Earth: 'internet',
  Cube: 'app',
  Layers: 'layers',
  Notification: 'notification',
  Tag: 'tag',
  Rocket: 'rocket',
  Server: 'server',
  Info: 'info-circle',
  Warning: 'error-circle',
  Heart: 'heart',
  Code: 'code',

  // 音视频
  Video: 'video',
  Videocam: 'video',
  Play: 'play-circle',
  Task: 'task',
  Mail: 'mail',
}

export function tdesignIcon(name?: string): string {
  if (!name) return 'app'
  return ICON_MAP[name] || ICON_MAP[name.trim()] || 'app'
}
