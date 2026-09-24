<template>
  <div>
    <n-space vertical size="large">
      <n-card title="CSR 管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:csr:create'"
              type="primary"
              @click="showCreate = true"
            >
              生成 CSR
            </n-button>
            <n-button @click="showImport = true">导入 CSR</n-button>
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

    <!-- 生成 CSR -->
    <n-modal
      v-model:show="showCreate"
      preset="card"
      title="生成 CSR"
      style="width: 720px"
    >
      <n-form :model="createForm" label-placement="left" label-width="110">
        <n-form-item label="算法">
          <n-select v-model:value="createForm.algorithm" :options="algorithmOptions" />
        </n-form-item>
        <n-form-item label="CN">
          <n-input v-model:value="createForm.subject.CN" placeholder="例如 test.example.com" />
        </n-form-item>
        <n-form-item label="O">
          <n-input v-model:value="createForm.subject.O" />
        </n-form-item>
        <n-form-item label="SAN 列表">
          <n-dynamic-input
            v-model:value="createForm.san"
            :on-create="() => ''"
            placeholder="例如 www.example.com"
          />
        </n-form-item>
        <n-form-item label="RSA 位数" v-if="createForm.algorithm === 'RSA'">
          <n-select v-model:value="createForm.key_params.key_size" :options="rsaSizeOptions" />
        </n-form-item>
        <n-form-item label="ECC 曲线" v-if="createForm.algorithm === 'ECC'">
          <n-select v-model:value="createForm.key_params.curve" :options="eccCurveOptions" />
        </n-form-item>
        <n-form-item label="密钥来源">
          <n-select v-model:value="createForm.key_source" :options="keySourceOptions" />
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

    <!-- 导入 CSR -->
    <n-modal
      v-model:show="showImport"
      preset="card"
      title="导入 CSR"
      style="width: 560px"
    >
      <n-form label-placement="left" label-width="110">
        <n-form-item label="CSR 路径">
          <n-input
            v-model:value="importPath"
            placeholder="例如 data/csr/imported.csr"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showImport = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleImport">
            导入
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 签发为终端证书 -->
    <n-modal
      v-model:show="showSign"
      preset="card"
      title="用 CSR 签发终端证书"
      style="width: 640px"
    >
      <n-form :model="signForm" label-placement="left" label-width="110">
        <n-form-item label="CSR ID">
          <n-input v-model:value="signForm.csr_id" disabled />
        </n-form-item>
        <n-form-item label="CA ID">
          <n-input v-model:value="signForm.ca_id" />
        </n-form-item>
        <n-form-item label="CA 密钥引用">
          <n-input v-model:value="signForm.ca_key_ref" />
        </n-form-item>
        <n-form-item label="证书类型">
          <n-select v-model:value="signForm.cert_type" :options="certTypeOptions" />
        </n-form-item>
        <n-form-item label="有效期天数">
          <n-input-number v-model:value="signForm.validity_days" :min="1" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showSign = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSign">
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
  NInputNumber,
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
const showImport = ref(false);
const showSign = ref(false);

const list = ref<any[]>([]);
const importPath = ref('');

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
});

const createForm = reactive({
  algorithm: 'SM2',
  subject: { CN: '', O: '' },
  san: [''],
  key_params: { key_size: 2048, curve: 'prime256v1' },
  key_source: 'generate',
});

const signForm = reactive({
  csr_id: '',
  ca_id: '',
  ca_key_ref: '',
  cert_type: 'server',
  validity_days: 365,
});

const algorithmOptions = [
  { label: 'SM2', value: 'SM2' },
  { label: 'RSA', value: 'RSA' },
  { label: 'ECC', value: 'ECC' },
  { label: 'ML-DSA', value: 'ML-DSA' },
];

const rsaSizeOptions = [
  { label: '2048', value: 2048 },
  { label: '3072', value: 3072 },
  { label: '4096', value: 4096 },
];

const eccCurveOptions = [
  { label: 'P-256 / prime256v1', value: 'prime256v1' },
  { label: 'P-384 / secp384r1', value: 'secp384r1' },
];

const keySourceOptions = [
  { label: '生成新密钥', value: 'generate' },
  { label: '使用已有密钥', value: 'existing' },
];

const certTypeOptions = [
  { label: '服务端 server', value: 'server' },
  { label: '客户端 client', value: 'client' },
];

const columns: DataTableColumns<any> = [
  { title: 'CSR ID', key: 'csr_id' },
  { title: '主题', key: 'subject', ellipsis: { tooltip: true } },
  { title: '算法', key: 'algorithm' },
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
            {
              size: 'small',
              type: 'primary',
              disabled: row.status === 'ISSUED' || row.status === 'REVOKED',
              onClick: () => openSign(row),
            },
            { default: () => '签发' }
          ),
        ],
      });
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCsrs({
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
  if (!createForm.subject.CN) {
    message.warning('请填写 CN');
    return;
  }
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      algorithm: createForm.algorithm,
      subject: { ...createForm.subject },
      san: createForm.san.filter((s) => s && s.trim()),
      key_source: createForm.key_source,
    };
    if (createForm.algorithm === 'RSA') {
      params.key_params = { key_size: createForm.key_params.key_size };
    } else if (createForm.algorithm === 'ECC') {
      params.key_params = { curve: createForm.key_params.curve };
    }

    await run('csr.create', params);
    showCreate.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

async function handleImport() {
  if (!importPath.value) {
    message.warning('请输入 CSR 路径');
    return;
  }
  // 导入走平台后端专用接口，不直接触发 core operation
  submitting.value = true;
  try {
    await cryptoApi.execute('csr.create', {
      key_source: 'existing',
      import_path: importPath.value,
    });
    showImport.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

function openSign(row: any) {
  signForm.csr_id = row.csr_id;
  signForm.ca_id = '';
  signForm.ca_key_ref = '';
  signForm.cert_type = 'server';
  signForm.validity_days = 365;
  showSign.value = true;
}

async function handleSign() {
  if (!signForm.ca_id || !signForm.ca_key_ref) {
    message.warning('请填写 CA ID 与 CA 密钥引用');
    return;
  }
  submitting.value = true;
  try {
    await run('cert.sign', {
      csr_id: signForm.csr_id,
      ca_id: signForm.ca_id,
      ca_key_ref: signForm.ca_key_ref,
      cert_type: signForm.cert_type,
      validity_days: signForm.validity_days,
      algorithm: 'SM2',
    });
    showSign.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

onMounted(fetchList);
</script>
