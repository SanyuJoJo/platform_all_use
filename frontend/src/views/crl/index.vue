<template>
  <t-card :title="'CRL 管理'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button
          v-permission="'crypto_console:crl:create'"
          theme="primary"
          @click="showCreate = true"
        >
          生成 CRL
        </t-button>
        <t-button @click="fetchList">刷新</t-button>
      </t-space>
    </template>

    <t-table
      row-key="crl_id"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      @page-change="handlePageChange"
    />
  </t-card>

  <t-dialog
    :visible="showCreate"
    :header="'生成 CRL'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="680"
    @close="showCreate = false"
  >
    <t-form :data="createForm" label-align="left" :label-width="120">
      <t-form-item label="CA ID">
        <t-input v-model="createForm.ca_id" />
      </t-form-item>
      <t-form-item label="CA 密钥引用">
        <t-input v-model="createForm.ca_key_ref" />
      </t-form-item>
      <t-form-item label="撤销序列号">
        <t-tag-input v-model="createForm.revoked_serials" placeholder="输入后回车，例如 0A1B2C..." clearable />
      </t-form-item>
      <t-form-item label="摘要算法">
        <t-select v-model="createForm.digest_algorithm" :options="digestOptions" />
      </t-form-item>
      <t-form-item label="本次更新">
        <t-input v-model="createForm.this_update" placeholder="可选，例如 2026-09-24T00:00:00Z" />
      </t-form-item>
      <t-form-item label="下次更新">
        <t-input v-model="createForm.next_update" placeholder="可选，例如 2026-10-24T00:00:00Z" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showCreate = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" @click="handleCreate">
          提交
        </t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import { Button, Space } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { cryptoApi } from '@/api/crypto';
import { useOperation } from '@/composables/useOperation';

const { run } = useOperation();

const loading = ref(false);
const submitting = ref(false);
const showCreate = ref(false);
const list = ref<any[]>([]);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const createForm = reactive({
  ca_id: '',
  ca_key_ref: '',
  revoked_serials: [''] as string[],
  digest_algorithm: 'SM3',
  this_update: '',
  next_update: '',
});

const digestOptions = [
  { label: 'SM3', value: 'SM3' },
  { label: 'SHA256', value: 'SHA256' },
  { label: 'SHA384', value: 'SHA384' },
  { label: 'SHA512', value: 'SHA512' },
];

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'crl_id', title: 'CRL ID' },
  { colKey: 'ca_id', title: 'CA ID' },
  { colKey: 'revoked_count', title: '撤销数量' },
  { colKey: 'next_update', title: '下次更新' },
  { colKey: 'status', title: '状态' },
  { colKey: 'created_at', title: '创建时间' },
  {
    colKey: 'actions',
    title: '操作',
    cell: (_h, { row }) =>
      h(Space, { size: 4 }, () => [
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleDownload(row) }, () => '下载'),
      ]),
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCrls({
      page: pagination.current,
      page_size: pagination.pageSize,
    });
    const data: any = res.data || {};
    list.value = data.items || data || [];
    pagination.total = data.total || list.value.length;
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

async function handleCreate() {
  if (!createForm.ca_id || !createForm.ca_key_ref) {
    MessagePlugin.warning('请填写 CA ID 与 CA 密钥引用');
    return;
  }
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      ca_id: createForm.ca_id,
      ca_key_ref: createForm.ca_key_ref,
      revoked_serials: createForm.revoked_serials.filter((s) => s && s.trim()),
      digest_algorithm: createForm.digest_algorithm,
    };
    if (createForm.this_update) params.this_update = createForm.this_update;
    if (createForm.next_update) params.next_update = createForm.next_update;

    await run('crl.create', params);
    showCreate.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

function handleDownload(row: any) {
  const url = `${import.meta.env.VITE_API_BASE_URL || ''}/api/v1/crls/${row.crl_id}/download`;
  const a = document.createElement('a');
  a.href = url;
  a.target = '_blank';
  a.click();
}

onMounted(fetchList);
</script>