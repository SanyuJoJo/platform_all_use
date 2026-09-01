import { loadMicroApp, type MicroApp } from 'qiankun';
import type { Module } from '@/types/module';
import { message } from '@/utils/naive';
import type { Router } from 'vue-router';
// 当前加载的子应用实例
let currentApp: MicroApp | null = null;
let currentModuleId: string | null = null;
// 存储所有模块的加载配置（用于热更新）
let moduleMap: Map<string, Module> = new Map();
/**
 * 初始化微前端加载器（单例）
 */
export function initMicroAppLoader() {
  // 空函数，保持接口一致性
}
/**
 * 获取当前加载的子应用模块 ID
 */
export function getCurrentModuleId(): string | null {
  return currentModuleId;
}
/**
 * 判断是否有子应用正在加载
 */
export function isSubAppLoaded(): boolean {
  return currentApp !== null && currentModuleId !== null;
}
/**
 * 注册模块列表（更新模块映射，检测被移除的模块）
 */
export function registerModules(modules: Module[]) {
  const oldModuleIds = Array.from(moduleMap.keys());
  
  moduleMap.clear();
  const newModuleIds: string[] = [];
  modules.forEach((m) => {
    if (m.status === 'active') {
      moduleMap.set(m.id, m);
      newModuleIds.push(m.id);
    }
  });
  // 检测被移除的模块，若当前正在加载则卸载
  const removedModules = oldModuleIds.filter(id => !newModuleIds.includes(id));
  if (removedModules.length > 0) {
    if (import.meta.env.MODE === 'development') {
      console.log('[MicroApp] 检测到模块停用:', removedModules);
    }
    if (currentModuleId && removedModules.includes(currentModuleId)) {
      // 使用 unloadCurrentApp 确保状态清理
      unloadCurrentApp();
    }
  }
  if (import.meta.env.MODE === 'development') {
    console.log('[MicroApp] 已注册模块:', Array.from(moduleMap.keys()));
  }
}
/**
 * 加载子应用
 */
export async function loadSubApp(module: Module, router: Router): Promise<MicroApp> {
  // 如果已有加载的应用且不是目标模块，先卸载
  if (currentApp && currentModuleId !== module.id) {
    await unloadCurrentApp();
  }
  // 如果已经加载了该模块，直接返回
  if (currentApp && currentModuleId === module.id) {
    return currentApp;
  }
  const entry = module.entry_frontend || 
    (import.meta.env.VITE_DEFAULT_ENTRY_PREFIX || '/sub-apps/') + module.id + '/';
  try {
    const app = loadMicroApp({
      name: module.id,
      entry,
      container: '#subapp-container',
      props: {
        mainRouter: router,
      },
    });
    await app.mountPromise;
    currentApp = app;
    currentModuleId = module.id;
    if (import.meta.env.MODE === 'development') {
      console.log(`[MicroApp] 加载子应用 ${module.id} 成功`);
    }
    return app;
  } catch (error) {
    console.error(`[MicroApp] 加载子应用 ${module.id} 失败:`, error);
    message.error(`加载模块 ${module.name} 失败，请检查网络`);
    throw error;
  }
}
/**
 * 卸载当前子应用（清理状态）
 */
export async function unloadCurrentApp() {
  if (!currentApp) return;
  try {
    await currentApp.unmount();
    if (import.meta.env.MODE === 'development') {
      console.log('[MicroApp] 卸载子应用成功');
    }
  } catch (error) {
    console.error('[MicroApp] 卸载子应用失败:', error);
  } finally {
    // ✅ 修复：无论成功与否，清空状态
    currentApp = null;
    currentModuleId = null;
  }
}
/**
 * 刷新模块（热更新）
 * 重新加载模块列表，并更新已加载的子应用
 */
export async function refreshModules(moduleStore: any, router: Router) {
  try {
    const modules = await moduleStore.refreshModules();
    registerModules(modules);
    
    // 如果当前有加载的子应用，检查其模块是否仍存在
    if (currentApp && currentModuleId) {
      const moduleExists = modules.some((m: Module) => m.id === currentModuleId && m.status === 'active');
      if (!moduleExists) {
        // 当前模块已被停用，卸载
        await unloadCurrentApp();
        // 重新触发路由守卫
        router.push(router.currentRoute.value.fullPath);
      } else {
        // 检查入口是否变更
        const newModule = modules.find((m: Module) => m.id === currentModuleId);
        if (newModule && newModule.entry_frontend !== moduleMap.get(currentModuleId)?.entry_frontend) {
          await unloadCurrentApp();
          router.push(router.currentRoute.value.fullPath);
        }
      }
    }
    return modules;
  } catch (error) {
    console.error('[MicroApp] 刷新模块失败:', error);
    throw error;
  }
}
