import request from './index';
import type { Module } from '@/types';
// 获取所有已激活模块（用于筛选下拉）
export async function getModulesForFilter(): Promise<Module[]> {
  try {
    const res = await request.get('/modules', { params: { status: 'active', page_size: 100 } });
    const items = res.data?.items || res.data || [];
    return Array.isArray(items) ? items : [];
  } catch (error) {
    console.warn('[AuditLog] 获取模块列表失败:', error);
    return [];
  }
}
// 获取所有用户（用于筛选下拉）
export async function getUsersForFilter(): Promise<{ id: number; username: string }[]> {
  try {
    const res = await request.get('/auth/users', { params: { page: 1, page_size: 100 } });
    const items = res.data?.items || [];
    return Array.isArray(items) ? items : [];
  } catch (error) {
    console.warn('[AuditLog] 获取用户列表失败:', error);
    return [];
  }
}