/**
 * 后端 GET /api/v1/menus/routers 返回的路由节点。
 * 字段与后端 menu.RouterVO 对齐。
 */
export interface RouterMeta {
  title: string;
  icon?: string;
  noCache?: boolean;
  link?: string;
}
export interface RouterVO {
  /** 路由名称（后端只对 C 类型给出，用于 Vue Router name） */
  name?: string;
  /** 路由地址，后端返回绝对路径（如 /system/users） */
  path: string;
  /** 是否在菜单中隐藏 */
  hidden?: boolean;
  redirect?: string;
  /**
   * 组件标识：
   * - 'Layout'：目录（M）
   * - 'system/user/index'：主应用页面（相对 views/）
   * - 子应用路径（如 /auth/users）时由前端根据 path 前缀识别，此项不使用
   */
  component?: string;
  query?: string;
  alwaysShow?: boolean;
  meta?: RouterMeta;
  children?: RouterVO[];
}
