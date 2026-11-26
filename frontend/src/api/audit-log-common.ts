import request from '@/utils/request';
import type { ModuleOption } from '@/types';

/** 后端 page_size 上限（与 module_manager / auth 模块 max=100 对齐） */
const BACKEND_PAGE_SIZE_MAX = 100;

/**
 * 通用分页拉全量工具。
 */
async function fetchAllPages<T>(
  url: string,
  baseParams: Record<string, unknown> = {},
  extractor: (res: any) => { items: T[]; total: number },
): Promise<T[]> {
  const MAX_PAGES = 50;
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
    console.warn(`[AuditLog] ${url} 超过 ${MAX_PAGES * BACKEND_PAGE_SIZE_MAX} 条，已截断`);
  }
  return all;
}

/** 获取所有已激活模块（用于筛选下拉） */
export async function getModulesForFilter(): Promise<ModuleOption[]> {
  try {
    return await fetchAllPages<ModuleOption>(
      '/modules',
      { status: 'active' },
      (res) => {
        const data = res.data;
        if (Array.isArray(data)) {
          return { items: data, total: data.length };
        }
        return {
          items: data?.items ?? [],
          total: data?.total ?? 0,
        };
      },
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
      },
    );
  } catch (error) {
    console.warn('[AuditLog] 获取用户列表失败:', error);
    return [];
  }
}