<template>
  <div>
    <n-space vertical size="large">
      <n-card title="证书管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:cert:sign'"
              type="primary"
              @click="showSign = true"
            >
              签发证书
            </n-button>
            <n-button @click="showParse = true">解析证书</n-button>
            <n-button @click="showConvert = true">格式转换</n-button>
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

    <!-- 签发终端证书 -->
    <n-modal
      v-model:show="showSign"
      preset="card"
      title="签发终端证书"
      style="width: 680px"
    >
      <n-form :model="signForm" label-placement="left" label-width="110">
        <n-form-item label="CSR ID">
          <n-input v-model:value="signForm.csr_id" placeholder="请输入 CSR ID" />
        </n-form-item>
        <n-form-item label="CA ID">
          <n-input v-model:value="signForm.ca_id" placeholder="请输入 CA ID" />
        </n-form-item>
        <n-form-item label="CA 密钥引用">
          <n-input
            v-model:value="signForm.ca_key_ref"
            placeholder="例如 key-20260101-xxx"
          />
        </n-form-item>
        <n-form-item label="证书类型">
          <n-select v-model:value="signForm.cert_type" :options="certTypeOptions" />
        </n-form-item>
        <n-form-item label="有效期天数">
          <n-input-number v-model:value="signForm.validity_days" :min="1" />
        </n-form-item>
        <n-form-item label="算法">
          <n-select v-model:value="signForm.algorithm" :options="algorithmOptions" />
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

    <!-- 解析证书 -->
    <n-modal
      v-model:show="showParse"
      preset="card"
      title="解析证书"
      style="width: 720px"
    >
      <n-form label-placement="left" label-width="110">
        <n-form-item label="证书路径">
          <n-input
            v-model:value="parseForm.cert_path"
            placeholder="例如 data/certs/cert-xxx.pem"
          />
        </n-form-item>
      </n-form>
      <n-divider v-if="parseResult" />
      <n-descriptions
        v-if="parseResult"
        bordered
        :column="1"
        label-placement="left"
        size="small"
      >
        <n-descriptions-item label="主题">
          {{ parseResult.subject }}
        </n-descriptions-item>
        <n-descriptions-item label="颁发者">
          {{ parseResult.issuer }}
        </n-descriptions-item>
        <n-descriptions-item label="序列号">
          {{ parseResult.serial }}
        </n-descriptions-item>
        <n-descriptions-item label="生效时间">
          {{ parseResult.not_before }}
        </n-descriptions-item>
        <n-descriptions-item label="失效时间">
          {{ parseResult.not_after }}
        </n-descriptions-item>
        <n-descriptions-item label="SHA-256 指纹">
          {{ parseResult.fingerprint_sha256 }}
        </n-descriptions-item>
        <n-descriptions-item label="公钥算法">
          {{ parseResult.public_key_algorithm }}
        </n-descriptions-item>
        <n-descriptions-item label="SAN">
          {{ parseResult.san || '—' }}
        </n-descriptions-item>
        <n-descriptions-item label="KeyUsage">
          {{ parseResult.key_usage || '—' }}
        </n-descriptions-item>
        <n-descriptions-item label="EKU">
          {{ parseResult.eku || '—' }}
        </n-descriptions-item>
      </n-descriptions>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showParse = false">关闭</n-button>
          <n-button type="primary" :loading="submitting" @click="handleParse">
            解析
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 格式转换 -->
    <n-modal
      v-model:show="showConvert"
      preset="card"
      title="证书格式转换"
      style="width: 640px"
    >
      <n-form :model="convertForm" label-placement="left" label-width="110">
        <n-form-item label="源格式">
          <n-select v-model:value="convertForm.source_format" :options="formatOptions" />
        </n-form-item>
        <n-form-item label="目标格式">
          <n-select v-model:value="convertForm.target_format" :options="formatOptions" />
        </n-form-item>
        <n-form-item label="源路径">
          <n-input v-model:value="convertForm.source_path" />
        </n-form-item>
        <n-form-item label="目标路径">
          <n-input
            v-model:value="convertForm.target_path"
            placeholder="留空则自动生成"
          />
        </n-form-item>
        <n-form-item
          v-if="
            convertForm.source_format === 'PKCS12' ||
            convertForm.target_format === 'PKCS12'
          "
          label="口令文件"
        >
          <n-input
            v-model:value="convertForm.password_file"
            placeholder="core/tmp 下 0600 权限文件"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showConvert = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleConvert">
            提交
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 详情抽屉 -->
    <n-drawer v-model:show="showDetail" :width="560">
      <n-drawer-content title="证书详情">
        <n-descriptions
          v-if="detail"
          bordered
          :column="1"
          label-placement="left"
          size="small"
        >
          <n-descriptions-item label="证书 ID">
            {{ detail.cert_id }}
          </n-descriptions-item>
          <n-descriptions-item label="序列号">
            {{ detail.serial }}
          </n-descriptions-item>
          <n-descriptions-item label="主题">
            {{ detail.subject }}
          </n-descriptions-item>
          <n-descriptions-item label="颁发者">
            {{ detail.issuer }}
          </n-descriptions-item>
          <n-descriptions-item label="算法">
            {{ detail.algorithm }}
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            {{ detail.status }}
          </n-descriptions-item>
          <n-descriptions-item label="创建时间">
            {{ detail.created_at }}
          </n-descriptions-item>
        </n-descriptions>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import {
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NDrawer,
  NDrawerContent,
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
const showSign = ref(false);
const showParse = ref(false);
const showConvert = ref(false);
const showDetail = ref(false);

const list = ref<any[]>([]);
const detail = ref<any>(null);
const parseResult = ref<any>(null);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
});

const signForm = reactive({
  csr_id: '',
  ca_id: '',
  ca_key_ref: '',
  cert_type: 'server',
  validity_days: 365,
  algorithm: 'SM2',
});

const parseForm = reactive({
  cert_path: '',
});

const convertForm = reactive({
  source_format: 'PEM',
  target_format: 'DER',
  source_path: '',
  target_path: '',
  password_file: '',
});

const certTypeOptions = [
  { label: '服务端 server', value: 'server' },
  { label: '客户端 client', value: 'client' },
];

const algorithmOptions = [
  { label: 'SM2', value: 'SM2' },
  { label: 'RSA', value: 'RSA' },
  { label: 'ECC', value: 'ECC' },
  { label: 'ML-DSA', value: 'ML-DSA' },
];

const formatOptions = [
  { label: 'PEM', value: 'PEM' },
  { label: 'DER', value: 'DER' },
  { label: 'PKCS12', value: 'PKCS12' },
];

const columns: DataTableColumns<any> = [
  { title: '证书 ID', key: 'cert_id' },
  { title: '序列号', key: 'serial', ellipsis: { tooltip: true } },
  { title: '主题', key: 'subject', ellipsis: { tooltip: true } },
  { title: '类型', key: 'cert_type' },
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
            { size: 'small', onClick: () => handleView(row) },
            { default: () => '详情' }
          ),
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
    const res = await cryptoApi.listCerts({
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

async function handleSign() {
  if (!signForm.csr_id || !signForm.ca_id || !signForm.ca_key_ref) {
    message.warning('请填写 CSR ID、CA ID 与 CA 密钥引用');
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
      algorithm: signForm.algorithm,
    });
    showSign.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

async function handleParse() {
  if (!parseForm.cert_path) {
    message.warning('请输入证书路径');
    return;
  }
  submitting.value = true;
  try {
    const res: any = await run('cert.parse', {
      cert_path: parseForm.cert_path,
    });
    parseResult.value = res || null;
  } finally {
    submitting.value = false;
  }
}

async function handleConvert() {
  if (!convertForm.source_path) {
    message.warning('请输入源路径');
    return;
  }
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      source_format: convertForm.source_format,
      target_format: convertForm.target_format,
      source_path: convertForm.source_path,
    };
    if (convertForm.target_path) params.target_path = convertForm.target_path;
    if (convertForm.password_file) params.password_file = convertForm.password_file;

    await run('cert.convert', params);
    showConvert.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

async function handleView(row: any) {
  try {
    const res = await cryptoApi.getCert(row.cert_id);
    detail.value = res.data;
    showDetail.value = true;
  } catch (e: any) {
    message.error(e.message || '加载详情失败');
  }
}

function handleDownload(row: any) {
  // 下载统一走后端静态资源接口，不在前端拼接 core 路径
  const url = `${import.meta.env.VITE_API_BASE_URL || ''}/api/v1/certs/${row.cert_id}/download`;
  const token = localStorage.getItem('token');
  const a = document.createElement('a');
  a.href = url;
  a.target = '_blank';
  if (token) {
    // 若后端需要 Bearer，可由后端提供一次性下载 token
    a.setAttribute('data-token', token);
  }
  a.click();
}

onMounted(fetchList);
</script>
