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
const columns: DataTableColumns<BatchItemResult> = [
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
