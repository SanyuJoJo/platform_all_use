import request from './index';
import type { Module } from '@/types';

/** 后端 page_size 上限（与 module_manager / auth 模块 max=100 对齐） */
const BACKEND_PAGE_SIZE_MAX = 100;

/**
 * 通用分页拉全量工具。
 *
 * - 后端 page_size 上限为 100（见《模块管理 v1.3》§ 5.6、
 *   《用户角色权限 v1.3》§ 5.2 的 binding:"max=100"）；
 * - 本函数按 100 每页循环拉取，直到累计条数达到 total；
 * - 设置安全上限（MAX_PAGES）避免异常数据导致死循环。
 */
async function fetchAllPages<T>(
  url: string,
  baseParams: Record<string, unknown> = {},
  extractor: (res: any) => { items: T[]; total: number }
): Promise<T[]> {
  const MAX_PAGES = 50; // 100 * 50 = 5000 条，超过则截断
  const all: T[] = [];
  let page = 1;

  while (page <= MAX_PAGES) {
    const res = await request.get(url, {
      params: {
        ...baseParams,
        page,
        page_size: BACKEND_PAGE_SIZE_MAX,
      },
    });
    const { items, total } = extractor(res);
    if (!Array.isArray(items) || items.length === 0) break;
    all.push(...items);
    if (all.length >= total) break;
    page += 1;
  }

  if (page > MAX_PAGES) {
    console.warn(
      `[AuditLog] ${url} 超过 ${MAX_PAGES * BACKEND_PAGE_SIZE_MAX} 条，已截断`
    );
  }
  return all;
}

/** 获取所有已激活模块（用于筛选下拉） */
export async function getModulesForFilter(): Promise<Module[]> {
  try {
    return await fetchAllPages<Module>(
      '/modules',
      { status: 'active' },
      (res) => {
        // 兼容分页结构 { items, total } 与直接数组 [...]
        const data = res.data;
        if (Array.isArray(data)) {
          return { items: data, total: data.length };
        }
        return {
          items: data?.items ?? [],
          total: data?.total ?? 0,
        };
      }
    );
  } catch (error) {
    console.warn('[AuditLog] 获取模块列表失败:', error);
    return [];
  }
}

/** 获取所有用户（用于筛选下拉） */
export async function getUsersForFilter(): Promise<
  { id: number; username: string }[]
> {
  try {
    return await fetchAllPages<{ id: number; username: string }>(
      '/auth/users',
      {},
      (res) => {
        const data = res.data;
        if (Array.isArray(data)) {
          return { items: data, total: data.length };
        }
        return {
          items: data?.items ?? [],
          total: data?.total ?? 0,
        };
      }
    );
  } catch (error) {
    console.warn('[AuditLog] 获取用户列表失败:', error);
    return [];
  }
}
