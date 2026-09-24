<template>
  <div>
    <n-space vertical size="large">
      <n-card title="密钥管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:key:manage'"
              type="primary"
              @click="openGenerate"
            >
              生成密钥
            </n-button>
            <n-button
              v-permission="'crypto_console:key:manage'"
              @click="openImport"
            >
              导入密钥
            </n-button>
            <n-button @click="fetchList">刷新</n-button>
          </n-space>
        </template>

        <n-alert type="info" :show-icon="true" style="margin-bottom: 12px">
          平台只保存密钥元数据与引用（key_ref），不保存私钥明文。导出默认禁止明文，
          需要平台 RBAC 放行并二次确认。
        </n-alert>

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

    <!-- 生成密钥 -->
    <n-modal
      v-model:show="showGenerate"
      preset="card"
      title="生成密钥"
      style="width: 620px"
    >
      <n-form :model="generateForm" label-placement="left" label-width="110">
        <n-form-item label="算法">
          <n-select v-model:value="generateForm.algorithm" :options="algorithmOptions" />
        </n-form-item>
        <n-form-item label="RSA 位数" v-if="generateForm.algorithm === 'RSA'">
          <n-select v-model:value="generateForm.key_params.key_size" :options="rsaSizeOptions" />
        </n-form-item>
        <n-form-item label="ECC 曲线" v-if="generateForm.algorithm === 'ECC'">
          <n-select v-model:value="generateForm.key_params.curve" :options="eccCurveOptions" />
        </n-form-item>
        <n-form-item label="PQC 参数" v-if="isPqc(generateForm.algorithm)">
          <n-input
            v-model:value="generateForm.key_params.parameter"
            :placeholder="pqcPlaceholder(generateForm.algorithm)"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showGenerate = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleGenerate">
            提交
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 导入密钥 -->
    <n-modal
      v-model:show="showImport"
      preset="card"
      title="导入密钥"
      style="width: 560px"
    >
      <n-form label-placement="left" label-width="110">
        <n-form-item label="密钥路径">
          <n-input
            v-model:value="importForm.key_path"
            placeholder="core/data 或 core/tmp 下受控路径"
          />
        </n-form-item>
        <n-form-item label="算法">
          <n-select v-model:value="importForm.algorithm" :options="algorithmOptions" />
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

    <!-- 导出确认 -->
    <n-modal
      v-model:show="showExport"
      preset="card"
      title="导出密钥（明文导出）"
      style="width: 520px"
    >
      <n-alert type="warning" :show-icon="true" style="margin-bottom: 12px">
        明文导出属于高风险操作，会写审计日志。导出后将私钥落到受控路径，
        请确认已完成 RBAC 授权与业务审批。
      </n-alert>
      <n-form label-placement="left" label-width="110">
        <n-form-item label="key_ref">
          <n-input v-model:value="exportForm.key_ref" disabled />
        </n-form-item>
        <n-form-item label="导出路径">
          <n-input
            v-model:value="exportForm.export_path"
            placeholder="留空则使用默认导出路径"
          />
        </n-form-item>
        <n-form-item label="确认明文导出">
          <n-checkbox v-model:checked="exportForm.confirm">
            我已确认完成 RBAC 授权与业务审批
          </n-checkbox>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showExport = false">取消</n-button>
          <n-button
            type="error"
            :disabled="!exportForm.confirm"
            :loading="submitting"
            @click="handleExport"
          >
            确认导出
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import {
  NAlert,
  NButton,
  NCard,
  NCheckbox,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSelect,
  NSpace,
  useDialog,
  useMessage,
  type DataTableColumns,
} from 'naive-ui';
import { cryptoApi } from '@/api/crypto';
import { useOperation } from '@/composables/useOperation';

const message = useMessage();
const dialog = useDialog();
const { run } = useOperation();

const loading = ref(false);
const submitting = ref(false);
const showGenerate = ref(false);
const showImport = ref(false);
const showExport = ref(false);

const list = ref<any[]>([]);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
});

const generateForm = reactive({
  algorithm: 'SM2',
  key_params: {
    key_size: 2048,
    curve: 'prime256v1',
    parameter: '',
  },
});

const importForm = reactive({
  key_path: '',
  algorithm: 'SM2',
});

const exportForm = reactive({
  key_ref: '',
  export_path: '',
  confirm: false,
});

const algorithmOptions = [
  { label: 'SM2', value: 'SM2' },
  { label: 'RSA', value: 'RSA' },
  { label: 'ECC', value: 'ECC' },
  { label: 'ML-KEM', value: 'ML-KEM' },
  { label: 'ML-DSA', value: 'ML-DSA' },
  { label: 'SLH-DSA', value: 'SLH-DSA' },
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

function isPqc(alg: string) {
  return alg === 'ML-KEM' || alg === 'ML-DSA' || alg === 'SLH-DSA';
}

function pqcPlaceholder(alg: string) {
  if (alg === 'ML-KEM') return 'ML-KEM-512 / ML-KEM-768 / ML-KEM-1024';
  if (alg === 'ML-DSA') return 'ML-DSA-44 / ML-DSA-65 / ML-DSA-87';
  if (alg === 'SLH-DSA') return 'SLH-DSA-SHA2-128s 等，按铜锁实测';
  return '';
}

const columns: DataTableColumns<any> = [
  { title: 'key_ref', key: 'key_ref', ellipsis: { tooltip: true } },
  { title: '算法', key: 'algorithm' },
  { title: '用途', key: 'usage' },
  { title: '状态', key: 'state' },
  { title: '创建时间', key: 'created_at' },
  {
    title: '操作',
    key: 'actions',
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            { size: 'small', onClick: () => openExport(row) },
            { default: () => '导出' }
          ),
          h(
            NButton,
            { size: 'small', type: 'error', onClick: () => confirmDelete(row) },
            { default: () => '删除' }
          ),
        ],
      });
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listKeys({
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

function openGenerate() {
  generateForm.algorithm = 'SM2';
  generateForm.key_params.key_size = 2048;
  generateForm.key_params.curve = 'prime256v1';
  generateForm.key_params.parameter = '';
  showGenerate.value = true;
}

async function handleGenerate() {
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      action: 'generate',
      algorithm: generateForm.algorithm,
    };
    if (generateForm.algorithm === 'RSA') {
      params.key_params = { key_size: generateForm.key_params.key_size };
    } else if (generateForm.algorithm === 'ECC') {
      params.key_params = { curve: generateForm.key_params.curve };
    } else if (isPqc(generateForm.algorithm)) {
      params.key_params = { parameter: generateForm.key_params.parameter };
    }

    await run('key.manage', params);
    showGenerate.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

function openImport() {
  importForm.key_path = '';
  importForm.algorithm = 'SM2';
  showImport.value = true;
}

async function handleImport() {
  if (!importForm.key_path) {
    message.warning('请输入密钥路径');
    return;
  }
  submitting.value = true;
  try {
    await run('key.manage', {
      action: 'import',
      algorithm: importForm.algorithm,
      key_path: importForm.key_path,
    });
    showImport.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

function openExport(row: any) {
  exportForm.key_ref = row.key_ref;
  exportForm.export_path = '';
  exportForm.confirm = false;
  showExport.value = true;
}

async function handleExport() {
  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      action: 'export',
      key_ref: exportForm.key_ref,
      allow_plain_export: true,
    };
    if (exportForm.export_path) params.export_path = exportForm.export_path;

    await run('key.manage', params);
    showExport.value = false;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

function confirmDelete(row: any) {
  dialog.warning({
    title: '删除密钥',
    content: `确认删除 key_ref=${row.key_ref}？该操作会写审计日志。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await run('key.manage', {
          action: 'delete',
          key_ref: row.key_ref,
        });
        await fetchList();
      } catch {
        // run 内部已提示
      }
    },
  });
}

onMounted(fetchList);
</script>
