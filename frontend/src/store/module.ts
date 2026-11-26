import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Module } from '@/types';

/**
 * 模块管理本地状态。
 * 权限判断统一走全局 useUserStore.hasPermission，
 * 这里仅保留列表缓存、加载态等 UI 状态。
 */
export const useModuleStore = defineStore('module', () => {
  const loading = ref(false);
  const list = ref<Module[]>([]);
  const total = ref(0);

  function setLoading(val: boolean) {
    loading.value = val;
  }
  function setList(data: Module[]) {
    list.value = data;
  }
  function setTotal(val: number) {
    total.value = val;
  }

  return {
    loading,
    list,
    total,
    setLoading,
    setList,
    setTotal,
  };
});