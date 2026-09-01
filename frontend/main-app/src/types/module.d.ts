/**
 * 菜单项接口（对应 manifest.json 中的 menus 数组元素）
 */
export interface Menu {
  id: string;
  parent_id: string | null;
  title: string;
  icon?: string;          // 图标名称（与 iconMap 对应）
  path: string;           // 相对路径，如 '/dashboard'
  component: string;      // 组件路径（子应用内部）
  permission?: string;    // 所需权限
  order?: number;         // 排序，数值越小越靠前
}

/**
 * 模块接口（对应后端 /modules 接口返回的模块对象）
 */
export interface Module {
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  homepage?: string;
  status: 'active' | 'inactive';
  entry_frontend: string; // 子应用入口 URL
  menus: Menu[];          // 菜单列表（已按权限过滤）
  dependencies?: string[];
  installed_at?: string;
  updated_at?: string;
  // 其他字段按需扩展
}
