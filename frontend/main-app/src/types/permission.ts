/**
 * 后端 /menus/routers 返回的路由节点。
 *
 * 字段说明：
 * - id：节点唯一 ID
 * - parentId：父节点 ID，顶级为 null（兼容后端 parent_id）
 * - name：路由名，全局唯一，用于 router.addRoute 和 router.hasRoute
 * - path：模块内相对路径（不带 /{microApp} 前缀），由主应用菜单拼前缀、子应用 basename 处理
 * - component：仅 microApp === 'main' 时由主应用解析为视图；子应用节点用于菜单生成
 * - microApp：'main' | 'platform' | 子应用 id（如 auth、module_manager）
 * - type：'directory' 目录（不注册路由）| 'menu' 菜单
 * - meta.permission：页面级权限码，与 user.permissions 对应
 * - meta.hidden：为 true 时不在菜单中显示
 */
export interface RouteMeta {
  title: string;
  icon?: string;
  permission?: string;
  hidden?: boolean;
  keepAlive?: boolean;
  [key: string]: unknown;
}

export interface RouteNode {
  id: string;
  parentId: string | null;
  name: string;
  path?: string;
  component?: string;
  microApp: string;
  type: 'directory' | 'menu';
  meta: RouteMeta;
  order?: number;
  children?: RouteNode[];
}

/** 主应用保留 ID：这些 id 下的节点不会注册为 qiankun 子应用，也不会下发给子应用 */
export const MAIN_APP_IDS: readonly string[] = ['main', 'platform', 'main-app', 'root'];
