import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import { createPinia } from 'pinia';
import naive from 'naive-ui';
import { setupPermission } from './composables/usePermission';
import { useUserStore } from './store/user';
import { useModuleStore } from './store/module';
import { useMenuStore } from './store/menu';
import { registerModules } from './micro-frontend/registry';
import { message } from './utils/naive';

console.log('[Main] 应用启动');

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(naive);
setupPermission(app);

// 等待路由就绪，检查登录状态
router.isReady().then(async () => {
  console.log('[Main] 路由准备就绪');
  const userStore = useUserStore();
  if (userStore.token) {
    try {
      const moduleStore = useModuleStore();
      const menuStore = useMenuStore();
      const modules = await moduleStore.fetchModules();
      // 注册子应用（qiankun）
      registerModules(modules, router);
      // 构建菜单
      await menuStore.buildMenus(modules);
      console.log('[Main] 模块注册和菜单构建完成');
    } catch (error) {
      console.error('[Main] 初始化失败:', error);
      message.error('初始化失败，请刷新页面重试');
    }
  } else {
    console.log('[Main] 未登录，跳过模块加载');
  }
});

app.mount('#app');
console.log('[Main] 应用挂载完成');