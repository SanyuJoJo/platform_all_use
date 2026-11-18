/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_MODULE_ID: string;
  readonly VITE_SUBAPP_DIR: string;
  readonly VITE_PORT: string;
  readonly VITE_PUBLIC_PATH: string;
  readonly VITE_QIANKUN_USE_DEV_MODE: string;
  readonly VITE_ZLM_BASE_URL: string;
  readonly DEV: boolean;
  readonly PROD: boolean;
  readonly MODE: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare global {
  interface Window {
    __POWERED_BY_QIANKUN__?: boolean;
  }
}

// ★ 新增：声明 flv.js 模块，欺骗 TypeScript
declare module 'flv.js' {
  const flvjs: any;
  export default flvjs;
}

export {};
