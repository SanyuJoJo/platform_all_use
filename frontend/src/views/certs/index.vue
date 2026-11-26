<template>
  <t-card :title="'证书管理'" :bordered="false">
    <template #actions>
      <t-space>
        <t-button v-permission="'crypto_console:cert:sign'" theme="primary" @click="openSign">申请证书</t-button>
        <t-button v-permission="'crypto_console:cert:import'" @click="openImport">导入证书</t-button>
        <t-button @click="fetchList">刷新</t-button>
      </t-space>
    </template>

    <t-table
      row-key="cert_id"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="pagination"
      :scroll="{ x: 1500 }"
      @page-change="handlePageChange"
    />
  </t-card>

  <!-- 申请证书 -->
  <t-dialog :visible="showSign" :header="'申请证书'" :confirm-btn="null" :cancel-btn="null" :width="900" @close="showSign = false">
    <t-form :data="signForm" label-align="left" :label-width="130">
      <t-form-item label="CA 来源">
        <t-radio-group v-model="signForm.ca_source">
          <t-radio value="local">使用本地 CA</t-radio>
          <t-radio value="manual">手动输入 CA</t-radio>
        </t-radio-group>
      </t-form-item>
      <template v-if="signForm.ca_source === 'local'">
        <t-form-item label="CA 使用者">
          <t-select v-model="signForm.ca_id" :options="caOptions" :loading="caLoading" filterable clearable placeholder="请选择本地 CA 的使用者" style="width: 100%" />
        </t-form-item>
        <t-form-item v-if="selectedCA" label="CA 信息">
          <t-descriptions :column="2" size="small" label-align="left" bordered style="width: 100%">
            <t-descriptions-item label="CA ID">{{ selectedCA.ca_id }}</t-descriptions-item>
            <t-descriptions-item label="算法">{{ selectedCaAlgLabel }}</t-descriptions-item>
            <t-descriptions-item label="状态">{{ selectedCA.status }}</t-descriptions-item>
            <t-descriptions-item label="证书路径"><span class="mono">{{ selectedCA.cert_path }}</span></t-descriptions-item>
          </t-descriptions>
        </t-form-item>
      </template>
      <template v-else>
        <t-form-item label="CA 证书 PEM">
          <t-textarea v-model="signForm.ca_cert_pem" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN CERTIFICATE-----" />
        </t-form-item>
        <t-form-item label="CA 私钥 PEM">
          <t-textarea v-model="signForm.ca_key_pem" :autosize="{ minRows: 4, maxRows: 8 }" placeholder="-----BEGIN PRIVATE KEY-----" />
        </t-form-item>
        <t-form-item label="私钥密码">
          <t-input v-model="signForm.ca_key_password" type="password" placeholder="私钥未加密可留空" />
        </t-form-item>
      </template>
      <t-form-item label="算法">
        <t-select v-model="signForm.algorithm" :options="algStore.options" :loading="algStore.loading" placeholder="选择算法" />
      </t-form-item>
      <t-form-item label="使用者 DN">
        <div class="dn-row">
          <t-textarea v-model="signForm.dn" :autosize="{ minRows: 3, maxRows: 6 }" :placeholder="DN_HELP_EXAMPLE" />
          <t-tooltip placement="left" :content="DN_HELP_KEYS">
            <span class="dn-help-btn">?</span>
          </t-tooltip>
        </div>
      </t-form-item>
      <t-form-item label="DN 分隔符">
        <t-input v-model="signForm.dnSeparator" placeholder="默认逗号 “,”，可输入 “;” 或 “\\n” 表示换行" style="max-width: 320px" />
      </t-form-item>
      <t-form-item v-if="signForm.dn.trim()" label="解析预览">
        <div class="dn-preview">
          <t-alert v-if="dnPreview.errors.length" theme="error" :message="dnPreview.errors.join('；')" style="width: 100%" />
          <template v-else>
            <t-tag v-for="(v, k) in dnPreview.subject" :key="k" size="small" theme="success" variant="light" style="margin: 2px 6px 2px 0">
              {{ k }} = {{ v }}
            </t-tag>
          </template>
        </div>
      </t-form-item>
      <t-form-item label="SAN（逗号分隔）">
        <t-input v-model="signForm.san" placeholder="例如 test.example.com,www.example.com" />
      </t-form-item>
      <t-form-item label="证书类型（可多选）">
        <t-checkbox-group v-model="signForm.cert_types">
          <t-checkbox value="signature">签名 signature</t-checkbox>
          <t-checkbox value="encryption">加密 encryption</t-checkbox>
        </t-checkbox-group>
      </t-form-item>
      <t-form-item label="有效期天数">
        <t-input-number v-model="signForm.validity_days" :min="1" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showSign = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="signDisabled" @click="handleSign">提交</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 导入证书 -->
  <t-dialog :visible="showImport" :header="'导入证书'" :confirm-btn="null" :cancel-btn="null" :width="720" @close="showImport = false">
    <t-alert theme="info" :message="'上传证书文件；如证书有对应私钥，可一并上传私钥文件（加密私钥需填写私钥密码）。'" style="margin-bottom: 12px" />
    <t-form label-align="top">
      <t-form-item label="证书文件（PEM/CRT/CER，必填）">
        <div class="file-row">
          <input ref="certFileInputRef" type="file" accept=".pem,.crt,.cer,.cert" style="display: none" @change="onCertFileChange" />
          <t-button @click="certFileInputRef?.click()">选择证书文件</t-button>
          <span v-if="importForm.cert_file_name" class="file-name">{{ importForm.cert_file_name }}</span>
          <span v-else class="file-placeholder">未选择文件</span>
        </div>
      </t-form-item>
      <t-form-item label="私钥文件（PEM/KEY，可选）">
        <div class="file-row">
          <input ref="keyFileInputRef" type="file" accept=".pem,.key" style="display: none" @change="onKeyFileChange" />
          <t-button @click="keyFileInputRef?.click()">选择私钥文件</t-button>
          <span v-if="importForm.key_file_name" class="file-name">{{ importForm.key_file_name }}</span>
          <span v-else class="file-placeholder">未选择文件</span>
          <t-button v-if="importForm.key_file_name" size="small" variant="text" @click="clearKeyFile">清除</t-button>
        </div>
      </t-form-item>
      <t-form-item label="私钥密码（仅加密私钥需要）">
        <t-input v-model="importForm.key_password" type="password" placeholder="私钥未加密可留空" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showImport = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="!importForm.cert_pem.trim()" @click="handleImport">导入</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 导出证书 -->
  <t-dialog :visible="showExport" :header="'导出证书'" :confirm-btn="null" :cancel-btn="null" :width="560" @close="showExport = false">
    <t-alert v-if="exportForm.type === 'key' || exportForm.type === 'pkcs12'" theme="warning" :message="'导出内容包含私钥，属于敏感操作，会写审计日志。请确保已完成 RBAC 授权与业务审批，并在安全环境中妥善保管导出文件。'" style="margin-bottom: 12px" />
    <t-alert v-if="pqcPkcs12Warning" theme="warning" :message="`当前证书算法为 PQC（${algStore.displayName(exportForm.algorithm)}），PKCS#12 导出依赖铜锁 8.5+；如失败请升级铜锁后重试。`" style="margin-bottom: 12px" />
    <t-form :data="exportForm" label-align="left" :label-width="130">
      <t-form-item label="导出内容">
        <t-radio-group v-model="exportForm.type">
          <t-radio value="cert">证书（PEM）</t-radio>
          <t-radio value="key">私钥（PEM）</t-radio>
          <t-radio value="pkcs12">带私钥证书（PKCS#12）</t-radio>
        </t-radio-group>
      </t-form-item>
      <t-form-item v-if="exportForm.type === 'key'" label="私钥口令（可选）">
        <t-input v-model="exportForm.password" type="password" placeholder="留空则导出明文私钥；填写则用 AES-256-CBC 加密" />
      </t-form-item>
      <t-form-item v-if="exportForm.type === 'pkcs12'" label="PKCS#12 口令">
        <t-input v-model="exportForm.password" type="password" placeholder="至少 6 位，导出后用于打开 .p12" />
      </t-form-item>
    </t-form>
    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="showExport = false">取消</t-button>
        <t-button theme="primary" :loading="submitting" :disabled="exportDisabled" @click="handleExport">导出</t-button>
      </t-space>
    </template>
  </t-dialog>

  <!-- 证书详情 -->
  <t-drawer :visible="showDetail" :size="760" placement="right" :footer="false" @close="showDetail = false">
    <template #header>证书详情</template>
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
import { ref, reactive, computed, onMounted, h } from 'vue';
import { Tag, Space, Button, Popconfirm } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  cryptoApi,
  type SignCertParams,
  type CertDetail,
  type CAListItem,
  type CertTypeItem,
} from '@/api/crypto';
import {
  DEFAULT_DN_SEPARATOR,
  DN_HELP_KEYS,
  DN_HELP_EXAMPLE,
  parseDN,
  validateDN,
} from '@/utils/dn';
import { useAlgorithmStore } from '@/store/algorithm';
import { publicKeyLabel } from '@/utils/crypto-algorithms';

const algStore = useAlgorithmStore();

const loading = ref(false);
const submitting = ref(false);
const detailLoading = ref(false);
const caLoading = ref(false);

const showSign = ref(false);
const showImport = ref(false);
const showExport = ref(false);
const showDetail = ref(false);

const list = ref<any[]>([]);
const caList = ref<CAListItem[]>([]);
const detail = ref<CertDetail | null>(null);

const certFileInputRef = ref<HTMLInputElement | null>(null);
const keyFileInputRef = ref<HTMLInputElement | null>(null);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const signForm = reactive({
  ca_source: 'local' as 'local' | 'manual',
  ca_id: '',
  ca_cert_pem: '',
  ca_key_pem: '',
  ca_key_password: '',
  algorithm: 'SM2',
  dn: '',
  dnSeparator: DEFAULT_DN_SEPARATOR,
  san: '',
  cert_types: ['signature'] as CertTypeItem[],
  validity_days: 365,
});

const importForm = reactive({
  cert_pem: '',
  cert_file_name: '',
  key_pem: '',
  key_file_name: '',
  key_password: '',
});

const exportForm = reactive({
  cert_id: '',
  type: 'cert' as 'cert' | 'key' | 'pkcs12',
  password: '',
  algorithm: '' as string,
});

const dnPreview = computed(() => parseDN(signForm.dn, signForm.dnSeparator));

const selectedCA = computed<CAListItem | null>(() => {
  if (!signForm.ca_id) return null;
  return caList.value.find((c) => c.ca_id === signForm.ca_id) || null;
});

const caOptions = computed(() =>
  caList.value.map((c) => ({
    label: `${c.subject_cn || c.ca_id} (${c.ca_id})`,
    value: c.ca_id,
  })),
);

const selectedCaAlgLabel = computed(() => {
  if (!selectedCA.value) return '';
  const c = selectedCA.value as any;
  if (c.algorithm) return algStore.label(c.algorithm);
  return publicKeyLabel(c.public_key_algorithm);
});

const detailAlgLabel = computed(() => {
  if (!detail.value) return '';
  const alg = (detail.value as any).algorithm;
  if (alg) return algStore.label(alg);
  return publicKeyLabel(detail.value.public_key_algorithm);
});

const signDisabled = computed(() => {
  if (!signForm.dn.trim() || dnPreview.value.errors.length > 0) return true;
  if (signForm.cert_types.length === 0) return true;
  if (!signForm.algorithm) return true;
  if (signForm.ca_source === 'local') {
    return !signForm.ca_id;
  }
  return !signForm.ca_cert_pem.trim() || !signForm.ca_key_pem.trim();
});

const exportDisabled = computed(() => {
  if (exportForm.type === 'pkcs12') {
    return !exportForm.password || exportForm.password.length < 6;
  }
  return false;
});

const pqcPkcs12Warning = computed(
  () => exportForm.type === 'pkcs12' && algStore.isPqc(exportForm.algorithm),
);

const CERT_TYPE_LABEL: Record<string, string> = {
  signature: '签名',
  encryption: '加密',
  imported: '导入',
  unknown: '未知',
  dual_sign: '双证签名',
  dual_enc: '双证加密',
  server: '服务端',
  client: '客户端',
};

function formatCertType(t: string | undefined | null): string {
  if (!t) return '';
  const segments = t
    .split(';')
    .map((s) => s.trim())
    .filter(Boolean);
  if (segments.length === 0) return '';
  const out: string[] = [];
  for (const seg of segments) {
    const parts = seg
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    if (parts.length === 0) {
      out.push(CERT_TYPE_LABEL[seg] || seg);
      continue;
    }
    for (const p of parts) {
      out.push(CERT_TYPE_LABEL[p] || p);
    }
  }
  return out.join('、');
}

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
  { colKey: 'cert_type', title: '类型', width: 200, cell: (_h, { row }) => formatCertType(row.cert_type) },
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
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openDetail(row) }, () => '详情'),
        h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => openExport(row) }, () => '导出'),
        h(
          Popconfirm,
          { content: `确认删除证书「${row.subject_cn || row.cert_id}」？该操作会写审计日志，且不可恢复。`, confirmBtn: '确认删除', cancelBtn: '取消', onConfirm: () => handleDelete(row) },
          () => h(Button, { size: 'small', variant: 'text', theme: 'danger' }, () => '删除'),
        ),
      ]),
  },
];

function truncate(s: string | undefined | null, n: number): string {
  if (!s) return '';
  return s.length > n ? s.slice(0, n) + '...' : s;
}

async function fetchCAOptions() {
  caLoading.value = true;
  try {
    const res = await cryptoApi.listCas({ page: 1, page_size: 200 });
    const data: any = res.data || {};
    const items: any[] = data.items || data || [];
    caList.value = items.map((c) => ({
      ca_id: c.ca_id,
      subject_cn: c.subject_cn || c.subject || '',
      subject: c.subject,
      public_key_algorithm: c.public_key_algorithm || c.algorithm || '',
      algorithm: c.algorithm,
      cert_path: c.cert_path,
      status: c.status,
      has_private_key: c.has_private_key,
      domain: c.domain,
      not_before: c.not_before,
      not_after: c.not_after,
    }));
  } catch (e: any) {
    console.warn('[cert] fetch CA options failed', e);
    caList.value = [];
  } finally {
    caLoading.value = false;
  }
}

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCerts({ page: pagination.current, page_size: pagination.pageSize });
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

function openSign() {
  signForm.ca_source = 'local';
  signForm.ca_id = '';
  signForm.ca_cert_pem = '';
  signForm.ca_key_pem = '';
  signForm.ca_key_password = '';
  signForm.algorithm = 'SM2';
  signForm.dn = '';
  signForm.dnSeparator = DEFAULT_DN_SEPARATOR;
  signForm.san = '';
  signForm.cert_types = ['signature'];
  signForm.validity_days = 365;
  if (caList.value.length === 0) {
    fetchCAOptions();
  }
  showSign.value = true;
}

async function handleSign() {
  if (signForm.cert_types.length === 0) {
    MessagePlugin.warning('请至少选择一种证书类型');
    return;
  }
  const r = validateDN(signForm.dn, signForm.dnSeparator);
  if (r.errors.length) {
    MessagePlugin.error(r.errors[0]);
    return;
  }

  submitting.value = true;
  try {
    const payload: SignCertParams = {
      ca_source: signForm.ca_source,
      subject: r.subject,
      algorithm: signForm.algorithm,
      cert_types: [...signForm.cert_types],
      validity_days: signForm.validity_days,
    };

    const sanList = signForm.san.split(',').map((s) => s.trim()).filter(Boolean);
    if (sanList.length) payload.san = sanList;

    if (signForm.ca_source === 'local') {
      if (!signForm.ca_id) {
        MessagePlugin.warning('请选择本地 CA');
        return;
      }
      payload.ca_id = signForm.ca_id;
    } else {
      if (!signForm.ca_cert_pem.trim() || !signForm.ca_key_pem.trim()) {
        MessagePlugin.warning('请填写 CA 证书与私钥');
        return;
      }
      payload.ca_cert_pem = signForm.ca_cert_pem;
      payload.ca_key_pem = signForm.ca_key_pem;
      if (signForm.ca_key_password) {
        payload.ca_key_password = signForm.ca_key_password;
      }
    }

    await cryptoApi.signCert(payload);
    MessagePlugin.success('签发成功');
    showSign.value = false;
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '签发失败');
  } finally {
    submitting.value = false;
  }
}

function openImport() {
  importForm.cert_pem = '';
  importForm.cert_file_name = '';
  importForm.key_pem = '';
  importForm.key_file_name = '';
  importForm.key_password = '';
  if (certFileInputRef.value) certFileInputRef.value.value = '';
  if (keyFileInputRef.value) keyFileInputRef.value.value = '';
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
    MessagePlugin.warning('请选择证书文件');
    return;
  }
  if (!importForm.cert_pem.includes('-----BEGIN CERTIFICATE-----')) {
    MessagePlugin.warning('证书不是有效 PEM 格式');
    return;
  }
  if (importForm.key_pem && !importForm.key_pem.includes('-----BEGIN')) {
    MessagePlugin.warning('私钥不是有效 PEM 格式');
    return;
  }

  submitting.value = true;
  try {
    const payload: Record<string, string> = {
      cert_pem: importForm.cert_pem.trim(),
    };
    if (importForm.key_pem.trim()) {
      payload.key_pem = importForm.key_pem.trim();
      if (importForm.key_password) {
        payload.key_password = importForm.key_password;
      }
    }
    await cryptoApi.importCert(payload as any);
    MessagePlugin.success('导入成功');
    showImport.value = false;
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '导入失败');
  } finally {
    submitting.value = false;
  }
}

function openExport(row: any) {
  exportForm.cert_id = row.cert_id;
  exportForm.type = 'cert';
  exportForm.password = '';
  exportForm.algorithm = row.algorithm || row.public_key_algorithm || '';
  showExport.value = true;
}

async function handleExport() {
  if (exportForm.type === 'pkcs12' && (!exportForm.password || exportForm.password.length < 6)) {
    MessagePlugin.warning('PKCS#12 口令至少 6 位');
    return;
  }

  submitting.value = true;
  try {
    const payload: { type: 'cert' | 'key' | 'pkcs12'; password?: string } = {
      type: exportForm.type,
    };
    if ((exportForm.type === 'key' || exportForm.type === 'pkcs12') && exportForm.password) {
      payload.password = exportForm.password;
    }

    const res: any = await cryptoApi.exportCert(exportForm.cert_id, payload);

    let filename = '';
    const cd = res?.headers?.['content-disposition'];
    if (cd) {
      const m = /filename="?([^"]+)"?/.exec(cd);
      if (m) filename = m[1];
    }
    if (!filename) {
      const ext = exportForm.type === 'pkcs12' ? '.p12' : exportForm.type === 'key' ? '.key.pem' : '.pem';
      filename = exportForm.cert_id + ext;
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
    submitting.value = false;
  }
}

async function openDetail(row: any) {
  showDetail.value = true;
  detail.value = null;
  detailLoading.value = true;
  try {
    const res = await cryptoApi.getCertDetail(row.cert_id);
    detail.value = (res.data as any) || null;
    if (detail.value && row.algorithm) {
      (detail.value as any).algorithm = row.algorithm;
    }
  } catch (e: any) {
    MessagePlugin.error(e.message || '详情加载失败');
    showDetail.value = false;
  } finally {
    detailLoading.value = false;
  }
}

async function handleDelete(row: any) {
  try {
    await cryptoApi.deleteCert(row.cert_id);
    MessagePlugin.success('删除成功');
    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '删除失败');
  }
}

onMounted(async () => {
  await Promise.all([fetchList(), fetchCAOptions(), algStore.ensureLoaded()]);
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