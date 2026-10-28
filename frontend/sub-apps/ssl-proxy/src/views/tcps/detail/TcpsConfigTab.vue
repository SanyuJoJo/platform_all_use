<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 8px">
      <span class="config-tab__title">{{ title }}</span>
      <n-space>
        <n-button size="small" @click="handleCopy">复制</n-button>
      </n-space>
    </n-space>
    <n-spin :show="loading">
      <n-code
        v-if="content"
        :code="content"
        language="nginx"
        show-line-numbers
        word-wrap
      />
      <n-empty v-else description="暂无内容" />
    </n-spin>
  </div>
</template>

<script setup lang="ts">
import { NButton, NCode, NEmpty, NSpace, NSpin, useMessage } from 'naive-ui';

const props = defineProps<{
  title: string;
  content: string;
  loading?: boolean;
}>();

const message = useMessage();

async function handleCopy() {
  if (!props.content) {
    message.warning('没有可复制的内容');
    return;
  }
  try {
    await navigator.clipboard.writeText(props.content);
    message.success('已复制到剪贴板');
  } catch {
    message.error('复制失败，请手动选择复制');
  }
}
</script>

<style scoped>
.config-tab__title {
  font-weight: 600;
  color: #374151;
}
</style>
