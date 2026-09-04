// src/store/module.ts
import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Module } from '@/types/module';
import { getActiveModules, createModule, updateModule, deleteModule } from '@/api/module';
import { message } from '@/utils/naive';
import { reRegister } from '@/micro-frontend/registry';
import router from '@/router';

const STORAGE_KEY = 'module_list_cache';

export const useModuleStore = defineStore('module', () => {
  const modules = ref<Module[]>([]);
  const loaded = ref(false);

  function loadCache(): Module[] {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      return data ? JSON.parse(data) : [];
    } catch {
      return [];
    }
  }

  function saveCache(data: Module[]) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  }

  async function fetchModules(force = false) {
    if (loaded.value && !force) {
      return modules.value;
    }
    try {
      const data = await getActiveModules();
      modules.value = data;
      loaded.value = true;
      saveCache(data);
      return data;
    } catch (error) {
      console.warn('[ModuleStore] 后端获取失败，使用缓存');
      const cached = loadCache();
      if (cached.length > 0) {
        modules.value = cached;
        loaded.value = true;
        return cached;
      }
      throw error;
    }
  }

  async function addModule(moduleData: Partial<Module>) {
    try {
      const newModule = await createModule(moduleData);
      await fetchModules(true);
      reRegister(router);
      return newModule;
    } catch (error) {
      console.error('[ModuleStore] 添加模块失败', error);
      throw error;
    }
  }

  async function updateModule(id: string, data: Partial<Module>) {
    try {
      const updated = await updateModule(id, data);
      await fetchModules(true);
      reRegister(router);
      return updated;
    } catch (error) {
      console.error('[ModuleStore] 更新模块失败', error);
      throw error;
    }
  }

  async function removeModule(id: string) {
    try {
      await deleteModule(id);
      await fetchModules(true);
      reRegister(router);
    } catch (error) {
      console.error('[ModuleStore] 删除模块失败', error);
      throw error;
    }
  }

  function getActiveModuleIds(): string[] {
    return modules.value.filter(m => m.status === 'active').map(m => m.id);
  }

  return {
    modules,
    loaded,
    fetchModules,
    addModule,
    updateModule,
    removeModule,
    getActiveModuleIds,
  };
});