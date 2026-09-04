<template>
  <div style="padding: 20px; background: #f0f9ff; border-radius: 8px;">
    <h2>🔐 Auth 子应用 - 仪表盘</h2>
    <p>这是认证授权模块的首页，路径为 <code>/auth/dashboard</code></p>
    <p>当前时间：{{ new Date().toLocaleString() }}</p>
  </div>
</template>
<script setup>
console.log('[SubApp] Dashboard 组件渲染');
</script>
