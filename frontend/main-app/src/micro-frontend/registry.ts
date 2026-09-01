import {
  registerMicroApps,
  start,
  initGlobalState,
  type MicroAppStateActions,
} from 'qiankun';
import type { Router } from 'vue-router';
import { message } from '@/utils/naive';
import type { Module } from '@/types/module';

let actions: MicroAppStateActions | null = null;
let isRegistered = false; // 防止重复注册

/**
 * 动态注册子应用
 * @param modules - 模块列表（需包含活跃模块）
 * @param router - 主应用路由实例
 */
export function registerModules(modules: Module[], router: Router) {
  if (isRegistered) {
    console.warn('[qiankun] 子应用已注册，跳过重复注册');
    return;
  }

  if (!modules || modules.length === 0) {
    console.warn('[qiankun] 没有可用的子应用模块');
    return;
  }

  const activeModules = modules.filter((m) => m.status === 'active');
  if (activeModules.length === 0) {
    console.warn('[qiankun] 没有激活的子应用');
    return;
  }

  const apps = activeModules.map((module) => {
    let entry = module.entry_frontend;
    if (!entry) {
      const prefix = import.meta.env.VITE_DEFAULT_ENTRY_PREFIX || '/sub-apps/';
      entry = `${prefix}${module.id}/`;
      console.warn(`[qiankun] 模块 ${module.id} 未指定 entry，使用备选 ${entry}`);
    }
    return {
      name: module.id,
      entry,
      container: '#subapp-container',
      activeRule: `/${module.id}`,
      props: {
        mainRouter: router,
      },
    };
  });

  registerMicroApps(apps, {
    beforeLoad: (app) => {
      console.log(`[qiankun] 开始加载子应用 ${app.name}`);
    },
    afterMount: (app) => {
      console.log(`[qiankun] 子应用 ${app.name} 加载完成`);
    },
    error: (err) => {
      console.error('[qiankun] 子应用加载失败:', err);
      message.error(`子应用加载失败，请检查网络或联系管理员`);
    },
  });

  start({
    prefetch: true,
    sandbox: {
      experimentalStyleIsolation: true,
    },
  });

  actions = initGlobalState({ user: null });
  actions.onGlobalStateChange((state, prev) => {
    console.log('[qiankun] 全局状态变更:', state, prev);
  });

  isRegistered = true;
}

export function getGlobalActions() {
  return actions;
}
