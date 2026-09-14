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
import { message } from '@/utils/naive';
import { useUserStore } from '@/store/user';

let actions: MicroAppStateActions | null = null;
let isStarted = false;
let isRegistered = false;
let errorHandlerBound = false;
let lastModules: Module[] = [];

/**
 * 子应用模块 ID → 部署目录名 映射
 */
const SUB_APP_DIR_MAP: Record<string, string> = {
  auth: 'auth',
  module_manager: 'module-manager',
  audit_log: 'audit-log',
  license: 'license',
};

function getSubAppDir(moduleId: string): string {
  return SUB_APP_DIR_MAP[moduleId] || moduleId.replace(/_/g, '-');
}

interface SubAppProps {
  mainRouter: Router;
  moduleId: string;
  permissions: string[];
  user: unknown;
  token: string;
}

function resolveEntry(module: Module): string {
  const raw = module.entry_frontend?.trim();

  if (raw) {
    let resolved = '';
    if (/^https?:\/\//i.test(raw)) {
      resolved = raw;
    } else if (raw.startsWith('//')) {
      resolved = `${window.location.protocol}${raw}`;
    } else if (raw.startsWith('/')) {
      resolved = new URL(raw, window.location.origin).href;
    } else {
      resolved = new URL(`/${raw}`, window.location.origin).href;
    }

    if (
      import.meta.env.PROD &&
      /(^https?:\/\/)(localhost|127\.0\.0\.1)(:|\/)/i.test(resolved)
    ) {
      console.warn(
        `[qiankun] 模块 ${module.id} 的 entry_frontend 指向开发地址：${resolved}，` +
          `请后端更新为生产路径（参考 backend-module-entry.patch.json）`
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

function ensureContainer(): HTMLElement {
  let container = document.getElementById('subapp-container');
  if (!container) {
    container = document.createElement('div');
    container.id = 'subapp-container';
    container.style.height = '100%';
    container.style.minHeight = '300px';
    container.style.background = '#fff';
    container.style.zIndex = '10';

    const content = document.querySelector('.n-layout-content');
    if (content) {
      content.prepend(container);
    } else {
      document.body.appendChild(container);
    }
    console.log('[qiankun] 已创建缺失的 #subapp-container');
  }
  return container;
}

/**
 * 绑定 qiankun 全局未捕获错误处理器。
 *
 * qiankun 的 registerMicroApps 第二个参数（FrameworkLifeCycles）
 * 不支持 error 回调，子应用加载错误通过 addGlobalUncaughtErrorHandler 捕获。
 */
function bindGlobalErrorHandler(): void {
  if (errorHandlerBound) return;
  addGlobalUncaughtErrorHandler((event) => {
    console.error('[qiankun] 未捕获异常:', event);
    // event 可能是 ErrorEvent / PromiseRejectionEvent / string
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

export function registerModules(modules: Module[], router: Router) {
  if (!modules || modules.length === 0) {
    console.warn('[qiankun] 没有模块可注册');
    return;
  }

  const activeModules = modules.filter((m) => m.status === 'active');
  if (activeModules.length === 0) {
    console.warn('[qiankun] 没有激活的模块');
    return;
  }

  if (isRegistered) {
    console.warn('[qiankun] 子应用已注册，跳过重复注册');
    return;
  }

  try {
    ensureContainer();
  } catch (err) {
    console.error('[qiankun] 容器创建失败:', err);
    return;
  }

  lastModules = modules;

  const userStore = useUserStore();
  const permissions = userStore.permissions || [];
  const user = userStore.user || null;
  const token = userStore.token || '';

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
    },
  }));

  console.log(
    '[qiankun] 注册子应用：',
    apps.map((a) => `${a.name} -> ${a.entry}`)
  );

  // 只传 qiankun 支持的 5 个生命周期钩子
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

  // 全局错误处理通过 addGlobalUncaughtErrorHandler
  bindGlobalErrorHandler();

  if (!isStarted) {
    start({
      prefetch: false,
      sandbox: {
        experimentalStyleIsolation: true,
      },
    });
    isStarted = true;
  }

  if (!actions) {
    actions = initGlobalState({ user: null });
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
