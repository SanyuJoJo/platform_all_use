<template>
  <div>
    <t-card title="仪表盘" :bordered="false">
      <t-alert theme="success" message="TDesign 单应用架构已就绪，各业务模块已合并。" />
      <t-space style="margin-top: 24px" size="large">
        <t-statistic title="业务模块" :value="7" />
        <t-statistic title="登录用户" :value="userStore.username || '-'" />
      </t-space>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
</script>
