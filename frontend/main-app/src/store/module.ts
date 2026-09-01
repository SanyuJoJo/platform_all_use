import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Module } from '@/types/module';
import { getActiveModules } from '@/api/module';
import { message } from '@/utils/naive';
import { eventBus } from '@/micro-frontend/event-bus';
export const useModuleStore = defineStore('module', () => {
  const modules = ref<Module[]>([]);
  const loaded = ref(false);
  async function fetchModules(force = false) {
    if (loaded.value && !force) {
      return modules.value;
    }
    try {
      const data = await getActiveModules();
      modules.value = data;
      loaded.value = true;
      return data;
    } catch (error) {
      console.error('[ModuleStore] 获取模块列表失败:', error);
      throw error;
    }
  }
  /**
   * 强制刷新模块列表
   * 发布事件通知菜单刷新（解耦）
   */
  async function refreshModules() {
    loaded.value = false;
    try {
      const data = await fetchModules(true);
      // ✅ 修复：通过事件总线通知菜单刷新，避免循环依赖
      eventBus.emit('modules:refreshed', data);
      return data;
    } catch (error) {
      throw error;
    }
  }
  function getActiveModuleIds(): string[] {
    return modules.value.filter((m) => m.status === 'active').map((m) => m.id);
  }
  return { modules, loaded, fetchModules, refreshModules, getActiveModuleIds };
});
