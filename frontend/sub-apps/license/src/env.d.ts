/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_MODULE_ID: string;
  readonly VITE_SUBAPP_DIR: string;
  readonly VITE_PORT: string;
  readonly VITE_PUBLIC_PATH: string;
  readonly VITE_QIANKUN_USE_DEV_MODE: string;
  readonly VITE_USE_MOCK: string;
  /** 独立运行时的 Mock 权限，逗号分隔 */
  readonly VITE_MOCK_PERMISSIONS: string;
  /** 是否开发模式，由 Vite 注入 */
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
