// main-app/src/micro-frontend/register.ts
import { registerMicroApps, start, addGlobalUncaughtErrorHandler } from 'qiankun';
import type { MicroAppConfig } from '@/api/modules';
/** 注册子应用 */
export function registerApps(apps: MicroAppConfig[]) {
  registerMicroApps(
    apps.map((app) => ({
      name: app.name,                    // 子应用唯一标识
      entry: app.entry,                  // 子应用入口 URL
      container: '#subapp-container',    // 挂载容器
      activeRule: app.activeRule,        // 激活规则
      props: {
        // 传递给子应用的数据
        getGlobalState: () => store.state,
        setGlobalState: (state) => store.setState(state),
        userInfo: store.user.userInfo,
        permissions: store.permission.permissions,
      },
    })),
    {
      // 全局生命周期钩子
      beforeLoad: [(app) => console.log('[qiankun] beforeLoad', app.name)],
      beforeMount: [(app) => console.log('[qiankun] beforeMount', app.name)],
      afterMount: [(app) => console.log('[qiankun] afterMount', app.name)],
      beforeUnmount: [(app) => console.log('[qiankun] beforeUnmount', app.name)],
      afterUnmount: [(app) => console.log('[qiankun] afterUnmount', app.name)],
    }
  );
  // 全局错误处理
  addGlobalUncaughtErrorHandler((event) => {
    console.error('[qiankun] 子应用加载失败:', event);
    // 展示友好错误提示
  });
}
/** 启动 qiankun */
export function startQiankun() {
  start({
    prefetch: true,              // 开启预加载
    sandbox: {                   // 沙箱配置
      strictStyleIsolation: false, // 不使用严格样式隔离（配合 Naive UI）
      experimentalStyleIsolation: true, // 启用实验性样式隔离
    },
  });
}
