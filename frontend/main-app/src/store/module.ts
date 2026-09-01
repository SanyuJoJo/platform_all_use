import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Module } from '@/types/module';
import { getActiveModules } from '@/api/module';

export const useModuleStore = defineStore('module', () => {
  const modules = ref<Module[]>([]);
  const loaded = ref(false);
  const loading = ref(false);

  async function fetchModules(force = false): Promise<Module[]> {
    if (loaded.value && !force) {
      return modules.value;
    }
    if (loading.value) {
      // 如果正在加载，等待当前加载完成（简单返回当前值）
      // 实际可以返回一个共享 Promise，但这里简单处理
      return modules.value;
    }
    loading.value = true;
    try {
      const data = await getActiveModules();
      modules.value = data;
      loaded.value = true;
      return data;
    } catch (error) {
      console.error('[ModuleStore] 获取模块列表失败:', error);
      throw error;
    } finally {
      loading.value = false;
    }
  }

  function getActiveModuleIds(): string[] {
    return modules.value.filter(m => m.status === 'active').map(m => m.id);
  }

  return { modules, loaded, loading, fetchModules, getActiveModuleIds };
});
