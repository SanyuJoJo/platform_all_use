import request from '@/utils/request'
import type { RouteNode } from '@/types'

/**
 * 获取当前用户的动态路由树。
 * 后端路径：GET /api/v1/menus/routers
 */
export async function getRouters(): Promise<RouteNode[]> {
  const res = await request.get<{ routers: RouteNode[] }>('/menus/routers')
  const data = res?.data as unknown

  if (Array.isArray(data)) return data as RouteNode[]
  if (data && typeof data === 'object') {
    const obj = data as Record<string, unknown>
    for (const key of ['routers', 'items', 'menus', 'list', 'records', 'rows']) {
      if (Array.isArray(obj[key])) return obj[key] as RouteNode[]
    }
  }
  console.warn('[api/menu] /menus/routers 返回格式非预期：', data)
  return []
}
