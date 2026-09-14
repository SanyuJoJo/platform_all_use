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
import { setupPermissionDirective } from './directives/permission';

let app: any = null;
let routerInstance: any = null;

function render(props: any = {}) {
  const { container, moduleId, targetPath, permissions, user, token } = props;
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'license'}` : '/';
  console.log('[License] 渲染，base:', base, 'targetPath:', targetPath);

  if (app) {
    app.unmount();
    app = null;
    routerInstance = null;
  }

  const pinia = createPinia();
  routerInstance = createAppRouter(moduleId || 'license', isQiankun);
  app = createApp(App);
  app.use(pinia);
  app.use(routerInstance);
  app.use(naive);
  setupPermissionDirective(app);

  // 处理权限：优先使用 props；独立运行时从环境变量读取 Mock 权限
  let finalPermissions: string[] = [];
  if (permissions && Array.isArray(permissions)) {
    finalPermissions = permissions;
  } else if (!isQiankun && import.meta.env.DEV) {
    const mockPerms = import.meta.env.VITE_MOCK_PERMISSIONS;
    if (mockPerms) {
      // 显式标注 s: string，避免隐式 any
      finalPermissions = mockPerms.split(',').map((s: string) => s.trim());
    } else {
      finalPermissions = ['license:license:view', 'license:license:create'];
    }
    console.warn('[License] 独立运行模式，使用 Mock 权限:', finalPermissions);
  }

  if (finalPermissions.length > 0) {
    const permissionStore = usePermissionStore();
    permissionStore.setPermissions(finalPermissions);
  }

  // 保留 user / token 供后续扩展
  void user;
  void token;

  const mountEl = container
    ? container.querySelector('#app')
    : document.getElementById('app');
  if (mountEl) {
    if (mountEl.hasChildNodes()) {
      mountEl.innerHTML = '';
    }
    app.mount(mountEl);
  } else {
    console.error('[License] 找不到挂载容器 #app');
  }

  if (targetPath && isQiankun) {
    const basePath = `/${moduleId || 'license'}`;
    const innerPath = targetPath.startsWith(basePath)
      ? targetPath.slice(basePath.length) || '/'
      : targetPath;
    routerInstance.replace(innerPath);
  }
}

if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[License] 独立运行模式');
  render();
}

renderWithQiankun({
  bootstrap() {
    console.log('[License] bootstrap');
    return Promise.resolve();
  },
  mount(props: any) {
    console.log('[License] mount', props);
    render(props);
    return Promise.resolve();
  },
  unmount() {
    console.log('[License] unmount');
    if (app) {
      app.unmount();
      app = null;
      routerInstance = null;
    }
    return Promise.resolve();
  },
  // 补齐 update 生命周期
  update(props: any) {
    console.log('[License] update', props);
    if (props?.permissions && Array.isArray(props.permissions)) {
      const permissionStore = usePermissionStore();
      permissionStore.setPermissions(props.permissions);
    }
    return Promise.resolve();
  },
});
