// src/api/module.ts
import request from './index';
import type { Module } from '@/types/module';

// 真实后端接口（若未实现，可用以下 Mock 版本）
export async function getActiveModules(): Promise<Module[]> {
  const res = await request.get('/modules', { params: { status: 'active' } });
  return res.data?.items || res.data || [];
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