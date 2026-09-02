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
  // 从环境变量读取后端地址，若未设置则使用默认值
  const apiTarget = env.VITE_API_BASE_URL || 'http://localhost:8000';

  console.log(`[Vite] Mock enabled: ${useMock}`);
  console.log(`[Vite] API proxy target: ${apiTarget}`);

  return {
    envDir,
    plugins: [
      vue(),
      qiankun('main-app', { useDevMode: true }),
      viteMockServe({
        mockPath: path.resolve(__dirname, 'mock'),
        enable: useMock,          // 仅在需要时启用 Mock
        logger: mode === 'development',
      }),
    ],
    server: {
      host,
      port,
      open: false,
      // 始终配置代理，不依赖 useMock
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
          // 可选：若后端返回的路径不带 /api 前缀，可添加 rewrite
          // rewrite: (path) => path.replace(/^\/api/, ''),
        },
      },
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
