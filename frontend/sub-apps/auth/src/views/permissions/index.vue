<template>
  <div>
    <h1>权限管理</h1>
    <n-space vertical size="large">
      <n-space>
        <n-select
          v-model:value="filterModule"
          placeholder="按模块筛选"
          clearable
          style="width: 200px"
          :options="moduleOptions"
          @update:value="fetchPermissions"
        />
        <n-select
          v-model:value="filterResource"
          placeholder="按资源筛选"
          clearable
          style="width: 200px"
          :options="resourceOptions"
          @update:value="fetchPermissions"
        />
        <n-button @click="resetFilter">重置</n-button>
      </n-space>
      <n-data-table
        :columns="columns"
        :data="permissionList"
        :loading="loading"
        :pagination="false"
        size="small"
        max-height="600"
      />
    </n-space>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { NDataTable, NSpace, NSelect, NButton, useMessage } from 'naive-ui';
import { permissionApi } from '@/api/permission';
import type { Permission } from '@/types';
const message = useMessage();
const permissionList = ref<Permission[]>([]);
const loading = ref(false);
const filterModule = ref<string | null>(null);
const filterResource = ref<string | null>(null);
const columns = [
  { title: 'ID', key: 'id' },
  { title: '权限编码', key: 'code' },
  { title: '名称', key: 'name' },
  { title: '模块', key: 'module_id' },
  { title: '资源', key: 'resource' },
  { title: '操作', key: 'action' },
];
const moduleOptions = computed(() => {
  const modules = new Set(permissionList.value.map(p => p.module_id));
  return Array.from(modules).map(m => ({ label: m, value: m }));
});
const resourceOptions = computed(() => {
  const resources = new Set(permissionList.value.map(p => p.resource));
  return Array.from(resources).map(r => ({ label: r, value: r }));
});
async function fetchPermissions() {
  loading.value = true;
  try {
    const params: any = {};
    if (filterModule.value) params.module_id = filterModule.value;
    if (filterResource.value) params.resource = filterResource.value;
    const res = await permissionApi.getList(params);
    permissionList.value = res.data;
  } catch (error: any) {
    message.error(error.message || '加载权限列表失败');
  } finally {
    loading.value = false;
  }
}
function resetFilter() {
  filterModule.value = null;
  filterResource.value = null;
  fetchPermissions();
}
onMounted(fetchPermissions);
</script>
