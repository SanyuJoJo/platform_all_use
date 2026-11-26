<template>
  <div class="port-check-field">
    <t-input-number
      :value="port"
      :min="1"
      :max="65535"
      style="width: 180px"
      @change="(v: number) => emit('update:port', v ?? 0)"
    />
    <t-button :loading="checking" :disabled="!ip || !port" @click="handleCheck">
      检测端口
    </t-button>

    <t-tag
      v-if="result"
      :theme="result.available ? 'success' : 'danger'"
      variant="light"
      size="small"
      style="margin-left: 4px"
    >
      {{ result.available ? '可用' : '不可用' }}
    </t-tag>

    <span v-if="result && !result.available" class="port-check-field__msg">
      {{ result.message }}
      <t-button
        v-if="result.suggested_port"
        variant="text"
        theme="primary"
        size="small"
        @click="applySuggested"
      >
        使用建议端口 {{ result.suggested_port }}
      </t-button>
    </span>
    <span v-else-if="result" class="port-check-field__msg">
      {{ result.message }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
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

const checking = ref(false);
const result = ref<PortCheckResult | null>(null);

watch(
  () => [props.ip, props.port],
  () => {
    result.value = null;
  },
);

async function handleCheck() {
  if (!props.ip || !props.port) {
    MessagePlugin.warning('请先填写监听 IP 与端口');
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
    MessagePlugin.error(e.message || '检测失败');
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