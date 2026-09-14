/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_APP_HOST: string;
  readonly VITE_APP_PORT: string;
  readonly VITE_USE_MOCK: string;
  readonly VITE_BASE_PATH: string;
  readonly VITE_SUBAPP_BASE_PREFIX: string;
  readonly VITE_DEFAULT_ENTRY_PREFIX: string;
  readonly VITE_QIANKUN_USE_DEV_MODE: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
