// frontend/sub-apps/crypto-console/src/api/crypto.ts
import axios, { type AxiosRequestConfig } from 'axios';
import request from './index';
import type { OperationResponse } from '@/types';

export interface OperationOptions {
  timeout_ms?: number;
  dry_run?: boolean;
}

export type CAExportType = 'cert' | 'key' | 'pkcs12';

export interface ImportCAParams {
  cert_pem: string;
  key_pem?: string;
  key_password?: string;
  key_ref?: string;
}

export interface CADetail {
  version: string;
  serial: string;
  issuer: string;
  subject: string;
  fingerprint: string;
  not_before: string;
  not_after: string;
  public_key_algorithm: string;
  signature_algorithm: string;
  signature_value: string;
  public_key_value: string;
}

/** 用于文件下载的独立 axios 实例（不走 JSON 拦截器）。 */
const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1';
const rawInstance = axios.create({
  baseURL,
  timeout: 60000,
  withCredentials: true,
});
rawInstance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const cryptoApi = {
  execute<T = unknown>(
    operationId: string,
    params: Record<string, unknown>,
    options?: OperationOptions
  ) {
    return request.post<OperationResponse<T>>(
      `/crypto/operations/${operationId}`,
      {
        params,
        options: {
          timeout_ms: 5000,
          dry_run: false,
          ...options,
        },
      }
    );
  },

  // ==========================================================================
  // CA
  // ==========================================================================

  listCas(params?: Record<string, unknown>) {
    return request.get('/cas', { params });
  },

  getCa(caId: string) {
    return request.get(`/cas/${caId}`);
  },

  /** 查询 CA 证书解析详情 */
  getCADetail(caId: string) {
    return request.get<CADetail>(`/cas/${caId}/detail`);
  },

  importCA(params: ImportCAParams) {
    return request.post('/cas/import', params);
  },

  exportCA(
    caId: string,
    payload: { type: CAExportType; password?: string },
    config?: AxiosRequestConfig
  ) {
    return rawInstance.post(`/cas/${caId}/export`, payload, {
      responseType: 'blob',
      ...config,
    });
  },

  deleteCA(caId: string) {
    return request.delete(`/cas/${caId}`);
  },

  // ==========================================================================
  // 证书 / CSR / CRL / 密钥
  // ==========================================================================

  listCerts(params?: Record<string, unknown>) {
    return request.get('/certs', { params });
  },
  getCert(certId: string) {
    return request.get(`/certs/${certId}`);
  },
  listCsrs(params?: Record<string, unknown>) {
    return request.get('/csrs', { params });
  },
  getCsr(csrId: string) {
    return request.get(`/csrs/${csrId}`);
  },
  listCrls(params?: Record<string, unknown>) {
    return request.get('/crls', { params });
  },
  getCrl(crlId: string) {
    return request.get(`/crls/${crlId}`);
  },
  listKeys(params?: Record<string, unknown>) {
    return request.get('/keys', { params });
  },
  getKey(keyRef: string) {
    return request.get(`/keys/${keyRef}`);
  },
};

export default cryptoApi;
