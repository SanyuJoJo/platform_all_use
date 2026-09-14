export interface Menu {
  id: string;
  parent_id: string | null;
  title: string;
  icon?: string;
  path: string;
  component: string;
  permission?: string;
  order?: number;
}
/**
 * 模块对象。
 *
 * P0-5：API 契约中 entry_frontend、menus 未必返回，
 * 因此这里定义为可选；前端在 getActiveModules 中做默认值兜底。
 */
export interface Module {
  id: string;
  name: string;
  version: string;
  description?: string;
  author?: string;
  homepage?: string;
  status: 'active' | 'inactive';
  entry_frontend?: string;
  menus?: Menu[];
  dependencies?: string[];
  installed_at?: string;
  updated_at?: string;
}
