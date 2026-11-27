<template>
  <router-view />
</template>

<script setup lang="ts">
// 根组件仅作为路由出口
</script>

<style>
:root {
  --td-brand-color: #2563eb;
  --td-brand-color-hover: #1d4ed8;
  --td-brand-color-active: #1e40af;
  --td-brand-color-focus: #3b82f6;
  --td-brand-color-light: #eff6ff;
  --td-brand-color-disabled: #bfdbfe;
  --app-bg: #f4f6fa;
}

* {
  box-sizing: border-box;
}

html,
body,
#app {
  height: 100%;
}

html,
body {
  margin: 0;
  padding: 0;
}

body {
  font-family:
    -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Arial, sans-serif;
  color: #1e293b;
  background: var(--app-bg);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
}

#app {
  min-width: 320px;
}

::selection {
  background: rgba(37, 99, 235, 0.18);
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  border: 2px solid transparent;
  border-radius: 8px;
  background: #cbd5e1;
  background-clip: content-box;
}

::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
  background-clip: content-box;
}
</style>
