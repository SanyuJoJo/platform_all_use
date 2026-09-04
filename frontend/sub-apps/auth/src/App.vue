<template>
  <div class="subapp-root">
    <router-view />
  </div>
</template>

<script setup lang="ts">
console.log('[SubApp] App 组件加载');
</script>

<style scoped>
/* 将子应用所有全局样式限制在此类名下 */
.subapp-root {
  /* 覆盖可能继承的样式 */
  color: inherit;
}
</style>