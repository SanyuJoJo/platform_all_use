import { createApp } from 'vue';
import { createPinia } from 'pinia';
import naive from 'naive-ui';
import App from './App.vue';
import router from './router';
import { useModuleStore } from './store/module';
import { useMenuStore } from './store/menu';
import { registerModules } from './micro-frontend/registry';

console.log('[Main] 应用启动');

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);
app.use(naive);

(async () => {
  const token = localStorage.getItem('token');

  if (token) {
    console.log('[Main] 检测到 token，提前加载模块');
    try {
      const moduleStore = useModuleStore();

      // 这里只是"提前"加载，真正决定能否渲染的还是守卫里的逻辑。
      // 即使这里失败，也不会阻塞下面的 router.isReady()。
      if (!moduleStore.loaded) {
        await moduleStore.fetchModules();
        console.log('[Main] 模块加载完成');
      }

      const modules = moduleStore.modules;
      if (modules && modules.length > 0) {
        const menuStore = useMenuStore();
        if (!menuStore.menuLoaded || menuStore.menuTree.length === 0) {
          await menuStore.buildMenus(modules);
          console.log('[Main] 菜单构建完成，数量:', menuStore.menuTree.length);
        }
      }
    } catch (error) {
      console.warn('[Main] 提前加载模块失败，但不阻塞应用启动:', error);
    }
  }

  // ★ 关键改动 1：等首次导航完成再挂载。
  //   首次导航会走守卫 → 加载模块 → 注册 subapp_* → 重定向 → 再次匹配。
  //   等它完成后再 mount，页面就不会先闪 404。
  await router.isReady();
  console.log('[Main] 路由准备就绪');

  // 挂载应用
  app.mount('#app');
  console.log('[Main] 应用挂载完成');

  // ★ 关键改动 2：挂载后再注册 qiankun 子应用。
  //   守卫里已经把所有 subapp_* 路由注册好了，这里只负责 qiankun 的 load/mount 生命周期。
  const moduleStore = useModuleStore();
  if (moduleStore.loaded && moduleStore.modules.length > 0) {
    console.log('[Main] 注册子应用');
    registerModules(moduleStore.modules, router);
  }
})();
