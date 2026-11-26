<template>
  <t-dialog
    v-model:visible="visible"
    :header="'日志详情'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="600"
    @close="handleClose"
  >
    <t-loading :loading="loading" :text="'加载中...'">
      <t-descriptions v-if="logData" :column="1" bordered>
        <t-descriptions-item label="ID">{{ logData.id }}</t-descriptions-item>
        <t-descriptions-item label="用户名">{{ logData.username || '-' }}</t-descriptions-item>
        <t-descriptions-item label="模块">{{ logData.module_id }}</t-descriptions-item>
        <t-descriptions-item label="操作类型">{{ logData.action }}</t-descriptions-item>
        <t-descriptions-item label="资源">{{ logData.resource || '-' }}</t-descriptions-item>
        <t-descriptions-item label="资源ID">{{ logData.resource_id || '-' }}</t-descriptions-item>
        <t-descriptions-item label="详情">{{ logData.detail || '-' }}</t-descriptions-item>
        <t-descriptions-item label="IP">{{ logData.ip || '-' }}</t-descriptions-item>
        <t-descriptions-item label="User-Agent">{{ logData.user_agent || '-' }}</t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag :theme="logData.status === 'success' ? 'success' : 'danger'" variant="light">
            {{ logData.status === 'success' ? '成功' : '失败' }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="错误码">{{ logData.error_code ?? '-' }}</t-descriptions-item>
        <t-descriptions-item label="请求ID">{{ logData.request_id || '-' }}</t-descriptions-item>
        <t-descriptions-item label="操作时间">{{ new Date(logData.created_at).toLocaleString() }}</t-descriptions-item>
      </t-descriptions>
    </t-loading>
    <template #footer>
      <t-button @click="handleClose">关闭</t-button>
    </template>
  </t-dialog>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { auditLogApi } from '@/api/audit-log';
import type { AuditLog } from '@/types';
const props = defineProps<{
  visible: boolean;
  logId: number;
}>();
const emit = defineEmits(['update:visible', 'success']);
const logData = ref<AuditLog | null>(null);
const loading = ref(false);
const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});
watch(
  () => props.visible,
  async (val) => {
    if (val && props.logId) {
      await fetchDetail();
    }
  },
  { immediate: true },
);
async function fetchDetail() {
  if (!props.logId) {
    console.warn('[DetailModal] logId 无效，跳过请求');
    return;
  }
  loading.value = true;
  try {
    const res = await auditLogApi.getDetail(props.logId);
    logData.value = res.data;
  } catch (err: any) {
    MessagePlugin.error(err.message || '获取日志详情失败');
    handleClose();
  } finally {
    loading.value = false;
  }
}
function handleClose() {
  visible.value = false;
  logData.value = null;
}
</script>