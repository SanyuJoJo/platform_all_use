import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import { createPinia } from 'pinia';
import naive from 'naive-ui';
import { setupPermission } from './composables/usePermission';
import { useUserStore } from './store/user';
import { useModuleStore } from './store/module';
import { useMenuStore } from './store/menu';
import { initMicroAppLoader, registerModules } from './micro-frontend/registry';
import { message } from './utils/naive';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(naive);
setupPermission(app);

initMicroAppLoader();

router.isReady().then(async () => {
  const userStore = useUserStore();
  if (userStore.token) {
    try {
      const moduleStore = useModuleStore();
      const menuStore = useMenuStore();
      const modules = await moduleStore.fetchModules();
      registerModules(modules);
      // ✅ 刷新页面时构建菜单
      await menuStore.buildMenus(modules);
      message.success('模块加载成功');
    } catch (error) {
      console.error('[Main] 加载模块失败:', error);
      message.error('加载模块失败，请刷新页面重试');
    }
  }
});

app.mount('#app');
