// frontend/sub-apps/crypto-console/src/api/crypto.ts
import axios, { type AxiosRequestConfig } from 'axios';
import request from './index';
import type { OperationResponse } from '@/types';

// =============================================================================
// 类型定义
// =============================================================================

export interface OperationOptions {
  timeout_ms?: number;
  dry_run?: boolean;
}

export type CAExportType = 'cert' | 'key' | 'pkcs12';
export type CertExportType = 'cert' | 'key' | 'pkcs12';

export interface ImportCAParams {
  cert_pem: string;
  key_pem?: string;
  key_password?: string;
  key_ref?: string;
}

export interface ImportCertParams {
  cert_pem: string;
  key_pem?: string;
  key_password?: string;
  key_ref?: string;
}

export interface ImportCsrParams {
  csr_pem: string;
  key_pem?: string;
  key_password?: string;
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
  key_usage?: string;
  extended_key_usage?: string;
}

export type CertDetail = CADetail;

export interface CAListItem {
  ca_id: string;
  subject_cn: string;
  subject?: string;
  algorithm: string;
  cert_path: string;
  status: string;
}

export type CertTypeItem = 'server' | 'client' | 'signature' | 'encryption';

export interface SignCertParams {
  ca_source: 'local' | 'manual';
  ca_id?: string;
  ca_cert_pem?: string;
  ca_key_pem?: string;
  ca_key_password?: string;

  csr_source?: 'existing' | 'upload';
  csr_id?: string;
  csr_pem?: string;
  csr_key_pem?: string;
  csr_key_password?: string;

  algorithm?: string;
  key_params?: Record<string, unknown>;
  subject?: Record<string, string>;
  san?: string[];
  cert_types?: CertTypeItem[];
  validity_days: number;
  cert_mode?: 'normal' | 'dual';

  return_key?: boolean;
  key_export_password?: string;
}

export interface SignCertResult {
  certificate?: Record<string, unknown>;
  cert_pem?: string;
  key_pem?: string;
  sign_cert_pem?: string;
  enc_cert_pem?: string;
  encrypted_envelope?: string;
}

/**
 * 查询信封信息请求。
 *
 * mode：
 *   - cert   ：从平台已签发的签名证书反查 key_ref，自动用签名私钥解密（推荐）
 *   - manual ：手动粘贴签名私钥 PEM 解密
 *
 * 两种模式都必须提供 encrypted_envelope。
 */
export interface QueryEnvelopeParams {
  format: 'pkcs10' | 'cfca';
  mode?: 'cert' | 'manual';
  encrypted_envelope: string;

  // mode=cert
  cert_id?: string;

  // mode=manual
  sign_key_pem?: string;
  sign_key_password?: string;
}

export interface QueryEnvelopeResult {
  format: string;
  mode: string;
  symmetric_key_cipher: string;
  iv: string;
  encrypted_private_key: string;
  decrypted_key_pem?: string;
}

// =============================================================================
// 文件下载用 raw 实例
// =============================================================================

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

// =============================================================================
// cryptoApi
// =============================================================================

export const cryptoApi = {
  execute<T = unknown>(
    operationId: string,
    params: Record<string, unknown>,
    options?: OperationOptions
  ) {
    return request.post<OperationResponse<T>>(
      `/crypto/operations/${operationId}`,
      { params, options: { timeout_ms: 5000, dry_run: false, ...options } }
    );
  },

  listCas(params?: Record<string, unknown>) {
    return request.get('/cas', { params });
  },
  getCa(caId: string) {
    return request.get(`/cas/${caId}`);
  },
  getCADetail(caId: string) {
    return request.get<CADetail>(`/cas/${caId}/detail`);
  },
  importCA(params: ImportCAParams) {
    return request.post('/cas/import', params);
  },
  exportCA(caId: string, payload: { type: CAExportType; password?: string }, config?: AxiosRequestConfig) {
    return rawInstance.post(`/cas/${caId}/export`, payload, { responseType: 'blob', ...config });
  },
  deleteCA(caId: string) {
    return request.delete(`/cas/${caId}`);
  },

  listCerts(params?: Record<string, unknown>) {
    return request.get('/certs', { params });
  },
  getCert(certId: string) {
    return request.get(`/certs/${certId}`);
  },
  getCertDetail(certId: string) {
    return request.get<CertDetail>(`/certs/${certId}/detail`);
  },
  signCert(payload: SignCertParams) {
    return request.post<SignCertResult>('/certs/sign', payload);
  },
  importCert(payload: ImportCertParams) {
    return request.post('/certs/import', payload);
  },
  queryEnvelope(payload: QueryEnvelopeParams) {
    return request.post<QueryEnvelopeResult>('/certs/envelope/query', payload);
  },
  exportCert(certId: string, payload: { type: CertExportType; password?: string }, config?: AxiosRequestConfig) {
    return rawInstance.post(`/certs/${certId}/export`, payload, { responseType: 'blob', ...config });
  },
  deleteCert(certId: string) {
    return request.delete(`/certs/${certId}`);
  },

  listCsrs(params?: Record<string, unknown>) {
    return request.get('/csrs', { params });
  },
  getCsr(csrId: string) {
    return request.get(`/csrs/${csrId}`);
  },
  importCsr(payload: ImportCsrParams) {
    return request.post('/csrs/import', payload);
  },
  downloadCsr(csrId: string) {
    return rawInstance.get(`/csrs/${csrId}/download`, { responseType: 'blob' });
  },
  downloadCsrKey(csrId: string, encrypt = false, password = '') {
    const params: Record<string, string> = {};
    if (encrypt) {
      params.encrypt = 'true';
      if (password) params.password = password;
    }
    return rawInstance.get(`/csrs/${csrId}/key`, { params, responseType: 'blob' });
  },
  deleteCsr(csrId: string) {
    return request.delete(`/csrs/${csrId}`);
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
