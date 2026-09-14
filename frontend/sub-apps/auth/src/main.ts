import { createApp } from 'vue';
import App from './App.vue';
import { createPinia } from 'pinia';
import { createAppRouter } from './router';
import {
  renderWithQiankun,
  qiankunWindow,
} from 'vite-plugin-qiankun/dist/helper';
import naive from 'naive-ui';
import { setupPermissionDirective } from './directives/permission';
import { useUserStore } from './store/user';

let app: any = null;
let routerInstance: any = null;

function render(props: any = {}) {
  const { container, moduleId, targetPath, permissions, user, token } = props;
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'auth'}` : '/';
  console.log('[Auth] 渲染，base:', base, 'targetPath:', targetPath);

  if (app) {
    app.unmount();
    app = null;
    routerInstance = null;
  }

  const pinia = createPinia();
  routerInstance = createAppRouter(moduleId || 'auth', isQiankun);
  app = createApp(App);
  app.use(pinia);
  app.use(routerInstance);
  app.use(naive);
  setupPermissionDirective(app);

  if (permissions) {
    const userStore = useUserStore();
    userStore.setPermissions(permissions);
  }
  if (user) {
    const userStore = useUserStore();
    userStore.setUser(user);
  }
  if (token) {
    const userStore = useUserStore();
    userStore.setToken(token);
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
    console.error('[Auth] 找不到挂载容器 #app');
  }

  if (targetPath && isQiankun) {
    const basePath = `/${moduleId || 'auth'}`;
    const innerPath = targetPath.startsWith(basePath)
      ? targetPath.slice(basePath.length) || '/'
      : targetPath;
    routerInstance.replace(innerPath);
  }
}

if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[Auth] 独立运行模式');
  render();
}

renderWithQiankun({
  bootstrap() {
    console.log('[Auth] bootstrap');
    return Promise.resolve();
  },
  mount(props: any) {
    console.log('[Auth] mount', props);
    render(props);
    return Promise.resolve();
  },
  unmount() {
    console.log('[Auth] unmount');
    if (app) {
      app.unmount();
      app = null;
      routerInstance = null;
    }
    return Promise.resolve();
  },
  // 补齐 update 生命周期
  update(props: any) {
    console.log('[Auth] update', props);
    if (props?.permissions) {
      const userStore = useUserStore();
      userStore.setPermissions(props.permissions);
    }
    if (props?.user) {
      const userStore = useUserStore();
      userStore.setUser(props.user);
    }
    if (props?.token) {
      const userStore = useUserStore();
      userStore.setToken(props.token);
    }
    return Promise.resolve();
  },
});
