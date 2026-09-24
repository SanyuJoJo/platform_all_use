<template>
  <div>
    <n-space vertical size="large">
      <n-card title="CA 管理">
        <template #header-extra>
          <n-space>
            <n-button
              v-permission="'crypto_console:ca:import'"
              @click="openImport"
            >
              导入 CA
            </n-button>
            <n-button
              v-permission="'crypto_console:ca:create'"
              type="primary"
              @click="openCreate"
            >
              创建根 CA
            </n-button>
          </n-space>
        </template>

        <n-data-table
          :columns="columns"
          :data="list"
          :loading="loading"
          :pagination="false"
        />
      </n-card>
    </n-space>

    <!-- ============ 创建根 CA ============ -->
    <n-modal
      v-model:show="showCreate"
      preset="card"
      title="创建根 CA"
      style="width: 720px"
    >
      <n-form :model="form" label-placement="left" label-width="120">
        <n-form-item label="摘要算法">
          <n-select v-model:value="form.algorithm" :options="algorithmOptions" />
        </n-form-item>

        <!-- 使用者 DN -->
        <n-form-item label="使用者 DN">
          <div class="dn-row">
            <n-input
              v-model:value="form.dn"
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
              v-model:value="form.dnSeparator"
              placeholder="默认逗号 “,”，可输入 “;” 或 “\n” 表示换行"
              style="max-width: 320px"
            />
            <span class="dn-sep-hint">
              用于拆分 DN 中各项；支持 “\n” 表示换行，“\t” 表示制表符
            </span>
          </div>
        </n-form-item>

        <n-form-item v-if="form.dn.trim()" label="解析预览">
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

        <n-form-item label="有效期天数">
          <n-input-number v-model:value="form.validity_days" :min="1" />
        </n-form-item>

        <n-form-item label="RSA 位数" v-if="form.algorithm === 'RSA'">
          <n-select v-model:value="form.key_params.key_size" :options="rsaSizeOptions" />
        </n-form-item>

        <n-form-item label="ECC 曲线" v-if="form.algorithm === 'ECC'">
          <n-select v-model:value="form.key_params.curve" :options="eccCurveOptions" />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreate = false">取消</n-button>
          <n-button
            type="primary"
            :loading="submitting"
            :disabled="dnPreview.errors.length > 0 || !form.dn.trim()"
            @click="handleCreate"
          >
            提交
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ============ 导入 CA ============ -->
    <n-modal
      v-model:show="showImport"
      preset="card"
      title="导入 CA"
      style="width: 680px"
    >
      <n-alert type="info" :show-icon="true" style="margin-bottom: 12px">
        支持两种导入方式：
        <br />① 仅导入 CA 证书（不选私钥文件）；
        <br />② 导入带私钥的 CA（选择证书 + 私钥文件，加密私钥需填写密码）。
      </n-alert>

      <n-form label-placement="top">
        <!-- CA 证书文件 -->
        <n-form-item label="CA 证书文件（PEM / CRT / CER，必填）">
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

        <!-- CA 私钥文件 -->
        <n-form-item label="CA 私钥文件（PEM，可选，不选则只导入证书）">
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

        <!-- 私钥密码（常显，加密私钥才需填写） -->
        <n-form-item label="私钥密码（仅加密私钥需要填写）">
          <n-input
            v-model:value="importForm.key_password"
            type="password"
            show-password-on="click"
            placeholder="若私钥未加密，留空即可"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showImport = false">取消</n-button>
          <n-button
            type="primary"
            :loading="importing"
            :disabled="!importForm.cert_pem.trim()"
            @click="handleImport"
          >
            导入
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ============ 导出 CA ============ -->
    <n-modal
      v-model:show="showExport"
      preset="card"
      title="导出 CA"
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

      <n-form :model="exportForm" label-placement="left" label-width="100">
        <n-form-item label="导出内容">
          <n-radio-group v-model:value="exportForm.type">
            <n-space vertical>
              <n-radio value="cert">CA 证书（PEM）</n-radio>
              <n-radio value="key">CA 私钥（PEM）</n-radio>
              <n-radio value="pkcs12">带私钥的 CA（PKCS#12）</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <n-form-item v-if="exportForm.type === 'pkcs12'" label="PKCS#12 密码">
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
            :loading="exporting"
            :disabled="
              exportForm.type === 'pkcs12' &&
              (!exportForm.password || exportForm.password.length < 6)
            "
            @click="handleExport"
          >
            导出
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- ============ CA 证书详情 ============ -->
    <n-drawer v-model:show="showDetail" :width="760">
      <n-drawer-content title="CA 证书详情" closable>
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
                  <span class="truncated mono">{{ truncate(detail.fingerprint, 64) }}</span>
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
                  <span class="truncated mono">{{ truncate(detail.signature_value, 64) }}</span>
                </template>
                <div class="full-value mono">{{ detail.signature_value }}</div>
              </n-popover>
            </n-descriptions-item>
            <n-descriptions-item label="公钥值">
              <n-popover trigger="click" :width="640" style="max-width: 90vw">
                <template #trigger>
                  <span class="truncated mono">{{ truncate(detail.public_key_value, 64) }}</span>
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
import { ref, onMounted, reactive, computed, h } from 'vue';
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NPopover,
  NAlert,
  NTag,
  NRadioGroup,
  NRadio,
  NPopconfirm,
  NDrawer,
  NDrawerContent,
  NDescriptions,
  NDescriptionsItem,
  NSpin,
  NEmpty,
  useMessage,
} from 'naive-ui';
import { cryptoApi, type CAExportType, type CADetail } from '@/api/crypto';
import { useOperation } from '@/composables/useOperation';
import {
  DEFAULT_DN_SEPARATOR,
  DN_HELP_KEYS,
  DN_HELP_EXAMPLE,
  validateDN,
  parseDN,
} from '@/utils/dn';

const message = useMessage();
const { run } = useOperation();

const loading = ref(false);
const submitting = ref(false);
const importing = ref(false);
const exporting = ref(false);
const detailLoading = ref(false);

const showCreate = ref(false);
const showImport = ref(false);
const showExport = ref(false);
const showDetail = ref(false);

const list = ref<any[]>([]);
const detail = ref<CADetail | null>(null);

const certFileInputRef = ref<HTMLInputElement | null>(null);
const keyFileInputRef = ref<HTMLInputElement | null>(null);

const form = reactive({
  algorithm: 'SM2',
  dn: '',
  dnSeparator: DEFAULT_DN_SEPARATOR,
  validity_days: 3650,
  key_params: { key_size: 2048, curve: 'prime256v1' },
});

const importForm = reactive({
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

const columns = [
  { title: 'CA ID', key: 'ca_id' },
  { title: '主题', key: 'subject_cn', ellipsis: { tooltip: true } },
  { title: '算法', key: 'algorithm' },
  { title: '序列号', key: 'serial', ellipsis: { tooltip: true } },
  { title: '状态', key: 'status' },
  { title: '创建时间', key: 'created_at' },
  {
    title: '操作',
    key: 'actions',
    width: 240,
    render(row: any) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            { size: 'small', onClick: () => handleDetail(row) },
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
                `确认删除 CA「${row.ca_id}」？该操作不可恢复，且会写审计日志。`,
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

async function fetchList() {
  loading.value = true;
  try {
    const res = await cryptoApi.listCas();
    list.value = (res.data as any)?.items || res.data || [];
  } catch (e: any) {
    message.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  form.algorithm = 'SM2';
  form.dn = '';
  form.dnSeparator = DEFAULT_DN_SEPARATOR;
  form.validity_days = 3650;
  form.key_params.key_size = 2048;
  form.key_params.curve = 'prime256v1';
  showCreate.value = true;
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

function openExport(row: any) {
  exportForm.ca_id = row.ca_id;
  exportForm.type = 'cert';
  exportForm.password = '';
  showExport.value = true;
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

async function handleDetail(row: any) {
  showDetail.value = true;
  detail.value = null;
  detailLoading.value = true;
  try {
    const res = await cryptoApi.getCADetail(row.ca_id);
    detail.value = (res.data as any) || null;
  } catch (e: any) {
    message.error(e.message || '加载详情失败');
    showDetail.value = false;
  } finally {
    detailLoading.value = false;
  }
}

async function handleCreate() {
  const result = validateDN(form.dn, form.dnSeparator);
  if (result.errors.length) {
    message.error(result.errors[0]);
    return;
  }

  submitting.value = true;
  try {
    const params: Record<string, unknown> = {
      algorithm: form.algorithm,
      subject: result.subject,
      validity_days: form.validity_days,
      key_protection: { cipher: 'AES-256-GCM' },
    };
    if (form.algorithm === 'RSA') {
      params.key_params = { key_size: form.key_params.key_size };
    } else if (form.algorithm === 'ECC') {
      params.key_params = { curve: form.key_params.curve };
    }
    await run('ca.create', params);
    showCreate.value = false;
    form.dn = '';
    form.dnSeparator = DEFAULT_DN_SEPARATOR;
    await fetchList();
  } finally {
    submitting.value = false;
  }
}

async function handleImport() {
  if (!importForm.cert_pem.trim()) {
    message.warning('请选择 CA 证书文件');
    return;
  }
  if (!importForm.cert_pem.includes('-----BEGIN CERTIFICATE-----')) {
    message.warning('证书文件不是有效的 PEM 格式');
    return;
  }
  if (importForm.key_pem && !importForm.key_pem.includes('-----BEGIN')) {
    message.warning('私钥文件不是有效的 PEM 格式');
    return;
  }

  importing.value = true;
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
    await cryptoApi.importCA(payload as any);
    message.success('导入成功');
    showImport.value = false;
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '导入失败');
  } finally {
    importing.value = false;
  }
}

async function handleExport() {
  if (!exportForm.ca_id) {
    message.warning('请选择要导出的 CA');
    return;
  }
  if (
    exportForm.type === 'pkcs12' &&
    (!exportForm.password || exportForm.password.length < 6)
  ) {
    message.warning('PKCS#12 密码至少 6 位');
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
      const ext =
        exportForm.type === 'cert'
          ? 'pem'
          : exportForm.type === 'key'
            ? 'key.pem'
            : 'p12';
      filename = `${exportForm.ca_id}.${ext}`;
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
    exporting.value = false;
  }
}

async function handleDelete(row: any) {
  try {
    await cryptoApi.deleteCA(row.ca_id);
    message.success('删除成功');
    await fetchList();
  } catch (e: any) {
    message.error(e.message || '删除失败');
  }
}

onMounted(fetchList);
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
