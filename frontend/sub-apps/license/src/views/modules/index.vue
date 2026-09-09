<template>
  <div>
    <h1>模块授权状态</h1>
    <n-space vertical size="large">
      <n-button @click="fetchModules">刷新</n-button>
      <n-data-table
        :columns="columns"
        :data="moduleList"
        :loading="loading"
        :pagination="false"
        size="small"
        max-height="600"
      />
      <n-empty v-if="!loading && moduleList.length === 0" description="暂无模块授权信息" />
    </n-space>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted, h } from 'vue';
import { NDataTable, NSpace, NButton, NTag, useMessage } from 'naive-ui';
import { licenseApi } from '@/api/license';
import type { ModuleAuthStatus } from '@/types';
const message = useMessage();
const moduleList = ref<ModuleAuthStatus[]>([]);
const loading = ref(false);
const columns = [
  { title: '模块ID', key: 'module_id' },
  { title: '模块名称', key: 'module_name' },
  {
    title: '授权状态',
    key: 'is_authorized',
    render(row: ModuleAuthStatus) {
      return h(
        NTag,
        { type: row.is_authorized ? 'success' : 'error' },
        { default: () => row.is_authorized ? '已授权' : '未授权' }
      );
    },
  },
  {
    title: '过期时间',
    key: 'expires_at',
    render(row: ModuleAuthStatus) {
      return row.expires_at ? new Date(row.expires_at).toLocaleString() : '无限制';
    },
  },
];
async function fetchModules() {
  loading.value = true;
  try {
    const res = await licenseApi.getModulesAuth();
    moduleList.value = res.data;
  } catch (error: any) {
    message.error(error.message || '获取模块授权状态失败');
  } finally {
    loading.value = false;
  }
}
onMounted(fetchModules);
</script>
