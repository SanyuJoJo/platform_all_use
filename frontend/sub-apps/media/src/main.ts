import { createApp, type App as VueApp } from 'vue';
import { createPinia } from 'pinia';
import { createRouter, createWebHistory, type Router } from 'vue-router';
import { renderWithQiankun, qiankunWindow } from 'vite-plugin-qiankun/dist/helper';
import App from './App.vue';
import { setupDynamicRoutes, resetDynamicRoutes, defaultRouters } from './router/dynamic';
import type { RouteNode } from './types';
const DEFAULT_MODULE_ID = 'media';
let app: VueApp | null = null;
let router: Router | null = null;
let currentModuleId = DEFAULT_MODULE_ID;
function createAppInstance(base: string) {
  const pinia = createPinia();
  const _router = createRouter({
    history: createWebHistory(base),
    routes: [
      {
        path: '/:pathMatch(.*)*',
        name: 'SubNotFound',
        component: () => import('./views/404.vue'),
      },
    ],
  });
  const _app = createApp(App);
  _app.use(pinia);
  _app.use(_router);
  return { app: _app, router: _router };
}
function resolveRouters(props: any): RouteNode[] {
  // 媒体子应用路由固定用 defaultRouters（媒体库 / 、直播 /live）。
  // 主应用下发的完整路由树在 qiankun 沙箱内深拷贝后，MediaLive 的
  // microApp/type 可能被破坏导致 isOwnedBy 过滤掉；固定路由最可靠。
  return defaultRouters;
}
async function bootstrapApp(props: any) {
  const moduleId = typeof props?.moduleId === 'string' ? props.moduleId : DEFAULT_MODULE_ID;
  currentModuleId = moduleId;
  const base = qiankunWindow.__POWERED_BY_QIANKUN__ ? `/${moduleId}` : '/';
  const { app: _app, router: _router } = createAppInstance(base);
  app = _app;
  router = _router;
  setupDynamicRoutes(_router, resolveRouters(props), moduleId);
  try {
    await _router.isReady();
  } catch (e) { console.error('[media] router.isReady 失败', e); }
  return { app: _app, router: _router };
}
async function mountApp(props: any) {
  const { app: _app, router: _router } = await bootstrapApp(props);
  const container = props?.container
    ? (props.container as HTMLElement).querySelector('#app')
    : document.querySelector('#app');
  if (!container) { console.error('[media] 未找到 #app'); return; }
  _app.mount(container);
  props?.onGlobalStateChange?.((state: any) => {
    if (!state?.routers?.length) return;
    resetDynamicRoutes(_router);
    setupDynamicRoutes(_router, state.routers, currentModuleId);
    const cur = _router.currentRoute.value;
    if (cur.name === 'SubNotFound') _router.replace(cur.fullPath);
  }, true);
}
function unmountApp() {
  app?.unmount();
  app = null;
  router = null;
}
renderWithQiankun({
  bootstrap() {},
  mount(props) { mountApp(props).catch((e) => console.error('[media] mountApp 异常', e)); },
  unmount() { unmountApp(); },
  update() {},
});
if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  mountApp({}).catch((e) => console.error('[media] 独立运行异常', e));
}

