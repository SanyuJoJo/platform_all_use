import { registerMicroApps, start, initGlobalState, type MicroAppStateActions } from 'qiankun';
import type { Router } from 'vue-router';
import { SUB_APP_PREFIXES } from './constants';
let actions: MicroAppStateActions | null = null;
/**
 * 初始化 qiankun 微前端
 *
 * 生产环境建议从后端动态获取已安装且启用的模块列表，示例实现：
 * 
 * import { getActiveModules } from '@/api/module';
 * 
 * async function registerDynamicModules(router: Router) {
 *   const modules = await getActiveModules();
 *   const apps = modules.map(mod => ({
 *     name: mod.id,
 *     entry: mod.entry_frontend,
 *     container: '#subapp-container',
 *     activeRule: `/${mod.id}`,
 *     props: { mainRouter: router },
 *   }));
 *   registerMicroApps(apps);
 * }
 */
export function setupQiankun(router: Router) {
  // 示例：静态列表（开发测试用）
  const modules = [
    {
      name: 'auth',
      entry: import.meta.env.VITE_SUBAPP_ENTRY_AUTH || '//localhost:3001/',
      container: '#subapp-container',
      activeRule: '/auth',
      props: { mainRouter: router },
    },
    // 更多子应用可在此添加，或从后端动态获取
  ];
  registerMicroApps(modules, {
    beforeLoad: (app) => console.log(`[qiankun] before load ${app.name}`),
    afterMount: (app) => console.log(`[qiankun] after mount ${app.name}`),
  });
  actions = initGlobalState({ user: null });
  actions.onGlobalStateChange((state, prev) => {
    console.log('global state changed', state, prev);
  });
  start({ prefetch: true });
}
export function getGlobalActions() {
  return actions;
}
