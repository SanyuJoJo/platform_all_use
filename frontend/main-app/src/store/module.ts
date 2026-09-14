import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Module } from '@/types/module';
import {
  getActiveModules,
  createModule,
  updateModule,
  deleteModule,
} from '@/api/module';
import { reRegister } from '@/micro-frontend/registry';
import { eventBus } from '@/micro-frontend/event-bus';

const STORAGE_KEY = 'module_list_cache';

export const useModuleStore = defineStore('module', () => {
  const modules = ref<Module[]>([]);
  const loaded = ref(false);
  const lastError = ref<string | null>(null);

  function loadCache(): Module[] {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      return data ? JSON.parse(data) : [];
    } catch {
      return [];
    }
  }

  function saveCache(data: Module[]) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
    } catch (e) {
      console.warn('[ModuleStore] 缓存模块列表失败', e);
    }
  }

  async function fetchModules(force = false): Promise<Module[]> {
    if (loaded.value && !force) {
      return modules.value;
    }

    try {
      const data = await getActiveModules();
      modules.value = data;
      loaded.value = true;
      lastError.value = null;
      saveCache(data);
      eventBus.emit('modules:changed', data);
      return data;
    } catch (error: unknown) {
      const code = (error as { code?: number })?.code;

      if (code === 403 || code === 20051) {
        console.warn(
          '[ModuleStore] 无权访问 /modules 接口（403），已降级。' +
            '建议后端提供已登录用户可访问的 /modules/active 接口。'
        );
        lastError.value = '无权访问模块列表接口';
        const cached = loadCache();
        modules.value = cached;
        loaded.value = true;
        return cached;
      }

      console.warn('[ModuleStore] 获取模块列表失败，尝试使用缓存', error);
      lastError.value =
        (error as { message?: string })?.message || '获取模块列表失败';
      const cached = loadCache();
      modules.value = cached.length > 0 ? cached : [];
      loaded.value = true;
      return modules.value;
    }
  }

  async function addModule(moduleData: Partial<Module>) {
    const newModule = await createModule(moduleData);
    await fetchModules(true);
    reRegister();
    return newModule;
  }

  async function updateModuleById(id: string, data: Partial<Module>) {
    const updated = await updateModule(id, data);
    await fetchModules(true);
    reRegister();
    return updated;
  }

  async function removeModule(id: string) {
    await deleteModule(id);
    await fetchModules(true);
    reRegister();
  }

  function getActiveModuleIds(): string[] {
    return modules.value.filter((m) => m.status === 'active').map((m) => m.id);
  }

  return {
    modules,
    loaded,
    lastError,
    fetchModules,
    addModule,
    updateModuleById,
    removeModule,
    getActiveModuleIds,
  };
});
