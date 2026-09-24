<template>
  <div>
    <n-space vertical size="large">
      <n-card title="任务中心">
        <template #header-extra>
          <n-space>
            <n-button @click="refreshAll">刷新</n-button>
          </n-space>
        </template>

        <n-alert type="info" :show-icon="true" style="margin-bottom: 12px">
          长任务由平台后端封装 task_id，前端按固定间隔轮询。非幂等密码操作默认不自动重试，
          超时、取消以平台后端为准。
        </n-alert>

        <n-data-table
          :columns="columns"
          :data="tasks"
          :loading="loading"
          :pagination="pagination"
          remote
          @update:page="handlePageChange"
        />
      </n-card>
    </n-space>

    <n-drawer v-model:show="showDetail" :width="560">
      <n-drawer-content title="任务详情">
        <n-descriptions
          v-if="detail"
          bordered
          :column="1"
          label-placement="left"
          size="small"
        >
          <n-descriptions-item label="task_id">
            {{ detail.task_id }}
          </n-descriptions-item>
          <n-descriptions-item label="operation_id">
            {{ detail.operation_id }}
          </n-descriptions-item>
          <n-descriptions-item label="request_id">
            {{ detail.request_id }}
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag :type="statusType(detail.status)">
              {{ detail.status }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="进度">
            {{ detail.progress ?? '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="结果引用">
            {{ detail.result_ref || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="错误码">
            {{ detail.error_code || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="错误消息">
            {{ detail.error_message || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="创建时间">
            {{ detail.created_at || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="开始时间">
            {{ detail.started_at || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="结束时间">
            {{ detail.finished_at || '—' }}
          </n-descriptions-item>
        </n-descriptions>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, h } from 'vue';
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NSpace,
  NTag,
  useMessage,
  type DataTableColumns,
} from 'naive-ui';
import { taskApi } from '@/api/task';
import { useTaskStore } from '@/store/task';

const message = useMessage();
const taskStore = useTaskStore();

const loading = ref(false);
const tasks = ref<any[]>([]);
const detail = ref<any>(null);
const showDetail = ref(false);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
});

function statusType(status: string) {
  switch (status) {
    case 'SUCCESS':
      return 'success';
    case 'FAILED':
      return 'error';
    case 'TIMEOUT':
      return 'warning';
    case 'CANCELLED':
      return 'default';
    case 'RUNNING':
      return 'info';
    default:
      return 'default';
  }
}

const TERMINAL = ['SUCCESS', 'FAILED', 'TIMEOUT', 'CANCELLED'];

const columns: DataTableColumns<any> = [
  { title: 'task_id', key: 'task_id', ellipsis: { tooltip: true } },
  { title: 'operation_id', key: 'operation_id' },
  { title: 'request_id', key: 'request_id', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    render(row) {
      return h(
        NTag,
        { type: statusType(row.status) as any },
        { default: () => row.status }
      );
    },
  },
  { title: '进度', key: 'progress' },
  { title: '创建时间', key: 'created_at' },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      const children = [
        h(
          NButton,
          { size: 'small', onClick: () => openDetail(row) },
          { default: () => '详情' }
        ),
      ];
      if (!TERMINAL.includes(row.status)) {
        children.push(
          h(
            NButton,
            {
              size: 'small',
              type: 'warning',
              onClick: () => cancelTask(row),
            },
            { default: () => '取消' }
          )
        );
      }
      return h(NSpace, null, { default: () => children });
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await taskApi.listTasks({
      page: pagination.page,
      page_size: pagination.pageSize,
    });
    const data: any = res.data || {};
    tasks.value = data.items || data || [];
    pagination.itemCount = data.total || tasks.value.length;

    // 对未终态的任务自动开启轮询
    tasks.value.forEach((t) => {
      if (!TERMINAL.includes(t.status)) {
        taskStore.watch(t.task_id);
      }
    });
  } catch (e: any) {
    message.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function handlePageChange(page: number) {
  pagination.page = page;
  fetchList();
}

async function refreshAll() {
  await fetchList();
}

async function openDetail(row: any) {
  try {
    const res = await taskApi.getTask(row.task_id);
    detail.value = res.data;
    showDetail.value = true;
  } catch (e: any) {
    message.error(e.message || '加载详情失败');
  }
}

async function cancelTask(row: any) {
  try {
    await taskApi.cancelTask(row.task_id);
    message.success('已提交取消请求');
    taskStore.stop(row.task_id);
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '取消失败');
  }
}

onMounted(fetchList);

onUnmounted(() => {
  // 离开页面时停止所有前端轮询，避免泄漏
  taskStore.stopAll();
});
</script>
