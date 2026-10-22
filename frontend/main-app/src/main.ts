import { createApp } from 'vue';
import { createPinia } from 'pinia';
import naive from 'naive-ui';
import App from './App.vue';
import router from './router';

console.log('[Main] 应用启动');

const app = createApp(App);
app.use(createPinia());
app.use(router);
app.use(naive);

(async () => {
  console.log('[Main] 等待 router.isReady()...');

  // ★ 关键：app.use(router) 已触发首次导航，守卫里会完成
  //   - 模块列表加载
  //   - 权限路由拉取 + 注册
  //   - 菜单构建
  //   - 子应用 catch-all 注册
  //   全部完成后 isReady() 才 resolve
  await router.isReady();

  console.log(
    '[Main] 路由就绪 | currentRoute =', router.currentRoute.value.fullPath,
    '| name =', String(router.currentRoute.value.name),
    '| matched =', router.currentRoute.value.matched.map((m) => String(m.name)).join('/')
  );

  app.mount('#app');
  console.log('[Main] 应用挂载完成');

  // qiankun 子应用注册在 Layout 的 onMounted 中完成（依赖 #subapp-container 已渲染）
})();
