import request from './index';
import type { RouteNode } from '@/types/permission';

/**
 * 获取当前用户的动态路由树。
 * 后端路径：GET /api/v1/menus/routers
 */
export async function getRouters(): Promise<RouteNode[]> {
  console.log('[DIAG] getRouters 请求开始');
  try {
    const res = await request.get('/menus/routers');
    console.log('[DIAG] getRouters 完整响应 =', res);
    console.log('[DIAG] getRouters res.data =', res?.data);

    const d = res?.data;
    let raw: unknown = d;
    let pickedKey = 'data(itself)';

    if (!Array.isArray(raw)) {
      const obj = d as Record<string, unknown> | null | undefined;
      const candidates: Array<[string, unknown]> = [
        ['items', obj?.items],
        ['routers', obj?.routers],
        ['menus', obj?.menus],
        ['list', obj?.list],
        ['records', obj?.records],
        ['data', obj?.data],
        ['rows', obj?.rows],
      ];
      for (const [k, v] of candidates) {
        if (Array.isArray(v)) {
          raw = v;
          pickedKey = k;
          break;
        }
      }
      if (!Array.isArray(raw)) {
        console.warn(
          '[DIAG] getRouters 未找到数组字段。data keys =',
          obj ? Object.keys(obj) : obj
        );
        return [];
      }
    }

    console.log(
      `[DIAG] getRouters 采用字段 = ${pickedKey}, 根节点数 = ${(raw as unknown[]).length}`
    );
    if ((raw as unknown[]).length > 0) {
      console.log(
        '[DIAG] getRouters 第 1 个根节点 =',
        (raw as unknown[])[0]
      );
      // 打印每个节点的关键字段，方便核对后端实际字段名
      (raw as unknown[]).slice(0, 3).forEach((n, i) => {
        const node = n as Record<string, unknown>;
        console.log(
          `[DIAG] 节点[${i}] 字段名 =`,
          Object.keys(node),
          '| 值 =',
          node
        );
      });
    }
    return raw as RouteNode[];
  } catch (e) {
    console.error('[DIAG] getRouters 请求异常 =', e);
    throw e;
  }
}
