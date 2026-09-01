import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import { viteMockServe } from 'vite-plugin-mock';
import path from 'path';

export default defineConfig(({ mode }) => {
  const envDir = path.resolve(__dirname, '..');
  const env = loadEnv(mode, envDir, '');
  const useMock = env.VITE_USE_MOCK === 'true';
  const host = env.VITE_APP_HOST || '0.0.0.0';
  const port = parseInt(env.VITE_APP_PORT || '3000', 10);

  console.log(`[Vite] Mock enabled: ${useMock}`);

  return {
    envDir,
    plugins: [
      vue(),
      qiankun('main-app', { useDevMode: true }),
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
      // 仅当启用 Mock 时才使用代理，否则请求直接发送到后端（由 axios baseURL 控制）
      proxy: useMock ? {
        '/api': {
          target: 'http://localhost:8000', // 任意地址，Mock 会拦截
          changeOrigin: true,
        },
      } : undefined,
    },
    resolve: {
      alias: {
        '@': '/src',
        'shared': path.resolve(__dirname, '../shared'),
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
