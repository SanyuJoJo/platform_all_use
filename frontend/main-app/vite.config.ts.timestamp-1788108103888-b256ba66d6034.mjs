// vite.config.ts
import { defineConfig, loadEnv } from "file:///code/platform_all_use/frontend/node_modules/.pnpm/vite@5.4.21_@types+node@20.19.43/node_modules/vite/dist/node/index.js";
import vue from "file:///code/platform_all_use/frontend/node_modules/.pnpm/@vitejs+plugin-vue@5.2.4_vite@5.4.21_vue@3.5.42/node_modules/@vitejs/plugin-vue/dist/index.mjs";
import qiankun from "file:///code/platform_all_use/frontend/node_modules/.pnpm/vite-plugin-qiankun@1.0.15_typescript@5.9.3_vite@5.4.21/node_modules/vite-plugin-qiankun/dist/index.js";
import { viteMockServe } from "file:///code/platform_all_use/frontend/node_modules/.pnpm/vite-plugin-mock@3.0.2_esbuild@0.28.2_mockjs@1.1.0_vite@5.4.21/node_modules/vite-plugin-mock/dist/index.mjs";
import path from "path";
var __vite_injected_original_dirname = "/code/platform_all_use/frontend/main-app";
var vite_config_default = defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiTarget = env.VITE_API_BASE_URL || "http://localhost:8000";
  return {
    plugins: [
      vue(),
      qiankun("main-app", { useDevMode: true }),
      viteMockServe({
        mockPath: "mock",
        enable: mode === "development",
        logger: mode === "development"
      })
    ],
    server: {
      host: "0.0.0.0",
      // 允许外部访问
      port: 3e3,
      open: true,
      allowedHosts: true,
      // ⬅️ 添加这一行，允许所有 Host 头
      proxy: {
        "/api": {
          target: apiTarget,
          changeOrigin: true
          // 后端路由以 /api/v1 开头，无需 rewrite
        }
      }
    },
    resolve: {
      alias: {
        "@": "/src",
        "shared": path.resolve(__vite_injected_original_dirname, "../shared")
      }
    },
    build: {
      outDir: "dist",
      rollupOptions: {
        output: {
          manualChunks: {
            vue: ["vue", "vue-router", "pinia"],
            naive: ["naive-ui"]
          }
        }
      }
    }
  };
});
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcudHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCIvY29kZS9wbGF0Zm9ybV9hbGxfdXNlL2Zyb250ZW5kL21haW4tYXBwXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ZpbGVuYW1lID0gXCIvY29kZS9wbGF0Zm9ybV9hbGxfdXNlL2Zyb250ZW5kL21haW4tYXBwL3ZpdGUuY29uZmlnLnRzXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ltcG9ydF9tZXRhX3VybCA9IFwiZmlsZTovLy9jb2RlL3BsYXRmb3JtX2FsbF91c2UvZnJvbnRlbmQvbWFpbi1hcHAvdml0ZS5jb25maWcudHNcIjtpbXBvcnQgeyBkZWZpbmVDb25maWcsIGxvYWRFbnYgfSBmcm9tICd2aXRlJztcbmltcG9ydCB2dWUgZnJvbSAnQHZpdGVqcy9wbHVnaW4tdnVlJztcbmltcG9ydCBxaWFua3VuIGZyb20gJ3ZpdGUtcGx1Z2luLXFpYW5rdW4nO1xuaW1wb3J0IHsgdml0ZU1vY2tTZXJ2ZSB9IGZyb20gJ3ZpdGUtcGx1Z2luLW1vY2snO1xuaW1wb3J0IHBhdGggZnJvbSAncGF0aCc7XG5leHBvcnQgZGVmYXVsdCBkZWZpbmVDb25maWcoKHsgbW9kZSB9KSA9PiB7XG4gIGNvbnN0IGVudiA9IGxvYWRFbnYobW9kZSwgcHJvY2Vzcy5jd2QoKSwgJycpO1xuICBjb25zdCBhcGlUYXJnZXQgPSBlbnYuVklURV9BUElfQkFTRV9VUkwgfHwgJ2h0dHA6Ly9sb2NhbGhvc3Q6ODAwMCc7XG4gIHJldHVybiB7XG4gICAgcGx1Z2luczogW1xuICAgICAgdnVlKCksXG4gICAgICBxaWFua3VuKCdtYWluLWFwcCcsIHsgdXNlRGV2TW9kZTogdHJ1ZSB9KSxcbiAgICAgIHZpdGVNb2NrU2VydmUoe1xuICAgICAgICBtb2NrUGF0aDogJ21vY2snLFxuICAgICAgICBlbmFibGU6IG1vZGUgPT09ICdkZXZlbG9wbWVudCcsXG4gICAgICAgIGxvZ2dlcjogbW9kZSA9PT0gJ2RldmVsb3BtZW50JyxcbiAgICAgIH0pLFxuICAgIF0sXG4gICAgc2VydmVyOiB7XG4gICAgICBob3N0OiAnMC4wLjAuMCcsICAgLy8gXHU1MTQxXHU4QkI4XHU1OTE2XHU5MEU4XHU4QkJGXHU5NUVFXG4gICAgICBwb3J0OiAzMDAwLFxuICAgICAgb3BlbjogdHJ1ZSxcbiAgICAgIGFsbG93ZWRIb3N0czogdHJ1ZSwgICAvLyBcdTJCMDVcdUZFMEYgXHU2REZCXHU1MkEwXHU4RkQ5XHU0RTAwXHU4ODRDXHVGRjBDXHU1MTQxXHU4QkI4XHU2MjQwXHU2NzA5IEhvc3QgXHU1OTM0XG4gICAgICBwcm94eToge1xuICAgICAgICAnL2FwaSc6IHtcbiAgICAgICAgICB0YXJnZXQ6IGFwaVRhcmdldCxcbiAgICAgICAgICBjaGFuZ2VPcmlnaW46IHRydWUsXG4gICAgICAgICAgLy8gXHU1NDBFXHU3QUVGXHU4REVGXHU3NTMxXHU0RUU1IC9hcGkvdjEgXHU1RjAwXHU1OTM0XHVGRjBDXHU2NUUwXHU5NzAwIHJld3JpdGVcbiAgICAgICAgfSxcbiAgICAgIH0sXG4gICAgfSxcbiAgICByZXNvbHZlOiB7XG4gICAgICBhbGlhczoge1xuICAgICAgICAnQCc6ICcvc3JjJyxcbiAgICAgICAgJ3NoYXJlZCc6IHBhdGgucmVzb2x2ZShfX2Rpcm5hbWUsICcuLi9zaGFyZWQnKSxcbiAgICAgIH0sXG4gICAgfSxcbiAgICBidWlsZDoge1xuICAgICAgb3V0RGlyOiAnZGlzdCcsXG4gICAgICByb2xsdXBPcHRpb25zOiB7XG4gICAgICAgIG91dHB1dDoge1xuICAgICAgICAgIG1hbnVhbENodW5rczoge1xuICAgICAgICAgICAgdnVlOiBbJ3Z1ZScsICd2dWUtcm91dGVyJywgJ3BpbmlhJ10sXG4gICAgICAgICAgICBuYWl2ZTogWyduYWl2ZS11aSddLFxuICAgICAgICAgIH0sXG4gICAgICAgIH0sXG4gICAgICB9LFxuICAgIH0sXG4gIH07XG59KTtcbiJdLAogICJtYXBwaW5ncyI6ICI7QUFBMFMsU0FBUyxjQUFjLGVBQWU7QUFDaFYsT0FBTyxTQUFTO0FBQ2hCLE9BQU8sYUFBYTtBQUNwQixTQUFTLHFCQUFxQjtBQUM5QixPQUFPLFVBQVU7QUFKakIsSUFBTSxtQ0FBbUM7QUFLekMsSUFBTyxzQkFBUSxhQUFhLENBQUMsRUFBRSxLQUFLLE1BQU07QUFDeEMsUUFBTSxNQUFNLFFBQVEsTUFBTSxRQUFRLElBQUksR0FBRyxFQUFFO0FBQzNDLFFBQU0sWUFBWSxJQUFJLHFCQUFxQjtBQUMzQyxTQUFPO0FBQUEsSUFDTCxTQUFTO0FBQUEsTUFDUCxJQUFJO0FBQUEsTUFDSixRQUFRLFlBQVksRUFBRSxZQUFZLEtBQUssQ0FBQztBQUFBLE1BQ3hDLGNBQWM7QUFBQSxRQUNaLFVBQVU7QUFBQSxRQUNWLFFBQVEsU0FBUztBQUFBLFFBQ2pCLFFBQVEsU0FBUztBQUFBLE1BQ25CLENBQUM7QUFBQSxJQUNIO0FBQUEsSUFDQSxRQUFRO0FBQUEsTUFDTixNQUFNO0FBQUE7QUFBQSxNQUNOLE1BQU07QUFBQSxNQUNOLE1BQU07QUFBQSxNQUNOLGNBQWM7QUFBQTtBQUFBLE1BQ2QsT0FBTztBQUFBLFFBQ0wsUUFBUTtBQUFBLFVBQ04sUUFBUTtBQUFBLFVBQ1IsY0FBYztBQUFBO0FBQUEsUUFFaEI7QUFBQSxNQUNGO0FBQUEsSUFDRjtBQUFBLElBQ0EsU0FBUztBQUFBLE1BQ1AsT0FBTztBQUFBLFFBQ0wsS0FBSztBQUFBLFFBQ0wsVUFBVSxLQUFLLFFBQVEsa0NBQVcsV0FBVztBQUFBLE1BQy9DO0FBQUEsSUFDRjtBQUFBLElBQ0EsT0FBTztBQUFBLE1BQ0wsUUFBUTtBQUFBLE1BQ1IsZUFBZTtBQUFBLFFBQ2IsUUFBUTtBQUFBLFVBQ04sY0FBYztBQUFBLFlBQ1osS0FBSyxDQUFDLE9BQU8sY0FBYyxPQUFPO0FBQUEsWUFDbEMsT0FBTyxDQUFDLFVBQVU7QUFBQSxVQUNwQjtBQUFBLFFBQ0Y7QUFBQSxNQUNGO0FBQUEsSUFDRjtBQUFBLEVBQ0Y7QUFDRixDQUFDOyIsCiAgIm5hbWVzIjogW10KfQo=
