<template>
  <t-card :title="'License 管理'" :bordered="false">
    <t-space direction="vertical" size="large" style="width: 100%">
      <t-space>
        <t-button theme="primary" v-permission="'license:license:create'" @click="showImportModal = true">
          导入 License
        </t-button>
        <t-button v-permission="'license:license:create'" @click="showActivateModal = true">
          在线激活
        </t-button>
        <t-button @click="fetchStatus">刷新</t-button>
      </t-space>

      <t-loading :loading="loading" :text="'加载中...'">
        <t-card v-if="licenseStatus" :title="'License 状态 (' + licenseStatus.license_type + ')'" :bordered="true" hoverable>
          <t-descriptions :column="2" bordered>
            <t-descriptions-item label="有效性">
              <t-tag :theme="licenseStatus.is_valid ? 'success' : 'danger'" variant="light">
                {{ licenseStatus.is_valid ? '有效' : '无效' }}
              </t-tag>
            </t-descriptions-item>
            <t-descriptions-item label="类型">{{ licenseStatus.license_type }}</t-descriptions-item>
            <t-descriptions-item label="过期时间">
              {{ licenseStatus.expires_at ? new Date(licenseStatus.expires_at).toLocaleString() : '无限制' }}
            </t-descriptions-item>
            <t-descriptions-item label="剩余天数">
              <span :style="{ color: licenseStatus.days_remaining !== null && licenseStatus.days_remaining <= 30 ? 'red' : 'inherit' }">
                {{ licenseStatus.days_remaining !== null ? licenseStatus.days_remaining + ' 天' : '-' }}
              </span>
            </t-descriptions-item>
            <t-descriptions-item label="最大用户数">{{ licenseStatus.max_users ?? '无限制' }}</t-descriptions-item>
            <t-descriptions-item label="当前用户数">{{ licenseStatus.current_users ?? '-' }}</t-descriptions-item>
            <t-descriptions-item label="授权模块" :span="2">
              <t-space wrap>
                <t-tag v-for="mod in licenseStatus.authorized_modules" :key="mod" theme="success" variant="light">
                  {{ mod }}
                </t-tag>
                <span v-if="!licenseStatus.authorized_modules.length">无</span>
              </t-space>
            </t-descriptions-item>
            <t-descriptions-item label="是否过期" :span="2">
              <t-tag :theme="licenseStatus.is_expired ? 'danger' : 'success'" variant="light">
                {{ licenseStatus.is_expired ? '已过期' : '未过期' }}
              </t-tag>
            </t-descriptions-item>
            <t-descriptions-item label="即将过期" :span="2">
              <t-tag :theme="licenseStatus.is_expiring_soon ? 'warning' : 'default'" variant="light">
                {{ licenseStatus.is_expiring_soon ? '是（剩余≤30天）' : '否' }}
              </t-tag>
            </t-descriptions-item>
          </t-descriptions>
        </t-card>
        <t-empty v-else description="暂无 License 信息" />
      </t-loading>
    </t-space>

    <ImportModal v-model:visible="showImportModal" @success="fetchStatus" />
    <ActivateModal v-model:visible="showActivateModal" @success="fetchStatus" />
  </t-card>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { licenseApi } from '@/api/license';
import ImportModal from './components/ImportModal.vue';
import ActivateModal from './components/ActivateModal.vue';
import type { LicenseStatus } from '@/types';

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
    MessagePlugin.error(error.message || '获取License状态失败');
  } finally {
    loading.value = false;
  }
}
onMounted(fetchStatus);
</script>