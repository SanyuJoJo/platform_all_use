// main-app/src/types/module.d.ts
/**
 * 菜单项接口（对应 frontend.yaml / manifest.json 中的 menus 数组元素）。
 *
 * 字段说明：
 * - id：菜单唯一 ID，建议 `module:子目录:页面`，如 auth:system:users
 * - parent_id：父菜单 ID，顶级为 null
 * - title：显示名称
 * - icon：图标语义名，对应 utils/icon-map.ts
 * - path：模块内相对路径，如 /users、/system/cert-ops/certs；目录可为空
 * - component：前端组件相对路径；目录可为空
 * - permission：所需权限码；目录通常不填
 * - order：同级排序，数值越小越靠前
 * - type：节点类型，directory（目录）/ menu（叶子菜单）
 */
export interface Menu {
  id: string;
  parent_id: string | null;
  title: string;
  icon?: string;
  path?: string;
  component?: string;
  permission?: string;
  order?: number;
  type?: 'directory' | 'menu';
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
