/**
 * 后端 GET /api/v1/menus/routers 返回的路由节点。
 *
 * - path：模块内相对路径（不带 /{microApp} 前缀）
 * - microApp：所属模块 id（main/auth/module_manager/audit_log/license/
 *   crypto_console/ssl_proxy/media）
 * - type：directory 目录（不注册路由）| menu 菜单
 */
export interface RouteMeta {
  title: string
  icon?: string
  permission?: string
  hidden?: boolean
  keepAlive?: boolean
  [key: string]: unknown
}

export interface RouteNode {
  id: string
  parentId: string | null
  name: string
  path?: string
  component?: string
  microApp: string
  type: 'directory' | 'menu'
  meta: RouteMeta
  order?: number
  children?: RouteNode[]
}

/** 主应用保留 id：这些节点不拼接 /{microApp} 前缀 */
export const MAIN_APP_IDS: readonly string[] = ['main', 'platform', 'main-app', 'root']

export function isMainAppId(id: string | undefined | null): boolean {
  return MAIN_APP_IDS.includes(String(id ?? ''))
}
