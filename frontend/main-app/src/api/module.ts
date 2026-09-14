import request from './index';
import type { Module } from '@/types/module';
/**
 * 获取已激活模块列表。
 *
 * 兼容后端返回格式：
 * - { code, data: { items: Module[], total, ... } }
 * - { code, data: Module[] }
 *
 * 兼容字段缺失：
 * - entry_frontend 缺省 → 空字符串（由 registry 兜底）
 * - menus 缺省 → 空数组
 * - status 缺省 → 'active'
 */
export async function getActiveModules(): Promise<Module[]> {
  const res = await request.get('/modules', {
    params: { status: 'active' },
  });
  const raw = res.data?.items || res.data || [];
  if (!Array.isArray(raw)) {
    console.warn('[API] /modules 返回格式非数组：', raw);
    return [];
  }
  return raw.map((m: any) => ({
    ...m,
    status: m.status || 'active',
    entry_frontend: typeof m.entry_frontend === 'string' ? m.entry_frontend : '',
    menus: Array.isArray(m.menus) ? m.menus : [],
    dependencies: Array.isArray(m.dependencies) ? m.dependencies : [],
  }));
}
export async function createModule(data: Partial<Module>): Promise<Module> {
  const res = await request.post('/modules', data);
  return res.data;
}
export async function updateModule(id: string, data: Partial<Module>): Promise<Module> {
  const res = await request.put(`/modules/${id}`, data);
  return res.data;
}
export async function deleteModule(id: string): Promise<void> {
  await request.delete(`/modules/${id}`);
}
