<template>
  <div>
    <h1>License 管理</h1>
    <n-space vertical size="large">
      <!-- 操作按钮 -->
      <n-space>
        <n-button type="primary" v-permission="'license:license:create'" @click="showImportModal = true">
          导入 License
        </n-button>
        <n-button type="info" v-permission="'license:license:create'" @click="showActivateModal = true">
          在线激活
        </n-button>
        <n-button @click="fetchStatus">刷新</n-button>
      </n-space>
      <!-- License 状态卡片 -->
      <n-spin :show="loading" description="加载中...">
        <n-card v-if="licenseStatus" :title="'License 状态 (' + licenseStatus.license_type + ')'" hoverable>
          <n-descriptions :column="2" bordered>
            <n-descriptions-item label="有效性">
              <n-tag :type="licenseStatus.is_valid ? 'success' : 'error'">
                {{ licenseStatus.is_valid ? '有效' : '无效' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="类型">
              {{ licenseStatus.license_type }}
            </n-descriptions-item>
            <n-descriptions-item label="过期时间">
              {{ licenseStatus.expires_at ? new Date(licenseStatus.expires_at).toLocaleString() : '无限制' }}
            </n-descriptions-item>
            <n-descriptions-item label="剩余天数">
              <span :style="{ color: (licenseStatus.days_remaining !== null && licenseStatus.days_remaining <= 30) ? 'red' : 'inherit' }">
                {{ licenseStatus.days_remaining !== null ? licenseStatus.days_remaining + ' 天' : '-' }}
              </span>
            </n-descriptions-item>
            <n-descriptions-item label="最大用户数">
              {{ licenseStatus.max_users ?? '无限制' }}
            </n-descriptions-item>
            <n-descriptions-item label="当前用户数">
              {{ licenseStatus.current_users ?? '-' }}
            </n-descriptions-item>
            <n-descriptions-item label="授权模块" :span="2">
              <n-space wrap>
                <n-tag v-for="mod in licenseStatus.authorized_modules" :key="mod" type="success">
                  {{ mod }}
                </n-tag>
                <span v-if="!licenseStatus.authorized_modules.length">无</span>
              </n-space>
            </n-descriptions-item>
            <n-descriptions-item label="是否过期" :span="2">
              <n-tag :type="licenseStatus.is_expired ? 'error' : 'success'">
                {{ licenseStatus.is_expired ? '已过期' : '未过期' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="即将过期" :span="2">
              <n-tag :type="licenseStatus.is_expiring_soon ? 'warning' : 'default'">
                {{ licenseStatus.is_expiring_soon ? '是（剩余≤30天）' : '否' }}
              </n-tag>
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
        <n-empty v-else description="暂无 License 信息" />
      </n-spin>
    </n-space>
    <!-- 导入弹窗 -->
    <ImportModal v-model:visible="showImportModal" @success="fetchStatus" />
    <!-- 激活弹窗 -->
    <ActivateModal v-model:visible="showActivateModal" @success="fetchStatus" />
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { NSpace, NButton, NCard, NDescriptions, NDescriptionsItem, NTag, NEmpty, NSpin, useMessage } from 'naive-ui';
import { licenseApi } from '@/api/license';
import ImportModal from './components/ImportModal.vue';
import ActivateModal from './components/ActivateModal.vue';
import type { LicenseStatus } from '@/types';
const message = useMessage();
const licenseStatus = ref<LicenseStatus | null>(null);
const loading = ref(false);
const showImportModal = ref(false);
const showActivateModal = ref(false);
async function fetchStatus() {
  loading.value = true;
  try {
    const res = await licenseApi.getStatus();
    licenseStatus.value = res.data;
  } catch (error: any) {
    message.error(error.message || '获取License状态失败');
  } finally {
    loading.value = false;
  }
}
onMounted(fetchStatus);
</script>
