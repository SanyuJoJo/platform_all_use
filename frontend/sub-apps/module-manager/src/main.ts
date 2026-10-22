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
import type { RouteNode } from './types';

const DEFAULT_MODULE_ID = 'module_manager';

let app: VueApp | null = null;
let router: Router | null = null;
let currentModuleId = DEFAULT_MODULE_ID;

function createAppInstance(base: string) {
  const pinia = createPinia();
  const _router = createRouter({
    history: createWebHistory(base),
    routes: [
      { path: '/', redirect: '/modules' },
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
  if (Array.isArray(props?.routers) && props.routers.length > 0) {
    return props.routers as RouteNode[];
  }
  return defaultRouters;
}

async function bootstrapApp(props: any) {
  const moduleId =
    typeof props?.moduleId === 'string' ? props.moduleId : DEFAULT_MODULE_ID;
  currentModuleId = moduleId;

  // qiankun 环境下 basename = /{moduleId}；独立运行时为 '/'
  const base = qiankunWindow.__POWERED_BY_QIANKUN__ ? `/${moduleId}` : '/';
  console.warn(
    `[module-manager] bootstrapApp: base = ${base}, moduleId = ${moduleId}, ` +
      `__POWERED_BY_QIANKUN__ = ${qiankunWindow.__POWERED_BY_QIANKUN__}`
  );

  const { app: _app, router: _router } = createAppInstance(base);
  app = _app;
  router = _router;

  const routers = resolveRouters(props);
  setupDynamicRoutes(_router, routers, moduleId);

  try {
    await _router.isReady();
    console.warn(
      '[module-manager] router.isReady() 完成, current =',
      _router.currentRoute.value.fullPath
    );
  } catch (e) {
    console.error('[module-manager] router.isReady() 失败', e);
  }

  return { app: _app, router: _router };
}

async function mountApp(props: any) {
  console.warn('[module-manager] mountApp 开始');
  const { app: _app, router: _router } = await bootstrapApp(props);

  const container = props?.container
    ? (props.container as HTMLElement).querySelector('#app')
    : document.querySelector('#app');
  if (!container) {
    console.error('[module-manager] 未找到挂载容器 #app');
    return;
  }
  _app.mount(container);
  console.warn('[module-manager] app.mount() 完成');

  props?.onGlobalStateChange?.((state: any) => {
    if (!state?.routers?.length) return;
    console.warn('[module-manager] 收到主应用下发的路由树，重建动态路由');
    resetDynamicRoutes(_router);
    setupDynamicRoutes(_router, state.routers, currentModuleId);
    const cur = _router.currentRoute.value;
    if (cur.name === 'SubNotFound') _router.replace(cur.fullPath);
  }, true);
}

function unmountApp() {
  console.warn('[module-manager] unmountApp 开始');
  app?.unmount();
  app = null;
  router = null;
}

console.warn(
  '[module-manager] main.ts 开始执行, __POWERED_BY_QIANKUN__ =',
  qiankunWindow.__POWERED_BY_QIANKUN__
);

/* ============================================================================
 * ★ 关键：必须用 renderWithQiankun 注册生命周期
 *   vite-plugin-qiankun 只识别 renderWithQiankun 里的生命周期，
 *   不支持 export function bootstrap 这种 webpack 写法。
 * ========================================================================== */
renderWithQiankun({
  bootstrap() {
    console.warn('[module-manager] bootstrap 生命周期被调用');
  },
  mount(props) {
    console.warn('[module-manager] mount 生命周期被调用');
    mountApp(props).catch((e) =>
      console.error('[module-manager] mountApp 异常', e)
    );
  },
  unmount() {
    console.warn('[module-manager] unmount 生命周期被调用');
    unmountApp();
  },
  update(props) {
    console.warn('[module-manager] update 生命周期被调用', props);
  },
});

/* ============ 独立运行入口 ============ */

if (!qiankunWindow.__POWERED_BY_QIANKUN__) {
  console.warn('[module-manager] 独立运行模式');
  mountApp({}).catch((e) =>
    console.error('[module-manager] 独立运行 mount 异常', e)
  );
}
