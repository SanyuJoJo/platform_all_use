import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import path from 'path';
/**
 * audit-log 子应用 Vite 配置。
 *
 * P1-3：MODULE_ID 使用后端模块 ID `audit_log`（下划线），
 * SUBAPP_DIR 使用部署目录名 `audit-log`（短横线）。
 */
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const MODULE_ID = env.VITE_MODULE_ID || 'audit_log';
  const SUBAPP_DIR = env.VITE_SUBAPP_DIR || 'audit-log';
  const PORT = parseInt(env.VITE_PORT || '3003', 10);
  const PUBLIC_PATH =
    env.VITE_PUBLIC_PATH ||
    (mode === 'production' ? `/sub-apps/${SUBAPP_DIR}/` : '/');
  const USE_DEV_MODE =
    mode === 'development' && env.VITE_QIANKUN_USE_DEV_MODE !== 'false';
  console.log(`[Vite][${MODULE_ID}] mode=${mode}`);
  console.log(`[Vite][${MODULE_ID}] publicPath=${PUBLIC_PATH}`);
  console.log(`[Vite][${MODULE_ID}] qiankunDevMode=${USE_DEV_MODE}`);
  return {
    base: PUBLIC_PATH,
    plugins: [vue(), qiankun(MODULE_ID, { useDevMode: USE_DEV_MODE })],
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
