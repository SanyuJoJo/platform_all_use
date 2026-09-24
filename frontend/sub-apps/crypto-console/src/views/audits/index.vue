<template>
  <div>
    <n-space vertical size="large">
      <n-card title="审计日志">
        <template #header-extra>
          <n-space>
            <n-button @click="fetchList">刷新</n-button>
          </n-space>
        </template>

        <n-alert type="info" :show-icon="true" style="margin-bottom: 12px">
          审计日志为只读展示。前端只展示 params_digest，不展示私钥、口令、主密钥、
          完整敏感参数；查看与导出受 RBAC 控制。
        </n-alert>

        <n-form inline :model="filters" label-placement="left">
          <n-form-item label="request_id">
            <n-input v-model:value="filters.request_id" clearable style="width: 220px" />
          </n-form-item>
          <n-form-item label="operation_id">
            <n-select
              v-model:value="filters.operation_id"
              :options="operationOptions"
              clearable
              style="width: 220px"
            />
          </n-form-item>
          <n-form-item label="结果">
            <n-select
              v-model:value="filters.result"
              :options="resultOptions"
              clearable
              style="width: 160px"
            />
          </n-form-item>
          <n-form-item>
            <n-button type="primary" @click="handleSearch">查询</n-button>
          </n-form-item>
        </n-form>

        <n-data-table
          :columns="columns"
          :data="list"
          :loading="loading"
          :pagination="pagination"
          remote
          @update:page="handlePageChange"
        />
      </n-card>
    </n-space>

    <n-drawer v-model:show="showDetail" :width="560">
      <n-drawer-content title="审计详情">
        <n-descriptions
          v-if="detail"
          bordered
          :column="1"
          label-placement="left"
          size="small"
        >
          <n-descriptions-item label="audit_id">
            {{ detail.audit_id }}
          </n-descriptions-item>
          <n-descriptions-item label="request_id">
            {{ detail.request_id }}
          </n-descriptions-item>
          <n-descriptions-item label="task_id">
            {{ detail.task_id || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="operation_id">
            {{ detail.operation_id }}
          </n-descriptions-item>
          <n-descriptions-item label="actor_type">
            {{ detail.actor_type }}
          </n-descriptions-item>
          <n-descriptions-item label="actor_id">
            {{ detail.actor_id }}
          </n-descriptions-item>
          <n-descriptions-item label="result">
            {{ detail.result }}
          </n-descriptions-item>
          <n-descriptions-item label="duration_ms">
            {{ detail.duration_ms }}
          </n-descriptions-item>
          <n-descriptions-item label="error_code">
            {{ detail.error_code || '—' }}
          </n-descriptions-item>
          <n-descriptions-item label="params_digest">
            {{ detail.params_digest }}
          </n-descriptions-item>
          <n-descriptions-item label="时间">
            {{ detail.ts }}
          </n-descriptions-item>
        </n-descriptions>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NSpace,
  useMessage,
  type DataTableColumns,
} from 'naive-ui';
import { auditApi } from '@/api/audit';

const message = useMessage();

const loading = ref(false);
const list = ref<any[]>([]);
const detail = ref<any>(null);
const showDetail = ref(false);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
});

const filters = reactive({
  request_id: '',
  operation_id: null as string | null,
  result: null as string | null,
});

const operationOptions = [
  { label: 'ca.create', value: 'ca.create' },
  { label: 'ca.intermediate.create', value: 'ca.intermediate.create' },
  { label: 'csr.create', value: 'csr.create' },
  { label: 'cert.sign', value: 'cert.sign' },
  { label: 'dual_cert.create', value: 'dual_cert.create' },
  { label: 'crl.create', value: 'crl.create' },
  { label: 'cert.convert', value: 'cert.convert' },
  { label: 'cert.parse', value: 'cert.parse' },
  { label: 'key.manage', value: 'key.manage' },
  { label: 'pqc.cert.create', value: 'pqc.cert.create' },
  { label: 'chain.verify', value: 'chain.verify' },
  { label: 'batch.execute', value: 'batch.execute' },
];

const resultOptions = [
  { label: 'SUCCESS', value: 'SUCCESS' },
  { label: 'FAILED', value: 'FAILED' },
  { label: 'TIMEOUT', value: 'TIMEOUT' },
  { label: 'CANCELLED', value: 'CANCELLED' },
];

const columns: DataTableColumns<any> = [
  { title: 'audit_id', key: 'audit_id', ellipsis: { tooltip: true } },
  { title: 'request_id', key: 'request_id', ellipsis: { tooltip: true } },
  { title: 'operation_id', key: 'operation_id' },
  { title: 'actor', key: 'actor_id', ellipsis: { tooltip: true } },
  { title: 'result', key: 'result' },
  { title: 'duration_ms', key: 'duration_ms' },
  { title: 'error_code', key: 'error_code' },
  { title: '时间', key: 'ts' },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(
        NButton,
        { size: 'small', onClick: () => openDetail(row) },
        { default: () => '详情' }
      );
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.pageSize,
    };
    if (filters.request_id) params.request_id = filters.request_id;
    if (filters.operation_id) params.operation_id = filters.operation_id;
    if (filters.result) params.result = filters.result;

    const res = await auditApi.listAudits(params);
    const data: any = res.data || {};
    list.value = data.items || data || [];
    pagination.itemCount = data.total || list.value.length;
  } catch (e: any) {
    message.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.page = 1;
  fetchList();
}

function handlePageChange(page: number) {
  pagination.page = page;
  fetchList();
}

async function openDetail(row: any) {
  try {
    const res = await auditApi.getAudit(row.audit_id);
    detail.value = res.data;
    showDetail.value = true;
  } catch (e: any) {
    message.error(e.message || '加载详情失败');
  }
}

onMounted(fetchList);
</script>
