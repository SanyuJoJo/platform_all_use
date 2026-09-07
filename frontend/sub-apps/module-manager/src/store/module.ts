import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Module } from '@/types';
export const useModuleStore = defineStore('module', () => {
  const permissions = ref<string[]>([]);
  const loading = ref(false);
  const list = ref<Module[]>([]);   // ✅ 修复 P2-4：从 any[] 改为 Module[]
  const total = ref(0);
  function setPermissions(perms: string[]) {
    permissions.value = perms;
  }
  function hasPermission(code: string): boolean {
    return permissions.value.includes(code);
  }
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
    permissions,
    loading,
    list,
    total,
    setPermissions,
    hasPermission,
    setLoading,
    setList,
    setTotal,
  };
});
