// main-app/src/main.ts
import { createApp } from 'vue';
import { createPinia } from 'pinia';
import NaiveUI from 'naive-ui';
import App from './App.vue';
import router from './router';
import { setupQiankun } from './micro-frontend';
import { setupGlobalComponents } from './components';
import { setupErrorHandler } from './utils/error';
const app = createApp(App);
// 插件注册
app.use(createPinia());
app.use(router);
app.use(NaiveUI);
// 全局组件注册
setupGlobalComponents(app);
// 全局错误处理
setupErrorHandler(app);
// 挂载应用
app.mount('#app');
// 启动 qiankun（在 DOM 渲染完成后）
setupQiankun();
