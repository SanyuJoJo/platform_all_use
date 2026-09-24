import { createApp } from 'vue';
import App from './App.vue';
import { createPinia } from 'pinia';
import { createAppRouter } from './router';
import {
  renderWithQiankun,
  qiankunWindow,
} from 'vite-plugin-qiankun/dist/helper';
import naive from 'naive-ui';
import { usePermissionStore } from './store/permission';
import { permissionDirective } from './directives/permission';
let app: any = null;
let routerInstance: any = null;
function render(props: any = {}) {
  const { container, moduleId, targetPath, permissions } = props;
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'crypto_console'}` : '/';
  console.log('[CryptoConsole] 渲染，base:', base, 'targetPath:', targetPath);
  if (app) {
    app.unmount();
    app = null;
    routerInstance = null;
  }
  const pinia = createPinia();
  routerInstance = createAppRouter(moduleId || 'crypto_console', isQiankun);
  app = createApp(App);
  app.use(pinia);
  app.use(routerInstance);
  app.use(naive);
  app.directive('permission', permissionDirective);
  if (permissions && Array.isArray(permissions)) {
    const permissionStore = usePermissionStore();
    permissionStore.setPermissions(permissions);
  } else if (!isQiankun && import.meta.env.DEV) {
    const mockPerms = import.meta.env.VITE_MOCK_PERMISSIONS as string | undefined;
    if (mockPerms) {
      const permissionStore = usePermissionStore();
      permissionStore.setPermissions(
        mockPerms.split(',').map((s: string) => s.trim()).filter(Boolean)
      );
    }
  }
  const mountEl = container
    ? container.querySelector('#app')
    : document.getElementById('app');
  if (mountEl) {
    if (mountEl.hasChildNodes()) {
      mountEl.innerHTML = '';
    }
    app.mount(mountEl);
  } else {
    console.error('[CryptoConsole] 找不到挂载容器 #app');
  }
  if (targetPath && isQiankun) {
    const basePath = `/${moduleId || 'crypto_console'}`;
    const innerPath = targetPath.startsWith(basePath)
      ? targetPath.slice(basePath.length) || '/'
      : targetPath;
    routerInstance.replace(innerPath);
  }
}
if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[CryptoConsole] 独立运行模式');
  render();
}
renderWithQiankun({
  bootstrap() {
    console.log('[CryptoConsole] bootstrap');
    return Promise.resolve();
  },
  mount(props: any) {
    console.log('[CryptoConsole] mount', props);
    render(props);
    return Promise.resolve();
  },
  unmount() {
    console.log('[CryptoConsole] unmount');
    if (app) {
      app.unmount();
      app = null;
      routerInstance = null;
    }
    return Promise.resolve();
  },
  update(props: any) {
    console.log('[CryptoConsole] update', props);
    if (props?.permissions && Array.isArray(props.permissions)) {
      const permissionStore = usePermissionStore();
      permissionStore.setPermissions(props.permissions);
    }
    return Promise.resolve();
  },
});
