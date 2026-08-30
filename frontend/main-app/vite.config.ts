import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import { viteMockServe } from 'vite-plugin-mock';
import path from 'path';
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const apiTarget = env.VITE_API_BASE_URL || 'http://localhost:8000';
  return {
    plugins: [
      vue(),
      qiankun('main-app', { useDevMode: true }),
      viteMockServe({
        mockPath: 'mock',
        enable: mode === 'development',
        logger: mode === 'development',
      }),
    ],
    server: {
      host: '0.0.0.0',   // 允许外部访问
      port: 3000,
      open: true,
      allowedHosts: true,   // ⬅️ 添加这一行，允许所有 Host 头
      proxy: {
        '/api': {
          target: apiTarget,
          changeOrigin: true,
          // 后端路由以 /api/v1 开头，无需 rewrite
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
