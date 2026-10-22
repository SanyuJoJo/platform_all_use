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
import { onMounted } from 'vue';
import {
  NMessageProvider,
  NDialogProvider,
  NNotificationProvider,
} from 'naive-ui';

onMounted(() => {
  console.warn('[auth] App.vue mounted');
});
</script>
