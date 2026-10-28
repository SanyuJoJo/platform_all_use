<template>
  <n-modal
    :show="show"
    preset="card"
    :title="title"
    style="width: 640px"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <template v-if="result">
      <n-space vertical size="medium">
        <n-alert
          :type="result.failed === 0 ? 'success' : 'warning'"
          :show-icon="true"
        >
          共 {{ result.total }} 项，成功 {{ result.succeeded }} 项，
          失败 {{ result.failed }} 项。
        </n-alert>
        <n-data-table
          :columns="columns"
          :data="result.results"
          :pagination="false"
          size="small"
          :max-height="360"
          virtual-scroll
        />
      </n-space>
    </template>
    <template #footer>
      <n-space justify="end">
        <n-button
          v-if="result && result.failed > 0"
          type="warning"
          @click="emit('retry-failed')"
        >
          仅重试失败项
        </n-button>
        <n-button type="primary" @click="emit('update:show', false)">
          关闭
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import {
  NAlert,
  NButton,
  NDataTable,
  NModal,
  NSpace,
  NTag,
  type DataTableColumns,
} from 'naive-ui';
import { h } from 'vue';
import { errorMessage } from '@/constants/errorCodes';
import type { TcpsBatchItemResult, TcpsBatchResp } from '@/types/tcps';

defineProps<{
  show: boolean;
  title: string;
  result: TcpsBatchResp | null;
}>();

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void;
  (e: 'retry-failed'): void;
}>();

const columns: DataTableColumns<TcpsBatchItemResult> = [
  { title: 'service_id', key: 'service_id', ellipsis: { tooltip: true } },
  {
    title: '结果',
    key: 'success',
    width: 90,
    render(row) {
      return h(
        NTag,
        { size: 'small', type: row.success ? 'success' : 'error' },
        { default: () => (row.success ? '成功' : '失败') }
      );
    },
  },
  {
    title: '错误码',
    key: 'code',
    width: 100,
    render(row) {
      return row.success ? '—' : row.code;
    },
  },
  {
    title: '消息',
    key: 'message',
    ellipsis: { tooltip: true },
    render(row) {
      return row.success ? row.message : errorMessage(row.code, row.message);
    },
  },
];
</script>
