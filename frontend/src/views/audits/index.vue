<template>
  <t-card :title="'审计日志'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button @click="fetchList">刷新</t-button>
      </t-space>
    </template>

    <t-alert
      theme="info"
      :message="'审计日志为只读展示。前端只展示 params_digest，不展示私钥、口令、主密钥、完整敏感参数；查看与导出受 RBAC 控制。'"
      style="margin-bottom: 12px"
    />

    <t-space align="center" wrap style="margin-bottom: 12px">
      <t-input v-model="filters.request_id" placeholder="request_id" clearable style="width: 220px" />
      <t-select
        v-model="filters.operation_id"
        placeholder="operation_id"
        :options="operationOptions"
        clearable
        style="width: 220px"
      />
      <t-select
        v-model="filters.result"
        placeholder="结果"
        :options="resultOptions"
        clearable
        style="width: 160px"
      />
      <t-button theme="primary" @click="handleSearch">查询</t-button>
    </t-space>

    <t-table
      row-key="audit_id"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      @page-change="handlePageChange"
    />
  </t-card>

  <t-drawer :visible="showDetail" :size="560" placement="right" :footer="false" @close="showDetail = false">
    <template #header>审计详情</template>
    <t-descriptions v-if="detail" :column="1" label-align="left" bordered size="small">
      <t-descriptions-item label="audit_id">{{ detail.audit_id }}</t-descriptions-item>
      <t-descriptions-item label="request_id">{{ detail.request_id }}</t-descriptions-item>
      <t-descriptions-item label="task_id">{{ detail.task_id || '—' }}</t-descriptions-item>
      <t-descriptions-item label="operation_id">{{ detail.operation_id }}</t-descriptions-item>
      <t-descriptions-item label="actor_type">{{ detail.actor_type }}</t-descriptions-item>
      <t-descriptions-item label="actor_id">{{ detail.actor_id }}</t-descriptions-item>
      <t-descriptions-item label="result">{{ detail.result }}</t-descriptions-item>
      <t-descriptions-item label="duration_ms">{{ detail.duration_ms }}</t-descriptions-item>
      <t-descriptions-item label="error_code">{{ detail.error_code || '—' }}</t-descriptions-item>
      <t-descriptions-item label="params_digest">{{ detail.params_digest }}</t-descriptions-item>
      <t-descriptions-item label="时间">{{ detail.ts }}</t-descriptions-item>
    </t-descriptions>
  </t-drawer>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import { Button } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { cryptoAuditApi } from '@/api/crypto-audit';

const loading = ref(false);
const list = ref<any[]>([]);
const detail = ref<any>(null);
const showDetail = ref(false);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
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

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'audit_id', title: 'audit_id', ellipsis: true },
  { colKey: 'request_id', title: 'request_id', ellipsis: true },
  { colKey: 'operation_id', title: 'operation_id' },
  { colKey: 'actor_id', title: 'actor', ellipsis: true },
  { colKey: 'result', title: 'result' },
  { colKey: 'duration_ms', title: 'duration_ms' },
  { colKey: 'error_code', title: 'error_code' },
  { colKey: 'ts', title: '时间' },
  {
    colKey: 'actions',
    title: '操作',
    cell: (_h, { row }) =>
      h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openDetail(row) }, () => '详情'),
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const params: Record<string, unknown> = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    if (filters.request_id) params.request_id = filters.request_id;
    if (filters.operation_id) params.operation_id = filters.operation_id;
    if (filters.result) params.result = filters.result;

    const res = await cryptoAuditApi.listAudits(params);
    const data: any = res.data || {};
    list.value = data.items || data || [];
    pagination.total = data.total || list.value.length;
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchList();
}

function handlePageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchList();
}

async function openDetail(row: any) {
  try {
    const res = await cryptoAuditApi.getAudit(row.audit_id);
    detail.value = res.data;
    showDetail.value = true;
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载详情失败');
  }
}

onMounted(fetchList);
</script>