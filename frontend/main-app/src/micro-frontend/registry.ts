import { registerMicroApps, start, initGlobalState, type MicroAppStateActions } from 'qiankun';
import type { Router } from 'vue-router';
import type { Module } from '@/types/module';
import { message } from '@/utils/naive';

let actions: MicroAppStateActions | null = null;
let isStarted = false;
let lastModules: Module[] = [];
let styleBackup: { head: string; body: string } | null = null;

export function registerModules(modules: Module[], router: Router) {
  if (!modules || modules.length === 0) {
    console.warn('[qiankun] 没有模块可注册');
    return;
  }

  const activeModules = modules.filter(m => m.status === 'active');
  if (activeModules.length === 0) {
    console.warn('[qiankun] 没有激活的模块');
    return;
  }

  lastModules = modules;

  const apps = activeModules.map(module => ({
    name: module.id,
    entry: module.entry_frontend || `//localhost:${module.id === 'auth' ? 3001 : 3002}`,
    container: '#subapp-container',
    activeRule: `/${module.id}`,
    props: { mainRouter: router, moduleId: module.id },
  }));

  registerMicroApps(apps, {
    beforeLoad: app => {
      console.log(`[qiankun] before load ${app.name}`);
      // 备份主应用 head 和 body 的样式属性（以防子应用修改）
      const appEl = document.getElementById('app');
      if (appEl) {
        styleBackup = {
          head: document.head.innerHTML,
          body: document.body.style.cssText,
        };
        // 强制重置主应用根元素样式
        appEl.style.setProperty('background', '#fff', 'important');
        appEl.style.setProperty('color', '#000', 'important');
      }
    },
    afterMount: app => {
      console.log(`[qiankun] after mount ${app.name}`);
      // 延迟执行，确保子应用所有样式加载完毕
      setTimeout(() => {
        // 恢复主应用样式（但保留必要的背景色）
        const appEl = document.getElementById('app');
        if (appEl) {
          appEl.style.background = '#fff';
          appEl.style.color = '#000';
        }
        // 强制修复菜单和顶部（也可通过 App.vue 的 observer 完成）
        const header = document.querySelector('.main-header') as HTMLElement;
        const sider = document.querySelector('.main-sider') as HTMLElement;
        if (header) {
          header.style.setProperty('background', '#fff', 'important');
          header.style.setProperty('color', '#000', 'important');
        }
        if (sider) {
          sider.style.setProperty('background', '#fff', 'important');
          sider.style.setProperty('color', '#000', 'important');
        }
        document.querySelectorAll('.n-menu-item-content').forEach(el => {
          (el as HTMLElement).style.setProperty('color', '#000', 'important');
        });
      }, 300);
    },
    error: err => {
      console.error('[qiankun] 子应用加载失败:', err);
      message.error('子应用加载失败，请检查网络');
    },
  });

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
}

export function reRegister(router: Router) {
  if (lastModules.length === 0) {
    console.warn('[qiankun] 没有缓存的模块，跳过重新注册');
    return;
  }
  window.location.reload();
}

export function getGlobalActions() {
  return actions;
}