import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
export function createBaseConfig(mode: string, dirname: string) {
  const env = loadEnv(mode, process.cwd(), '');
  return defineConfig({
    plugins: [vue()],
    resolve: {
      alias: {
        '@': resolve(dirname, 'src'),
        '@shared': resolve(__dirname, 'shared'),
      },
    },
    css: {
      preprocessorOptions: {
        scss: { additionalData: `@import "@/styles/variables.scss";` },
      },
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL || 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, ''),
        },
      },
    },
    build: {
      target: 'es2020',
      minify: 'terser',
      sourcemap: false,
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ['vue', 'vue-router', 'pinia'],
          },
        },
      },
    },
  });
}
