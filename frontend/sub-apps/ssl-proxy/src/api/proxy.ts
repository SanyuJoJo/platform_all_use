import request from './index';
import type {
  AvailableCertsResp,
  BatchReq,
  BatchResp,
  DomainsResp,
  PlatformSSLProxy,
  PortCheckReq,
  PortCheckResult,
  ProxyCreateReq,
  ProxyListQuery,
  ProxyListItem,
  ProxyProtocol,
  ProxyUpdateReq,
} from '@/types';

export const proxyApi = {
  // ---------- CRUD ----------
  list(params: ProxyListQuery) {
    return request.get<{
      items: ProxyListItem[];
      total: number;
      page: number;
      size: number;
    }>('/ssl-proxies', { params });
  },

  create(req: ProxyCreateReq) {
    return request.post<PlatformSSLProxy>('/ssl-proxies', req);
  },

  get(serviceId: string) {
    return request.get<PlatformSSLProxy>(`/ssl-proxies/${serviceId}`);
  },

  update(serviceId: string, req: ProxyUpdateReq) {
    return request.put<PlatformSSLProxy>(`/ssl-proxies/${serviceId}`, req);
  },

  remove(serviceId: string) {
    return request.delete<{ service_id: string }>(`/ssl-proxies/${serviceId}`);
  },

  // ---------- 生命周期 ----------
  start(serviceId: string) {
    return request.post<{ service_id: string; status: string }>(
      `/ssl-proxies/${serviceId}/start`
    );
  },

  stop(serviceId: string) {
    return request.post<{ service_id: string; status: string }>(
      `/ssl-proxies/${serviceId}/stop`
    );
  },

  restart(serviceId: string) {
    return request.post<{ service_id: string; status: string }>(
      `/ssl-proxies/${serviceId}/restart`
    );
  },

  // ---------- 端口检测 ----------
  checkPort(req: PortCheckReq) {
    return request.post<PortCheckResult>('/ssl-proxies/check-port', req);
  },

  // ---------- 证书目录 ----------
  listAvailableCerts(params: {
    purpose: 'sign' | 'enc';
    protocol: ProxyProtocol;
  }) {
    return request.get<AvailableCertsResp>(
      '/ssl-proxies/available-certs',
      { params }
    );
  },

  // ★ CA 信任域列表（复用一期证书模块的 /cas/domains）
  listCADomains() {
    return request.get<DomainsResp | { data?: { domains?: string[] } }>(
      '/cas/domains'
    );
  },

  // ---------- 配置与日志 ----------
  getConfig(serviceId: string) {
    return request.get<{ kv: string; nginx: string }>(
      `/ssl-proxies/${serviceId}/config`
    );
  },

  getLogs(serviceId: string, tail: number = 200) {
    return request.get<{ path: string; content?: string; message?: string }>(
      `/ssl-proxies/${serviceId}/logs`,
      { params: { tail } }
    );
  },

  // ---------- 批量 ----------
  batchStart(req: BatchReq) {
    return request.post<BatchResp>('/ssl-proxies/batch/start', req);
  },

  batchStop(req: BatchReq) {
    return request.post<BatchResp>('/ssl-proxies/batch/stop', req);
  },
};

export default proxyApi;
