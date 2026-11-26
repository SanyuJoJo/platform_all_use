/* ============================================================================
 * 菜单管理类型（对应后端 models.Menu）
 *
 * 具体 Menu / MenuForm 类型定义在 src/api/system-menu.ts（与菜单 API 放在一起，
 * 以便调用方一条 import 引入）。这里仅为类型化导出提供入口。
 * ==========================================================================*/
export type {
  Menu,
  MenuForm,
  MenuQuery,
  MicroAppOption,
} from '@/api/system-menu';