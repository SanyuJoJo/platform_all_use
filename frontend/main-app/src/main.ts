import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import { createPinia } from 'pinia';
import naive from 'naive-ui';
import { setupPermission } from './composables/usePermission';
import { useUserStore } from './store/user';
import { useModuleStore } from './store/module';
import { registerModules } from './micro-frontend/registry';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(naive);
setupPermission(app);

// 等待路由准备就绪后，检查用户是否已登录（刷新页面场景）
router.isReady().then(async () => {
  const userStore = useUserStore();
  // 如果用户已登录（token 存在），则加载模块并注册子应用
  if (userStore.token) {
    try {
      const moduleStore = useModuleStore();
      const modules = await moduleStore.fetchModules();
      registerModules(modules, router);
    } catch (error) {
      console.error('[Main] 登录状态下加载子应用失败:', error);
      // 可选：显示提示，但不阻塞主流程
    }
  }
});

app.mount('#app');
