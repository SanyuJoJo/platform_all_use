<template>
  <t-card :title="'CA 管理'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button v-permission="'crypto_console:ca:import'" @click="openImport">导入 CA</t-button>
        <t-button v-permission="'crypto_console:ca:create'" theme="primary" @click="openCreate">创建根 CA</t-button>
        <t-button @click="openDomainManager">域管理</t-button>
      </t-space>
    </template>

    <t-table
      row-key="ca_id"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="false"
      :scroll="{ x: 1500 }"
    />
  </t-card>

  <!-- 创建根 CA -->
  <t-dialog :visible="showCreate" :header="'创建根 CA'" :confirm-btn="null" :cancel-btn="null" :width="720" @close="showCreate = false">
    <t-form :data="form" label-align="left" :label-width="120">
      <t-form-item label="域">
        <t-select v-model="form.domain" :options="domainOptions" :loading="domainLoading" placeholder="选择域" style="width: 100%" />
      </t-form-item>
      <t-form-item label="公钥算法">
        <t-select v-model="form.algorithm" :options="algStore.options" :loading="algStore.loading" placeholder="选择算法" />
      </t-form-item>
      <t-form-item label="使用者 DN">
        <div class="dn-row">
          <t-textarea v-model="form.dn" :autosize="{ minRows: 3, maxRows: 6 }" :placeholder="DN_HELP_EXAMPLE" />
          <t-tooltip placement="left" :content="DN_HELP_KEYS">
            <span class="dn-help-btn">?</span>
          </t-tooltip>
        </div>
      </t-form-item>
      <t-form-item label="DN 分隔符">
        <t-input v-model="form.dnSeparator" placeholder="默认逗号 “,”，可输入 “;” 或 “\\n” 表示换行" style="max-width: 320px" />
      </t-form-item>
      <t-form-item v-if="form.dn.trim()" label="解析预览">
        <div class="dn-preview">
          <t-alert v-if="dnPreview.errors.length" theme="error" :message="dnPreview.errors.join('；')" style="width: 100%" />
          <template v-else>
            <t-tag v-for="(v, k) in dnPreview.subject" :key="k" size="small" theme="success" variant="light" style="margin: 2px 6px 2px 0">
              {{ k }} = {{ v }}
            </t-tag>
          </template>
        </div>
      </t-form-item>
      <t-form-item label="有效期天数">
        <t-input-number v-model="form.validity_days" :min="1" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showCreate = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="dnPreview.errors.length > 0 || !form.dn.trim() || !form.domain || !form.algorithm" @click="handleCreate">提交</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 导入 CA -->
  <t-dialog :visible="showImport" :header="'导入 CA'" :confirm-btn="null" :cancel-btn="null" :width="680" @close="showImport = false">
    <t-alert theme="info" :message="'支持两种导入方式：① 仅导入 CA 证书（不选私钥文件）；② 导入带私钥的 CA（选择证书 + 私钥文件，加密私钥需填写密码）。'" style="margin-bottom: 12px" />
    <t-form label-align="top">
      <t-form-item label="域">
        <t-select v-model="importForm.domain" :options="domainOptions" :loading="domainLoading" placeholder="选择域（默认 safe）" style="width: 100%" />
      </t-form-item>
      <t-form-item label="CA 证书文件（PEM / CRT / CER，必填）">
        <div class="file-row">
          <input ref="certFileInputRef" type="file" accept=".pem,.crt,.cer,.cert" style="display: none" @change="onCertFileChange" />
          <t-button @click="certFileInputRef?.click()">选择证书文件</t-button>
          <span v-if="importForm.cert_file_name" class="file-name">{{ importForm.cert_file_name }}</span>
          <span v-else class="file-placeholder">未选择文件</span>
        </div>
      </t-form-item>
      <t-form-item label="CA 私钥文件（PEM，可选）">
        <div class="file-row">
          <input ref="keyFileInputRef" type="file" accept=".pem,.key" style="display: none" @change="onKeyFileChange" />
          <t-button @click="keyFileInputRef?.click()">选择私钥文件</t-button>
          <span v-if="importForm.key_file_name" class="file-name">{{ importForm.key_file_name }}</span>
          <span v-else class="file-placeholder">未选择文件</span>
          <t-button v-if="importForm.key_file_name" size="small" variant="text" @click="clearKeyFile">清除</t-button>
        </div>
      </t-form-item>
      <t-form-item label="私钥密码（仅加密私钥需要填写）">
        <t-input v-model="importForm.key_password" type="password" placeholder="若私钥未加密，留空即可" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showImport = false">取消</t-button>
        <t-button theme="primary" :loading="importing" :disabled="!importForm.cert_pem.trim() || !importForm.domain" @click="handleImport">导入</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 域管理 -->
  <t-dialog :visible="showDomainManager" :header="'域管理'" :confirm-btn="null" :cancel-btn="null" :width="520" @close="showDomainManager = false">
    <t-space direction="vertical" style="width: 100%">
      <t-alert theme="info" :message="'域用于隔离不同业务线的 CA 与证书。safe 为默认域，不可删除。'" />
      <div class="domain-list">
        <div v-for="d in domains" :key="d" class="domain-item">
          <span class="mono">{{ d }}</span>
          <t-space align="center">
            <t-tag v-if="d === 'safe'" size="small" theme="success" variant="light">默认</t-tag>
            <t-popconfirm v-else :content="`确认删除域「${d}」？该操作不可恢复，且要求域目录为空。`" confirm-btn="确认删除" cancel-btn="取消" @confirm="handleDeleteDomain(d)">
              <t-button size="small" variant="text" theme="danger">删除</t-button>
            </t-popconfirm>
          </t-space>
        </div>
      </div>
      <t-divider style="margin: 4px 0" />
      <t-form-item label="新建域">
        <t-input v-model="newDomainName" placeholder="仅支持字母、数字、下划线、短横线；长度 1-64" />
      </t-form-item>
      <t-alert v-if="newDomainError" theme="error" :message="newDomainError" />
    </t-space>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showDomainManager = false">关闭</t-button>
        <t-button theme="primary" :loading="creatingDomain" :disabled="!newDomainValid" @click="handleCreateDomain">创建</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 导出 CA -->
  <t-dialog :visible="showExport" :header="'导出 CA'" :confirm-btn="null" :cancel-btn="null" :width="560" @close="showExport = false">
    <t-alert v-if="exportForm.type === 'key' || exportForm.type === 'pkcs12'" theme="warning" :message="'导出内容包含私钥，属于敏感操作，会写审计日志。请确保已完成 RBAC 授权与业务审批，并在安全环境中妥善保管导出文件。'" style="margin-bottom: 12px" />
    <t-form :data="exportForm" label-align="left" :label-width="100">
      <t-form-item label="导出内容">
        <t-radio-group v-model="exportForm.type">
          <t-radio value="cert">CA 证书（PEM）</t-radio>
          <t-radio value="key">CA 私钥（PEM）</t-radio>
          <t-radio value="pkcs12">带私钥的 CA（PKCS#12）</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item v-if="exportForm.type === 'pkcs12'" label="PKCS#12 密码">
        <t-input v-model="exportForm.password" type="password" placeholder="至少 6 位，导出后用于打开 .p12" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showExport = false">取消</t-button>
        <t-button theme="primary" :loading="exporting" :disabled="exportForm.type === 'pkcs12' && (!exportForm.password || exportForm.password.length < 6)" @click="handleExport">导出</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- CA 证书详情 -->
  <t-drawer :visible="showDetail" :size="760" placement="right" :footer="false" @close="showDetail = false">
    <template #header>CA 证书详情</template>
    <t-loading :loading="detailLoading">
      <t-descriptions v-if="detail" :column="1" label-align="left" bordered size="small">
        <t-descriptions-item label="版本号">{{ detail.version }}</t-descriptions-item>
        <t-descriptions-item label="序列号"><span class="mono">{{ detail.serial }}</span></t-descriptions-item>
        <t-descriptions-item label="颁发者">{{ detail.issuer }}</t-descriptions-item>
        <t-descriptions-item label="使用者">{{ detail.subject }}</t-descriptions-item>
        <t-descriptions-item label="摘要值">
          <t-tooltip placement="left" :content="detail.fingerprint">
            <span class="truncated mono">{{ truncate(detail.fingerprint, 64) }}</span>
          </t-tooltip>
        </t-descriptions-item>
        <t-descriptions-item label="有效期（起）">{{ detail.not_before }}</t-descriptions-item>
        <t-descriptions-item label="有效期（止）">{{ detail.not_after }}</t-descriptions-item>
        <t-descriptions-item label="公钥算法">{{ detailAlgLabel }}</t-descriptions-item>
        <t-descriptions-item label="签名摘要算法">{{ detail.signature_algorithm }}</t-descriptions-item>
        <t-descriptions-item label="签名摘要值">
          <t-tooltip placement="left" :content="detail.signature_value">
            <span class="truncated mono">{{ truncate(detail.signature_value, 64) }}</span>
          </t-tooltip>
        </t-descriptions-item>
        <t-descriptions-item label="公钥值">
          <t-tooltip placement="left" :content="detail.public_key_value">
            <span class="truncated mono">{{ truncate(detail.public_key_value, 64) }}</span>
          </t-tooltip>
        </t-descriptions-item>
      </t-descriptions>
      <t-empty v-else-if="!detailLoading" description="暂无详情" />
    </t-loading>
  </t-drawer>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed, h } from 'vue';
import { Tag, Space, Button, Popconfirm } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { cryptoApi, type CAExportType, type CADetail } from '@/api/crypto';
import { useOperation } from '@/composables/useOperation';
import { DEFAULT_DN_SEPARATOR, DN_HELP_KEYS, DN_HELP_EXAMPLE, validateDN, parseDN } from '@/utils/dn';
import { useAlgorithmStore } from '@/store/algorithm';
import { publicKeyLabel } from '@/utils/crypto-algorithms';

const { run } = useOperation();
const algStore = useAlgorithmStore();

const loading = ref(false);
const submitting = ref(false);
const importing = ref(false);
const exporting = ref(false);
const detailLoading = ref(false);

const showCreate = ref(false);
const showImport = ref(false);
const showExport = ref(false);
const showDetail = ref(false);
const showDomainManager = ref(false);

const list = ref<any[]>([]);
const detail = ref<CADetail | null>(null);

const domains = ref<string[]>(['safe']);
const domainLoading = ref(false);
const newDomainName = ref('');
const creatingDomain = ref(false);

const certFileInputRef = ref<HTMLInputElement | null>(null);
const keyFileInputRef = ref<HTMLInputElement | null>(null);

const form = reactive({
  domain: 'safe',
  algorithm: 'SM2',
  dn: '',
  dnSeparator: DEFAULT_DN_SEPARATOR,
  validity_days: 3650,
});

const importForm = reactive({
  domain: 'safe',
  cert_pem: '',
  cert_file_name: '',
  key_pem: '',
  key_file_name: '',
  key_password: '',
});

const exportForm = reactive({
  ca_id: '',
  type: 'cert' as CAExportType,
  password: '',
});

const dnPreview = computed(() => parseDN(form.dn, form.dnSeparator));

const domainOptions = computed(() => domains.value.map((d) => ({ label: d, value: d })));

const detailAlgLabel = computed(() => {
  if (!detail.value) return '';
  const alg = (detail.value as any).algorithm;
  if (alg) return algStore.label(alg);
  return publicKeyLabel(detail.value.public_key_algorithm);
});

const newDomainError = computed(() => {
  const v = newDomainName.value.trim();
  if (!v) return '';
  if (v.length > 64) return '长度不能超过 64';
  if (!/^[A-Za-z0-9_-]+$/.test(v)) return '仅支持字母、数字、下划线、短横线';
  if (domains.value.includes(v)) return '该域已存在';
  return '';
});

const newDomainValid = computed(() => newDomainName.value.trim() !== '' && newDomainError.value === '');

function fmtTime(t: any): string {
  if (!t) return '-';
  if (typeof t !== 'string') {
    const d = new Date(t as any);
    if (isNaN(d.getTime())) return '-';
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  }
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(t)) return t;
  const d = new Date(t);
  if (isNaN(d.getTime())) return t;
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'subject_cn', title: '使用者', width: 180, ellipsis: true },
  {
    colKey: 'domain',
    title: '信任域',
    width: 100,
    cell: (_h, { row }) => {
      const d = row.domain || '-';
      if (d === '-') return d;
      return h(Tag, { size: 'small', theme: 'primary', variant: 'light' }, () => d);
    },
  },
  {
    colKey: 'public_key_algorithm',
    title: '公钥算法',
    width: 130,
    cell: (_h, { row }) => {
      if (row.algorithm) {
        return algStore.label(row.algorithm);
      }
      return publicKeyLabel(row.public_key_algorithm);
    },
  },
  { colKey: 'serial', title: '序列号', width: 160, ellipsis: true },
  { colKey: 'not_before', title: '生效时间', width: 170, cell: (_h, { row }) => fmtTime(row.not_before) },
  { colKey: 'not_after', title: '失效时间', width: 170, cell: (_h, { row }) => fmtTime(row.not_after) },
  { colKey: 'status', title: '状态', width: 90 },
  { colKey: 'created_at', title: '创建时间', width: 170 },
  {
    colKey: 'has_private_key',
    title: '私钥文件',
    width: 100,
    cell: (_h, { row }) =>
      h(Tag, { theme: row.has_private_key ? 'success' : 'default', variant: 'light' }, () => (row.has_private_key ? '有' : '无')),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 220,
    fixed: 'right',
    cell: (_h, { row }) =>
      h(Space, { size: 4 }, () => [
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleDetail(row) }, () => '详情'),
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openExport(row) }, () => '导出'),
        h(
          Popconfirm,
          { content: `确认删除 CA「${row.subject_cn || row.ca_id}」？该操作不可恢复，且会写审计日志。`, confirmBtn: '确认删除', cancelBtn: '取消', onConfirm: () => handleDelete(row) },
          () => h(Button, { size: 'small', variant: 'text', theme: 'danger' }, () => '删除'),
        ),
      ]),
  },
];

function truncate(s: string | undefined | null, n: number): string {
  if (!s) return '';
  return s.length > n ? s.slice(0, n) + '...' : s;
}

async function fetchDomains() {
  domainLoading.value = true;
  try {
    const res: any = await cryptoApi.listDomains();
    const arr = (res.data?.domains as string[]) || [];
    domains.value = arr.length ? arr : ['safe'];
  } catch (e: any) {
    console.warn('[ca] fetch domains failed', e);
    domains.value = ['safe'];
  } finally {
    domainLoading.value = false;
  }
}

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCas();
    list.value = (res.data as any)?.items || res.data || [];
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  form.domain = 'safe';
  form.algorithm = 'SM2';
  form.dn = '';
  form.dnSeparator = DEFAULT_DN_SEPARATOR;
  form.validity_days = 3650;
  if (domains.value.length === 0) fetchDomains();
  showCreate.value = true;
}

async function handleCreate() {
  const result = validateDN(form.dn, form.dnSeparator);
  if (result.errors.length) {
    MessagePlugin.error(result.errors[0]);
    return;
  }
  submitting.value = true;
  try {
    await run('ca.create', {
      domain: form.domain,
      algorithm: form.algorithm,
      subject: result.subject,
      validity_days: form.validity_days,
      key_protection: { cipher: 'AES-256-GCM' },
    });
    showCreate.value = false;
    form.dn = '';
    form.dnSeparator = DEFAULT_DN_SEPARATOR;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

function openImport() {
  importForm.domain = 'safe';
  importForm.cert_pem = '';
  importForm.cert_file_name = '';
  importForm.key_pem = '';
  importForm.key_file_name = '';
  importForm.key_password = '';
  if (certFileInputRef.value) certFileInputRef.value.value = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
  if (domains.value.length === 0) fetchDomains();
  showImport.value = true;
}

async function onCertFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    importForm.cert_pem = '';
    importForm.cert_file_name = '';
    return;
  }
  try {
    importForm.cert_pem = await file.text();
    importForm.cert_file_name = file.name;
  } catch (err: any) {
    MessagePlugin.error('读取证书文件失败: ' + (err?.message || ''));
    importForm.cert_pem = '';
    importForm.cert_file_name = '';
  }
}

async function onKeyFileChange(e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    importForm.key_pem = '';
    importForm.key_file_name = '';
    return;
  }
  try {
    importForm.key_pem = await file.text();
    importForm.key_file_name = file.name;
  } catch (err: any) {
    MessagePlugin.error('读取私钥文件失败: ' + (err?.message || ''));
    importForm.key_pem = '';
    importForm.key_file_name = '';
  }
}

function clearKeyFile() {
  importForm.key_pem = '';
  importForm.key_file_name = '';
  importForm.key_password = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
}

async function handleImport() {
  if (!importForm.cert_pem.trim()) {
    MessagePlugin.warning('请选择 CA 证书文件');
    return;
  }
  if (!importForm.cert_pem.includes('-----BEGIN CERTIFICATE-----')) {
    MessagePlugin.warning('证书文件不是有效的 PEM 格式');
    return;
  }
  if (importForm.key_pem && !importForm.key_pem.includes('-----BEGIN')) {
    MessagePlugin.warning('私钥文件不是有效的 PEM 格式');
    return;
  }
  importing.value = true;
  try {
    const payload: Record<string, string> = {
      domain: importForm.domain,
      cert_pem: importForm.cert_pem.trim(),
    };
    if (importForm.key_pem.trim()) {
      payload.key_pem = importForm.key_pem.trim();
      if (importForm.key_password) {
        payload.key_password = importForm.key_password;
      }
    }
    await cryptoApi.importCA(payload as any);
    MessagePlugin.success('导入成功');
    showImport.value = false;
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '导入失败');
  } finally {
    importing.value = false;
  }
}

function openDomainManager() {
  newDomainName.value = '';
  fetchDomains();
  showDomainManager.value = true;
}

async function handleCreateDomain() {
  if (!newDomainValid.value) return;
  creatingDomain.value = true;
  try {
    await cryptoApi.createDomain(newDomainName.value.trim());
    MessagePlugin.success('域创建成功');
    newDomainName.value = '';
    await fetchDomains();
  } catch (e: any) {
    MessagePlugin.error(e.message || '创建域失败');
  } finally {
    creatingDomain.value = false;
  }
}

async function handleDeleteDomain(name: string) {
  if (!name || name === 'safe') {
    MessagePlugin.warning('默认域 safe 不允许删除');
    return;
  }
  try {
    await cryptoApi.deleteDomain(name);
    MessagePlugin.success('删除成功');
    await fetchDomains();
  } catch (e: any) {
    MessagePlugin.error(e.message || '删除域失败');
  }
}

function openExport(row: any) {
  exportForm.ca_id = row.ca_id;
  exportForm.type = 'cert';
  exportForm.password = '';
  showExport.value = true;
}

async function handleExport() {
  if (!exportForm.ca_id) {
    MessagePlugin.warning('请选择要导出的 CA');
    return;
  }
  if (exportForm.type === 'pkcs12' && (!exportForm.password || exportForm.password.length < 6)) {
    MessagePlugin.warning('PKCS#12 密码至少 6 位');
    return;
  }
  exporting.value = true;
  try {
    const payload: { type: CAExportType; password?: string } = {
      type: exportForm.type,
    };
    if (exportForm.type === 'pkcs12') {
      payload.password = exportForm.password;
    }
    const res: any = await cryptoApi.exportCA(exportForm.ca_id, payload);
    let filename = '';
    const cd = res?.headers?.['content-disposition'];
    if (cd) {
      const m = /filename="?([^"]+)"?/.exec(cd);
      if (m) filename = m[1];
    }
    if (!filename) {
      const ext = exportForm.type === 'cert' ? 'pem' : exportForm.type === 'key' ? 'key.pem' : 'p12';
      filename = `${exportForm.ca_id}.${ext}`;
    }
    const blob: Blob = res?.data instanceof Blob ? res.data : new Blob([res.data]);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    MessagePlugin.success('导出成功');
    showExport.value = false;
    exportForm.password = '';
  } catch (e: any) {
    MessagePlugin.error(e.message || '导出失败');
  } finally {
    exporting.value = false;
  }
}

async function handleDetail(row: any) {
  showDetail.value = true;
  detail.value = null;
  detailLoading.value = true;
  try {
    const res = await cryptoApi.getCADetail(row.ca_id);
    detail.value = (res.data as any) || null;
    if (detail.value && row.algorithm) {
      (detail.value as any).algorithm = row.algorithm;
    }
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载详情失败');
    showDetail.value = false;
  } finally {
    detailLoading.value = false;
  }
}

async function handleDelete(row: any) {
  try {
    await cryptoApi.deleteCA(row.ca_id);
    MessagePlugin.success('删除成功');
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '删除失败');
  }
}

onMounted(async () => {
  await Promise.all([fetchList(), fetchDomains(), algStore.ensureLoaded()]);
});
</script>

<style scoped>
.dn-row {
  display: flex;
  gap: 8px;
  width: 100%;
  align-items: flex-start;
}
.dn-help-btn {
  margin-top: 6px;
  flex-shrink: 0;
  font-weight: 700;
  cursor: help;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 1px solid #909399;
  color: #909399;
  font-size: 11px;
}
.dn-sep-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.dn-sep-hint {
  color: #6b7280;
  font-size: 12px;
}
.dn-preview {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  width: 100%;
}
.file-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  width: 100%;
}
.file-name {
  color: #2563eb;
  font-size: 13px;
  word-break: break-all;
}
.file-placeholder {
  color: #9ca3af;
  font-size: 13px;
}
.domain-list {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
}
.domain-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f0f0;
}
.domain-item:last-child {
  border-bottom: none;
}
.truncated {
  display: inline-block;
  max-width: 100%;
  cursor: pointer;
  color: #2563eb;
  border-bottom: 1px dashed #2563eb;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>