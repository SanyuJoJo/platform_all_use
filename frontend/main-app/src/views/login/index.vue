<template>
  <div style="display:flex; justify-content:center; align-items:center; height:100vh; background:#f0f2f5;">
    <n-card style="width:400px;">
      <h2 style="text-align:center;margin-bottom:24px;">登录</h2>
      <n-form :model="form" ref="formRef">
        <n-form-item label="用户名">
          <n-input v-model:value="form.username" placeholder="请输入用户名" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="请输入密码"
            @keyup.enter="handleLogin"
          />
        </n-form-item>
        <n-button type="primary" block size="large" :loading="loading" @click="handleLogin">
          登 录
        </n-button>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useUserStore } from '@/store/user';
import { useModuleStore } from '@/store/module';
import { useMenuStore } from '@/store/menu';
import { registerModules } from '@/micro-frontend/registry';
import { login } from '@/api/auth';
import { message } from '@/utils/naive';

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const moduleStore = useModuleStore();
const menuStore = useMenuStore();
const loading = ref(false);

const form = reactive({
  username: 'admin',
  password: '123456',
});

async function handleLogin() {
  console.log('🔥 点击登录');
  loading.value = true;
  try {
    const res = await login(form);
    console.log('✅ 登录响应:', res);
    const { access_token, user } = res.data;
    userStore.setToken(access_token);
    userStore.setUser(user);

    // 加载模块列表
    const modules = await moduleStore.fetchModules();
    console.log('📦 模块列表:', modules);
    registerModules(modules);

    // ✅ 传入数据构建菜单
    await menuStore.buildMenus(modules);
    console.log('📋 菜单树:', menuStore.menuTree);

    message.success('登录成功');
    const redirect = route.query.redirect as string || '/dashboard';
    await router.push(redirect);
  } catch (error: any) {
    console.error('❌ 登录失败:', error);
    message.error(error.message || '登录失败，请检查用户名和密码');
  } finally {
    loading.value = false;
  }
}
</script>
