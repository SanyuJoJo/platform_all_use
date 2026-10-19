<template>
  <div>
    <n-space vertical size="large">
      <n-card>
        <template #header>
          <n-space align="center">
            <span>代理详情</span>
            <n-tag
              v-if="detail"
              :type="STATUS_TAG_TYPE[detail.status] || 'default'"
              size="small"
            >
              {{ detail.status }}
            </n-tag>
          </n-space>
        </template>
        <template #header-extra>
          <n-space>
            <n-button
              v-if="detail && !isRunning(detail.status)"
              v-permission="'ssl_proxy:proxy:start'"
              type="primary"
              :disabled="isTransitioning(detail.status)"
              @click="handleStart"
            >
              启动
            </n-button>
            <n-button
              v-if="detail && isRunning(detail.status)"
              v-permission="'ssl_proxy:proxy:stop'"
              type="warning"
              @click="handleStop"
            >
              停止
            </n-button>
            <n-button
              v-if="detail && isRunning(detail.status)"
              @click="handleRestart"
            >
              重启
            </n-button>
            <n-button
              v-permission="'ssl_proxy:proxy:update'"
              :disabled="detail ? isRunning(detail.status) : true"
              @click="goEdit"
            >
              编辑
            </n-button>
            <n-button @click="goBack">返回列表</n-button>
          </n-space>
        </template>
        <n-spin :show="loading">
          <n-descriptions
            v-if="detail"
            bordered
            :column="2"
            label-placement="left"
            size="small"
          >
            <n-descriptions-item label="service_id">
              <span class="mono">{{ detail.service_id }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="服务名称">
              {{ detail.service_name }}
            </n-descriptions-item>
            <n-descriptions-item label="模式">
              {{ detail.mode === 'client' ? '客户端（nchttp）' : '服务端（nshttp）' }}
            </n-descriptions-item>
            <n-descriptions-item label="协议">
              {{ PROTOCOL_LABEL[detail.protocol] || detail.protocol }}
            </n-descriptions-item>
            <n-descriptions-item label="监听地址">
              <span class="mono">
                {{ detail.listen_ip }}:{{ detail.listen_port }}
              </span>
            </n-descriptions-item>
            <n-descriptions-item label="后端地址">
              <span class="mono">
                {{ detail.backend_scheme }}://{{ detail.backend_ip }}:{{ detail.backend_port }}
              </span>
            </n-descriptions-item>
            <n-descriptions-item label="SSL Server Name">
              {{ detail.ssl_server_name || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="PID">
              {{ detail.pid > 0 ? detail.pid : '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="服务器证书">
              {{ detail.cert_id || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="CA 证书">
              {{ detail.ca_cert_id || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="客户端证书">
              {{ detail.client_cert_id || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="客户端私钥引用">
              {{ detail.client_key_ref || '—' }}
            </n-descriptions-item>
            <n-descriptions-item label="创建时间">
              {{ detail.created_at }}
            </n-descriptions-item>
            <n-descriptions-item label="更新时间">
              {{ detail.updated_at }}
            </n-descriptions-item>
          </n-descriptions>
        </n-spin>
      </n-card>
      <n-card>
        <n-tabs v-model:value="activeTab" type="line" @update:value="handleTabChange">
          <n-tab-pane name="kv" tab="KV 配置">
            <ConfigTab
              v-if="activeTab === 'kv'"
              title="KV 配置（用户视图）"
              :content="config.kv"
              :loading="configLoading"
            />
          </n-tab-pane>
          <n-tab-pane name="nginx" tab="Nginx 配置">
            <ConfigTab
              v-if="activeTab === 'nginx'"
              title="Nginx 配置（渲染结果）"
              :content="config.nginx"
              :loading="configLoading"
            />
          </n-tab-pane>
          <n-tab-pane name="logs" tab="日志">
            <LogsTab
              v-if="activeTab === 'logs'"
              :service-id="serviceId"
            />
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-space>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  NButton,
  NCard,
  NDescriptions,
  NDescriptionsItem,
  NSpace,
  NSpin,
  NTabPane,
  NTabs,
  NTag,
  useMessage,
} from 'naive-ui';
import proxyApi from '@/api/proxy';
import { useProxyAction } from '@/composables/useProxyAction';
import { PROTOCOL_LABEL, STATUS_TAG_TYPE } from '@/constants/proxyOptions';
import ConfigTab from './ConfigTab.vue';
import LogsTab from './LogsTab.vue';
import type { PlatformSSLProxy } from '@/types';
const route = useRoute();
const router = useRouter();
const message = useMessage();
const { start, stop, restart, isRunning, isTransitioning } = useProxyAction();
const serviceId = computed(() => route.params.id as string);
const loading = ref(false);
const detail = ref<PlatformSSLProxy | null>(null);
const activeTab = ref<'kv' | 'nginx' | 'logs'>('kv');
const configLoading = ref(false);
const config = reactive({ kv: '', nginx: '' });
async function fetchDetail() {
  loading.value = true;
  try {
    const res = await proxyApi.get(serviceId.value);
    detail.value = res.data;
  } catch (e: any) {
    message.error(e.message || '加载失败');
    goBack();
  } finally {
    loading.value = false;
  }
}
async function fetchConfig() {
  configLoading.value = true;
  try {
    const res = await proxyApi.getConfig(serviceId.value);
    config.kv = res.data.kv || '';
    config.nginx = res.data.nginx || '';
  } catch (e: any) {
    message.error(e.message || '加载配置失败');
  } finally {
    configLoading.value = false;
  }
}
function handleTabChange(name: string | number) {
  if (name === 'kv' || name === 'nginx') {
    if (!config.kv && !config.nginx) {
      fetchConfig();
    }
  }
}
async function handleStart() {
  const ok = await start(serviceId.value);
  if (ok) {
    await fetchDetail();
    await fetchConfig();
  }
}
async function handleStop() {
  const ok = await stop(serviceId.value);
  if (ok) {
    await fetchDetail();
  }
}
async function handleRestart() {
  const ok = await restart(serviceId.value);
  if (ok) {
    await fetchDetail();
    await fetchConfig();
  }
}
function goEdit() {
  router.push(`/${serviceId.value}/edit`);
}
function goBack() {
  router.push('/list');
}
onMounted(async () => {
  await fetchDetail();
  await fetchConfig();
});
</script>
<style scoped>
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>
