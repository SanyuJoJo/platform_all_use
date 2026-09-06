<template>
  <n-message-provider>
    <n-dialog-provider>
      <n-notification-provider>
        <router-view />
      </n-notification-provider>
    </n-dialog-provider>
  </n-message-provider>
</template>

<script setup lang="ts">
import { NMessageProvider, NDialogProvider, NNotificationProvider } from 'naive-ui';
console.log('[Auth] App 组件加载（含 Provider）');
</script>
