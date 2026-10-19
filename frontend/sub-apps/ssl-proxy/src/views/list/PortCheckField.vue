<template>
  <div class="port-check-field">
    <n-input-number
      :value="port"
      :min="1"
      :max="65535"
      style="width: 180px"
      @update:value="(v: number | null) => emit('update:port', v ?? 0)"
    />
    <n-button
      :loading="checking"
      :disabled="!ip || !port"
      @click="handleCheck"
    >
      检测端口
    </n-button>

    <n-tag
      v-if="result"
      :type="result.available ? 'success' : 'error'"
      size="small"
      round
      style="margin-left: 4px"
    >
      {{ result.available ? '可用' : '不可用' }}
    </n-tag>

    <span v-if="result && !result.available" class="port-check-field__msg">
      {{ result.message }}
      <n-button
        v-if="result.suggested_port"
        text
        type="primary"
        size="tiny"
        @click="applySuggested"
      >
        使用建议端口 {{ result.suggested_port }}
      </n-button>
    </span>
    <span v-else-if="result" class="port-check-field__msg">
      {{ result.message }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { NButton, NInputNumber, NTag, useMessage } from 'naive-ui';
import proxyApi from '@/api/proxy';
import type { PortCheckResult } from '@/types';

const props = defineProps<{
  ip: string;
  port: number;
  serviceId?: string;
}>();

const emit = defineEmits<{
  (e: 'update:ip', value: string): void;
  (e: 'update:port', value: number): void;
}>();

const message = useMessage();
const checking = ref(false);
const result = ref<PortCheckResult | null>(null);

// ip/port 变化后清除旧结果，避免误导
watch(
  () => [props.ip, props.port],
  () => {
    result.value = null;
  }
);

async function handleCheck() {
  if (!props.ip || !props.port) {
    message.warning('请先填写监听 IP 与端口');
    return;
  }
  checking.value = true;
  try {
    const res = await proxyApi.checkPort({
      service_id: props.serviceId,
      listen_ip: props.ip,
      listen_port: props.port,
    });
    result.value = res.data;
  } catch (e: any) {
    message.error(e.message || '检测失败');
  } finally {
    checking.value = false;
  }
}

function applySuggested() {
  if (result.value?.suggested_port) {
    emit('update:port', result.value.suggested_port);
  }
}
</script>

<style scoped>
.port-check-field {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  width: 100%;
}
.port-check-field__msg {
  color: #6b7280;
  font-size: 12px;
  margin-left: 4px;
}
</style>
