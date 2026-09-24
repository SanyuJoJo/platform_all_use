<template>
  <div>
    <n-space vertical size="large">
      <n-card title="CRL 管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:crl:create'"
              type="primary"
              @click="showCreate = true"
            >
              生成 CRL
            </n-button>
            <n-button @click="fetchList">刷新</n-button>
          </n-space>
        </template>

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

    <n-modal
      v-model:show="showCreate"
      preset="card"
      title="生成 CRL"
      style="width: 680px"
    >
      <n-form :model="createForm" label-placement="left" label-width="120">
        <n-form-item label="CA ID">
          <n-input v-model:value="createForm.ca_id" />
        </n-form-item>
        <n-form-item label="CA 密钥引用">
          <n-input v-model:value="createForm.ca_key_ref" />
        </n-form-item>
        <n-form-item label="撤销序列号">
          <n-dynamic-input
            v-model:value="createForm.revoked_serials"
            :on-create="() => ''"
            placeholder="例如 0A1B2C..."
          />
        </n-form-item>
        <n-form-item label="摘要算法">
          <n-select v-model:value="createForm.digest_algorithm" :options="digestOptions" />
        </n-form-item>
        <n-form-item label="本次更新">
          <n-input
            v-model:value="createForm.this_update"
            placeholder="可选，例如 2026-09-24T00:00:00Z"
          />
        </n-form-item>
        <n-form-item label="下次更新">
          <n-input
            v-model:value="createForm.next_update"
            placeholder="可选，例如 2026-10-24T00:00:00Z"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleCreate">
            提交
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import {
  NButton,
  NCard,
  NDataTable,
  NDynamicInput,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpace,
  useMessage,
  type DataTableColumns,
} from 'naive-ui';
import { cryptoApi } from '@/api/crypto';
import { useOperation } from '@/composables/useOperation';

const message = useMessage();
const { run } = useOperation();

const loading = ref(false);
const submitting = ref(false);
const showCreate = ref(false);
const list = ref<any[]>([]);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
});

const createForm = reactive({
  ca_id: '',
  ca_key_ref: '',
  revoked_serials: [''],
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

const columns: DataTableColumns<any> = [
  { title: 'CRL ID', key: 'crl_id' },
  { title: 'CA ID', key: 'ca_id' },
  { title: '撤销数量', key: 'revoked_count' },
  { title: '下次更新', key: 'next_update' },
  { title: '状态', key: 'status' },
  { title: '创建时间', key: 'created_at' },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            { size: 'small', onClick: () => handleDownload(row) },
            { default: () => '下载' }
          ),
        ],
      });
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCrls({
      page: pagination.page,
      page_size: pagination.pageSize,
    });
    const data: any = res.data || {};
    list.value = data.items || data || [];
    pagination.itemCount = data.total || list.value.length;
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

async function handleCreate() {
  if (!createForm.ca_id || !createForm.ca_key_ref) {
    message.warning('请填写 CA ID 与 CA 密钥引用');
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
