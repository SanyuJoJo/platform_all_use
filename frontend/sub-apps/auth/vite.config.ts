import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import path from 'path';
/**
 * 子应用 Vite 配置模板（以 auth 为例）。
 *
 * 关键变量：
 * - VITE_MODULE_ID：后端模块 ID（qiankun 名称），默认 auth
 * - VITE_SUBAPP_DIR：部署目录名，默认与模块 ID 相同
 * - VITE_PUBLIC_PATH：生产静态资源基础路径，例如 /sub-apps/auth/
 * - VITE_PORT：开发端口
 * - VITE_QIANKUN_USE_DEV_MODE：是否开启 qiankun 开发模式，生产由 build-all.mjs 传 false
 */
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const MODULE_ID = env.VITE_MODULE_ID || 'auth';
  const SUBAPP_DIR = env.VITE_SUBAPP_DIR || 'auth';
  const PORT = parseInt(env.VITE_PORT || '3001', 10);
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
