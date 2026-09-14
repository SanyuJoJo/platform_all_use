import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import { viteMockServe } from 'vite-plugin-mock';
import path from 'path';
/**
 * 主应用 Vite 配置
 *
 * 说明：
 * - 主应用为 qiankun 基座，不需要 vite-plugin-qiankun（该插件用于子应用导出生命周期）
 * - 生产构建通过 `pnpm --filter main-app build` 调用，环境变量由 build-all.mjs 注入
 */
export default defineConfig(({ mode }) => {
  const envDir = path.resolve(__dirname, '..');
  const env = loadEnv(mode, envDir, '');
  const useMock = env.VITE_USE_MOCK === 'true';
  const host = env.VITE_APP_HOST || '0.0.0.0';
  const port = parseInt(env.VITE_APP_PORT || '3000', 10);
  const apiTarget = env.VITE_API_BASE_URL || 'http://localhost:8000';
  const base = env.VITE_BASE_PATH || '/';
  console.log(`[Vite][main-app] mode=${mode}`);
  console.log(`[Vite][main-app] base=${base}`);
  console.log(`[Vite][main-app] Mock enabled: ${useMock}`);
  console.log(`[Vite][main-app] API proxy target: ${apiTarget}`);
  return {
    envDir,
    base,
    plugins: [
      vue(),
      viteMockServe({
        mockPath: path.resolve(__dirname, 'mock'),
        enable: useMock,
        logger: mode === 'development',
      }),
    ],
    server: {
      host,
      port,
      open: false,
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
        },
      },
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
        shared: path.resolve(__dirname, '../shared'),
      },
    },
    build: {
      outDir: 'dist',
      rollupOptions: {
        output: {
          manualChunks: {
            vue: ['vue', 'vue-router', 'pinia'],
            naive: ['naive-ui'],
          },
        },
      },
    },
  };
});
