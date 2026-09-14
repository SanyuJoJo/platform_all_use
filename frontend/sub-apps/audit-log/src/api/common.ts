import request from './index';
import type { Module } from '@/types';

/**
 * 通用分页响应结构。
 * 与 /types 中的 PaginatedResponse 保持一致，这里单独声明避免循环依赖。
 */
interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  pages: number;
}

/**
 * 获取所有已激活模块（用于筛选下拉）。
 * 兼容后端两种返回：
 * - 分页结构 { items: [...] }
 * - 直接数组 [...]
 */
export async function getModulesForFilter(): Promise<Module[]> {
  try {
    const res = await request.get<
      PaginatedResponse<Module> | Module[]
    >('/modules', {
      params: { status: 'active', page_size: 999 },
    });

    const data = res.data;
    if (Array.isArray(data)) {
      return data;
    }
    if (data && Array.isArray(data.items)) {
      return data.items;
    }
    return [];
  } catch (error) {
    console.warn('[AuditLog] 获取模块列表失败:', error);
    return [];
  }
}

/**
 * 获取所有用户（用于筛选下拉）。
 */
export async function getUsersForFilter(): Promise<
  { id: number; username: string }[]
> {
  try {
    const res = await request.get<
      PaginatedResponse<{ id: number; username: string }>
    >('/auth/users', {
      params: { page: 1, page_size: 999 },
    });

    const data = res.data;
    if (Array.isArray(data)) {
      return data;
    }
    if (data && Array.isArray(data.items)) {
      return data.items;
    }
    return [];
  } catch (error) {
    console.warn('[AuditLog] 获取用户列表失败:', error);
    return [];
  }
}
