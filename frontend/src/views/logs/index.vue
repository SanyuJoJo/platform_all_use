<template>
  <t-card :title="'日志审计'" :bordered="false">
    <t-space direction="vertical" size="large" style="width: 100%">
      <t-space align="center" wrap>
        <t-select
          v-model="filters.module_id"
          placeholder="模块"
          clearable
          style="width: 150px"
          :options="moduleOptions"
          @change="handleSearch"
        />
        <t-select
          v-model="filters.user_id"
          placeholder="用户"
          clearable
          style="width: 150px"
          :options="userOptions"
          @change="handleSearch"
        />
        <t-select
          v-model="filters.action"
          placeholder="操作类型"
          clearable
          style="width: 150px"
          :options="actionOptions"
          @change="handleSearch"
        />
        <t-date-picker
          v-model="filters.start_time"
          enable-time-picker
          value-type="timestamp"
          placeholder="开始时间"
          clearable
          style="width: 200px"
          @change="handleSearch"
        />
        <t-date-picker
          v-model="filters.end_time"
          enable-time-picker
          value-type="timestamp"
          placeholder="结束时间"
          clearable
          style="width: 200px"
          @change="handleSearch"
        />
        <t-input
          v-model="filters.keyword"
          placeholder="搜索关键词"
          clearable
          style="width: 200px"
          @input="handleSearchInput"
        />
        <t-button theme="primary" @click="handleSearch">搜索</t-button>
        <t-button @click="resetFilters">重置</t-button>
        <t-button
          v-if="canExport"
          theme="success"
          :loading="exportLoading"
          @click="handleExport"
        >
          导出 CSV
        </t-button>
      </t-space>

      <t-table
        row-key="id"
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        size="small"
        @page-change="onPageChange"
      />
    </t-space>

    <DetailModal
      v-model:visible="detailVisible"
      :log-id="detailLogId"
      @success="fetchLogs"
    />
  </t-card>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted, computed, h } from 'vue';
import { Tag, Button, MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { useDebounceFn } from '@vueuse/core';
import { auditLogApi } from '@/api/audit-log';
import { getModulesForFilter, getUsersForFilter } from '@/api/audit-log-common';
import DetailModal from './components/DetailModal.vue';
import type { AuditLog } from '@/types';
import { useUserStore } from '@/store/user';

const userStore = useUserStore();
const canExport = computed(() => userStore.hasPermission('audit_log:log:export'));

const logList = ref<AuditLog[]>([]);
const loading = ref(false);
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const filters = reactive({
  module_id: null as string | null,
  user_id: null as number | null,
  action: null as string | null,
  start_time: null as number | null,
  end_time: null as number | null,
  keyword: '',
});

const moduleOptions = ref<{ label: string; value: string }[]>([]);
const userOptions = ref<{ label: string; value: number }[]>([]);
const actionOptions = [
  { label: '登录', value: 'login' },
  { label: '登出', value: 'logout' },
  { label: '创建', value: 'create' },
  { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '查看', value: 'view' },
  { label: '导出', value: 'export' },
  { label: '其他', value: 'other' },
];

const detailVisible = ref(false);
const detailLogId = ref<number>(0);

const columns: PrimaryTableCol<AuditLog>[] = [
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'username', title: '用户名', width: 100 },
  { colKey: 'module_id', title: '模块', width: 120 },
  {
    colKey: 'action',
    title: '操作类型',
    width: 100,
    cell: (_h, { row }) => {
      const map: Record<string, string> = {
        login: '登录',
        logout: '登出',
        create: '创建',
        update: '更新',
        delete: '删除',
        view: '查看',
        export: '导出',
      };
      return map[row.action] || row.action;
    },
  },
  { colKey: 'resource', title: '资源', width: 100 },
  {
    colKey: 'status',
    title: '状态',
    width: 80,
    cell: (_h, { row }) =>
      h(Tag, { theme: row.status === 'success' ? 'success' : 'danger', variant: 'light' }, () =>
        row.status === 'success' ? '成功' : '失败',
      ),
  },
  { colKey: 'detail', title: '详情', ellipsis: true },
  {
    colKey: 'created_at',
    title: '操作时间',
    width: 180,
    cell: (_h, { row }) => new Date(row.created_at).toLocaleString(),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 100,
    cell: (_h, { row }) =>
      h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openDetail(row.id) }, () => '详情'),
  },
];

async function fetchOptions() {
  try {
    const modules = await getModulesForFilter();
    moduleOptions.value = modules.map((m) => ({ label: m.name, value: m.id }));
    const users = await getUsersForFilter();
    userOptions.value = users.map((u) => ({ label: u.username, value: u.id }));
  } catch (error) {
    // handled in common
  }
}

async function fetchLogs() {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    if (filters.module_id) params.module_id = filters.module_id;
    if (filters.user_id) params.user_id = filters.user_id;
    if (filters.action) params.action = filters.action;
    if (filters.start_time) params.start_time = new Date(filters.start_time).toISOString();
    if (filters.end_time) params.end_time = new Date(filters.end_time).toISOString();
    if (filters.keyword) params.keyword = filters.keyword;
    const res = await auditLogApi.getList(params);
    logList.value = res.data.items;
    pagination.total = res.data.total;
  } catch (err: any) {
    MessagePlugin.error(err.message || '加载日志失败');
  } finally {
    loading.value = false;
  }
}

const handleSearch = useDebounceFn(() => {
  pagination.current = 1;
  fetchLogs();
}, 300);

function handleSearchInput() {
  handleSearch();
}

function resetFilters() {
  filters.module_id = null;
  filters.user_id = null;
  filters.action = null;
  filters.start_time = null;
  filters.end_time = null;
  filters.keyword = '';
  pagination.current = 1;
  fetchLogs();
}

function onPageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchLogs();
}

function openDetail(id: number) {
  detailLogId.value = id;
  detailVisible.value = true;
}

const exportLoading = ref(false);
async function handleExport() {
  if (!canExport.value) {
    MessagePlugin.warning('您没有导出权限');
    return;
  }
  exportLoading.value = true;
  try {
    const params: any = {};
    if (filters.module_id) params.module_id = filters.module_id;
    if (filters.user_id) params.user_id = filters.user_id;
    if (filters.action) params.action = filters.action;
    if (filters.start_time) params.start_time = new Date(filters.start_time).toISOString();
    if (filters.end_time) params.end_time = new Date(filters.end_time).toISOString();
    if (filters.keyword) params.keyword = filters.keyword;
    const response = await auditLogApi.exportLogs(params);
    const blob = new Blob([response.data], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `audit_logs_${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
    MessagePlugin.success('导出成功');
  } catch (err: any) {
    MessagePlugin.error(err.message || '导出失败');
  } finally {
    exportLoading.value = false;
  }
}

onMounted(() => {
  fetchOptions();
  fetchLogs();
});
</script>