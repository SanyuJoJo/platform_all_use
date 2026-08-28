// sub-apps/_template/vite.config.ts
import { createBaseConfig } from '../../vite.base.config';
import { defineConfig } from 'vite';
export default defineConfig(({ mode }) => {
  const config = createBaseConfig(mode, __dirname);
  return {
    ...config,
    base: process.env.VITE_SUB_APP_PREFIX || '/subapp/',
    server: {
      ...config.server,
      port: 3001, // 各子应用分配不同端口
      host: true,
      headers: {
        'Access-Control-Allow-Origin': '*', // qiankun 跨域支持
      },
    },
    build: {
      ...config.build,
      outDir: 'dist',
      rollupOptions: {
        output: {
          entryFileNames: 'assets/[name]-[hash].js',
        },
      },
    },
  };
});
