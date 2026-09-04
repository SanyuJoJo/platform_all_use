// src/main.ts
import { createApp } from 'vue';
import App from './App.vue';
import { createPinia } from 'pinia';
import { createAppRouter } from './router';
import { renderWithQiankun, qiankunWindow } from 'vite-plugin-qiankun/dist/helper';

let app: any = null;
let routerInstance: any = null;

function render(props: any = {}) {
  const { container, moduleId, targetPath } = props; // 获取 targetPath
  const isQiankun = !!qiankunWindow.__POWERED_BY_QIANKUN__;
  const base = isQiankun ? `/${moduleId || 'auth'}` : '/';
  console.log('[SubApp] 渲染，base:', base, 'targetPath:', targetPath);

  if (app) {
    app.unmount();
    app = null;
  }

  routerInstance = createAppRouter(moduleId || 'auth', isQiankun);
  app = createApp(App);
  app.use(routerInstance);
  app.use(createPinia());
  app.mount(container ? container.querySelector('#app') : '#app');

  // 🔥 关键：如果提供了 targetPath，则同步跳转
  if (targetPath && isQiankun) {
    // targetPath 是完整路径，如 '/auth/users'
    // 需要提取子应用内部路径（去掉 base 前缀）
    const basePath = `/${moduleId || 'auth'}`;
    const innerPath = targetPath.startsWith(basePath) 
      ? targetPath.slice(basePath.length) || '/' 
      : targetPath;
    console.log('[SubApp] 跳转到内部路径:', innerPath);
    routerInstance.replace(innerPath);
  }
}

if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.log('[SubApp] 独立运行模式');
  render();
}

renderWithQiankun({
  bootstrap() {
    console.log('[SubApp] bootstrap');
    return Promise.resolve();
  },
  mount(props: any) {
    console.log('[SubApp] mount', props);
    render(props);
    return Promise.resolve();
  },
  unmount() {
    console.log('[SubApp] unmount');
    if (app) {
      app.unmount();
      app = null;
      routerInstance = null;
    }
    return Promise.resolve();
  }
});