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

// 在挂载前，如果已登录，提前加载模块并添加子应用路由
(async () => {
  const token = localStorage.getItem('token');

  if (token) {
    console.log('[Main] 检测到 token，提前加载模块');

    // ★ 唯一改动：整个"提前加载"逻辑用 try-catch 包裹，
    //    无论成功失败，都不阻塞下面的 app.mount('#app')
    try {
      const moduleStore = useModuleStore();

      if (!moduleStore.loaded) {
        await moduleStore.fetchModules();
      }

      const modules = moduleStore.modules;
      if (modules && modules.length > 0) {
        const menuStore = useMenuStore();
        if (!menuStore.menuLoaded || menuStore.menuTree.length === 0) {
          await menuStore.buildMenus(modules);
          console.log('[Main] 菜单构建完成，数量:', menuStore.menuTree.length);
        }

        // 添加子应用路由到 router
        modules.forEach(m => {
          const routeName = `subapp_${m.id}`;
          if (!router.hasRoute(routeName)) {
            router.addRoute('Layout', {
              path: `/${m.id}/:pathMatch(.*)*`,
              name: routeName,
              component: { render: () => null },
              meta: { ignoreAuth: true, isSubApp: true },
            });
            console.log(`[Main] 添加子应用路由: ${routeName}`);
          }
        });
      }
    } catch (error) {
      console.warn('[Main] 提前加载模块失败，但不阻塞应用启动:', error);
    }
  }

  // 挂载应用
  app.mount('#app');
  console.log('[Main] 应用挂载完成');

  // 挂载后，注册子应用（qiankun）
  router.isReady().then(() => {
    console.log('[Main] 路由准备就绪，注册子应用');
    const moduleStore = useModuleStore();
    if (moduleStore.loaded && moduleStore.modules.length > 0) {
      registerModules(moduleStore.modules, router);
    }
  });
})();
