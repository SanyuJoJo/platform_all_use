// src/types/tcps.ts
export interface TcpsApiResponse<T = unknown> {
  code: number | string;
  message: string;
  data: T;
  request_id?: string;
}

export type TcpsMode = 'client' | 'server';
export type TcpsProtocol = 'tcp' | 'tls' | 'gmssl' | 'pqc';
export type TcpsStatus = 'STOPPED' | 'STARTING' | 'RUNNING' | 'STOPPING' | 'FAILED';
export type TcpsCiphersMode = 'default' | 'custom' | 'input';

export interface TcpsProxy {
  id: number;
  service_id: string;
  service_name: string;
  mode: TcpsMode;
  protocol: TcpsProtocol;
  listen_ip: string;
  listen_port: number;
  backend_ip: string;
  backend_port: number;
  backend_tls?: boolean;
  ssl_server_name: string;

  /** 服务端监听是否走 SSL；客户端恒为 false（与 HTTP 模型对齐，便于详情页编辑） */
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
  ssl_ciphers_mode?: TcpsCiphersMode;
  ssl_session_cache_enabled?: boolean;
  ssl_verify_client?: boolean;
  ssl_session_tickets?: boolean;

  status: TcpsStatus;
  pid: number;
  config_kv_path: string;
  config_nginx_path: string;
  log_dir: string;
  created_at: string;
  updated_at: string;
}

export interface TcpsListItem {
  id: number;
  service_id: string;
  service_name: string;
  mode: TcpsMode;
  protocol: TcpsProtocol;
  listen_address: string;
  backend_address: string;
  cert_ref: string;
  status: TcpsStatus;
  pid: number;
  updated_at: string;
  cert_id?: string;
  enc_cert_id?: string;
  client_cert_id?: string;
  client_enc_cert_id?: string;
  intl_cert_id?: string;
  intl_client_cert_id?: string;
}

export interface TcpsCreateReq {
  service_name: string;
  mode: TcpsMode;
  protocol: TcpsProtocol;
  listen_ip: string;
  listen_port: number;
  backend_ip: string;
  backend_port: number;
  backend_tls?: boolean;
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
  ssl_ciphers_mode?: TcpsCiphersMode;
  ssl_session_cache_enabled?: boolean;
  ssl_verify_client?: boolean;
  ssl_session_tickets?: boolean;
}

export type TcpsUpdateReq = TcpsCreateReq;

export interface TcpsListQuery {
  mode?: TcpsMode | '';
  status?: TcpsStatus | '';
  page?: number;
  size?: number;
}

export interface TcpsPortCheckReq {
  service_id?: string;
  listen_ip: string;
  listen_port: number;
}

export interface TcpsPortCheckResult {
  available: boolean;
  conflict_type?: 'system' | 'config';
  occupied_by?: string;
  suggested_port?: number;
  message: string;
}

export interface TcpsBatchReq {
  service_ids: string[];
}

export interface TcpsBatchItemResult {
  service_id: string;
  success: boolean;
  code: number;
  message: string;
}

export interface TcpsBatchResp {
  total: number;
  succeeded: number;
  failed: number;
  results: TcpsBatchItemResult[];
}

export interface TcpsCertBrief {
  cert_id: string;
  subject_cn: string;
  algorithm: string;
  cert_type: string;
  signature_algorithm?: string;
  not_after: string;
  has_private_key: boolean;
}

export interface TcpsAvailableCertsResp {
  items: TcpsCertBrief[];
}

export interface TcpsDomainsResp {
  domains: string[];
}
