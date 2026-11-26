<template>
  <t-dialog
    :visible="show"
    :header="title"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="700"
    @close="emit('update:show', false)"
  >
    <t-alert
      v-if="result && result.failed > 0"
      theme="warning"
      :message="`共 ${result.total} 个，成功 ${result.succeeded} 个，失败 ${result.failed} 个。可点击下方「重试失败项」再次执行。`"
      style="margin-bottom: 12px"
    />

    <t-table
      row-key="service_id"
      :columns="columns"
      :data="result?.results || []"
      :pagination="false"
      size="small"
      :scroll="{ y: 300 }"
    />

    <template #footer>
      <t-space style="display: flex; justify-content: flex-end">
        <t-button @click="emit('update:show', false)">关闭</t-button>
        <t-button
          v-if="result && result.failed > 0"
          theme="warning"
          @click="emit('retry-failed')"
        >
          重试失败项
        </t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { h } from 'vue';
import { Tag } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { errorMessage } from '@/constants/ssl-proxy-errorCodes';
import type { BatchItemResult, BatchResp } from '@/types';

defineProps<{
  show: boolean;
  title: string;
  result: BatchResp | null;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'retry-failed'): void;
}>();

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'service_id', title: 'service_id', ellipsis: true },
  {
    colKey: 'success',
    title: '结果',
    width: 90,
    cell: (_h, { row }) =>
      h(Tag, { theme: row.success ? 'success' : 'danger', variant: 'light' }, () =>
        row.success ? '成功' : '失败',
      ),
  },
  {
    colKey: 'code',
    title: '错误码',
    width: 100,
    cell: (_h, { row }) => (row.success ? '—' : row.code),
  },
  {
    colKey: 'message',
    title: '消息',
    ellipsis: true,
    cell: (_h, { row }) => (row.success ? row.message : errorMessage(row.code, row.message)),
  },
];
</script>