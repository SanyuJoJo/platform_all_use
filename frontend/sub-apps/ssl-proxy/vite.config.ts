import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import path from 'path';
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const MODULE_ID = env.VITE_MODULE_ID || 'ssl_proxy';
  const SUBAPP_DIR = env.VITE_SUBAPP_DIR || 'ssl-proxy';
  const PORT = parseInt(env.VITE_PORT || '3011', 10);
  const PUBLIC_PATH =
    env.VITE_PUBLIC_PATH ||
    (mode === 'production' ? `/sub-apps/${SUBAPP_DIR}/` : '/');
  const USE_DEV_MODE =
    mode === 'development' && env.VITE_QIANKUN_USE_DEV_MODE !== 'false';
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
      },
    },
    build: {
      outDir: 'dist',
    },
  };
});
