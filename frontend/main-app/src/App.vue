<template>
  <router-view />
</template>

<script setup lang="ts">
// 根组件仅作为路由出口，无需额外逻辑
</script>

<style>
/* 全局基础样式，保持 #app 占满视口 */
#app {
  height: 100vh;
  position: relative;
}
</style>