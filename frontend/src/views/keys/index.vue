<template>
  <t-card :title="'密钥管理'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button v-permission="'crypto_console:key:manage'" theme="primary" @click="openGenerate">生成密钥</t-button>
        <t-button v-permission="'crypto_console:key:manage'" @click="openImport">导入密钥</t-button>
        <t-button @click="fetchList">刷新</t-button>
      </t-space>
    </template>

    <t-alert theme="info" :message="'平台只保存密钥元数据与引用（key_ref），不保存私钥明文。导出默认禁止明文，需要平台 RBAC 放行并二次确认。'" style="margin-bottom: 12px" />

    <t-table
      row-key="key_ref"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      @page-change="handlePageChange"
    />
  </t-card>

  <t-dialog :visible="showGenerate" :header="'生成密钥'" :confirm-btn="null" :cancel-btn="null" :width="620" @close="showGenerate = false">
    <t-form :data="generateForm" label-align="left" :label-width="110">
      <t-form-item label="算法">
        <t-select v-model="generateForm.algorithm" :options="algorithmOptions" />
      </t-form-item>
      <t-form-item label="RSA 位数" v-if="generateForm.algorithm === 'RSA'">
        <t-select v-model="generateForm.key_params.key_size" :options="rsaSizeOptions" />
      </t-form-item>
      <t-form-item label="ECC 曲线" v-if="generateForm.algorithm === 'ECC'">
        <t-select v-model="generateForm.key_params.curve" :options="eccCurveOptions" />
      </t-form-item>
      <t-form-item label="PQC 参数" v-if="isPqc(generateForm.algorithm)">
        <t-input v-model="generateForm.key_params.parameter" :placeholder="pqcPlaceholder(generateForm.algorithm)" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showGenerate = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" @click="handleGenerate">提交</t-button>
      </t-space>
    </template>
  </t-dialog>

  <t-dialog :visible="showImport" :header="'导入密钥'" :confirm-btn="null" :cancel-btn="null" :width="560" @close="showImport = false">
    <t-form label-align="left" :label-width="110">
      <t-form-item label="密钥路径">
        <t-input v-model="importForm.key_path" placeholder="core/data 或 core/tmp 下受控路径" />
      </t-form-item>
      <t-form-item label="算法">
        <t-select v-model="importForm.algorithm" :options="algorithmOptions" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showImport = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" @click="handleImport">导入</t-button>
      </t-space>
    </template>
  </t-dialog>

  <t-dialog :visible="showExport" :header="'导出密钥（明文导出）'" :confirm-btn="null" :cancel-btn="null" :width="520" @close="showExport = false">
    <t-alert theme="warning" :message="'明文导出属于高风险操作，会写审计日志。导出后将私钥落到受控路径，请确认已完成 RBAC 授权与业务审批。'" style="margin-bottom: 12px" />
    <t-form label-align="left" :label-width="110">
      <t-form-item label="key_ref">
        <t-input v-model="exportForm.key_ref" disabled />
      </t-form-item>
      <t-form-item label="导出路径">
        <t-input v-model="exportForm.export_path" placeholder="留空则使用默认导出路径" />
      </t-form-item>
      <t-form-item label="确认明文导出">
        <t-checkbox v-model="exportForm.confirm">我已确认完成 RBAC 授权与业务审批</t-checkbox>
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showExport = false">取消</t-button>
        <t-button theme="danger" :disabled="!exportForm.confirm" :loading="submitting" @click="handleExport">确认导出</t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import { Button, Space, DialogPlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { cryptoApi } from '@/api/crypto';
import { useOperation } from '@/composables/useOperation';

const { run } = useOperation();

const loading = ref(false);
const submitting = ref(false);
const showGenerate = ref(false);
const showImport = ref(false);
const showExport = ref(false);

const list = ref<any[]>([]);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
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

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'key_ref', title: 'key_ref', ellipsis: true },
  { colKey: 'algorithm', title: '算法' },
  { colKey: 'usage', title: '用途' },
  { colKey: 'state', title: '状态' },
  { colKey: 'created_at', title: '创建时间' },
  {
    colKey: 'actions',
    title: '操作',
    cell: (_h, { row }) =>
      h(Space, { size: 4 }, () => [
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openExport(row) }, () => '导出'),
        h(Button, { size: 'small', variant: 'text', theme: 'danger', onClick: () => confirmDelete(row) }, () => '删除'),
      ]),
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listKeys({ page: pagination.current, page_size: pagination.pageSize });
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
    MessagePlugin.warning('请输入密钥路径');
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
  const dialog = DialogPlugin.confirm({
    header: '删除密钥',
    body: `确认删除 key_ref=${row.key_ref}？该操作会写审计日志。`,
    confirmBtn: '确认删除',
    cancelBtn: '取消',
    theme: 'danger',
    onConfirm: async () => {
      try {
        await run('key.manage', {
          action: 'delete',
          key_ref: row.key_ref,
        });
        await fetchList();
      } catch {
        // run 内部已提示
      }
      dialog.destroy();
    },
    onClose: () => dialog.destroy(),
  });
}

onMounted(fetchList);
</script>