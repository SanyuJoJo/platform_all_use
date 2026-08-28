

import { createBaseConfig } from '../vite.base.config';
import { defineConfig } from 'vite';
export default defineConfig(({ mode }) => {
  const config = createBaseConfig(mode, __dirname);
  return {
    ...config,
    server: {
      ...config.server,
      port: 3000,
      host: true,
    },
    build: {
      ...config.build,
      outDir: 'dist',
    },
  };
});
