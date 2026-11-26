<template>
  <div class="login-wrap">
    <t-card class="login-card" :bordered="false">
      <h2 class="login-title">多功能管理平台</h2>
      <t-form :data="form" :rules="rules" label-align="top" @submit="handleLogin">
        <t-form-item label="用户名" name="username">
          <t-input v-model="form.username" placeholder="请输入用户名" clearable />
        </t-form-item>
        <t-form-item label="密码" name="password">
          <t-input v-model="form.password" type="password" placeholder="请输入密码" clearable />
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" type="submit" block :loading="loading">登 录</t-button>
        </t-form-item>
      </t-form>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { MessagePlugin, type SubmitContext } from 'tdesign-vue-next'
import { useUserStore } from '@/store/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)
const form = ref({ username: 'admin', password: '' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' as const }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' as const }],
}

async function handleLogin({ validateResult }: SubmitContext) {
  if (validateResult !== true) return
  loading.value = true
  try {
    await userStore.login(form.value.username, form.value.password)
    MessagePlugin.success('登录成功')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.push(redirect)
  } catch (e: any) {
    MessagePlugin.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f5f7fa;
}
.login-card {
  width: 380px;
  padding: 12px 12px 0;
}
.login-title {
  text-align: center;
  color: #0052d9;
  margin-bottom: 24px;
}
</style>
