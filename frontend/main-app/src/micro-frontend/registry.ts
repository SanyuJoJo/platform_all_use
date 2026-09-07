// main-app/src/micro-frontend/registry.ts

import { registerMicroApps, start, initGlobalState, type MicroAppStateActions } from 'qiankun';
import type { Router } from 'vue-router';
import type { Module } from '@/types/module';
import { message } from '@/utils/naive';
import { useUserStore } from '@/store/user';

let actions: MicroAppStateActions | null = null;
let isStarted = false;
let isRegistered = false; // 防止重复注册
let lastModules: Module[] = [];
let styleBackup: { head: string; body: string } | null = null;

/**
 * 确保子应用容器存在
 */
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

  // 防止重复注册
  if (isRegistered) {
    console.warn('[qiankun] 子应用已注册，跳过');
    return;
  }

  // 确保容器存在
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

  const apps = activeModules.map(module => ({
    name: module.id,
    entry: module.entry_frontend || `//${window.location.hostname}:${module.id === 'auth' ? 3001 : 3002}`,
    container: '#subapp-container',
    activeRule: `/${module.id}`,
    props: {
      mainRouter: router,
      moduleId: module.id,
      permissions: permissions,
      user: user,
      token: token,
    },
  }));

  registerMicroApps(apps, {
    beforeLoad: app => {
      console.log(`[qiankun] before load ${app.name}`);
      ensureContainer();
      const appEl = document.getElementById('app');
      if (appEl) {
        styleBackup = {
          head: document.head.innerHTML,
          body: document.body.style.cssText,
        };
        appEl.style.setProperty('background', '#fff', 'important');
        appEl.style.setProperty('color', '#000', 'important');
      }
    },
    afterMount: app => {
      console.log(`[qiankun] after mount ${app.name}`);
      setTimeout(() => {
        const appEl = document.getElementById('app');
        if (appEl) {
          appEl.style.background = '#fff';
          appEl.style.color = '#000';
        }
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

  isRegistered = true;
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