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
  domain?: string;
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
  public_key_algorithm: string;
  algorithm?: string;
  cert_path: string;
  status: string;
  has_private_key?: boolean;
  /** ★ 域 */
  domain?: string;
  /** ★ 生效时间（北京时间字符串） */
  not_before?: string;
  /** ★ 失效时间（北京时间字符串） */
  not_after?: string;
}

export interface CreateDomainResponse {
  name: string;
}

export interface DomainListResponse {
  domains: string[];
}

// ★ 证书类型只保留签名 / 加密两类
export type CertTypeItem = 'signature' | 'encryption';

// =============================================================================
// ★ CSR 申请（普通 / CFCA）
// =============================================================================

export type P10Format = 'pkcs10' | 'cfca';

export interface CSRCreateParams {
  algorithm?: string;
  subject: Record<string, string>;
  key_source?: 'generate';
  /** ★ 新增：CFCA P10 会在 CSR 中携带临时公钥 attribute */
  p10_format?: P10Format;
  san?: string[];
}

export interface CSRCreateResult {
  csr_id?: string;
  csr_pem?: string;
  /** @deprecated 平台内置固定临时密钥，不再逐次返回；保留字段兼容旧后端 */
  tmp_key_pem?: string;
  /** 可选：签名私钥（若平台在本次请求中可提供） */
  sign_key_pem?: string;
  /** 可选：平台生成的签名私钥的封装形式（视后端实现） */
  csr_key_pem?: string;
}

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
  domain?: string;

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
  /** ★ CFCA 双证专用（后端 cfca/sign 返回） */
  tmp_key_pem?: string;
  sign_key_pem?: string;
  sign_cert_id?: string;
  enc_cert_id?: string;
}

export interface QueryEnvelopeParams {
  format: 'pkcs10' | 'cfca';
  mode?: 'cert' | 'manual';
  encrypted_envelope: string;
  cert_id?: string;
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
// ★ CFCA 双证签发（与后端 CFCASignRequest 字段对齐）
// =============================================================================

export interface CFCASignParams {
  ca_source?: 'local' | 'manual';
  ca_id?: string;
  ca_cert_pem?: string;
  ca_key_pem?: string;
  ca_key_password?: string;

  csr_source?: 'generate' | 'upload' | 'existing';
  csr_id?: string;
  csr_pem?: string;
  csr_key_pem?: string;
  csr_key_password?: string;
  tmp_pub_pem?: string;

  subject?: Record<string, string>;
  san?: string[];
  validity_days?: number;
  domain?: string;

  return_sign_key?: boolean;
  key_export_password?: string;
}

export interface CFCASignResult {
  sign_cert_pem: string;
  enc_cert_pem: string;
  encrypted_envelope: string;
  sign_cert_id?: string;
  enc_cert_id?: string;
  /** @deprecated 平台内置固定临时密钥，不再逐次返回；保留字段兼容旧后端 */
  tmp_key_pem?: string;
  /** 仅 return_sign_key=true 且本次请求确实提供了签名私钥时返回 */
  sign_key_pem?: string;
}

export interface CFCADecodeEnvelopeParams {
  encrypted_envelope: string;
  /** @deprecated 平台使用内置固定临时密钥解密；仅用于兼容旧后端 */
  tmp_key_pem?: string;
}

export interface CFCADecodeEnvelopeResult {
  decrypted_key_pem: string;
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

  // ★ 算法注册表（国密 / RSA / 后量子）——供前端下拉与能力位使用
  listAlgorithms() {
    return request.get<{ items: Array<Record<string, unknown>> }>(
      '/crypto/algorithms'
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

  // ★ 域管理
  listDomains() {
    return request.get<DomainListResponse>('/cas/domains');
  },
  createDomain(name: string) {
    return request.post<CreateDomainResponse>('/cas/domains', { name });
  },
  deleteDomain(name: string) {
    return request.delete(`/cas/domains/${encodeURIComponent(name)}`);
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

  // ★ CFCA 双证签发 / 解信封
  cfcaSign(payload: CFCASignParams) {
    return request.post<CFCASignResult>('/cfca/sign', payload);
  },
  cfcaDecodeEnvelope(payload: CFCADecodeEnvelopeParams) {
    return request.post<CFCADecodeEnvelopeResult>('/cfca/decode-envelope', payload);
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
