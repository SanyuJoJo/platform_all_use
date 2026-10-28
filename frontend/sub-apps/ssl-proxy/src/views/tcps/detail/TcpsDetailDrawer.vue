<template>
  <n-drawer
    :show="show"
    :width="760"
    placement="right"
    :mask-closable="false"
    @update:show="handleUpdateShow"
  >
    <n-drawer-content closable>
      <template #header>
        <div class="drawer-header">
          <n-space align="center">
            <span>{{ isEditing ? '编辑配置' : '配置详情' }}</span>
            <n-tag v-if="detail" :type="TCPS_STATUS_TAG_TYPE[detail.status] || 'default'" size="small">
              {{ detail.status }}
            </n-tag>
          </n-space>

          <n-space class="drawer-header__actions">
            <template v-if="!isEditing">
              <n-button v-if="hasPerm('ssl_proxy:tcps:update')" size="small" type="primary" @click="enterEdit">
                编辑
              </n-button>
            </template>
            <template v-else>
              <n-button size="small" @click="cancelEdit">取消</n-button>
              <n-button size="small" type="primary" :loading="submitting" @click="handleSave">
                保存
              </n-button>
            </template>
          </n-space>
        </div>
      </template>

      <n-spin :show="loading">
        <n-space style="margin-bottom: 16px">
          <n-button :type="activeSection === 'basic' ? 'primary' : 'default'" size="small" @click="activeSection = 'basic'">基础配置</n-button>
          <n-button :type="activeSection === 'backend' ? 'primary' : 'default'" size="small" @click="activeSection = 'backend'">后端代理配置</n-button>
          <n-button :type="activeSection === 'ssl' ? 'primary' : 'default'" size="small" @click="activeSection = 'ssl'">SSL配置</n-button>
        </n-space>

        <!-- 基础配置 -->
        <div v-show="activeSection === 'basic'">
          <n-descriptions v-if="!isEditing && detail" bordered :column="1" label-placement="left" size="small">
            <n-descriptions-item label="配置名称">{{ detail.service_name }}</n-descriptions-item>
            <n-descriptions-item label="本地监听地址"><span class="mono">{{ detail.listen_ip }}</span></n-descriptions-item>
            <n-descriptions-item label="本地端口">{{ detail.listen_port }}</n-descriptions-item>
            <n-descriptions-item label="后台服务地址"><span class="mono">{{ detail.backend_ip }}</span></n-descriptions-item>
            <n-descriptions-item label="后台服务端口">{{ detail.backend_port }}</n-descriptions-item>
            <!-- ★ 新增查看模式 SNI -->
            <n-descriptions-item label="SSL Server Name（SNI）">
              {{ detail.ssl_server_name || '—' }}
            </n-descriptions-item>
            
            <n-descriptions-item label="证书类型">{{ certTypeLabel(detail) }}</n-descriptions-item>
            <n-descriptions-item label="配置证书">
              <n-tag :type="hasAnyCert(detail) ? 'success' : 'default'" size="small">{{ hasAnyCert(detail) ? '是' : '否' }}</n-tag>
            </n-descriptions-item>

            <template v-if="showCertInDetail(detail)">
              <n-descriptions-item v-if="showIntlInDetail(detail)" label="国际签名证书">{{ certName(pickIntlCert(detail)) || '—' }}</n-descriptions-item>
              <n-descriptions-item v-if="showSignInDetail(detail)" label="国密签名证书">{{ certName(pickSignCert(detail)) || '—' }}</n-descriptions-item>
              <n-descriptions-item v-if="showEncInDetail(detail)" label="国密加密证书">{{ certName(pickEncCert(detail)) || '—' }}</n-descriptions-item>
            </template>

            <n-descriptions-item label="信任域(验证服务端证书)">
              <n-tag v-if="detail.ca_domain" size="small" type="info">{{ detail.ca_domain }}</n-tag>
              <span v-else>—</span>
            </n-descriptions-item>

            <n-descriptions-item label="代理模式">{{ detail.mode === 'client' ? '客户端' : '服务端' }}</n-descriptions-item>
            <n-descriptions-item label="service_id"><span class="mono">{{ detail.service_id }}</span></n-descriptions-item>
            <n-descriptions-item label="PID">{{ detail.pid > 0 ? detail.pid : '—' }}</n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ detail.created_at }}</n-descriptions-item>
            <n-descriptions-item label="更新时间">{{ detail.updated_at }}</n-descriptions-item>
          </n-descriptions>

          <n-form v-else ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="180">
            <n-form-item label="配置名称" path="service_name">
              <n-input v-model:value="form.service_name" placeholder="请填写配置名称" maxlength="64" disabled />
              <span class="form-hint-block">配置名称创建后不可修改</span>
            </n-form-item>
            <n-form-item label="本地监听地址" path="listen_ip">
              <n-select v-model:value="form.listen_ip" :options="TCPS_LISTEN_IP_OPTIONS" placeholder="请选择" filterable tag />
            </n-form-item>
            <n-form-item label="本地端口" path="listen_port">
              <n-input-number v-model:value="form.listen_port" :min="1" :max="65535" placeholder="请填写本地端口" style="width: 100%" />
            </n-form-item>
            <n-form-item label="后台服务地址" path="backend_ip">
              <n-input v-model:value="form.backend_ip" placeholder="例如 127.0.0.1 或 0.0.0.0" />
            </n-form-item>
            <n-form-item label="后台服务端口" path="backend_port">
              <n-input-number v-model:value="form.backend_port" :min="1" :max="65535" placeholder="请填写后台服务端口" style="width: 100%" />
            </n-form-item>

            <!-- ★ 新增编辑模式 SNI，带 Tooltip 提示 -->
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

            <n-form-item label="证书类型">
              <n-select v-model:value="certTypeUI" :options="TCPS_CERT_TYPE_OPTIONS" placeholder="请选择证书类型" @update:value="handleCertTypeChange" />
            </n-form-item>
            <n-form-item label="配置证书">
              <n-space align="center">
                <n-switch v-model:value="configureCertUI" :disabled="certSwitchDisabled">
                  <template #checked>是</template>
                  <template #unchecked>否</template>
                </n-switch>
                <span class="form-hint">{{ certSwitchHint }}</span>
              </n-space>
            </n-form-item>

            <template v-if="showCertFields">
              <n-form-item v-if="certTypeUI === 'intl' || certTypeUI === 'intl_gmssl'" label="国际签名证书" :required="certRequired">
                <n-select v-model:value="form.intl_cert_id" :options="intlCertOptions" :loading="intlCertLoading" placeholder="请选择国际签名证书" clearable filterable />
              </n-form-item>
              <n-form-item v-if="certTypeUI === 'gmssl' || certTypeUI === 'intl_gmssl'" label="国密签名证书" :required="certRequired">
                <n-select v-model:value="form.cert_id" :options="signCertOptions" :loading="signCertLoading" placeholder="请选择国密签名证书" clearable filterable />
              </n-form-item>
              <n-form-item v-if="certTypeUI === 'gmssl' || certTypeUI === 'intl_gmssl'" label="国密加密证书" :required="certRequired">
                <n-select v-model:value="form.enc_cert_id" :options="encCertOptions" :loading="encCertLoading" placeholder="请选择国密加密证书" clearable filterable />
              </n-form-item>
            </template>

            <n-form-item label="信任域(验证服务端证书)">
              <n-select v-model:value="form.ca_domain" :options="domainOptions" :loading="domainLoading" placeholder="请选择 CA 信任域，留空表示不校验" clearable filterable />
            </n-form-item>
          </n-form>
        </div>

        <!-- 后端代理配置 -->
        <div v-show="activeSection === 'backend'">
          <n-descriptions v-if="!isEditing && detail" bordered :column="1" label-placement="left" size="small">
            <n-descriptions-item label="开启 SSL 通信">
              <n-tag :type="detail.backend_tls ? 'success' : 'default'" size="small">{{ detail.backend_tls ? '是' : '否' }}</n-tag>
            </n-descriptions-item>
          </n-descriptions>
          <n-form v-else ref="backendFormRef" :model="form" label-placement="left" label-width="140">
            <n-form-item label="开启 SSL 通信">
              <n-space align="center">
                <n-switch v-model:value="form.backend_tls">
                  <template #checked>是</template>
                  <template #unchecked>否</template>
                </n-switch>
                <span class="form-hint">{{ form.backend_tls ? '使用 TLS 访问后台服务' : '使用明文 TCP 访问后台服务' }}</span>
              </n-space>
            </n-form-item>
          </n-form>
        </div>

        <!-- SSL配置 -->
        <div v-show="activeSection === 'ssl'">
          <n-descriptions v-if="!isEditing && detail" bordered :column="1" label-placement="left" size="small">
            <n-descriptions-item label="SSL协议">{{ formatProtocolsForDisplay(detail.ssl_protocols) }}</n-descriptions-item>
            <n-descriptions-item label="SSL加密套件"><div class="mono" style="word-break: break-all">{{ detail.ssl_ciphers || tcpsDefaultCiphers() }}</div></n-descriptions-item>
            <n-descriptions-item label="SSL 扩展指令"><div class="mono" style="word-break: break-all">{{ detail.ssl_conf_command || '—' }}</div></n-descriptions-item>
            <template v-if="detail.mode === 'server'">
              <n-descriptions-item label="SSL会话缓存"><n-tag :type="detail.ssl_session_cache_enabled ? 'success' : 'default'" size="small">{{ detail.ssl_session_cache_enabled ? '开启' : '关闭' }}</n-tag></n-descriptions-item>
              <n-descriptions-item label="SSL双向认证"><n-tag :type="detail.ssl_verify_client ? 'success' : 'default'" size="small">{{ detail.ssl_verify_client ? '开启' : '关闭' }}</n-tag></n-descriptions-item>
              <n-descriptions-item label="SSL会话票据"><n-tag :type="detail.ssl_session_tickets ? 'success' : 'default'" size="small">{{ detail.ssl_session_tickets ? '开启' : '关闭' }}</n-tag></n-descriptions-item>
            </template>
          </n-descriptions>
          <n-form v-else ref="sslFormRef" :model="form" label-placement="left" label-width="140">
            <n-form-item label="SSL协议">
              <n-checkbox-group v-model:value="sslProtocolsUI">
                <n-space>
                  <n-checkbox v-for="opt in sslProtocolOptions" :key="opt.value" :value="opt.value" :label="opt.label" />
                </n-space>
              </n-checkbox-group>
            </n-form-item>
            <n-form-item label="SSL加密套件">
              <n-input v-model:value="form.ssl_ciphers" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" :placeholder="tcpsDefaultCiphers()" />
            </n-form-item>
            <n-form-item label="SSL 扩展指令">
              <n-input v-model:value="form.ssl_conf_command" type="textarea" :autosize="{ minRows: 2, maxRows: 4 }" placeholder="例如 Groups X25519MLKEM768:X25519" />
            </n-form-item>
            <template v-if="detail?.mode === 'server'">
              <n-form-item label="SSL会话缓存"><n-switch v-model:value="form.ssl_session_cache_enabled" /></n-form-item>
              <n-form-item label="SSL双向认证"><n-switch v-model:value="form.ssl_verify_client" /></n-form-item>
              <n-form-item label="SSL会话票据"><n-switch v-model:value="form.ssl_session_tickets" /></n-form-item>
            </template>
          </n-form>
        </div>
      </n-spin>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { ref, reactive, watch, nextTick, computed } from 'vue';
import {
  NButton, NCheckbox, NCheckboxGroup, NDescriptions, NDescriptionsItem,
  NDrawer, NDrawerContent, NForm, NFormItem, NInput, NInputNumber,
  NSelect, NSpace, NSpin, NSwitch, NTag, NTooltip, useMessage, type FormInst, type FormRules,
} from 'naive-ui';
import tcpsApi from '@/api/tcps';
import { useTcpsCertMap } from '@/composables/useTcpsCertMap';
import { usePermissionStore } from '@/store/permission';
import {
  TCPS_CERT_TYPE_OPTIONS, TCPS_LISTEN_IP_OPTIONS, TCPS_STATUS_TAG_TYPE,
  TCPS_SSL_PROTOCOL_OPTIONS, TCPS_GM_SSL_PROTOCOL_VALUE,
  tcpsCertTypeToProtocol, inferTcpsCertType, tcpsDefaultProtocolsFor,
  tcpsDefaultCiphers, tcpsProtocolsFromUI, tcpsProtocolsToUI, type TcpsCertTypeOption,
} from '@/constants/tcpsOptions';
import type { TcpsCertBrief, TcpsProxy, TcpsCreateReq } from '@/types/tcps';

const props = defineProps<{ show: boolean; serviceId?: string; }>();
const emit = defineEmits<{ (e: 'update:show', value: boolean): void; (e: 'saved'): void; }>();

const message = useMessage();
const { load: loadCertMap, certName } = useTcpsCertMap();
const permissionStore = usePermissionStore();
const hasPerm = (code: string) => permissionStore.hasPermission(code);

const loading = ref(false);
const submitting = ref(false);
const detail = ref<TcpsProxy | null>(null);
const isEditing = ref(false);
const activeSection = ref<'basic' | 'backend' | 'ssl'>('basic');

const formRef = ref<FormInst | null>(null);
const backendFormRef = ref<FormInst | null>(null);
const sslFormRef = ref<FormInst | null>(null);

const certTypeUI = ref<TcpsCertTypeOption>('intl');
const configureCertUI = ref<boolean>(false);
const sslProtocolsUI = ref<string[]>([]);

const domainOptions = ref<{ label: string; value: string }[]>([]);
const domainLoading = ref(false);

const sslProtocolOptions = computed(() => {
  return TCPS_SSL_PROTOCOL_OPTIONS.filter((opt) => {
    if (opt.value === TCPS_GM_SSL_PROTOCOL_VALUE) return form.protocol === 'gmssl';
    return true;
  });
});

const intlCertOptions = ref<{ label: string; value: string }[]>([]);
const signCertOptions = ref<{ label: string; value: string }[]>([]);
const encCertOptions = ref<{ label: string; value: string }[]>([]);
const intlCertLoading = ref(false);
const signCertLoading = ref(false);
const encCertLoading = ref(false);

function buildEmptyForm(mode: 'client' | 'server' = 'client'): TcpsCreateReq {
  const protocol = 'tcp';
  return {
    service_name: '', mode, protocol,
    listen_ip: '0.0.0.0', listen_port: 9000,
    backend_ip: '127.0.0.1', backend_port: 9000,
    backend_tls: false, ssl_server_name: '',
    listen_ssl: mode === 'server' ? true : undefined,
    cert_id: '', enc_cert_id: '', intl_cert_id: '',
    ca_cert_id: '', ca_domain: '', client_cert_id: '',
    client_enc_cert_id: '', intl_client_cert_id: '', client_key_ref: '',
    ssl_protocols: tcpsDefaultProtocolsFor(protocol),
    ssl_ciphers: tcpsDefaultCiphers(),
    ssl_conf_command: '',
    ssl_ciphers_mode: 'default',
    ssl_session_cache_enabled: mode === 'server' ? false : undefined,
    ssl_verify_client: mode === 'server' ? true : undefined,
    ssl_session_tickets: mode === 'server' ? false : undefined,
  };
}

const form = reactive<TcpsCreateReq>(buildEmptyForm());

const rules: FormRules = {
  service_name: { required: true, message: '请填写配置名称', trigger: 'blur' },
  listen_ip: { required: true, message: '请选择本地监听地址', trigger: ['blur', 'change'] },
  listen_port: { required: true, type: 'number', message: '请填写本地端口', trigger: ['blur', 'change'] },
  backend_ip: { required: true, message: '请填写后台服务地址', trigger: 'blur' },
  backend_port: { required: true, type: 'number', message: '请填写后台服务端口', trigger: ['blur', 'change'] },
};

const showCertFields = computed(() => configureCertUI.value);
const certRequired = computed(() => configureCertUI.value);
const certSwitchDisabled = computed(() => detail.value?.mode === 'server');

const certSwitchHint = computed(() => {
  if (detail.value?.mode === 'server') return '服务端默认开启 SSL 通信，必须配置证书';
  return configureCertUI.value ? '本代理加载证书' : '本代理不加载证书';
});

const sniPlaceholder = computed(() =>
  form.mode === 'client' && form.protocol !== 'tcp'
    ? '留空自动探测，或填写后端证书的 CN/SAN'
    : 'SNI，可选'
);

async function loadDomains() {
  if (domainOptions.value.length > 0) return;
  domainLoading.value = true;
  try {
    const res: any = await tcpsApi.listCADomains();
    const raw = res.data;
    let domains: string[] = [];
    if (Array.isArray(raw?.domains)) domains = raw.domains;
    else if (Array.isArray(raw?.data?.domains)) domains = raw.data.domains;
    else if (Array.isArray(raw)) domains = raw;
    if (!domains.includes('safe')) domains.unshift('safe');
    domainOptions.value = domains.map((d) => ({ label: d, value: d }));
  } catch (e: any) {
    message.error(e.message || '加载 CA 信任域失败');
    domainOptions.value = [{ label: 'safe', value: 'safe' }];
  } finally {
    domainLoading.value = false;
  }
}

function clearAllCertSlots() {
  form.intl_cert_id = '';
  form.cert_id = '';
  form.enc_cert_id = '';
}

function handleCertTypeChange(t: TcpsCertTypeOption) {
  if (t === 'intl') {
    form.cert_id = ''; form.enc_cert_id = '';
  } else if (t === 'gmssl') {
    form.intl_cert_id = '';
  }
  const protocol = tcpsCertTypeToProtocol(t);
  form.ssl_protocols = tcpsDefaultProtocolsFor(protocol);
  sslProtocolsUI.value = tcpsProtocolsToUI(form.ssl_protocols);
  if (form.ssl_ciphers_mode === 'default') form.ssl_ciphers = tcpsDefaultCiphers();
  preloadCertsFor(t);
}

function toOptions(items: TcpsCertBrief[]) {
  return items.map((c) => ({
    label: `${c.subject_cn || c.cert_id || '未知证书'}（${c.algorithm || '未知算法'}）`,
    value: c.cert_id,
  })).filter(opt => !!opt.value);
}

async function loadIntlCerts() {
  intlCertLoading.value = true;
  try {
    const res = await tcpsApi.listAvailableCerts({ purpose: 'sign', protocol: 'tls' });
    const items = res.data?.items || [];
    console.log('[TCPS 详情] 国际证书返回:', items);
    intlCertOptions.value = toOptions(items);
  } catch (e: any) {
    message.error(e.message || '加载国际签名证书失败');
  } finally {
    intlCertLoading.value = false;
  }
}

async function loadSignCerts() {
  signCertLoading.value = true;
  try {
    const res = await tcpsApi.listAvailableCerts({ purpose: 'sign', protocol: 'gmssl' });
    const items = res.data?.items || [];
    signCertOptions.value = toOptions(items);
  } catch (e: any) {
    message.error(e.message || '加载国密签名证书失败');
  } finally {
    signCertLoading.value = false;
  }
}

async function loadEncCerts() {
  encCertLoading.value = true;
  try {
    const res = await tcpsApi.listAvailableCerts({ purpose: 'enc', protocol: 'gmssl' });
    const items = res.data?.items || [];
    encCertOptions.value = toOptions(items);
  } catch (e: any) {
    message.error(e.message || '加载国密加密证书失败');
  } finally {
    encCertLoading.value = false;
  }
}

function preloadCertsFor(t: TcpsCertTypeOption) {
  if (t === 'intl') loadIntlCerts();
  else if (t === 'gmssl') { loadSignCerts(); loadEncCerts(); }
  else { loadIntlCerts(); loadSignCerts(); loadEncCerts(); }
}

watch(configureCertUI, (val) => {
  if (val) preloadCertsFor(certTypeUI.value);
});

function listenSslOf(d: TcpsProxy): boolean { return d.listen_ssl !== false; }
function certTypeLabel(d: TcpsProxy): string {
  const t = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));
  return TCPS_CERT_TYPE_OPTIONS.find((o) => o.value === t)?.label || t;
}
function hasAnyCert(d: TcpsProxy): boolean {
  return !!(d.intl_cert_id || d.intl_client_cert_id || d.cert_id || d.client_cert_id || d.enc_cert_id || d.client_enc_cert_id);
}
function showCertInDetail(d: TcpsProxy): boolean {
  if (d.mode === 'server') return listenSslOf(d) && hasAnyCert(d);
  return hasAnyCert(d);
}
function showIntlInDetail(d: TcpsProxy): boolean {
  const t = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));
  return t === 'intl' || t === 'intl_gmssl';
}
function showSignInDetail(d: TcpsProxy): boolean {
  const t = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));
  return t === 'gmssl' || t === 'intl_gmssl';
}
function showEncInDetail(d: TcpsProxy): boolean { return showSignInDetail(d); }
function pickIntlCert(d: TcpsProxy): string {
  const t = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));
  if (t === 'intl') return d.mode === 'client' ? d.client_cert_id || '' : d.cert_id || '';
  return d.mode === 'client' ? d.intl_client_cert_id || '' : d.intl_cert_id || '';
}
function pickSignCert(d: TcpsProxy): string {
  const t = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));
  if (t === 'intl') return '';
  return d.mode === 'client' ? d.client_cert_id || '' : d.cert_id || '';
}
function pickEncCert(d: TcpsProxy): string {
  const t = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));
  if (t === 'intl') return '';
  return d.mode === 'client' ? d.client_enc_cert_id || '' : d.enc_cert_id || '';
}
function formatProtocolsForDisplay(raw: string | undefined): string {
  const list = tcpsProtocolsToUI(raw);
  if (list.length === 0) return '—';
  return list.map((v) => (v === TCPS_GM_SSL_PROTOCOL_VALUE ? 'GMTLSv1.1' : v)).join(' ');
}

async function fetchDetail(serviceId: string) {
  loading.value = true;
  try {
    const res = await tcpsApi.get(serviceId);
    detail.value = res.data;
  } catch (e: any) {
    message.error(e.message || '加载失败');
    emit('update:show', false);
  } finally {
    loading.value = false;
  }
}

function syncFormFromDetail() {
  if (!detail.value) return;
  const d = detail.value;
  Object.assign(form, {
    service_name: d.service_name, mode: d.mode, protocol: d.protocol,
    listen_ip: d.listen_ip, listen_port: d.listen_port,
    backend_ip: d.backend_ip, backend_port: d.backend_port,
    backend_tls: d.backend_tls ?? false, ssl_server_name: d.ssl_server_name,
    listen_ssl: d.listen_ssl !== false, ca_cert_id: d.ca_cert_id || '',
    ca_domain: d.ca_domain || '', client_key_ref: d.client_key_ref,
    ssl_protocols: d.ssl_protocols || tcpsDefaultProtocolsFor(d.protocol),
    ssl_ciphers: d.ssl_ciphers || tcpsDefaultCiphers(),
    ssl_conf_command: d.ssl_conf_command || '',
    ssl_ciphers_mode: d.ssl_ciphers_mode || 'default',
    ssl_session_cache_enabled: d.ssl_session_cache_enabled ?? (d.mode === 'server' ? false : undefined),
    ssl_verify_client: d.ssl_verify_client ?? (d.mode === 'server' ? true : undefined),
    ssl_session_tickets: d.ssl_session_tickets ?? (d.mode === 'server' ? false : undefined),
  });

  sslProtocolsUI.value = tcpsProtocolsToUI(form.ssl_protocols);
  clearAllCertSlots();

  certTypeUI.value = inferTcpsCertType(d.protocol, !!(d.intl_cert_id || d.intl_client_cert_id));

  const isClient = d.mode === 'client';
  const backCert = isClient ? d.client_cert_id || '' : d.cert_id || '';
  const backEnc = isClient ? d.client_enc_cert_id || '' : d.enc_cert_id || '';
  const backIntl = isClient ? d.intl_client_cert_id || '' : d.intl_cert_id || '';

  if (certTypeUI.value === 'intl') form.intl_cert_id = backCert;
  else if (certTypeUI.value === 'gmssl') { form.cert_id = backCert; form.enc_cert_id = backEnc; }
  else { form.intl_cert_id = backIntl; form.cert_id = backCert; form.enc_cert_id = backEnc; }

  if (!isClient && d.listen_ssl !== false) configureCertUI.value = true;
  else configureCertUI.value = hasAnyCert(d);
}

watch(
  () => [props.show, props.serviceId] as const,
  async ([show, sid]) => {
    if (!show) {
      isEditing.value = false;
      activeSection.value = 'basic';
      detail.value = null;
      certTypeUI.value = 'intl';
      configureCertUI.value = false;
      sslProtocolsUI.value = [];
      clearAllCertSlots();
      return;
    }
    isEditing.value = false;
    activeSection.value = 'basic';
    
    intlCertOptions.value = [];
    signCertOptions.value = [];
    encCertOptions.value = [];

    if (sid) {
      await Promise.all([fetchDetail(sid), loadCertMap(), loadDomains()]);
      syncFormFromDetail();
    }
    await nextTick();
    formRef.value?.restoreValidation();
    backendFormRef.value?.restoreValidation();
    sslFormRef.value?.restoreValidation();
  },
  { immediate: true }
);

function handleUpdateShow(val: boolean) { emit('update:show', val); }

function enterEdit() {
  if (!detail.value) return;
  syncFormFromDetail();
  isEditing.value = true;
  if (showCertFields.value) preloadCertsFor(certTypeUI.value);
  loadDomains();
  nextTick(() => {
    formRef.value?.restoreValidation();
    backendFormRef.value?.restoreValidation();
    sslFormRef.value?.restoreValidation();
  });
}

function cancelEdit() {
  isEditing.value = false;
  if (detail.value) syncFormFromDetail();
}

function validateCertSelection(): string | null {
  if (!configureCertUI.value) return null;
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

function buildPayload(): TcpsCreateReq {
  const protocol = tcpsCertTypeToProtocol(certTypeUI.value);
  const payload: TcpsCreateReq = { ...form, protocol, service_name: detail.value?.service_name || form.service_name };
  if (payload.mode === 'client') payload.listen_ssl = undefined;

  payload.cert_id = ''; payload.enc_cert_id = ''; payload.intl_cert_id = '';
  payload.client_cert_id = ''; payload.client_enc_cert_id = ''; payload.intl_client_cert_id = '';

  if (configureCertUI.value) {
    const isClient = payload.mode === 'client';
    if (certTypeUI.value === 'intl') {
      if (isClient) payload.client_cert_id = form.intl_cert_id || '';
      else payload.cert_id = form.intl_cert_id || '';
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

  payload.ssl_protocols = tcpsProtocolsFromUI(sslProtocolsUI.value, protocol);
  if (payload.mode === 'client') {
    payload.ssl_session_cache_enabled = undefined;
    payload.ssl_verify_client = undefined;
    payload.ssl_session_tickets = undefined;
  }
  if (payload.ssl_ciphers_mode !== 'default' && !(payload.ssl_ciphers || '').trim()) {
    payload.ssl_ciphers_mode = 'default';
    payload.ssl_ciphers = tcpsDefaultCiphers();
  }
  return payload;
}

async function handleSave() {
  try {
    if (activeSection.value === 'basic') await formRef.value?.validate();
    else if (activeSection.value === 'backend') await backendFormRef.value?.validate();
    else await sslFormRef.value?.validate();
  } catch { return; }

  if (!props.serviceId) return;
  const certErr = validateCertSelection();
  if (certErr) { message.error(certErr); return; }

  const wasRunning = !!detail.value && detail.value.status === 'RUNNING';
  const payload = buildPayload();

  submitting.value = true;
  try {
    await tcpsApi.update(props.serviceId, payload);
    isEditing.value = false;
    await Promise.all([fetchDetail(props.serviceId), loadCertMap(true)]);
    syncFormFromDetail();
    emit('saved');
    if (wasRunning) message.warning('配置已修改，请重启服务使其生效');
    else message.success('保存成功');
  } catch (e: any) {
    message.error(e.message || '保存失败');
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.drawer-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; width: 100%; padding-right: 28px; }
.drawer-header__actions { flex-shrink: 0; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.form-hint { color: #6b7280; font-size: 12px; }
.form-hint-block { display: block; color: #9ca3af; font-size: 12px; margin-top: 4px; }
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
