/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue';
  const component: DefineComponent<{}, {}, any>;
  export default component;
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_MODULE_ID: string;
  readonly VITE_SUBAPP_DIR: string;
  readonly VITE_PORT: string;
  readonly VITE_PUBLIC_PATH: string;
  readonly VITE_MOCK_PERMISSIONS: string;
  readonly VITE_TASK_POLL_INTERVAL_MS: string;
  readonly VITE_QIANKUN_USE_DEV_MODE: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare const __MODULE_ID__: string;

// qiankun 运行时注入的全局标记
declare interface Window {
  __POWERED_BY_QIANKUN__?: boolean;
  __POWERED_BY_QIANKUN__: boolean;
  __INJECTED_PUBLIC_PATH_BY_QIANKUN__?: string;
}
