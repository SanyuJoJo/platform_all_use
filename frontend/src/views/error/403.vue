<template>
  <div class="error-page">
    <h1 class="code">403</h1>
    <p class="desc">抱歉，您没有权限访问该页面</p>
    <t-space>
      <t-button theme="primary" @click="router.push('/')">返回首页</t-button>
      <t-button @click="router.back()">返回上一页</t-button>
    </t-space>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.error-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-top: 120px;
}
.code {
  font-size: 80px;
  color: #d54941;
  margin: 0;
}
.desc {
  color: #666;
  margin: 12px 0 24px;
}
</style>
