import { createApp } from 'vue';
import App from './App.vue';
import { createPinia } from 'pinia';
import { createAppRouter } from './router';
import { renderWithQiankun, qiankunWindow } from 'vite-plugin-qiankun/dist/helper';
import naive from 'naive-ui';
import { usePermissionStore } from './store/permission';
let app: any = null;
let routerInstance: any = null;
function render(props: any = {}) {
  const { container, moduleId, targetPath, permissions, user, token } = props;
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'audit-log'}` : '/';
  console.log('[AuditLog] 渲染，base:', base, 'targetPath:', targetPath);
  if (app) {
    app.unmount();
    app = null;
    routerInstance = null;
  }
  const pinia = createPinia();
  routerInstance = createAppRouter(moduleId || 'audit-log', isQiankun);
  app = createApp(App);
  app.use(pinia);
  app.use(routerInstance);
  app.use(naive);
  // 将权限存入 Pinia store
  if (permissions) {
    const permissionStore = usePermissionStore();
    permissionStore.setPermissions(permissions);
  }
  const mountEl = container ? container.querySelector('#app') : document.getElementById('app');
  if (mountEl) {
    if (mountEl.hasChildNodes()) {
      mountEl.innerHTML = '';
    }
    app.mount(mountEl);
  } else {
    console.error('[AuditLog] 找不到挂载容器 #app');
  }
  if (targetPath && isQiankun) {
    const basePath = `/${moduleId || 'audit-log'}`;
    const innerPath = targetPath.startsWith(basePath)
      ? targetPath.slice(basePath.length) || '/'
      : targetPath;
    routerInstance.replace(innerPath);
  }
}
if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[AuditLog] 独立运行模式');
  render();
}
renderWithQiankun({
  bootstrap() {
    console.log('[AuditLog] bootstrap');
    return Promise.resolve();
  },
  mount(props: any) {
    console.log('[AuditLog] mount', props);
    render(props);
    return Promise.resolve();
  },
  unmount() {
    console.log('[AuditLog] unmount');
    if (app) {
      app.unmount();
      app = null;
      routerInstance = null;
    }
    return Promise.resolve();
  }
});