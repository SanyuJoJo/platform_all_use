import {
  registerMicroApps,
  start,
  initGlobalState,
  addGlobalUncaughtErrorHandler,
  type MicroAppStateActions,
  type RegistrableApp,
} from 'qiankun';
import type { Router } from 'vue-router';
import type { Module } from '@/types/module';
import type { RouteNode } from '@/types/permission';
import { message } from '@/utils/naive';
import { useUserStore } from '@/store/user';
import { usePermissionStore } from '@/store/permission';
import { MAIN_APP_IDS } from '@/types/permission';

let actions: MicroAppStateActions | null = null;
let isStarted = false;
let isRegistered = false;
let errorHandlerBound = false;
let lastModules: Module[] = [];

const SUB_APP_DIR_MAP: Record<string, string> = {
  auth: 'auth',
  module_manager: 'module-manager',
  audit_log: 'audit-log',
  license: 'license',
  crypto_console: 'crypto-console',
  ssl_proxy: 'ssl-proxy',
};

function getSubAppDir(moduleId: string): string {
  return SUB_APP_DIR_MAP[moduleId] || moduleId.replace(/_/g, '-');
}

function isMainAppModule(moduleId: string): boolean {
  return MAIN_APP_IDS.includes(moduleId);
}

interface SubAppProps {
  mainRouter: Router;
  moduleId: string;
  permissions: string[];
  user: unknown;
  token: string;
  routers: RouteNode[];
}

function resolveEntry(module: Module): string {
  const raw = module.entry_frontend?.trim();

  if (raw) {
    let resolved = '';
    if (/^https?:\/\//i.test(raw)) resolved = raw;
    else if (raw.startsWith('//')) resolved = `${window.location.protocol}${raw}`;
    else if (raw.startsWith('/')) resolved = new URL(raw, window.location.origin).href;
    else resolved = new URL(`/${raw}`, window.location.origin).href;

    if (
      import.meta.env.PROD &&
      /(^https?:\/\/)(localhost|127\.0\.0\.1)(:|\/)/i.test(resolved)
    ) {
      console.warn(
        `[qiankun] 模块 ${module.id} 的 entry_frontend 指向开发地址：${resolved}`
      );
    }
    return resolved;
  }

  const prefix = import.meta.env.VITE_SUBAPP_BASE_PREFIX || '/sub-apps';
  const dir = getSubAppDir(module.id);
  const fallback = new URL(`${prefix}/${dir}/`, window.location.origin).href;
  console.warn(
    `[qiankun] 模块 ${module.id} 缺少 entry_frontend，使用兜底入口：${fallback}`
  );
  return fallback;
}

function ensureContainer(): HTMLElement | null {
  const el = document.getElementById('subapp-container');
  if (!el) {
    console.warn('[qiankun] 未找到 #subapp-container，Layout 可能尚未渲染');
  }
  return el;
}

function bindGlobalErrorHandler(): void {
  if (errorHandlerBound) return;
  addGlobalUncaughtErrorHandler((event) => {
    console.error('[qiankun] 未捕获异常:', event);
    const errMsg =
      (event as ErrorEvent)?.message ||
      (event as PromiseRejectionEvent)?.reason?.message ||
      String(event);
    if (/Failed to fetch|import.*module|Loading chunk/i.test(errMsg)) {
      message.error('子应用加载失败，请检查网络或联系管理员');
    }
  });
  errorHandlerBound = true;
}

/**
 * 深拷贝：把 Pinia reactive Proxy 转成纯 JSON。
 * qiankun 跨沙箱传递 props 时，Proxy 可能导致部分字段丢失，
 * 深拷贝后再传可确保子应用拿到完整数据。
 */
function toPlain<T>(value: T): T {
  try {
    return JSON.parse(JSON.stringify(value)) as T;
  } catch (e) {
    console.error('[qiankun] toPlain 失败', e);
    return value;
  }
}

export function registerModules(modules: Module[], router: Router) {
  if (!modules || modules.length === 0) {
    console.warn('[qiankun] 没有模块可注册');
    return;
  }

  const activeModules = modules.filter(
    (m) => m.status === 'active' && !isMainAppModule(m.id)
  );
  if (activeModules.length === 0) {
    console.warn('[qiankun] 过滤后没有需要注册的子应用');
    return;
  }

  if (isRegistered) {
    console.warn('[qiankun] 子应用已注册，跳过重复注册');
    return;
  }

  lastModules = modules;

  const userStore = useUserStore();
  const permissionStore = usePermissionStore();
  // ★ 深拷贝，确保跨 qiankun 沙箱能完整传递
  const permissions = toPlain(userStore.permissions || []);
  const user = userStore.user ? toPlain(userStore.user) : null;
  const token = userStore.token || '';
  const routers = toPlain(permissionStore.routeTree);

  console.warn(
    `[qiankun] 准备下发 routers: 根节点数 = ${routers.length}, ` +
      `第 1 个根节点 name = ${routers[0]?.name ?? '-'}`
  );

  const apps: RegistrableApp<SubAppProps>[] = activeModules.map((module) => ({
    name: module.id,
    entry: resolveEntry(module),
    container: '#subapp-container',
    activeRule: `/${module.id}`,
    props: {
      mainRouter: router,
      moduleId: module.id,
      permissions,
      user,
      token,
      routers,
    },
  }));

  console.warn(
    '[qiankun] 注册子应用：',
    apps.map((a) => `${a.name} -> ${a.entry}`)
  );

  registerMicroApps<SubAppProps>(apps, {
    beforeLoad: (app) => {
      console.log(`[qiankun] before load ${app.name}`);
      ensureContainer();
      return Promise.resolve();
    },
    beforeMount: (app) => {
      console.log(`[qiankun] before mount ${app.name}`);
      return Promise.resolve();
    },
    afterMount: (app) => {
      console.log(`[qiankun] after mount ${app.name}`);
      return Promise.resolve();
    },
    beforeUnmount: (app) => {
      console.log(`[qiankun] before unmount ${app.name}`);
      return Promise.resolve();
    },
    afterUnmount: (app) => {
      console.log(`[qiankun] after unmount ${app.name}`);
      return Promise.resolve();
    },
  });

  bindGlobalErrorHandler();

  if (!isStarted) {
    start({
      prefetch: false,
      sandbox: { experimentalStyleIsolation: true },
    });
    isStarted = true;
  }

  if (!actions) {
    actions = initGlobalState({ user, permissions, token, routers });
  } else {
    actions.setGlobalState({ user, permissions, token, routers });
  }

  isRegistered = true;
}

export function reRegister() {
  if (lastModules.length === 0) {
    console.warn('[qiankun] 没有缓存的模块，跳过重新注册');
    return;
  }
  window.location.reload();
}

export function getGlobalActions() {
  return actions;
}
