<template>
  <t-card :title="'权限管理'" :bordered="false">
    <t-space direction="vertical" size="large" style="width: 100%">
      <t-space>
        <t-select
          v-model="filterModule"
          placeholder="按模块筛选"
          clearable
          style="width: 200px"
          :options="moduleOptions"
          @change="fetchPermissions"
        />
        <t-select
          v-model="filterResource"
          placeholder="按资源筛选"
          clearable
          style="width: 200px"
          :options="resourceOptions"
          @change="fetchPermissions"
        />
        <t-button @click="resetFilter">重置</t-button>
      </t-space>
      <t-table
        row-key="id"
        :columns="columns"
        :data="permissionList"
        :loading="loading"
        :pagination="false"
        size="small"
        :max-height="600"
      />
    </t-space>
  </t-card>
</template>
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { permissionApi } from '@/api/permission';
import type { Permission } from '@/types';
const permissionList = ref<Permission[]>([]);
const loading = ref(false);
const filterModule = ref<string | null>(null);
const filterResource = ref<string | null>(null);
const columns: PrimaryTableCol<Permission>[] = [
  { colKey: 'id', title: 'ID' },
  { colKey: 'code', title: '权限编码' },
  { colKey: 'name', title: '名称' },
  { colKey: 'module_id', title: '模块' },
  { colKey: 'resource', title: '资源' },
  { colKey: 'action', title: '操作' },
];
const moduleOptions = computed(() => {
  const modules = new Set(permissionList.value.map((p) => p.module_id));
  return Array.from(modules).map((m) => ({ label: m, value: m }));
});
const resourceOptions = computed(() => {
  const resources = new Set(permissionList.value.map((p) => p.resource));
  return Array.from(resources).map((r) => ({ label: r, value: r }));
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
    // no-op
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