<template>
  <div style="display:flex; justify-content:center; align-items:center; height:100vh; background:#f0f2f5;">
    <n-card style="width:400px;">
      <h2 style="text-align:center;margin-bottom:24px;">登录</h2>
      <n-form :model="form">
        <n-form-item label="用户名">
          <n-input v-model:value="form.username" placeholder="请输入用户名" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input v-model:value="form.password" type="password" placeholder="请输入密码" />
        </n-form-item>
        <n-button type="primary" block size="large" @click="handleLogin">
          登 录
        </n-button>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/store/user';
import { login } from '@/api/auth';
import { message } from '@/utils/naive';

const router = useRouter();
const userStore = useUserStore();

const form = reactive({
  username: 'admin',
  password: '123456'
});

async function handleLogin() {
  console.log('🔥 点击登录按钮');
  console.log('🔄 登录前 userStore.token:', userStore.token);
  try {
    const res = await login(form);
    console.log('✅ 登录成功，响应:', res);
    userStore.setToken(res.data.access_token);
    console.log('🔄 登录后 userStore.token:', userStore.token);
    userStore.setUser(res.data.user);
    console.log('🔄 登录后 userStore.user:', userStore.user);
    message.success('登录成功');
    const redirect = router.currentRoute.value.query.redirect as string || '/dashboard';
    console.log('🔀 跳转到:', redirect);
    await router.push(redirect);
  } catch (error: any) {
    console.error('❌ 登录失败:', error);
    message.error(error.message || '登录失败');
  }
}

</script>
