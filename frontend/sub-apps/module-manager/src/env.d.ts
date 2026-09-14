/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_MODULE_ID: string;
  readonly VITE_SUBAPP_DIR: string;
  readonly VITE_PORT: string;
  readonly VITE_PUBLIC_PATH: string;
  readonly VITE_QIANKUN_USE_DEV_MODE: string;
  readonly VITE_USE_MOCK: string;
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

export {};
