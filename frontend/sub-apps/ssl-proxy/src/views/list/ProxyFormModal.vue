<template>
  <n-modal
    :show="show"
    preset="card"
    :title="isEdit ? '编辑配置' : '新建配置'"
    style="width: 620px"
    :mask-closable="false"
    @update:show="handleUpdateShow"
  >
    <n-spin :show="loading">
      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-placement="left"
        label-width="180"
      >
        <n-form-item label="配置名称" path="service_name">
          <n-input
            v-model:value="form.service_name"
            placeholder="请填写配置名称"
            maxlength="64"
            :disabled="isEdit"
          />
          <span v-if="isEdit" class="form-hint-block">
            配置名称创建后不可修改
          </span>
        </n-form-item>

        <n-form-item label="本地监听地址" path="listen_ip">
          <n-select
            v-model:value="form.listen_ip"
            :options="LISTEN_IP_OPTIONS"
            placeholder="请选择"
            filterable
            tag
          />
        </n-form-item>

        <n-form-item label="本地端口" path="listen_port">
          <n-input-number
            v-model:value="form.listen_port"
            :min="1"
            :max="65535"
            placeholder="请填写本地端口"
            style="width: 100%"
          />
        </n-form-item>

        <n-form-item label="后台服务地址" path="backend_ip">
          <n-input
            v-model:value="form.backend_ip"
            placeholder="例如 127.0.0.1 或 0.0.0.0"
          />
        </n-form-item>

        <n-form-item label="后台服务端口" path="backend_port">
          <n-input-number
            v-model:value="form.backend_port"
            :min="1"
            :max="65535"
            placeholder="请填写后台服务端口"
            style="width: 100%"
          />
        </n-form-item>

        <!-- ★ 新增：SSL Server Name（SNI）字段 -->
        <n-form-item>
          <template #label>
            <n-space align="center" :size="4">
              <span>SSL Server Name（SNI）</span>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <div class="help-icon">?</div>
                </template>
                留空时后端会自动从证书的 SAN/CN 获取并填入，无需手动指定。<br/>
                仅当后端依赖 SNI 路由时才需显式填写。
              </n-tooltip>
            </n-space>
          </template>
          <n-input
            v-model:value="form.ssl_server_name"
            :placeholder="sniPlaceholder"
          />
        </n-form-item>

        <n-divider title-placement="left" style="margin: 8px 0 16px">
          配置证书
        </n-divider>

        <n-form-item label="证书类型">
          <n-select
            v-model:value="certTypeUI"
            :options="CERT_TYPE_OPTIONS"
            placeholder="请选择证书类型"
            @update:value="handleCertTypeChange"
          />
        </n-form-item>

        <!-- 客户端：配置证书开关 -->
        <n-form-item v-if="form.mode === 'client'" label="配置证书">
          <n-space align="center">
            <n-switch v-model:value="configureCertUI">
              <template #checked>是</template>
              <template #unchecked>否</template>
            </n-switch>
            <span class="form-hint">
              {{ configureCertUI ? '本代理加载证书' : '本代理不加载证书' }}
            </span>
          </n-space>
        </n-form-item>

        <!-- 服务端：只读提示必配 -->
        <n-form-item v-if="form.mode === 'server'" label="配置证书">
          <n-space align="center">
            <n-tag type="success" size="small">是</n-tag>
            <span class="form-hint">
              服务端默认开启 SSL 通信，必须配置证书
            </span>
          </n-space>
        </n-form-item>

        <!-- 证书选择框 -->
        <template v-if="showCertFields">
          <n-form-item
            v-if="certTypeUI === 'intl' || certTypeUI === 'intl_gmssl'"
            label="国际签名证书"
            :required="certRequired"
          >
            <n-select
              v-model:value="form.intl_cert_id"
              :options="intlCertOptions"
              :loading="intlCertLoading"
              placeholder="请选择国际签名证书"
              clearable
              filterable
            />
          </n-form-item>

          <n-form-item
            v-if="certTypeUI === 'gmssl' || certTypeUI === 'intl_gmssl'"
            label="国密签名证书"
            :required="certRequired"
          >
            <n-select
              v-model:value="form.cert_id"
              :options="signCertOptions"
              :loading="signCertLoading"
              placeholder="请选择国密签名证书"
              clearable
              filterable
            />
          </n-form-item>

          <n-form-item
            v-if="certTypeUI === 'gmssl' || certTypeUI === 'intl_gmssl'"
            label="国密加密证书"
            :required="certRequired"
          >
            <n-select
              v-model:value="form.enc_cert_id"
              :options="encCertOptions"
              :loading="encCertLoading"
              placeholder="请选择国密加密证书"
              clearable
              filterable
            />
          </n-form-item>
        </template>
      </n-form>
    </n-spin>

    <template #footer>
      <n-space justify="end">
        <n-button @click="handleCancel">取消</n-button>
        <n-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ isEdit ? '保存' : '创建' }}
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick } from 'vue';
import {
  NButton,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NSpin,
  NSwitch,
  NTag,
  NTooltip, // ★ 引入 NTooltip
  useMessage,
  type FormInst,
  type FormRules,
} from 'naive-ui';
import proxyApi from '@/api/proxy';
import {
  CERT_TYPE_OPTIONS,
  LISTEN_IP_OPTIONS,
  certTypeToProtocol,
  inferCertType,
  defaultCiphersForMode,
  defaultConfCmdFor,
  defaultProtocolsFor,
  protocolsFromUI,
  protocolsToUI,
  type CertTypeOption,
} from '@/constants/proxyOptions';
import type {
  CertBrief,
  PlatformSSLProxy,
  ProxyCreateReq,
  ProxyMode,
} from '@/types';

const props = defineProps<{
  show: boolean;
  serviceId?: string;
  defaultMode?: ProxyMode;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'success'): void;
}>();

const message = useMessage();
const isEdit = computed(() => !!props.serviceId);
const formRef = ref<FormInst | null>(null);
const loading = ref(false);
const submitting = ref(false);

const certTypeUI = ref<CertTypeOption>('intl');
const configureCertUI = ref<boolean>(false);

// 新建不展示协议选择 UI，但保留状态供 buildPayload 统一处理
const sslProtocolsUI = ref<string[]>([]);

const intlCertOptions = ref<{ label: string; value: string }[]>([]);
const signCertOptions = ref<{ label: string; value: string }[]>([]);
const encCertOptions = ref<{ label: string; value: string }[]>([]);
const intlCertLoading = ref(false);
const signCertLoading = ref(false);
const encCertLoading = ref(false);

// ★ SNI 动态提示文案
const sniPlaceholder = computed(() =>
  form.mode === 'client' && form.backend_scheme === 'https'
    ? '留空自动探测，或填写后端证书的 CN/SAN'
    : 'SNI，可选'
);

function buildEmptyForm(mode: ProxyMode = 'client'): ProxyCreateReq {
  const protocol = 'http';
  return {
    service_name: '',
    mode,
    protocol,
    listen_ip: '0.0.0.0',
    listen_port: 8080,
    backend_ip: '127.0.0.1',
    backend_port: 8080,
    backend_scheme: mode === 'client' ? 'https' : 'http',
    ssl_server_name: '',
    listen_ssl: mode === 'server' ? true : undefined,
    cert_id: '',
    enc_cert_id: '',
    intl_cert_id: '',
    ca_cert_id: '',
    ca_domain: '',
    client_cert_id: '',
    client_enc_cert_id: '',
    intl_client_cert_id: '',
    client_key_ref: '',
    ssl_protocols: defaultProtocolsFor(protocol),
    ssl_ciphers: defaultCiphersForMode(mode),
    ssl_conf_command: defaultConfCmdFor(protocol),
    ssl_ciphers_mode: 'default',
    ssl_session_cache_enabled: mode === 'server' ? false : undefined,
    ssl_verify_client: mode === 'server' ? true : undefined,
    ssl_session_tickets: mode === 'server' ? false : undefined,
  };
}

const form = reactive<ProxyCreateReq>(buildEmptyForm());

const rules: FormRules = {
  service_name: { required: true, message: '请填写配置名称', trigger: 'blur' },
  listen_ip: {
    required: true,
    message: '请选择本地监听地址',
    trigger: ['blur', 'change'],
  },
  listen_port: {
    required: true,
    type: 'number',
    message: '请填写本地端口',
    trigger: ['blur', 'change'],
  },
  backend_ip: {
    required: true,
    message: '请填写后台服务地址',
    trigger: 'blur',
  },
  backend_port: {
    required: true,
    type: 'number',
    message: '请填写后台服务端口',
    trigger: ['blur', 'change'],
  },
};

const showCertFields = computed<boolean>(() => {
  if (form.mode === 'server') return true;
  return configureCertUI.value;
});

const certRequired = computed<boolean>(() => {
  if (form.mode === 'server') return true;
  return configureCertUI.value;
});

function clearAllCertSlots() {
  form.intl_cert_id = '';
  form.cert_id = '';
  form.enc_cert_id = '';
}

function handleCertTypeChange(t: CertTypeOption) {
  if (t === 'intl') {
    form.cert_id = '';
    form.enc_cert_id = '';
  } else if (t === 'gmssl') {
    form.intl_cert_id = '';
  }
  const protocol = certTypeToProtocol(t);
  form.ssl_protocols = defaultProtocolsFor(protocol);
  form.ssl_conf_command = defaultConfCmdFor(protocol);
  sslProtocolsUI.value = protocolsToUI(form.ssl_protocols);
  preloadCertsFor(t);
}

function toOptions(items: CertBrief[]) {
  return items.map((c) => ({
    label: `${c.subject_cn || c.cert_id}（${c.algorithm}）`,
    value: c.cert_id,
  }));
}

async function loadIntlCerts() {
  if (intlCertOptions.value.length > 0) return;
  intlCertLoading.value = true;
  try {
    const res = await proxyApi.listAvailableCerts({
      purpose: 'sign',
      protocol: 'http',
    });
    intlCertOptions.value = toOptions(res.data.items || []);
  } catch (e: any) {
    message.error(e.message || '加载国际签名证书失败');
  } finally {
    intlCertLoading.value = false;
  }
}

async function loadSignCerts() {
  if (signCertOptions.value.length > 0) return;
  signCertLoading.value = true;
  try {
    const res = await proxyApi.listAvailableCerts({
      purpose: 'sign',
      protocol: 'gmssl',
    });
    signCertOptions.value = toOptions(res.data.items || []);
  } catch (e: any) {
    message.error(e.message || '加载国密签名证书失败');
  } finally {
    signCertLoading.value = false;
  }
}

async function loadEncCerts() {
  if (encCertOptions.value.length > 0) return;
  encCertLoading.value = true;
  try {
    const res = await proxyApi.listAvailableCerts({
      purpose: 'enc',
      protocol: 'gmssl',
    });
    encCertOptions.value = toOptions(res.data.items || []);
  } catch (e: any) {
    message.error(e.message || '加载国密加密证书失败');
  } finally {
    encCertLoading.value = false;
  }
}

function preloadCertsFor(t: CertTypeOption) {
  if (t === 'intl') {
    loadIntlCerts();
  } else if (t === 'gmssl') {
    loadSignCerts();
    loadEncCerts();
  } else {
    loadIntlCerts();
    loadSignCerts();
    loadEncCerts();
  }
}

function resetForm(mode: ProxyMode = 'client') {
  Object.assign(form, buildEmptyForm(mode));
  certTypeUI.value = 'intl';
  configureCertUI.value = mode === 'server';
  sslProtocolsUI.value = protocolsToUI(form.ssl_protocols);
}

// ★ 名称唯一性校验：按 (mode, service_name) 判定
async function checkNameUnique(
  name: string,
  mode: ProxyMode,
  excludeServiceId?: string
): Promise<boolean> {
  if (!name) return true;
  try {
    const res = await proxyApi.list({ mode, page: 1, size: 500 });
    const items = res.data.items || [];
    const dup = items.find(
      (it) =>
        it.service_name === name &&
        (!excludeServiceId || it.service_id !== excludeServiceId)
    );
    return !dup;
  } catch {
    return true;
  }
}

function fillFormFromDetail(d: PlatformSSLProxy) {
  Object.assign(form, {
    service_name: d.service_name,
    mode: d.mode,
    protocol: d.protocol,
    listen_ip: d.listen_ip,
    listen_port: d.listen_port,
    backend_ip: d.backend_ip,
    backend_port: d.backend_port,
    backend_scheme: d.backend_scheme,
    ssl_server_name: d.ssl_server_name,
    listen_ssl: d.listen_ssl,
    ca_cert_id: d.ca_cert_id || '',
    ca_domain: d.ca_domain || '',
    client_key_ref: d.client_key_ref,
    ssl_protocols: d.ssl_protocols || defaultProtocolsFor(d.protocol),
    ssl_ciphers: d.ssl_ciphers || defaultCiphersForMode(d.mode),
    ssl_conf_command: d.ssl_conf_command || '',
    ssl_ciphers_mode: d.ssl_ciphers_mode || 'default',
    ssl_session_cache_enabled:
      d.ssl_session_cache_enabled ?? (d.mode === 'server' ? false : undefined),
    ssl_verify_client:
      d.ssl_verify_client ?? (d.mode === 'server' ? true : undefined),
    ssl_session_tickets:
      d.ssl_session_tickets ?? (d.mode === 'server' ? false : undefined),
  });

  sslProtocolsUI.value = protocolsToUI(form.ssl_protocols);

  clearAllCertSlots();

  certTypeUI.value = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );

  const isClient = d.mode === 'client';
  const backCert = isClient ? d.client_cert_id || '' : d.cert_id || '';
  const backEnc = isClient ? d.client_enc_cert_id || '' : d.enc_cert_id || '';
  const backIntl = isClient
    ? d.intl_client_cert_id || ''
    : d.intl_cert_id || '';

  if (certTypeUI.value === 'intl') {
    form.intl_cert_id = backCert;
  } else if (certTypeUI.value === 'gmssl') {
    form.cert_id = backCert;
    form.enc_cert_id = backEnc;
  } else {
    form.intl_cert_id = backIntl;
    form.cert_id = backCert;
    form.enc_cert_id = backEnc;
  }

  if (isClient) {
    configureCertUI.value = !!(
      d.intl_client_cert_id ||
      d.client_cert_id ||
      d.client_enc_cert_id
    );
  } else {
    configureCertUI.value = true;
  }
}

async function fetchDetail(serviceId: string) {
  loading.value = true;
  try {
    const res = await proxyApi.get(serviceId);
    fillFormFromDetail(res.data);
    if (showCertFields.value) {
      preloadCertsFor(certTypeUI.value);
    }
  } catch (e: any) {
    message.error(e.message || '加载失败');
    emit('update:show', false);
  } finally {
    loading.value = false;
  }
}

watch(
  () => [props.show, props.serviceId] as const,
  async ([show]) => {
    if (!show) return;
    await nextTick();
    formRef.value?.restoreValidation();
    if (props.serviceId) {
      await fetchDetail(props.serviceId);
    } else {
      const m = props.defaultMode || 'client';
      resetForm(m);
      preloadCertsFor(certTypeUI.value);
    }
  }
);

function handleUpdateShow(val: boolean) {
  emit('update:show', val);
}

function handleCancel() {
  emit('update:show', false);
}

function validateCertSelection(): string | null {
  if (!certRequired.value) return null;

  if (certTypeUI.value === 'intl') {
    if (!form.intl_cert_id) return '请选择国际签名证书';
  } else if (certTypeUI.value === 'gmssl') {
    if (!form.cert_id) return '请选择国密签名证书';
    if (!form.enc_cert_id) return '请选择国密加密证书';
  } else {
    if (!form.intl_cert_id) return '请选择国际签名证书';
    if (!form.cert_id) return '请选择国密签名证书';
    if (!form.enc_cert_id) return '请选择国密加密证书';
  }
  return null;
}

function buildPayload(): ProxyCreateReq {
  const protocol = certTypeToProtocol(certTypeUI.value);

  const payload: ProxyCreateReq = {
    ...form,
    protocol,
  };

  if (payload.mode === 'server' && payload.listen_ssl === undefined) {
    payload.listen_ssl = true;
  }
  if (payload.mode === 'client') {
    payload.listen_ssl = undefined;
  }

  // 清空所有证书槽位，再按类型赋值
  payload.cert_id = '';
  payload.enc_cert_id = '';
  payload.intl_cert_id = '';
  payload.client_cert_id = '';
  payload.client_enc_cert_id = '';
  payload.intl_client_cert_id = '';

  const certEnabled =
    payload.mode === 'server' ? true : configureCertUI.value;
  if (certEnabled) {
    const isClient = payload.mode === 'client';

    if (certTypeUI.value === 'intl') {
      if (isClient) {
        payload.client_cert_id = form.intl_cert_id || '';
      } else {
        payload.cert_id = form.intl_cert_id || '';
      }
    } else if (certTypeUI.value === 'gmssl') {
      if (isClient) {
        payload.client_cert_id = form.cert_id || '';
        payload.client_enc_cert_id = form.enc_cert_id || '';
      } else {
        payload.cert_id = form.cert_id || '';
        payload.enc_cert_id = form.enc_cert_id || '';
      }
    } else {
      if (isClient) {
        payload.intl_client_cert_id = form.intl_cert_id || '';
        payload.client_cert_id = form.cert_id || '';
        payload.client_enc_cert_id = form.enc_cert_id || '';
      } else {
        payload.intl_cert_id = form.intl_cert_id || '';
        payload.cert_id = form.cert_id || '';
        payload.enc_cert_id = form.enc_cert_id || '';
      }
    }
  }

  // ★ ssl_protocols：纯 join，不做剔除，后端 sanitizeSSLProtocols 负责过滤
  payload.ssl_protocols = protocolsFromUI(sslProtocolsUI.value, protocol);

  // SSL 扩展字段：按 mode 归一化
  if (payload.mode === 'client') {
    payload.ssl_session_cache_enabled = undefined;
    payload.ssl_verify_client = undefined;
    payload.ssl_session_tickets = undefined;
  }

  if (
    payload.ssl_ciphers_mode !== 'default' &&
    !(payload.ssl_ciphers || '').trim()
  ) {
    payload.ssl_ciphers_mode = 'default';
    payload.ssl_ciphers = defaultCiphersForMode(payload.mode);
  }

  return payload;
}

async function handleSubmit() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  const certErr = validateCertSelection();
  if (certErr) {
    message.error(certErr);
    return;
  }

  const unique = await checkNameUnique(
    form.service_name,
    form.mode,
    isEdit.value ? props.serviceId : undefined
  );
  if (!unique) {
    message.error(
      `配置名称「${form.service_name}」在当前${
        form.mode === 'client' ? '客户端' : '服务端'
      }下已存在，请换一个`
    );
    return;
  }

  const payload = buildPayload();

  submitting.value = true;
  try {
    if (isEdit.value && props.serviceId) {
      await proxyApi.update(props.serviceId, payload);
      message.success('保存成功');
    } else {
      await proxyApi.create(payload);
      message.success('创建成功');
    }
    emit('update:show', false);
    emit('success');
  } catch (e: any) {
    if (e?.code === 40903 || /已存在|duplicate|unique/i.test(e?.message || '')) {
      message.error(
        `配置名称「${form.service_name}」在当前${
          form.mode === 'client' ? '客户端' : '服务端'
        }下已存在，请换一个`
      );
      return;
    }
    message.error(e.message || '提交失败');
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.form-hint {
  color: #6b7280;
  font-size: 12px;
}
.form-hint-block {
  display: block;
  color: #9ca3af;
  font-size: 12px;
  margin-top: 4px;
}
/* 问号提示图标样式 */
.help-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1px solid #909399;
  color: #909399;
  font-size: 11px;
  cursor: help;
  user-select: none;
}
</style>
