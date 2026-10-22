// ============================================================
// 通用响应
// ============================================================
export interface ApiResponse<T = unknown> {
  code: number | string;
  message: string;
  data: T;
  request_id?: string;
}

// ============================================================
// 代理模型
// ============================================================
export type ProxyMode = 'client' | 'server';
export type ProxyProtocol = 'http' | 'pqc' | 'gmssl';
export type BackendScheme = 'http' | 'https';
export type ProxyStatus =
  | 'STOPPED'
  | 'STARTING'
  | 'RUNNING'
  | 'STOPPING'
  | 'FAILED';

export type SSLCiphersMode = 'default' | 'custom' | 'input';

export interface PlatformSSLProxy {
  id: number;
  service_id: string;
  service_name: string;
  mode: ProxyMode;
  protocol: ProxyProtocol;
  listen_ip: string;
  listen_port: number;
  backend_ip: string;
  backend_port: number;
  backend_scheme: BackendScheme;
  ssl_server_name: string;
  listen_ssl?: boolean;

  cert_id: string;
  enc_cert_id: string;
  intl_cert_id?: string;

  ca_cert_id: string;
  ca_domain?: string;
  client_cert_id: string;
  client_enc_cert_id: string;
  intl_client_cert_id?: string;
  client_key_ref: string;

  ssl_protocols: string;
  ssl_ciphers: string;
  ssl_conf_command: string;

  ssl_ciphers_mode?: SSLCiphersMode;
  ssl_session_cache_enabled?: boolean;
  ssl_verify_client?: boolean;
  ssl_session_tickets?: boolean;

  status: ProxyStatus;
  pid: number;
  config_kv_path: string;
  config_nginx_path: string;
  log_dir: string;
  created_at: string;
  updated_at: string;
}

export interface ProxyListItem {
  id: number;
  service_id: string;
  service_name: string;
  mode: ProxyMode;
  protocol: ProxyProtocol;
  listen_address: string;
  backend_address: string;
  cert_ref: string;
  status: ProxyStatus;
  pid: number;
  updated_at: string;

  cert_id?: string;
  enc_cert_id?: string;
  client_cert_id?: string;
  client_enc_cert_id?: string;
  intl_cert_id?: string;
  intl_client_cert_id?: string;
}

// ============================================================
// 请求 / 响应
// ============================================================
export interface ProxyCreateReq {
  service_name: string;
  mode: ProxyMode;
  protocol: ProxyProtocol;
  listen_ip: string;
  listen_port: number;
  backend_ip: string;
  backend_port: number;
  backend_scheme?: BackendScheme;
  ssl_server_name?: string;
  listen_ssl?: boolean;

  cert_id?: string;
  enc_cert_id?: string;
  intl_cert_id?: string;

  ca_cert_id?: string;
  ca_domain?: string;
  client_cert_id?: string;
  client_enc_cert_id?: string;
  intl_client_cert_id?: string;
  client_key_ref?: string;

  ssl_protocols?: string;
  ssl_ciphers?: string;
  ssl_conf_command?: string;

  ssl_ciphers_mode?: SSLCiphersMode;
  ssl_session_cache_enabled?: boolean;
  ssl_verify_client?: boolean;
  ssl_session_tickets?: boolean;
}

export type ProxyUpdateReq = ProxyCreateReq;

export interface ProxyListQuery {
  mode?: ProxyMode | '';
  status?: ProxyStatus | '';
  page?: number;
  size?: number;
}

export interface PortCheckReq {
  service_id?: string;
  listen_ip: string;
  listen_port: number;
}

export interface PortCheckResult {
  available: boolean;
  conflict_type?: 'system' | 'config';
  occupied_by?: string;
  suggested_port?: number;
  message: string;
}

export interface BatchReq {
  service_ids: string[];
}

export interface BatchItemResult {
  service_id: string;
  success: boolean;
  code: number;
  message: string;
}

export interface BatchResp {
  total: number;
  succeeded: number;
  failed: number;
  results: BatchItemResult[];
}

// ============================================================
// 证书目录
// ============================================================
export interface CertBrief {
  cert_id: string;
  subject_cn: string;
  algorithm: string;
  cert_type: string;
  signature_algorithm?: string;
  not_after: string;
  has_private_key: boolean;
}

export interface AvailableCertsResp {
  items: CertBrief[];
}

// ★ CA 信任域列表响应
export interface DomainsResp {
  domains: string[];
}

/* ========== 动态路由 ========== */
export interface RouteMeta {
  title: string;
  icon?: string;
  permission?: string;
  hidden?: boolean;
  keepAlive?: boolean;
  [key: string]: unknown;
  [key: symbol]: unknown;
}
export interface RouteNode {
  id: string;
  parentId: string | null;
  name: string;
  path?: string;
  component?: string;
  microApp: string;
  type: 'directory' | 'menu';
  meta: RouteMeta;
  order?: number;
  children?: RouteNode[];
}
