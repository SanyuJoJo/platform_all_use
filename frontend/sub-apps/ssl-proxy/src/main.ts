import { createApp, type App as VueApp } from 'vue';
import { createPinia } from 'pinia';
import { createRouter, createWebHistory, type Router } from 'vue-router';
import {
  renderWithQiankun,
  qiankunWindow,
} from 'vite-plugin-qiankun/dist/helper';
import App from './App.vue';
import {
  setupDynamicRoutes,
  resetDynamicRoutes,
  defaultRouters,
} from './router/dynamic';
import { usePermissionStore } from './store/permission'; // ★ 引入权限Store
import type { RouteNode } from './types';

const DEFAULT_MODULE_ID = 'ssl_proxy';

let app: VueApp | null = null;
let router: Router | null = null;
let currentModuleId = DEFAULT_MODULE_ID;

function createAppInstance(base: string, props: any = {}) {
  const pinia = createPinia();
  const _router = createRouter({
    history: createWebHistory(base),
    routes: [
      { path: '/', redirect: '/list' },
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
  
  // ★ 关键：将主应用传来的权限注入 Pinia
  const permissionStore = usePermissionStore(pinia);
  permissionStore.setPermissions(props?.permissions || []);

  return { app: _app, router: _router };
}

function resolveRouters(props: any): RouteNode[] {
  const r = props?.routers;
  if (Array.isArray(r) && r.length > 0) {
    console.warn('[ssl-proxy] resolveRouters 使用 props.routers, 长度 =', r.length);
    return r as RouteNode[];
  }
  console.warn('[ssl-proxy] resolveRouters props.routers 为空，使用 defaultRouters');
  return defaultRouters;
}

async function bootstrapApp(props: any) {
  const moduleId =
    typeof props?.moduleId === 'string' ? props.moduleId : DEFAULT_MODULE_ID;
  currentModuleId = moduleId;

  const base = qiankunWindow.__POWERED_BY_QIANKUN__ ? `/${moduleId}` : '/';
  console.warn(
    `[ssl-proxy] bootstrapApp: base = ${base}, moduleId = ${moduleId}, ` +
      `__POWERED_BY_QIANKUN__ = ${qiankunWindow.__POWERED_BY_QIANKUN__}`
  );

  // ★ 传入 props 用于初始化权限 Store
  const { app: _app, router: _router } = createAppInstance(base, props);
  app = _app;
  router = _router;

  const routers = resolveRouters(props);
  setupDynamicRoutes(_router, routers, moduleId);

  try {
    await _router.isReady();
    console.warn(
      '[ssl-proxy] router.isReady() 完成, current =',
      _router.currentRoute.value.fullPath
    );
  } catch (e) {
    console.error('[ssl-proxy] router.isReady() 失败', e);
  }

  return { app: _app, router: _router };
}

async function mountApp(props: any) {
  console.warn('[ssl-proxy] mountApp 开始');
  const { app: _app, router: _router } = await bootstrapApp(props);

  const container = props?.container
    ? (props.container as HTMLElement).querySelector('#app')
    : document.querySelector('#app');
  if (!container) {
    console.error('[ssl-proxy] 未找到挂载容器 #app');
    return;
  }
  _app.mount(container);
  console.warn('[ssl-proxy] app.mount() 完成');

  props?.onGlobalStateChange?.((state: any) => {
    // ★ 如果主应用动态更新了权限，同步更新 Pinia
    if (state?.permissions) {
      const permissionStore = usePermissionStore(_app.config.globalProperties.$pinia as any);
      permissionStore.setPermissions(state.permissions);
    }

    if (!state?.routers?.length) return;
    console.warn('[ssl-proxy] 收到主应用下发的路由树，重建动态路由');
    resetDynamicRoutes(_router);
    setupDynamicRoutes(_router, state.routers, currentModuleId);
    const cur = _router.currentRoute.value;
    if (cur.name === 'SubNotFound') _router.replace(cur.fullPath);
  }, true);
}

function unmountApp() {
  console.warn('[ssl-proxy] unmountApp 开始');
  app?.unmount();
  app = null;
  router = null;
}

console.warn(
  '[ssl-proxy] main.ts 开始执行, __POWERED_BY_QIANKUN__ =',
  qiankunWindow.__POWERED_BY_QIANKUN__
);

renderWithQiankun({
  bootstrap() {
    console.warn('[ssl-proxy] bootstrap 生命周期被调用');
  },
  mount(props) {
    console.warn('[ssl-proxy] mount 生命周期被调用');
    mountApp(props).catch((e) =>
      console.error('[ssl-proxy] mountApp 异常', e)
    );
  },
  unmount() {
    console.warn('[ssl-proxy] unmount 生命周期被调用');
    unmountApp();
  },
  update(props) {
    console.warn('[ssl-proxy] update 生命周期被调用', props);
  },
});

if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.warn('[ssl-proxy] 独立运行模式');
  mountApp({}).catch((e) =>
    console.error('[ssl-proxy] 独立运行 mount 异常', e)
  );
}
