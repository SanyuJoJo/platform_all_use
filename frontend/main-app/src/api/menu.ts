import request from './index';
import type { RouterVO } from '@/types/menu';
/**
 * 获取当前登录用户可见的前端路由树。
 *
 * 对应后端 GET /api/v1/menus/routers。
 * - 返回的路由树已按用户角色过滤（后端 menu.BuildRouters 完成）；
 * - 超管返回全部 status=0 的 M/C 菜单；
 * - 该接口本身不做权限校验（所有登录用户可访问），用于前端首屏构建。
 */
export async function getRouters(): Promise<RouterVO[]> {
  const res = await request.get<{ routers: RouterVO[] }>('/menus/routers');
  const routers = res.data?.routers;
  if (!Array.isArray(routers)) {
    console.warn('[API] /menus/routers 返回格式非数组：', res.data);
    return [];
  }
  return routers;
}
