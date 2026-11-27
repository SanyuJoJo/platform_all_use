import axios, { type AxiosRequestConfig } from 'axios';
import request from '@/utils/request';
import type {
  OperationResponse,
  OperationOptions,
  CAExportType,
  CertExportType,
  ImportCAParams,
  ImportCertParams,
  ImportCsrParams,
  CADetail,
  CertDetail,
  CreateDomainResponse,
  DomainListResponse,
  CSRCreateParams,
  CSRCreateResult,
  SignCertParams,
  SignCertResult,
  QueryEnvelopeParams,
  QueryEnvelopeResult,
  CFCASignParams,
  CFCASignResult,
  CFCADecodeEnvelopeParams,
  CFCADecodeEnvelopeResult,
  CertTypeItem,
  P10Format,
  CAListItem,
} from '@/types';

export type {
  OperationOptions,
  CAExportType,
  CertExportType,
  ImportCAParams,
  ImportCertParams,
  ImportCsrParams,
  CADetail,
  CertDetail,
  CreateDomainResponse,
  DomainListResponse,
  CSRCreateParams,
  CSRCreateResult,
  SignCertParams,
  SignCertResult,
  QueryEnvelopeParams,
  QueryEnvelopeResult,
  CFCASignParams,
  CFCASignResult,
  CFCADecodeEnvelopeParams,
  CFCADecodeEnvelopeResult,
  CertTypeItem,
  P10Format,
  CAListItem,
  OperationResponse,
}

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
    options?: OperationOptions,
  ) {
    return request.post<OperationResponse<T>>(
      `/crypto/operations/${operationId}`,
      { params, options: { timeout_ms: 5000, dry_run: false, ...options } },
    );
  },

  listAlgorithms() {
    return request.get<{ items: Array<Record<string, unknown>> }>(
      '/crypto/algorithms',
    );
  },
  listDigests() {
    return request.get<{
      all: string[];
      by_family: Record<string, string[]>;
      weak: string[];
    }>('/crypto/digests');
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
  exportCA(
    caId: string,
    payload: { type: CAExportType; password?: string },
    config?: AxiosRequestConfig,
  ) {
    return rawInstance.post(`/cas/${caId}/export`, payload, {
      responseType: 'blob',
      ...config,
    });
  },
  deleteCA(caId: string) {
    return request.delete(`/cas/${caId}`);
  },

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
  exportCert(
    certId: string,
    payload: { type: CertExportType; password?: string },
    config?: AxiosRequestConfig,
  ) {
    return rawInstance.post(`/certs/${certId}/export`, payload, {
      responseType: 'blob',
      ...config,
    });
  },
  deleteCert(certId: string) {
    return request.delete(`/certs/${certId}`);
  },

  cfcaSign(payload: CFCASignParams) {
    return request.post<CFCASignResult>('/cfca/sign', payload);
  },
  cfcaDecodeEnvelope(payload: CFCADecodeEnvelopeParams) {
    return request.post<CFCADecodeEnvelopeResult>(
      '/cfca/decode-envelope',
      payload,
    );
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
    return rawInstance.get(`/csrs/${csrId}/key`, {
      params,
      responseType: 'blob',
    });
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