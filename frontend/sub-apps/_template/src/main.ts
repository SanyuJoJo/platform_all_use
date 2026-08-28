// sub-apps/_template/src/main.ts
import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
let app: any = null;
/** 渲染函数 */
function render(props: any = {}) {
  const { container } = props;
  app = createApp(App);
  app.use(createPinia());
  app.use(router);
  app.mount(container ? container.querySelector('#app') : '#app');
}
/** 独立运行（非 qiankun 环境） */
if (!window.__POWERED_BY_QIANKUN__) {
  render();
}
/** qiankun 生命周期：bootstrap */
export async function bootstrap() {
  console.log('[子应用] bootstrap');
}
/** qiankun 生命周期：mount */
export async function mount(props: any) {
  console.log('[子应用] mount', props);
  render(props);
}
/** qiankun 生命周期：unmount */
export async function unmount() {
  console.log('[子应用] unmount');
  app?.unmount();
  app = null;
}
