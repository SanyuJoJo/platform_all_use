<template>
  <div>
    <n-space vertical size="large">
      <n-card title="证书管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:cert:sign'"
              type="primary"
              @click="openSign"
            >
              申请证书
            </n-button>
            <n-button
              v-permission="'crypto_console:cert:import'"
              @click="openImport"
            >
              导入证书
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

    <!-- ================= 申请证书 ================= -->
    <n-modal
      v-model:show="showSign"
      preset="card"
      title="申请证书"
      style="width: 900px"
    >
      <n-form
        :model="signForm"
        label-placement="left"
        label-width="130"
      >
        <n-form-item label="CA 来源">
          <n-radio-group v-model:value="signForm.ca_source">
            <n-radio value="local">使用本地 CA</n-radio>
            <n-radio value="manual">手动输入 CA</n-radio>
          </n-radio-group>
        </n-form-item>

        <template v-if="signForm.ca_source === 'local'">
          <n-form-item label="CA 使用者">
            <n-select
              v-model:value="signForm.ca_id"
              :options="caOptions"
              :loading="caLoading"
              filterable
              clearable
              placeholder="请选择本地 CA 的使用者"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item v-if="selectedCA" label="CA 信息">
            <n-descriptions
              bordered
              :column="2"
              size="small"
              label-placement="left"
              style="width: 100%"
            >
              <n-descriptions-item label="CA ID">
                {{ selectedCA.ca_id }}
              </n-descriptions-item>
              <n-descriptions-item label="算法">
                {{ selectedCA.algorithm }}
              </n-descriptions-item>
              <n-descriptions-item label="状态">
                {{ selectedCA.status }}
              </n-descriptions-item>
              <n-descriptions-item label="证书路径">
                <span class="mono">{{ selectedCA.cert_path }}</span>
              </n-descriptions-item>
            </n-descriptions>
          </n-form-item>
        </template>

        <template v-else>
          <n-form-item label="CA 证书 PEM">
            <n-input
              v-model:value="signForm.ca_cert_pem"
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 8 }"
              placeholder="-----BEGIN CERTIFICATE-----"
            />
          </n-form-item>
          <n-form-item label="CA 私钥 PEM">
            <n-input
              v-model:value="signForm.ca_key_pem"
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 8 }"
              placeholder="-----BEGIN PRIVATE KEY-----"
            />
          </n-form-item>
          <n-form-item label="私钥密码">
            <n-input
              v-model:value="signForm.ca_key_password"
              type="password"
              show-password-on="click"
              placeholder="私钥未加密可留空"
            />
          </n-form-item>
        </template>

        <n-form-item label="算法">
          <n-select
            v-model:value="signForm.algorithm"
            :options="algorithmOptions"
          />
        </n-form-item>

        <n-form-item label="使用者 DN">
          <div class="dn-row">
            <n-input
              v-model:value="signForm.dn"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              :placeholder="DN_HELP_EXAMPLE"
            />
            <n-popover
              trigger="click"
              placement="left-start"
              :width="440"
              style="max-width: 90vw"
            >
              <template #trigger>
                <n-button
                  circle
                  size="small"
                  quaternary
                  type="info"
                  class="dn-help-btn"
                  title="DN 项说明"
                >
                  ?
                </n-button>
              </template>
              <div class="dn-help">
                <p class="dn-help__title">DN 项支持</p>
                <p class="dn-help__keys">{{ DN_HELP_KEYS }}</p>
                <p class="dn-help__title">示例</p>
                <p class="dn-help__example">{{ DN_HELP_EXAMPLE }}</p>
                <p class="dn-help__tip">
                  提示：如某字段值本身包含分隔符（例如 O 中出现逗号），
                  请改用其他分隔符（如 “;” 或 “\n” 表示换行）。
                </p>
              </div>
            </n-popover>
          </div>
        </n-form-item>

        <n-form-item label="DN 分隔符">
          <div class="dn-sep-row">
            <n-input
              v-model:value="signForm.dnSeparator"
              placeholder="默认逗号 “,”，可输入 “;” 或 “\n” 表示换行"
              style="max-width: 320px"
            />
            <span class="dn-sep-hint">
              用于拆分 DN 中各项；支持 “\n” 表示换行，“\t” 表示制表符
            </span>
          </div>
        </n-form-item>

        <n-form-item v-if="signForm.dn.trim()" label="解析预览">
          <div class="dn-preview">
            <n-alert
              v-if="dnPreview.errors.length"
              type="error"
              :show-icon="true"
              style="width: 100%"
            >
              <div v-for="(e, i) in dnPreview.errors" :key="i">{{ e }}</div>
            </n-alert>
            <template v-else>
              <n-tag
                v-for="(v, k) in dnPreview.subject"
                :key="k"
                size="small"
                type="success"
                style="margin: 2px 6px 2px 0"
              >
                {{ k }} = {{ v }}
              </n-tag>
            </template>
          </div>
        </n-form-item>

        <n-form-item label="SAN（逗号分隔）">
          <n-input
            v-model:value="signForm.san"
            placeholder="例如 test.example.com,www.example.com"
          />
        </n-form-item>

        <n-form-item label="证书类型（可多选）">
          <n-checkbox-group v-model:value="signForm.cert_types">
            <n-space>
              <n-checkbox value="server">服务端 server</n-checkbox>
              <n-checkbox value="client">客户端 client</n-checkbox>
              <n-checkbox value="signature">签名 signature</n-checkbox>
              <n-checkbox value="encryption">加密 encryption</n-checkbox>
            </n-space>
          </n-checkbox-group>
        </n-form-item>

        <n-form-item label="有效期天数">
          <n-input-number v-model:value="signForm.validity_days" :min="1" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showSign = false">取消</n-button>
          <n-button
            type="primary"
            :loading="submitting"
            :disabled="signDisabled"
            @click="handleSign"
          >
            提交
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 导入证书（文件上传） ================= -->
    <n-modal
      v-model:show="showImport"
      preset="card"
      title="导入证书"
      style="width: 720px"
    >
      <n-alert type="info" :show-icon="true" style="margin-bottom: 12px">
        上传证书文件；如证书有对应私钥，可一并上传私钥文件（加密私钥需填写私钥密码）。
      </n-alert>

      <n-form label-placement="top">
        <n-form-item label="证书文件（PEM/CRT/CER，必填）">
          <div class="file-row">
            <input
              ref="certFileInputRef"
              type="file"
              accept=".pem,.crt,.cer,.cert"
              style="display: none"
              @change="onCertFileChange"
            />
            <n-button @click="certFileInputRef?.click()">选择证书文件</n-button>
            <span v-if="importForm.cert_file_name" class="file-name">
              {{ importForm.cert_file_name }}
            </span>
            <span v-else class="file-placeholder">未选择文件</span>
          </div>
        </n-form-item>

        <n-form-item label="私钥文件（PEM/KEY，可选）">
          <div class="file-row">
            <input
              ref="keyFileInputRef"
              type="file"
              accept=".pem,.key"
              style="display: none"
              @change="onKeyFileChange"
            />
            <n-button @click="keyFileInputRef?.click()">选择私钥文件</n-button>
            <span v-if="importForm.key_file_name" class="file-name">
              {{ importForm.key_file_name }}
            </span>
            <span v-else class="file-placeholder">未选择文件</span>
            <n-button
              v-if="importForm.key_file_name"
              size="tiny"
              quaternary
              @click="clearKeyFile"
            >
              清除
            </n-button>
          </div>
        </n-form-item>

        <n-form-item label="私钥密码（仅加密私钥需要）">
          <n-input
            v-model:value="importForm.key_password"
            type="password"
            show-password-on="click"
            placeholder="私钥未加密可留空"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showImport = false">取消</n-button>
          <n-button
            type="primary"
            :loading="submitting"
            :disabled="!importForm.cert_pem.trim()"
            @click="handleImport"
          >
            导入
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 导出证书 ================= -->
    <n-modal
      v-model:show="showExport"
      preset="card"
      title="导出证书"
      style="width: 560px"
    >
      <n-alert
        v-if="exportForm.type === 'key' || exportForm.type === 'pkcs12'"
        type="warning"
        :show-icon="true"
        style="margin-bottom: 12px"
      >
        导出内容包含私钥，属于敏感操作，会写审计日志。请确保已完成 RBAC 授权
        与业务审批，并在安全环境中妥善保管导出文件。
      </n-alert>

      <n-form :model="exportForm" label-placement="left" label-width="130">
        <n-form-item label="导出内容">
          <n-radio-group v-model:value="exportForm.type">
            <n-space vertical>
              <n-radio value="cert">证书（PEM）</n-radio>
              <n-radio value="key">私钥（PEM）</n-radio>
              <n-radio value="pkcs12">带私钥证书（PKCS#12）</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <n-form-item
          v-if="exportForm.type === 'key'"
          label="私钥口令（可选）"
        >
          <n-input
            v-model:value="exportForm.password"
            type="password"
            show-password-on="click"
            placeholder="留空则导出明文私钥；填写则用 AES-256-CBC 加密"
          />
        </n-form-item>

        <n-form-item
          v-if="exportForm.type === 'pkcs12'"
          label="PKCS#12 口令"
        >
          <n-input
            v-model:value="exportForm.password"
            type="password"
            show-password-on="click"
            placeholder="至少 6 位，导出后用于打开 .p12"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showExport = false">取消</n-button>
          <n-button
            type="primary"
            :loading="submitting"
            :disabled="exportDisabled"
            @click="handleExport"
          >
            导出
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ================= 证书详情 ================= -->
    <n-drawer v-model:show="showDetail" :width="760">
      <n-drawer-content title="证书详情" closable>
        <n-spin :show="detailLoading">
          <n-descriptions
            v-if="detail"
            bordered
            :column="1"
            label-placement="left"
            size="small"
          >
            <n-descriptions-item label="版本号">
              {{ detail.version }}
            </n-descriptions-item>
            <n-descriptions-item label="序列号">
              <span class="mono">{{ detail.serial }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="颁发者">
              {{ detail.issuer }}
            </n-descriptions-item>
            <n-descriptions-item label="使用者">
              {{ detail.subject }}
            </n-descriptions-item>
            <n-descriptions-item label="摘要值">
              <n-popover trigger="click" :width="640" style="max-width: 90vw">
                <template #trigger>
                  <span class="truncated mono">
                    {{ truncate(detail.fingerprint, 64) }}
                  </span>
                </template>
                <div class="full-value mono">{{ detail.fingerprint }}</div>
              </n-popover>
            </n-descriptions-item>
            <n-descriptions-item label="有效期（起）">
              {{ detail.not_before }}
            </n-descriptions-item>
            <n-descriptions-item label="有效期（止）">
              {{ detail.not_after }}
            </n-descriptions-item>
            <n-descriptions-item label="公钥算法">
              {{ detail.public_key_algorithm }}
            </n-descriptions-item>
            <n-descriptions-item label="签名摘要算法">
              {{ detail.signature_algorithm }}
            </n-descriptions-item>
            <n-descriptions-item label="签名摘要值">
              <n-popover trigger="click" :width="640" style="max-width: 90vw">
                <template #trigger>
                  <span class="truncated mono">
                    {{ truncate(detail.signature_value, 64) }}
                  </span>
                </template>
                <div class="full-value mono">{{ detail.signature_value }}</div>
              </n-popover>
            </n-descriptions-item>
            <n-descriptions-item label="公钥值">
              <n-popover trigger="click" :width="640" style="max-width: 90vw">
                <template #trigger>
                  <span class="truncated mono">
                    {{ truncate(detail.public_key_value, 64) }}
                  </span>
                </template>
                <div class="full-value mono">{{ detail.public_key_value }}</div>
              </n-popover>
            </n-descriptions-item>
          </n-descriptions>
          <n-empty v-else-if="!detailLoading" description="暂无详情" />
        </n-spin>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, h } from 'vue';
import {
  NAlert,
  NButton,
  NCard,
  NCheckbox,
  NCheckboxGroup,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NPopover,
  NRadio,
  NRadioGroup,
  NSelect,
  NSpace,
  NSpin,
  NTag,
  useMessage,
  type DataTableColumns,
} from 'naive-ui';
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

const message = useMessage();

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
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
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
  cert_types: ['server'] as CertTypeItem[],
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
  }))
);

const signDisabled = computed(() => {
  if (!signForm.dn.trim() || dnPreview.value.errors.length > 0) return true;
  if (signForm.cert_types.length === 0) return true;
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

const algorithmOptions = [
  { label: 'SM2', value: 'SM2' },
  { label: 'RSA', value: 'RSA' },
  { label: 'ECC', value: 'ECC' },
  { label: 'ML-DSA', value: 'ML-DSA' },
];

const CERT_TYPE_LABEL: Record<string, string> = {
  server: '服务端',
  client: '客户端',
  signature: '签名',
  encryption: '加密',
  imported: '导入',
  ca: 'CA',
  dual_sign: '双证签名',
  dual_enc: '双证加密',
};

function formatCertType(t: string | undefined | null): string {
  if (!t) return '';
  const parts = t
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
  if (parts.length === 0) return '';
  return parts.map((p) => CERT_TYPE_LABEL[p] || p).join('、');
}

const columns: DataTableColumns<any> = [
  {
    title: '使用者',
    key: 'subject_cn',
    ellipsis: { tooltip: true },
  },
  {
    title: '类型',
    key: 'cert_type',
    render(row: any) {
      return formatCertType(row.cert_type);
    },
  },
  {
    title: '序列号',
    key: 'serial',
    ellipsis: { tooltip: true },
  },
  { title: '算法', key: 'algorithm' },
  { title: '状态', key: 'status' },
  { title: '创建时间', key: 'created_at' },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render(row: any) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            { size: 'small', onClick: () => openDetail(row) },
            { default: () => '详情' }
          ),
          h(
            NButton,
            { size: 'small', onClick: () => openExport(row) },
            { default: () => '导出' }
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDelete(row),
              positiveText: '确认删除',
              negativeText: '取消',
            },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', type: 'error' },
                  { default: () => '删除' }
                ),
              default: () =>
                `确认删除证书「${row.subject_cn || row.cert_id}」？该操作会写审计日志，且不可恢复。`,
            }
          ),
        ],
      });
    },
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
      algorithm: c.algorithm,
      cert_path: c.cert_path,
      status: c.status,
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
  signForm.cert_types = ['server'];
  signForm.validity_days = 365;

  if (caList.value.length === 0) {
    fetchCAOptions();
  }
  showSign.value = true;
}

async function handleSign() {
  if (signForm.cert_types.length === 0) {
    message.warning('请至少选择一种证书类型');
    return;
  }
  const r = validateDN(signForm.dn, signForm.dnSeparator);
  if (r.errors.length) {
    message.error(r.errors[0]);
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

    const sanList = signForm.san
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    if (sanList.length) payload.san = sanList;

    if (signForm.ca_source === 'local') {
      if (!signForm.ca_id) {
        message.warning('请选择本地 CA');
        return;
      }
      payload.ca_id = signForm.ca_id;
    } else {
      if (!signForm.ca_cert_pem.trim() || !signForm.ca_key_pem.trim()) {
        message.warning('请填写 CA 证书与私钥');
        return;
      }
      payload.ca_cert_pem = signForm.ca_cert_pem;
      payload.ca_key_pem = signForm.ca_key_pem;
      if (signForm.ca_key_password) {
        payload.ca_key_password = signForm.ca_key_password;
      }
    }

    await cryptoApi.signCert(payload);
    message.success('签发成功');
    showSign.value = false;
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '签发失败');
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
    const text = await file.text();
    importForm.cert_pem = text;
    importForm.cert_file_name = file.name;
  } catch (err: any) {
    message.error('读取证书文件失败: ' + (err?.message || ''));
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
    const text = await file.text();
    importForm.key_pem = text;
    importForm.key_file_name = file.name;
  } catch (err: any) {
    message.error('读取私钥文件失败: ' + (err?.message || ''));
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
    message.warning('请选择证书文件');
    return;
  }
  if (!importForm.cert_pem.includes('-----BEGIN CERTIFICATE-----')) {
    message.warning('证书不是有效 PEM 格式');
    return;
  }
  if (importForm.key_pem && !importForm.key_pem.includes('-----BEGIN')) {
    message.warning('私钥不是有效 PEM 格式');
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
    message.success('导入成功');
    showImport.value = false;
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '导入失败');
  } finally {
    submitting.value = false;
  }
}

function openExport(row: any) {
  exportForm.cert_id = row.cert_id;
  exportForm.type = 'cert';
  exportForm.password = '';
  showExport.value = true;
}

async function handleExport() {
  if (
    exportForm.type === 'pkcs12' &&
    (!exportForm.password || exportForm.password.length < 6)
  ) {
    message.warning('PKCS#12 口令至少 6 位');
    return;
  }

  submitting.value = true;
  try {
    const payload: {
      type: 'cert' | 'key' | 'pkcs12';
      password?: string;
    } = {
      type: exportForm.type,
    };
    if (
      (exportForm.type === 'key' || exportForm.type === 'pkcs12') &&
      exportForm.password
    ) {
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
      const ext =
        exportForm.type === 'pkcs12'
          ? '.p12'
          : exportForm.type === 'key'
            ? '.key.pem'
            : '.pem';
      filename = exportForm.cert_id + ext;
    }

    const blob: Blob =
      res?.data instanceof Blob ? res.data : new Blob([res.data]);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    message.success('导出成功');
    showExport.value = false;
    exportForm.password = '';
  } catch (e: any) {
    message.error(e.message || '导出失败');
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
  } catch (e: any) {
    message.error(e.message || '详情加载失败');
    showDetail.value = false;
  } finally {
    detailLoading.value = false;
  }
}

async function handleDelete(row: any) {
  try {
    await cryptoApi.deleteCert(row.cert_id);
    message.success('删除成功');
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '删除失败');
  }
}

onMounted(async () => {
  await Promise.all([fetchList(), fetchCAOptions()]);
});
</script>

<style scoped>
.dn-row {
  display: flex;
  gap: 8px;
  width: 100%;
  align-items: flex-start;
}
.dn-row .n-input {
  flex: 1;
  min-width: 0;
}
.dn-help-btn {
  margin-top: 6px;
  flex-shrink: 0;
  font-weight: 700;
}
.dn-help {
  line-height: 1.7;
  font-size: 13px;
}
.dn-help__title {
  margin: 0 0 4px;
  font-weight: 600;
  color: #111827;
}
.dn-help__keys {
  margin: 0 0 10px;
  color: #2563eb;
  word-break: break-all;
}
.dn-help__example {
  margin: 0 0 10px;
  color: #374151;
  background: #f3f4f6;
  padding: 6px 8px;
  border-radius: 4px;
  word-break: break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}
.dn-help__tip {
  margin: 0;
  color: #6b7280;
  font-size: 12px;
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
.full-value {
  word-break: break-all;
  line-height: 1.6;
  font-size: 12px;
  max-height: 400px;
  overflow-y: auto;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>
