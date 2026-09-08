<template>
  <n-modal v-model:show="visible" title="日志详情" preset="dialog" style="width: 600px;" @close="handleClose">
    <n-spin :show="loading" description="加载中...">
      <n-descriptions v-if="logData" :column="1" bordered>
        <n-descriptions-item label="ID">{{ logData.id }}</n-descriptions-item>
        <n-descriptions-item label="用户名">{{ logData.username || '-' }}</n-descriptions-item>
        <n-descriptions-item label="模块">{{ logData.module_id }}</n-descriptions-item>
        <n-descriptions-item label="操作类型">{{ logData.action }}</n-descriptions-item>
        <n-descriptions-item label="资源">{{ logData.resource || '-' }}</n-descriptions-item>
        <n-descriptions-item label="资源ID">{{ logData.resource_id || '-' }}</n-descriptions-item>
        <n-descriptions-item label="详情">{{ logData.detail || '-' }}</n-descriptions-item>
        <n-descriptions-item label="IP">{{ logData.ip || '-' }}</n-descriptions-item>
        <n-descriptions-item label="User-Agent">{{ logData.user_agent || '-' }}</n-descriptions-item>
        <n-descriptions-item label="状态">
          <n-tag :type="logData.status === 'success' ? 'success' : 'error'">
            {{ logData.status === 'success' ? '成功' : '失败' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="错误码">{{ logData.error_code ?? '-' }}</n-descriptions-item>
        <n-descriptions-item label="请求ID">{{ logData.request_id || '-' }}</n-descriptions-item>
        <n-descriptions-item label="操作时间">{{ new Date(logData.created_at).toLocaleString() }}</n-descriptions-item>
      </n-descriptions>
    </n-spin>
    <template #action>
      <n-button @click="handleClose">关闭</n-button>
    </template>
  </n-modal>
</template>
<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { NModal, NDescriptions, NDescriptionsItem, NTag, NButton, NSpin, useMessage } from 'naive-ui';
import { auditLogApi } from '@/api/audit-log';
import type { AuditLog } from '@/types';
const props = defineProps<{
  visible: boolean;
  logId: number;
}>();
const emit = defineEmits(['update:visible', 'success']);
const message = useMessage();
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
  { immediate: true }
);
async function fetchDetail() {
  // ✅ 守卫：防止 logId 为 0 或无效时发起请求
  if (!props.logId) {
    console.warn('[DetailModal] logId 无效，跳过请求');
    return;
  }
  loading.value = true;
  try {
    const res = await auditLogApi.getDetail(props.logId);
    logData.value = res.data;
  } catch (err: any) {
    message.error(err.message || '获取日志详情失败');
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