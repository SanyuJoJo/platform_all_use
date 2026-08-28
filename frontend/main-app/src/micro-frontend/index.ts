// main-app/src/micro-frontend/index.ts
import { registerApps, startQiankun } from './register';
import { useAppStore } from '@/store/app';
export async function setupQiankun() {
  const appStore = useAppStore();
  
  // 从后端获取已安装且启用的模块列表
  await appStore.fetchModules();
  
  const activeModules = appStore.activeModules;
  const apps = activeModules.map((module) => ({
    name: module.code,
    entry: module.entryUrl,
    activeRule: module.activeRule, // 如 '/auth', '/module-manager'
  }));
  
  registerApps(apps);
  startQiankun();
}
/** 动态添加新模块（模块安装后调用） */
export function addModule(module: ModuleConfig) {
  // 重新注册所有应用（qiankun 支持动态添加）
  // 或调用 loadMicroApp 手动加载
}
