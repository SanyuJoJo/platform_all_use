<template>
  <t-card :title="'任务中心'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button @click="refreshAll">刷新</t-button>
      </t-space>
    </template>

    <t-alert
      theme="info"
      :message="'长任务由平台后端封装 task_id，前端按固定间隔轮询。非幂等密码操作默认不自动重试，超时、取消以平台后端为准。'"
      style="margin-bottom: 12px"
    />

    <t-table
      row-key="task_id"
      :columns="columns"
      :data="tasks"
      :loading="loading"
      :pagination="pagination"
      @page-change="handlePageChange"
    />
  </t-card>

  <t-drawer :visible="showDetail" :size="560" placement="right" :footer="false" @close="showDetail = false">
    <template #header>任务详情</template>
    <t-descriptions v-if="detail" :column="1" label-align="left" bordered size="small">
      <t-descriptions-item label="task_id">{{ detail.task_id }}</t-descriptions-item>
      <t-descriptions-item label="operation_id">{{ detail.operation_id }}</t-descriptions-item>
      <t-descriptions-item label="request_id">{{ detail.request_id }}</t-descriptions-item>
      <t-descriptions-item label="状态">
        <t-tag :theme="statusType(detail.status)" variant="light">{{ detail.status }}</t-tag>
      </t-descriptions-item>
      <t-descriptions-item label="进度">{{ detail.progress ?? '—' }}</t-descriptions-item>
      <t-descriptions-item label="结果引用">{{ detail.result_ref || '—' }}</t-descriptions-item>
      <t-descriptions-item label="错误码">{{ detail.error_code || '—' }}</t-descriptions-item>
      <t-descriptions-item label="错误消息">{{ detail.error_message || '—' }}</t-descriptions-item>
      <t-descriptions-item label="创建时间">{{ detail.created_at || '—' }}</t-descriptions-item>
      <t-descriptions-item label="开始时间">{{ detail.started_at || '—' }}</t-descriptions-item>
      <t-descriptions-item label="结束时间">{{ detail.finished_at || '—' }}</t-descriptions-item>
    </t-descriptions>
  </t-drawer>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, h } from 'vue';
import { Button, Space, Tag } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { cryptoTaskApi } from '@/api/crypto-task';
import { useTaskStore } from '@/store/task';

const taskStore = useTaskStore();

const loading = ref(false);
const tasks = ref<any[]>([]);
const detail = ref<any>(null);
const showDetail = ref(false);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

function statusType(status: string): 'success' | 'danger' | 'warning' | 'default' | 'primary' {
  switch (status) {
    case 'SUCCESS':
      return 'success';
    case 'FAILED':
      return 'danger';
    case 'TIMEOUT':
      return 'warning';
    case 'CANCELLED':
      return 'default';
    case 'RUNNING':
      return 'primary';
    default:
      return 'default';
  }
}

const TERMINAL = ['SUCCESS', 'FAILED', 'TIMEOUT', 'CANCELLED'];

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'task_id', title: 'task_id', ellipsis: true },
  { colKey: 'operation_id', title: 'operation_id' },
  { colKey: 'request_id', title: 'request_id', ellipsis: true },
  {
    colKey: 'status',
    title: '状态',
    cell: (_h, { row }) =>
      h(Tag, { theme: statusType(row.status), variant: 'light' }, () => row.status),
  },
  { colKey: 'progress', title: '进度' },
  { colKey: 'created_at', title: '创建时间' },
  {
    colKey: 'actions',
    title: '操作',
    cell: (_h, { row }) => {
      const children = [
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openDetail(row) }, () => '详情'),
      ];
      if (!TERMINAL.includes(row.status)) {
        children.push(
          h(
            Button,
            { size: 'small', variant: 'text', theme: 'warning', onClick: () => cancelTask(row) },
            () => '取消',
          ),
        );
      }
      return h(Space, { size: 4 }, () => children);
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoTaskApi.listTasks({
      page: pagination.current,
      page_size: pagination.pageSize,
    });
    const data: any = res.data || {};
    tasks.value = data.items || data || [];
    pagination.total = data.total || tasks.value.length;

    tasks.value.forEach((t) => {
      if (!TERMINAL.includes(t.status)) {
        taskStore.watch(t.task_id);
      }
    });
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function handlePageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchList();
}

async function refreshAll() {
  await fetchList();
}

async function openDetail(row: any) {
  try {
    const res = await cryptoTaskApi.getTask(row.task_id);
    detail.value = res.data;
    showDetail.value = true;
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载详情失败');
  }
}

async function cancelTask(row: any) {
  try {
    await cryptoTaskApi.cancelTask(row.task_id);
    MessagePlugin.success('已提交取消请求');
    taskStore.stop(row.task_id);
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '取消失败');
  }
}

onMounted(fetchList);

onUnmounted(() => {
  taskStore.stopAll();
});
</script>