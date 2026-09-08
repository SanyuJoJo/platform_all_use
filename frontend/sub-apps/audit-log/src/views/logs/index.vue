<template>
  <div>
    <h1>日志审计</h1>
    <n-space vertical size="large">
      <!-- 搜索栏 -->
      <n-space align="center" wrap>
        <n-select
          v-model:value="filters.module_id"
          placeholder="模块"
          clearable
          style="width: 150px"
          :options="moduleOptions"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="filters.user_id"
          placeholder="用户"
          clearable
          style="width: 150px"
          :options="userOptions"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="filters.action"
          placeholder="操作类型"
          clearable
          style="width: 150px"
          :options="actionOptions"
          @update:value="handleSearch"
        />
        <n-date-picker
          v-model:value="filters.start_time"
          type="datetime"
          placeholder="开始时间"
          clearable
          style="width: 200px"
          @update:value="handleSearch"
        />
        <n-date-picker
          v-model:value="filters.end_time"
          type="datetime"
          placeholder="结束时间"
          clearable
          style="width: 200px"
          @update:value="handleSearch"
        />
        <n-input
          v-model:value="filters.keyword"
          placeholder="搜索关键词"
          clearable
          style="width: 200px"
          @input="handleSearchInput"
        />
        <n-button type="primary" @click="handleSearch">搜索</n-button>
        <n-button @click="resetFilters">重置</n-button>
        <n-button
          v-if="canExport"
          type="success"
          :loading="exportLoading"
          @click="handleExport"
        >
          导出 CSV
        </n-button>
      </n-space>
      <!-- 表格 -->
      <n-data-table
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
        size="small"
        remote
      />
    </n-space>
    <!-- 详情弹窗 -->
    <DetailModal
      v-model:visible="detailVisible"
      :log-id="detailLogId"
      @success="fetchLogs"
    />
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted, computed, h } from 'vue';
import {
  NDataTable,
  NSpace,
  NInput,
  NSelect,
  NButton,
  NDatePicker,
  useMessage,
  NTag,
} from 'naive-ui';
import { useDebounceFn } from '@vueuse/core';
import { auditLogApi } from '@/api/audit-log';
import { getModulesForFilter, getUsersForFilter } from '@/api/common';
import DetailModal from './components/DetailModal.vue';
import type { AuditLog } from '@/types';
import { usePermissionStore } from '@/store/permission';
const message = useMessage();
const permissionStore = usePermissionStore();
// 权限
const canExport = computed(() => permissionStore.hasPermission('audit_log:log:export'));
// 列表数据
const logList = ref<AuditLog[]>([]);
const loading = ref(false);
const total = ref(0);
const pagination = reactive({
  page: 1,
  pageSize: 20,
});
// 筛选条件
const filters = reactive({
  module_id: null as string | null,
  user_id: null as number | null,
  action: null as string | null,
  start_time: null as number | null,
  end_time: null as number | null,
  keyword: '',
});
// 下拉选项
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
// 详情弹窗
const detailVisible = ref(false);
const detailLogId = ref<number>(0);
// 表格列定义
const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户名', key: 'username', width: 100 },
  { title: '模块', key: 'module_id', width: 120 },
  {
    title: '操作类型',
    key: 'action',
    width: 100,
    render(row: AuditLog) {
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
  { title: '资源', key: 'resource', width: 100 },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row: AuditLog) {
      return h(
        NTag,
        { type: row.status === 'success' ? 'success' : 'error' },
        { default: () => (row.status === 'success' ? '成功' : '失败') }
      );
    },
  },
  { title: '详情', key: 'detail', ellipsis: true },
  {
    title: '操作时间',
    key: 'created_at',
    width: 180,
    render: (row: AuditLog) => new Date(row.created_at).toLocaleString(),
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render(row: AuditLog) {
      return h(
        NButton,
        { size: 'small', onClick: () => openDetail(row.id) },
        { default: () => '详情' }
      );
    },
  },
];
// 获取模块和用户选项
async function fetchOptions() {
  try {
    const modules = await getModulesForFilter();
    moduleOptions.value = modules.map(m => ({ label: m.name, value: m.id }));
    const users = await getUsersForFilter();
    userOptions.value = users.map(u => ({ label: u.username, value: u.id }));
  } catch (error) {
    // 已在 common.ts 中处理
  }
}
// 获取日志列表
async function fetchLogs() {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
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
    total.value = res.data.total;
  } catch (err: any) {
    message.error(err.message || '加载日志失败');
  } finally {
    loading.value = false;
  }
}
// 搜索（防抖）
const handleSearch = useDebounceFn(() => {
  pagination.page = 1;
  fetchLogs();
}, 300);
// 输入时触发防抖搜索
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
  pagination.page = 1;
  fetchLogs();
}
function onPageChange(page: number) {
  pagination.page = page;
  fetchLogs();
}
function onPageSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.page = 1;
  fetchLogs();
}
function openDetail(id: number) {
  detailLogId.value = id;
  detailVisible.value = true;
}
// 导出
const exportLoading = ref(false);
async function handleExport() {
  if (!canExport.value) {
    message.warning('您没有导出权限');
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
    // 正常情况 response.data 是 Blob
    const blob = new Blob([response.data], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `audit_logs_${new Date().toISOString().slice(0,10)}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
    message.success('导出成功');
  } catch (err: any) {
    // 拦截器已经将业务错误 reject，err.message 包含后端返回的 message
    message.error(err.message || '导出失败');
  } finally {
    exportLoading.value = false;
  }
}
onMounted(() => {
  fetchOptions();
  fetchLogs();
});
</script>