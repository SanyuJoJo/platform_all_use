// src/main.ts
import { createApp } from 'vue';
import App from './App.vue';
import { createPinia } from 'pinia';
import { createAppRouter } from './router';
import { renderWithQiankun, qiankunWindow } from 'vite-plugin-qiankun/dist/helper';
import naive from 'naive-ui';
import { setupPermissionDirective } from './directives/permission';
import { useUserStore } from './store/user';

let app: any = null;
let routerInstance: any = null;

function render(props: any = {}) {
  const { container, moduleId, targetPath } = props;
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'auth'}` : '/';
  console.log('[Auth] 渲染，base:', base, 'targetPath:', targetPath);

  // 清理已有实例
  if (app) {
    app.unmount();
    app = null;
    routerInstance = null;
  }

  // 创建 Pinia 实例
  const pinia = createPinia();

  // 创建路由
  routerInstance = createAppRouter(moduleId || 'auth', isQiankun);

  // 创建 Vue 应用
  app = createApp(App);
  app.use(pinia);
  app.use(routerInstance);
  app.use(naive);
  setupPermissionDirective(app);

  // 同步主应用传递的权限等信息
  if (props.permissions) {
    const userStore = useUserStore();
    userStore.setPermissions(props.permissions);
  }
  if (props.user) {
    const userStore = useUserStore();
    userStore.setUser(props.user);
  }
  if (props.token) {
    const userStore = useUserStore();
    userStore.setToken(props.token);
  }

  // 挂载
  const mountEl = container ? container.querySelector('#app') : document.getElementById('app');
  if (mountEl) {
    // 清空容器（防止重复挂载）
    if (mountEl.hasChildNodes()) {
      mountEl.innerHTML = '';
    }
    app.mount(mountEl);
  } else {
    console.error('[Auth] 找不到挂载容器 #app');
  }

  // 如果提供了 targetPath，则同步跳转
  if (targetPath && isQiankun) {
    const basePath = `/${moduleId || 'auth'}`;
    const innerPath = targetPath.startsWith(basePath)
      ? targetPath.slice(basePath.length) || '/'
      : targetPath;
    console.log('[Auth] 跳转到内部路径:', innerPath);
    routerInstance.replace(innerPath);
  }
}

// 独立运行时直接渲染
if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[Auth] 独立运行模式');
  render();
}

// qiankun 生命周期
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
  }
});
