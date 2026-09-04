// vite.config.ts
import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import qiankun from 'vite-plugin-qiankun';
import path from 'path';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const MODULE_ID = env.VITE_MODULE_ID || 'auth';
  const PORT = parseInt(env.VITE_PORT || '3001', 10);

  return {
    plugins: [
      vue(),
      qiankun(MODULE_ID, { useDevMode: true }),
    ],
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
        'shared': path.resolve(__dirname, '../../shared'),
      },
    },
    build: {
      outDir: 'dist',
    },
    //base: '/', // 明确设为根路径
  };
});