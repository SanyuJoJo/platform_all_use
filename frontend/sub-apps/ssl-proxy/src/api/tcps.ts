// src/api/tcps.ts
import request from './index';
import type {
  TcpsAvailableCertsResp,
  TcpsBatchReq,
  TcpsBatchResp,
  TcpsCreateReq,
  TcpsDomainsResp,
  TcpsListQuery,
  TcpsListItem,
  TcpsPortCheckReq,
  TcpsPortCheckResult,
  TcpsProxy,
  TcpsUpdateReq,
} from '@/types/tcps';
const BASE = '/tcps-proxies';
export const tcpsApi = {
  list(params: TcpsListQuery) {
    return request.get<{ items: TcpsListItem[]; total: number; page: number; size: number }>(
      BASE,
      { params },
    );
  },
  create(req: TcpsCreateReq) {
    return request.post<TcpsProxy>(BASE, req);
  },
  get(serviceId: string) {
    return request.get<TcpsProxy>(`${BASE}/${serviceId}`);
  },
  update(serviceId: string, req: TcpsUpdateReq) {
    return request.put<TcpsProxy>(`${BASE}/${serviceId}`, req);
  },
  remove(serviceId: string) {
    return request.delete<{ service_id: string }>(`${BASE}/${serviceId}`);
  },
  start(serviceId: string) {
    return request.post<{ service_id: string; status: string }>(`${BASE}/${serviceId}/start`);
  },
  stop(serviceId: string) {
    return request.post<{ service_id: string; status: string }>(`${BASE}/${serviceId}/stop`);
  },
  restart(serviceId: string) {
    return request.post<{ service_id: string; status: string }>(`${BASE}/${serviceId}/restart`);
  },
  checkPort(req: TcpsPortCheckReq) {
    return request.post<TcpsPortCheckResult>(`${BASE}/check-port`, req);
  },
  listAvailableCerts(params: { purpose: 'sign' | 'enc'; protocol: 'tls' | 'gmssl' | 'pqc' }) {
    return request.get<TcpsAvailableCertsResp>(`${BASE}/available-certs`, { params });
  },
  listCADomains() {
    return request.get<TcpsDomainsResp | { data?: { domains?: string[] } }>('/cas/domains');
  },
  getConfig(serviceId: string) {
    return request.get<{ kv: string; nginx: string }>(`${BASE}/${serviceId}/config`);
  },
  getLogs(serviceId: string, tail = 200) {
    return request.get<{ path: string; content?: string; message?: string }>(
      `${BASE}/${serviceId}/logs`,
      { params: { tail } },
    );
  },
  batchStart(req: TcpsBatchReq) {
    return request.post<TcpsBatchResp>(`${BASE}/batch/start`, req);
  },
  batchStop(req: TcpsBatchReq) {
    return request.post<TcpsBatchResp>(`${BASE}/batch/stop`, req);
  },
};
export default tcpsApi;
