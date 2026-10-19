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
            <n-tag
              v-if="detail"
              :type="STATUS_TAG_TYPE[detail.status] || 'default'"
              size="small"
            >
              {{ detail.status }}
            </n-tag>
          </n-space>

          <n-space class="drawer-header__actions">
            <template v-if="!isEditing">
              <n-button
                v-permission="'ssl_proxy:proxy:update'"
                size="small"
                type="primary"
                @click="enterEdit"
              >
                编辑
              </n-button>
            </template>
            <template v-else>
              <n-button size="small" @click="cancelEdit">取消</n-button>
              <n-button
                size="small"
                type="primary"
                :loading="submitting"
                @click="handleSave"
              >
                保存
              </n-button>
            </template>
          </n-space>
        </div>
      </template>

      <n-spin :show="loading">
        <n-space style="margin-bottom: 16px">
          <n-button
            :type="activeSection === 'basic' ? 'primary' : 'default'"
            size="small"
            @click="activeSection = 'basic'"
          >
            基础配置
          </n-button>
          <n-button
            :type="activeSection === 'backend' ? 'primary' : 'default'"
            size="small"
            @click="activeSection = 'backend'"
          >
            后端代理配置
          </n-button>
          <n-button
            :type="activeSection === 'ssl' ? 'primary' : 'default'"
            size="small"
            @click="activeSection = 'ssl'"
          >
            SSL配置
          </n-button>
        </n-space>

        <!-- ============ 基础配置 ============ -->
        <div v-show="activeSection === 'basic'">
          <n-descriptions
            v-if="!isEditing && detail"
            bordered
            :column="1"
            label-placement="left"
            size="small"
          >
            <n-descriptions-item label="配置名称">
              {{ detail.service_name }}
            </n-descriptions-item>
            <n-descriptions-item label="本地监听地址">
              <span class="mono">{{ detail.listen_ip }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="本地端口">
              {{ detail.listen_port }}
            </n-descriptions-item>
            <n-descriptions-item
              v-if="detail.mode === 'server'"
              label="开启SSL通信"
            >
              <n-tag
                :type="listenSslOf(detail) ? 'success' : 'default'"
                size="small"
              >
                {{ listenSslOf(detail) ? '是' : '否' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="后台服务地址">
              <span class="mono">{{ detail.backend_ip }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="后台服务端口">
              {{ detail.backend_port }}
            </n-descriptions-item>
            <n-descriptions-item label="证书类型">
              {{ certTypeLabel(detail) }}
            </n-descriptions-item>
            <n-descriptions-item label="配置证书">
              <n-tag
                :type="hasAnyCert(detail) ? 'success' : 'default'"
                size="small"
              >
                {{ hasAnyCert(detail) ? '是' : '否' }}
              </n-tag>
            </n-descriptions-item>

            <template v-if="showCertInDetail(detail)">
              <n-descriptions-item
                v-if="showIntlInDetail(detail)"
                label="国际签名证书"
              >
                {{ certName(pickIntlCert(detail)) || '—' }}
              </n-descriptions-item>
              <n-descriptions-item
                v-if="showSignInDetail(detail)"
                label="国密签名证书"
              >
                {{ certName(pickSignCert(detail)) || '—' }}
              </n-descriptions-item>
              <n-descriptions-item
                v-if="showEncInDetail(detail)"
                label="国密加密证书"
              >
                {{ certName(pickEncCert(detail)) || '—' }}
              </n-descriptions-item>
            </template>

            <n-descriptions-item label="信任域(验证服务端证书)">
              <n-tag v-if="detail.ca_domain" size="small" type="info">
                {{ detail.ca_domain }}
              </n-tag>
              <span v-else>—</span>
            </n-descriptions-item>

            <n-descriptions-item label="代理模式">
              {{ detail.mode === 'client' ? '客户端' : '服务端' }}
            </n-descriptions-item>
            <n-descriptions-item label="service_id">
              <span class="mono">{{ detail.service_id }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="PID">
              {{ detail.pid > 0 ? detail.pid : '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="创建时间">
              {{ detail.created_at }}
            </n-descriptions-item>
            <n-descriptions-item label="更新时间">
              {{ detail.updated_at }}
            </n-descriptions-item>
          </n-descriptions>

          <n-form
            v-else
            ref="formRef"
            :model="form"
            :rules="rules"
            label-placement="left"
            label-width="160"
          >
            <n-form-item label="配置名称" path="service_name">
              <n-input
                v-model:value="form.service_name"
                placeholder="请填写配置名称"
                maxlength="64"
                disabled
              />
              <span class="form-hint-block">配置名称创建后不可修改</span>
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

            <n-form-item v-if="form.mode === 'server'" label="开启SSL通信">
              <n-space align="center">
                <n-switch v-model:value="listenSslUI">
                  <template #checked>是</template>
                  <template #unchecked>否</template>
                </n-switch>
                <span class="form-hint">
                  {{
                    listenSslUI
                      ? '服务端监听使用 SSL 加密，必须配置证书'
                      : '服务端监听使用明文 HTTP'
                  }}
                </span>
              </n-space>
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

            <n-form-item label="证书类型">
              <n-select
                v-model:value="certTypeUI"
                :options="CERT_TYPE_OPTIONS"
                placeholder="请选择证书类型"
                @update:value="handleCertTypeChange"
              />
            </n-form-item>

            <n-form-item label="配置证书">
              <n-space align="center">
                <n-switch
                  v-model:value="configureCertUI"
                  :disabled="certSwitchDisabled"
                >
                  <template #checked>是</template>
                  <template #unchecked>否</template>
                </n-switch>
                <span class="form-hint">{{ certSwitchHint }}</span>
              </n-space>
            </n-form-item>

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

            <n-form-item label="信任域(验证服务端证书)">
              <n-select
                v-model:value="form.ca_domain"
                :options="domainOptions"
                :loading="domainLoading"
                placeholder="请选择 CA 信任域，留空表示不校验"
                clearable
                filterable
              />
            </n-form-item>
          </n-form>
        </div>

        <!-- ============ 后端代理配置 ============ -->
        <div v-show="activeSection === 'backend'">
          <n-descriptions
            v-if="!isEditing && detail"
            bordered
            :column="1"
            label-placement="left"
            size="small"
          >
            <n-descriptions-item label="开启 SSL 通信">
              <n-tag
                :type="detail.backend_scheme === 'https' ? 'success' : 'default'"
                size="small"
              >
                {{ detail.backend_scheme === 'https' ? '是' : '否' }}
              </n-tag>
            </n-descriptions-item>
          </n-descriptions>

          <n-form
            v-else
            ref="backendFormRef"
            :model="form"
            label-placement="left"
            label-width="140"
          >
            <n-form-item label="开启 SSL 通信">
              <n-space align="center">
                <n-switch v-model:value="backendSslUI">
                  <template #checked>是</template>
                  <template #unchecked>否</template>
                </n-switch>
                <span class="form-hint">
                  {{
                    backendSslUI
                      ? '使用 HTTPS 访问后台服务'
                      : '使用 HTTP 访问后台服务'
                  }}
                </span>
              </n-space>
            </n-form-item>
          </n-form>
        </div>

        <!-- ============ SSL 配置 ============ -->
        <div v-show="activeSection === 'ssl'">
          <n-descriptions
            v-if="!isEditing && detail"
            bordered
            :column="1"
            label-placement="left"
            size="small"
          >
            <n-descriptions-item label="SSL协议">
              {{ formatProtocolsForDisplay(detail.ssl_protocols) }}
            </n-descriptions-item>

            <template v-if="detail.mode === 'client'">
              <n-descriptions-item label="SSL会话缓存">
                <n-tag
                  :type="detail.ssl_session_cache_enabled ? 'success' : 'default'"
                  size="small"
                >
                  {{ detail.ssl_session_cache_enabled ? '开启' : '关闭' }}
                </n-tag>
              </n-descriptions-item>
            </template>

            <template v-if="detail.mode === 'server'">
              <n-descriptions-item label="SSL会话缓存">
                <n-tag
                  :type="detail.ssl_session_cache_enabled ? 'success' : 'default'"
                  size="small"
                >
                  {{ detail.ssl_session_cache_enabled ? '开启' : '关闭' }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="SSL双向认证">
                <n-tag
                  :type="detail.ssl_verify_client ? 'success' : 'default'"
                  size="small"
                >
                  {{ detail.ssl_verify_client ? '开启' : '关闭' }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="SSL会话票据">
                <n-tag
                  :type="detail.ssl_session_tickets ? 'success' : 'default'"
                  size="small"
                >
                  {{ detail.ssl_session_tickets ? '开启' : '关闭' }}
                </n-tag>
              </n-descriptions-item>
            </template>

            <n-descriptions-item label="SSL加密套件">
              <div class="mono" style="word-break: break-all">
                {{
                  detail.ssl_ciphers_mode === 'default' || !detail.ssl_ciphers
                    ? defaultCiphersForMode(detail.mode)
                    : detail.ssl_ciphers
                }}
              </div>
            </n-descriptions-item>
          </n-descriptions>

          <n-form
            v-else
            ref="sslFormRef"
            :model="form"
            label-placement="left"
            label-width="140"
          >
            <n-divider title-placement="left" style="margin: 0 0 16px">
              SSL协议配置
            </n-divider>

            <n-form-item label="SSL协议">
              <n-checkbox-group v-model:value="sslProtocolsUI">
                <n-space>
                  <n-checkbox
                    v-for="opt in sslProtocolOptions"
                    :key="opt.value"
                    :value="opt.value"
                    :label="opt.label"
                  />
                </n-space>
              </n-checkbox-group>
            </n-form-item>

            <template v-if="form.mode === 'client'">
              <n-form-item label="SSL会话缓存">
                <n-space align="center">
                  <n-switch v-model:value="clientSessionCacheUI">
                    <template #checked>开启</template>
                    <template #unchecked>关闭</template>
                  </n-switch>
                  <span class="form-hint">
                    {{
                      clientSessionCacheUI
                        ? '启用共享 SSL 会话缓存（shared:SSL:10m）'
                        : '不启用会话缓存'
                    }}
                  </span>
                </n-space>
              </n-form-item>
            </template>

            <template v-if="form.mode === 'server'">
              <n-form-item label="SSL会话缓存">
                <n-space align="center">
                  <n-switch v-model:value="serverSessionCacheUI">
                    <template #checked>开启</template>
                    <template #unchecked>关闭</template>
                  </n-switch>
                  <span class="form-hint">
                    {{
                      serverSessionCacheUI
                        ? '启用共享 SSL 会话缓存（shared:SSL:10m）'
                        : '不启用会话缓存'
                    }}
                  </span>
                </n-space>
              </n-form-item>

              <n-form-item label="SSL双向认证">
                <n-space align="center">
                  <n-switch v-model:value="serverVerifyClientUI">
                    <template #checked>开启</template>
                    <template #unchecked>关闭</template>
                  </n-switch>
                  <span class="form-hint">
                    {{
                      serverVerifyClientUI
                        ? '要求客户端提供证书（需配置 CA 信任域）'
                        : '不校验客户端证书'
                    }}
                  </span>
                </n-space>
              </n-form-item>

              <n-form-item label="SSL会话票据">
                <n-space align="center">
                  <n-switch v-model:value="serverSessionTicketsUI">
                    <template #checked>开启</template>
                    <template #unchecked>关闭</template>
                  </n-switch>
                  <span class="form-hint">
                    {{
                      serverSessionTicketsUI
                        ? '允许会话票据'
                        : '禁止会话票据'
                    }}
                  </span>
                </n-space>
              </n-form-item>
            </template>

            <n-divider title-placement="left" style="margin: 16px 0">
              SSL加密套件
            </n-divider>

            <n-form-item label="SSL加密套件">
              <n-space vertical style="width: 100%">
                <n-radio-group v-model:value="sslCiphersModeUI">
                  <n-space>
                    <n-radio value="default">默认套件</n-radio>
                    <n-radio value="custom">自定义套件</n-radio>
                    <n-radio value="input">输入套件</n-radio>
                  </n-space>
                </n-radio-group>

                <n-input
                  v-if="sslCiphersModeUI === 'default'"
                  :value="defaultCiphersForMode(form.mode)"
                  type="textarea"
                  :autosize="{ minRows: 2, maxRows: 4 }"
                  readonly
                />

                <div
                  v-else-if="sslCiphersModeUI === 'custom'"
                  class="ciphers-group"
                >
                  <div class="ciphers-group__section">
                    <div class="ciphers-group__title">TLS 1.3</div>
                    <n-checkbox-group v-model:value="sslCiphersCustomUI">
                      <n-space>
                        <n-checkbox
                          v-for="c in ciphersGroupTLS13"
                          :key="c.value"
                          :value="c.value"
                          :label="c.label"
                        />
                      </n-space>
                    </n-checkbox-group>
                  </div>

                  <div class="ciphers-group__section">
                    <div class="ciphers-group__title">TLS 1.2</div>
                    <n-checkbox-group v-model:value="sslCiphersCustomUI">
                      <n-space>
                        <n-checkbox
                          v-for="c in ciphersGroupTLS12"
                          :key="c.value"
                          :value="c.value"
                          :label="c.label"
                        />
                      </n-space>
                    </n-checkbox-group>
                  </div>

                  <div class="ciphers-group__section">
                    <div class="ciphers-group__title">国密 NTLS</div>
                    <n-checkbox-group v-model:value="sslCiphersCustomUI">
                      <n-space>
                        <n-checkbox
                          v-for="c in ciphersGroupGM"
                          :key="c.value"
                          :value="c.value"
                          :label="c.label"
                        />
                      </n-space>
                    </n-checkbox-group>
                  </div>

                  <div class="ciphers-group__preview">
                    <span class="form-hint">已选：</span>
                    <span class="mono ciphers-group__preview-text">
                      {{
                        sslCiphersCustomUI.length > 0
                          ? sslCiphersCustomUI.join(':')
                          : '（未选择，保存时使用默认套件）'
                      }}
                    </span>
                  </div>
                </div>

                <n-input
                  v-else
                  v-model:value="sslCiphersInputUI"
                  type="textarea"
                  :autosize="{ minRows: 2, maxRows: 6 }"
                  :placeholder="ciphersPlaceholderFor(form.mode)"
                />
              </n-space>
            </n-form-item>
          </n-form>
        </div>
      </n-spin>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup lang="ts">
import { ref, reactive, watch, nextTick, computed } from 'vue';
import {
  NButton,
  NCheckbox,
  NCheckboxGroup,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NRadio,
  NRadioGroup,
  NSelect,
  NSpace,
  NSpin,
  NSwitch,
  NTag,
  useMessage,
  type FormInst,
  type FormRules,
} from 'naive-ui';
import proxyApi from '@/api/proxy';
import { useCertMap } from '@/composables/useCertMap';
import {
  CERT_TYPE_OPTIONS,
  LISTEN_IP_OPTIONS,
  STATUS_TAG_TYPE,
  SSL_PROTOCOL_OPTIONS,
  CIPHERS_OPTIONS,
  DEFAULT_SSL_CIPHERS,
  GM_SSL_PROTOCOL_VALUE,
  certTypeToProtocol,
  inferCertType,
  defaultProtocolsFor,
  defaultCiphersForMode,
  defaultConfCmdFor,
  ciphersPlaceholderFor,
  protocolsFromUI,
  protocolsToUI,
  type CertTypeOption,
} from '@/constants/proxyOptions';
import type {
  CertBrief,
  PlatformSSLProxy,
  ProxyCreateReq,
  ProxyMode,
  SSLCiphersMode,
} from '@/types';

const props = defineProps<{
  show: boolean;
  serviceId?: string;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'saved'): void;
}>();

const message = useMessage();
const { load: loadCertMap, certName } = useCertMap();

const loading = ref(false);
const submitting = ref(false);
const detail = ref<PlatformSSLProxy | null>(null);
const isEditing = ref(false);
const activeSection = ref<'basic' | 'backend' | 'ssl'>('basic');

const formRef = ref<FormInst | null>(null);
const backendFormRef = ref<FormInst | null>(null);
const sslFormRef = ref<FormInst | null>(null);

const certTypeUI = ref<CertTypeOption>('intl');
const configureCertUI = ref<boolean>(false);

const sslProtocolsUI = ref<string[]>([]);
const sslCiphersModeUI = ref<SSLCiphersMode>('default');
const sslCiphersCustomUI = ref<string[]>([]);
const sslCiphersInputUI = ref<string>('');

const clientSessionCacheUI = ref<boolean>(false);
const serverSessionCacheUI = ref<boolean>(false);
const serverVerifyClientUI = ref<boolean>(true);
const serverSessionTicketsUI = ref<boolean>(false);

const domainOptions = ref<{ label: string; value: string }[]>([]);
const domainLoading = ref(false);

const listenSslUI = computed<boolean>({
  get: () => form.listen_ssl !== false,
  set: (v) => {
    form.listen_ssl = v;
  },
});

const backendSslUI = computed<boolean>({
  get: () => form.backend_scheme === 'https',
  set: (v) => {
    form.backend_scheme = v ? 'https' : 'http';
  },
});

// GMTLSv1.1 仅在国密协议下展示（value 仍为 NTLSv1.1）
const sslProtocolOptions = computed(() => {
  return SSL_PROTOCOL_OPTIONS.filter((opt) => {
    if (opt.value === GM_SSL_PROTOCOL_VALUE) {
      return form.protocol === 'gmssl';
    }
    return true;
  });
});

const ciphersGroupTLS13 = CIPHERS_OPTIONS.filter((c) => c.group === 'tls13');
const ciphersGroupTLS12 = CIPHERS_OPTIONS.filter((c) => c.group === 'tls12');
const ciphersGroupGM = CIPHERS_OPTIONS.filter((c) => c.group === 'gm');

const intlCertOptions = ref<{ label: string; value: string }[]>([]);
const signCertOptions = ref<{ label: string; value: string }[]>([]);
const encCertOptions = ref<{ label: string; value: string }[]>([]);
const intlCertLoading = ref(false);
const signCertLoading = ref(false);
const encCertLoading = ref(false);

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
  return configureCertUI.value;
});

const certRequired = computed<boolean>(() => {
  return configureCertUI.value;
});

const certSwitchDisabled = computed<boolean>(() => {
  return form.mode === 'server' && listenSslUI.value;
});

const certSwitchHint = computed<string>(() => {
  if (form.mode === 'server' && listenSslUI.value) {
    return '开启 SSL 通信时必须配置证书';
  }
  return configureCertUI.value ? '本代理加载证书' : '本代理不加载证书';
});

watch(listenSslUI, (val) => {
  if (form.mode === 'server' && val) {
    configureCertUI.value = true;
  }
});

watch(sslCiphersModeUI, (val) => {
  if (val === 'input') {
    if (!sslCiphersInputUI.value.trim()) {
      sslCiphersInputUI.value = DEFAULT_SSL_CIPHERS;
    }
  }
});

async function loadDomains() {
  if (domainOptions.value.length > 0) return;
  domainLoading.value = true;
  try {
    const res: any = await proxyApi.listCADomains();
    const raw = res.data;
    let domains: string[] = [];
    if (Array.isArray(raw?.domains)) {
      domains = raw.domains;
    } else if (Array.isArray(raw?.data?.domains)) {
      domains = raw.data.domains;
    } else if (Array.isArray(raw)) {
      domains = raw;
    }
    if (!domains.includes('safe')) {
      domains.unshift('safe');
    }
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

// 切换证书类型：按 protocol 更新默认协议
function handleCertTypeChange(t: CertTypeOption) {
  if (t === 'intl') {
    form.cert_id = '';
    form.enc_cert_id = '';
  } else if (t === 'gmssl') {
    form.intl_cert_id = '';
  }
  const protocol = certTypeToProtocol(t);
  const nextProtocols = defaultProtocolsFor(protocol);
  form.ssl_protocols = nextProtocols;
  sslProtocolsUI.value = protocolsToUI(nextProtocols);
  if (sslCiphersModeUI.value === 'default') {
    form.ssl_ciphers = defaultCiphersForMode(form.mode);
  }
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

// ============================================================
// 详情展示辅助
// ============================================================
function listenSslOf(d: PlatformSSLProxy): boolean {
  return d.listen_ssl !== false;
}

function certTypeLabel(d: PlatformSSLProxy): string {
  const t = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );
  return CERT_TYPE_OPTIONS.find((o) => o.value === t)?.label || t;
}

function hasAnyCert(d: PlatformSSLProxy): boolean {
  return !!(
    d.intl_cert_id ||
    d.intl_client_cert_id ||
    d.cert_id ||
    d.client_cert_id ||
    d.enc_cert_id ||
    d.client_enc_cert_id
  );
}

function showCertInDetail(d: PlatformSSLProxy): boolean {
  if (d.mode === 'server') {
    return listenSslOf(d) && hasAnyCert(d);
  }
  return hasAnyCert(d);
}

function showIntlInDetail(d: PlatformSSLProxy): boolean {
  const t = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );
  return t === 'intl' || t === 'intl_gmssl';
}

function showSignInDetail(d: PlatformSSLProxy): boolean {
  const t = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );
  return t === 'gmssl' || t === 'intl_gmssl';
}

function showEncInDetail(d: PlatformSSLProxy): boolean {
  return showSignInDetail(d);
}

function pickIntlCert(d: PlatformSSLProxy): string {
  const t = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );
  if (t === 'intl') {
    return d.mode === 'client' ? d.client_cert_id || '' : d.cert_id || '';
  }
  return d.mode === 'client'
    ? d.intl_client_cert_id || ''
    : d.intl_cert_id || '';
}

function pickSignCert(d: PlatformSSLProxy): string {
  const t = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );
  if (t === 'intl') return '';
  return d.mode === 'client' ? d.client_cert_id || '' : d.cert_id || '';
}

function pickEncCert(d: PlatformSSLProxy): string {
  const t = inferCertType(
    d.protocol,
    !!(d.intl_cert_id || d.intl_client_cert_id)
  );
  if (t === 'intl') return '';
  return d.mode === 'client'
    ? d.client_enc_cert_id || ''
    : d.enc_cert_id || '';
}

// ★ 只读态展示：把 NTLSv1.1 显示为 GMTLSv1.1
function formatProtocolsForDisplay(raw: string | undefined): string {
  const list = protocolsToUI(raw);
  if (list.length === 0) return '—';
  return list
    .map((v) => (v === GM_SSL_PROTOCOL_VALUE ? 'GMTLSv1.1' : v))
    .join(' ');
}

// ============================================================
// 数据拉取
// ============================================================
async function fetchDetail(serviceId: string) {
  loading.value = true;
  try {
    const res = await proxyApi.get(serviceId);
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
    service_name: d.service_name,
    mode: d.mode,
    protocol: d.protocol,
    listen_ip: d.listen_ip,
    listen_port: d.listen_port,
    backend_ip: d.backend_ip,
    backend_port: d.backend_port,
    backend_scheme: d.backend_scheme,
    ssl_server_name: d.ssl_server_name,
    listen_ssl: d.listen_ssl !== false,
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

  // ★ 直接 split，不做任何补全/剔除
  sslProtocolsUI.value = protocolsToUI(form.ssl_protocols);

  sslCiphersModeUI.value = (d.ssl_ciphers_mode as SSLCiphersMode) || 'default';
  if (sslCiphersModeUI.value === 'custom') {
    sslCiphersCustomUI.value = (d.ssl_ciphers || '')
      .split(':')
      .map((s) => s.trim())
      .filter(Boolean);
    sslCiphersInputUI.value = '';
  } else if (sslCiphersModeUI.value === 'input') {
    sslCiphersCustomUI.value = [];
    sslCiphersInputUI.value =
      (d.ssl_ciphers || '').trim() || DEFAULT_SSL_CIPHERS;
  } else {
    sslCiphersCustomUI.value = [];
    sslCiphersInputUI.value = '';
  }

  clientSessionCacheUI.value = !!d.ssl_session_cache_enabled;
  serverSessionCacheUI.value = !!d.ssl_session_cache_enabled;
  serverVerifyClientUI.value = d.ssl_verify_client !== false;
  serverSessionTicketsUI.value = !!d.ssl_session_tickets;

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

  if (!isClient && d.listen_ssl !== false) {
    configureCertUI.value = true;
  } else {
    configureCertUI.value = hasAnyCert(d);
  }
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
      sslCiphersModeUI.value = 'default';
      sslCiphersCustomUI.value = [];
      sslCiphersInputUI.value = '';
      clearAllCertSlots();
      return;
    }
    isEditing.value = false;
    activeSection.value = 'basic';
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

function handleUpdateShow(val: boolean) {
  emit('update:show', val);
}

function enterEdit() {
  if (!detail.value) return;
  syncFormFromDetail();
  isEditing.value = true;
  if (showCertFields.value) {
    preloadCertsFor(certTypeUI.value);
  }
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

// ============================================================
// 保存
// ============================================================
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

function buildPayload(): ProxyCreateReq {
  const protocol = certTypeToProtocol(certTypeUI.value);

  const payload: ProxyCreateReq = {
    ...form,
    protocol,
    service_name: detail.value?.service_name || form.service_name,
  };

  if (payload.mode === 'client') {
    payload.listen_ssl = undefined;
  }

  payload.cert_id = '';
  payload.enc_cert_id = '';
  payload.intl_cert_id = '';
  payload.client_cert_id = '';
  payload.client_enc_cert_id = '';
  payload.intl_client_cert_id = '';

  if (configureCertUI.value) {
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

  // ★ 纯 join，NTLSv1.1 原样提交，后端 sanitizeSSLProtocols 处理
  payload.ssl_protocols = protocolsFromUI(sslProtocolsUI.value, protocol);

  if (sslCiphersModeUI.value === 'default') {
    payload.ssl_ciphers = defaultCiphersForMode(payload.mode);
  } else if (sslCiphersModeUI.value === 'custom') {
    payload.ssl_ciphers = sslCiphersCustomUI.value.join(':');
  } else {
    payload.ssl_ciphers = sslCiphersInputUI.value.trim();
  }
  payload.ssl_ciphers_mode = sslCiphersModeUI.value;

  if (
    payload.ssl_ciphers_mode !== 'default' &&
    !(payload.ssl_ciphers || '').trim()
  ) {
    payload.ssl_ciphers_mode = 'default';
    payload.ssl_ciphers = defaultCiphersForMode(payload.mode);
  }

  if (payload.mode === 'client') {
    payload.ssl_session_cache_enabled = clientSessionCacheUI.value;
    payload.ssl_verify_client = undefined;
    payload.ssl_session_tickets = undefined;
  } else {
    payload.ssl_session_cache_enabled = serverSessionCacheUI.value;
    payload.ssl_verify_client = serverVerifyClientUI.value;
    payload.ssl_session_tickets = serverSessionTicketsUI.value;
  }

  return payload;
}

async function handleSave() {
  try {
    if (activeSection.value === 'basic') {
      await formRef.value?.validate();
    } else if (activeSection.value === 'backend') {
      await backendFormRef.value?.validate();
    } else {
      await sslFormRef.value?.validate();
    }
  } catch {
    return;
  }

  if (!props.serviceId) return;

  const certErr = validateCertSelection();
  if (certErr) {
    message.error(certErr);
    return;
  }

  const wasRunning = !!detail.value && detail.value.status === 'RUNNING';

  const payload = buildPayload();

  submitting.value = true;
  try {
    await proxyApi.update(props.serviceId, payload);
    isEditing.value = false;
    await Promise.all([fetchDetail(props.serviceId), loadCertMap(true)]);
    syncFormFromDetail();
    emit('saved');

    if (wasRunning) {
      message.warning('配置已修改，请重启服务使其生效');
    } else {
      message.success('保存成功');
    }
  } catch (e: any) {
    message.error(e.message || '保存失败');
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding-right: 28px;
}
.drawer-header__actions {
  flex-shrink: 0;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
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

.ciphers-group {
  width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 12px;
  background: #fafafa;
}
.ciphers-group__section {
  margin-bottom: 12px;
}
.ciphers-group__section:last-of-type {
  margin-bottom: 8px;
}
.ciphers-group__title {
  font-size: 12px;
  color: #374151;
  font-weight: 600;
  margin-bottom: 6px;
}
.ciphers-group__preview {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #e5e7eb;
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
}
.ciphers-group__preview-text {
  color: #374151;
  word-break: break-all;
  line-height: 1.5;
  flex: 1;
}
</style>
