<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 8px">
      <n-space align="center">
        <span class="logs-tab__title">日志</span>
        <n-input-number
          v-model:value="tail"
          :min="10"
          :max="2000"
          size="small"
          style="width: 120px"
        />
        <span class="logs-tab__hint">最近 N 行</span>
      </n-space>
      <n-space>
        <n-button size="small" :loading="loading" @click="fetchLogs">刷新</n-button>
      </n-space>
    </n-space>
    <n-alert
      v-if="logPath"
      type="default"
      :show-icon="false"
      style="margin-bottom: 8px"
    >
      <span class="mono">{{ logPath }}</span>
    </n-alert>
    <n-spin :show="loading">
      <pre v-if="content" class="logs-tab__pre">{{ content }}</pre>
      <n-empty v-else description="暂无日志（实例可能未启动）" />
    </n-spin>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import {
  NAlert,
  NButton,
  NEmpty,
  NInputNumber,
  NSpace,
  NSpin,
  useMessage,
} from 'naive-ui';
import proxyApi from '@/api/proxy';
const props = defineProps<{
  serviceId: string;
}>();
const message = useMessage();
const loading = ref(false);
const tail = ref(200);
const content = ref('');
const logPath = ref('');
async function fetchLogs() {
  loading.value = true;
  try {
    const res = await proxyApi.getLogs(props.serviceId, tail.value);
    content.value = res.data.content || '';
    logPath.value = res.data.path || '';
    if (!content.value && res.data.message) {
      message.info(res.data.message);
    }
  } catch (e: any) {
    message.error(e.message || '加载日志失败');
  } finally {
    loading.value = false;
  }
}
onMounted(fetchLogs);
</script>
<style scoped>
.logs-tab__title {
  font-weight: 600;
  color: #374151;
}
.logs-tab__hint {
  color: #6b7280;
  font-size: 12px;
}
.logs-tab__pre {
  background: #0f172a;
  color: #e2e8f0;
  padding: 12px;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  line-height: 1.6;
  max-height: 600px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}
</style>
