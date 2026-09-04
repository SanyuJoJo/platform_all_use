import { createApp } from 'vue';
import { createPinia } from 'pinia';
import naive from 'naive-ui';
import App from './App.vue';
import router from './router';
import { registerModules } from './micro-frontend/registry';
import { useModuleStore } from './store/module';
import { useMenuStore } from './store/menu';

console.log('[Main] 应用启动');

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(naive);

app.mount('#app');

// 在路由准备就绪后，加载模块并构建菜单，再注册子应用
router.isReady().then(async () => {
  console.log('[Main] 路由准备就绪');
  const moduleStore = useModuleStore();
  if (!moduleStore.loaded) {
    await moduleStore.fetchModules();
  }
  const modules = moduleStore.modules;
  if (modules && modules.length > 0) {
    // 先构建菜单（确保在子应用加载前菜单数据存在）
    const menuStore = useMenuStore();
    if (!menuStore.menuLoaded || menuStore.menuTree.length === 0) {
      await menuStore.buildMenus(modules);
      console.log('[Main] 菜单构建完成，数量:', menuStore.menuTree.length);
    }
    // 再注册子应用
    registerModules(modules, router);
  }
  console.log('[Main] 模块注册和菜单构建完成');
});

console.log('[Main] 应用挂载完成');