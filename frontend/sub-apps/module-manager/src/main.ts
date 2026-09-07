import { createApp } from 'vue';
import App from './App.vue';
import { createPinia } from 'pinia';
import { createAppRouter } from './router';
import { renderWithQiankun, qiankunWindow } from 'vite-plugin-qiankun/dist/helper';
import naive from 'naive-ui';
import { useModuleStore } from './store/module';
let app: any = null;
let routerInstance: any = null;
function render(props: any = {}) {
  const { container, moduleId, targetPath, permissions, user, token } = props;
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'module-manager'}` : '/';
  console.log('[ModuleManager] 渲染，base:', base, 'targetPath:', targetPath);
  if (app) {
    app.unmount();
    app = null;
    routerInstance = null;
  }
  const pinia = createPinia();
  routerInstance = createAppRouter(moduleId || 'module-manager', isQiankun);
  app = createApp(App);
  app.use(pinia);
  app.use(routerInstance);
  app.use(naive);
  // 从主应用接收权限并存入 store
  if (permissions) {
    const moduleStore = useModuleStore();
    moduleStore.setPermissions(permissions);
  }
  // 可选：存储用户/token（如需）
  if (user) { /* 可扩展 */ }
  if (token) { /* 可扩展 */ }
  const mountEl = container ? container.querySelector('#app') : document.getElementById('app');
  if (mountEl) {
    if (mountEl.hasChildNodes()) {
      mountEl.innerHTML = '';
    }
    app.mount(mountEl);
  } else {
    console.error('[ModuleManager] 找不到挂载容器 #app');
  }
  if (targetPath && isQiankun) {
    const basePath = `/${moduleId || 'module-manager'}`;
    const innerPath = targetPath.startsWith(basePath)
      ? targetPath.slice(basePath.length) || '/'
      : targetPath;
    routerInstance.replace(innerPath);
  }
}
if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[ModuleManager] 独立运行模式');
  render();
}
renderWithQiankun({
  bootstrap() {
    console.log('[ModuleManager] bootstrap');
    return Promise.resolve();
  },
  mount(props: any) {
    console.log('[ModuleManager] mount', props);
    render(props);
    return Promise.resolve();
  },
  unmount() {
    console.log('[ModuleManager] unmount');
    if (app) {
      app.unmount();
      app = null;
      routerInstance = null;
    }
    return Promise.resolve();
  }
});
