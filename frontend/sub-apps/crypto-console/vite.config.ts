import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import path from 'path';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const MODULE_ID = env.VITE_MODULE_ID || 'crypto_console';
  const SUBAPP_DIR = env.VITE_SUBAPP_DIR || 'crypto-console';
  const PORT = parseInt(env.VITE_PORT || '3010', 10);
  const PUBLIC_PATH =
    env.VITE_PUBLIC_PATH ||
    (mode === 'production' ? `/sub-apps/${SUBAPP_DIR}/` : '/');

  // 只有以下两种情况才启用 qiankun 插件：
  //   1) 显式声明 QIANKUN_USE_DEV_MODE=true（挂主应用联调）
  //   2) 生产构建（需要 qiankun 生命周期入口）
  // 独立开发（pnpm dev）默认不加载，避免干扰 SFC 编译
  const USE_DEV_MODE =
    mode === 'development' && env.VITE_QIANKUN_USE_DEV_MODE === 'true';
  const NEED_QIANKUN = USE_DEV_MODE || mode === 'production';

  const plugins: any[] = [vue()];
  if (NEED_QIANKUN) {
    plugins.push(qiankun(MODULE_ID, { useDevMode: USE_DEV_MODE }));
  }

  console.log(
    `[vite] mode=${mode} needQiankun=${NEED_QIANKUN} useDevMode=${USE_DEV_MODE}`
  );

  return {
    base: PUBLIC_PATH,
    plugins,
    define: {
      __MODULE_ID__: JSON.stringify(MODULE_ID),
    },
    server: {
      host: '0.0.0.0',
      cors: true,
      port: PORT,
      headers: {
        'Access-Control-Allow-Origin': '*',
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
        shared: path.resolve(__dirname, '../../shared'),
      },
    },
    build: {
      outDir: 'dist',
    },
  };
});
