import request from './index';
import type { Module } from '@/types/module';
/**
 * 获取已激活的模块列表（包含 menus）
 * @returns Promise<Module[]>
 */
export async function getActiveModules(): Promise<Module[]> {
  try {
    const res = await request.get('/modules', {
      params: { status: 'active' },
    });
    // 兼容响应格式
    const items = res.data?.items || res.data || [];
    return Array.isArray(items) ? items : [];
  } catch (error) {
    console.error('[API] 获取模块列表失败:', error);
    throw error;
  }
}
