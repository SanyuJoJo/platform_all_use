<template>
  <div style="display:flex; justify-content:center; align-items:center; height:100vh; background:#f0f2f5;">
    <n-card style="width:400px;">
      <h2 style="text-align:center;margin-bottom:24px;">登录</h2>
      <n-form :model="form" ref="formRef">
        <n-form-item label="用户名">
          <n-input v-model:value="form.username" placeholder="请输入用户名" autocomplete="username" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="请输入密码"
            autocomplete="current-password"
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
import { useRouter, useRoute } from 'vue-router'; // ★ 补回 useRouter
import { useUserStore } from '@/store/user';
import { useModuleStore } from '@/store/module';
import { useMenuStore } from '@/store/menu';
import { usePermissionStore } from '@/store/permission';
import { registerModules } from '@/micro-frontend/registry';
import { login } from '@/api/auth';
import { message } from '@/utils/naive';

const router = useRouter(); // ★ 补回 router 实例
const route = useRoute();
const userStore = useUserStore();
const moduleStore = useModuleStore();
const menuStore = useMenuStore();
const permissionStore = usePermissionStore();
const loading = ref(false);
const formRef = ref();

const form = reactive({
  username: 'admin',
  password: '123456',
});

async function handleLogin() {
  loading.value = true;
  try {
    // 1. 登录（只有这一步失败才算登录失败）
    const res = await login(form);
    const { access_token, user } = res.data;
    userStore.setToken(access_token);
    userStore.setUser(user);

    // 2. 拉取路由树 → 注册主应用动态路由 → 构建菜单
    const tree = await permissionStore.fetchRouters(true);
    permissionStore.registerRoutes(router); // 这里需要 router
    menuStore.buildMenus(tree);

    // 3. 加载模块并注册 qiankun 子应用（失败不阻塞登录）
    try {
      const modules = await moduleStore.fetchModules();
      registerModules(modules, router); // 这里需要 router
    } catch (moduleError) {
      console.warn('[Login] 模块加载失败，子应用暂不可用', moduleError);
    }

    message.success('登录成功');
    const redirect = (route.query.redirect as string) || '/';
    
    // ★ 核心修改：使用 window.location.href 强制刷新页面跳转
    setTimeout(() => {
      window.location.href = redirect;
    }, 300);
    
  } catch (error: any) {
    console.error('[Login] 登录失败', error);
    message.error(error?.message || '登录失败，请检查用户名和密码');
  } finally {
    loading.value = false;
  }
}
</script>
