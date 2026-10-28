// src/constants/tcpsOptions.ts
import type { TcpsMode, TcpsProtocol, TcpsStatus } from '@/types/tcps';

export const TCPS_MODE_OPTIONS: { label: string; value: TcpsMode }[] = [
  { label: '客户端代理', value: 'client' },
  { label: '服务端代理', value: 'server' },
];

export type TcpsCertTypeOption = 'intl' | 'gmssl' | 'intl_gmssl';

export const TCPS_CERT_TYPE_OPTIONS: { label: string; value: TcpsCertTypeOption }[] = [
  { label: '国际', value: 'intl' },
  { label: '国密', value: 'gmssl' },
  { label: '国际+国密', value: 'intl_gmssl' },
];

/**
 * 证书类型 → 后端协议。
 *
 * 表单不再暴露协议选择，统一由证书类型推断：
 *   - intl       → tls   （国际 TLS）
 *   - gmssl      → gmssl （国密 GMTLS）
 *   - intl_gmssl → gmssl （双栈：gmssl 承载 NTLS，intl_cert_id 承载国际证书，
 *                          后端通过 enable_ntls + ssl_certificate 同时开启两条通道）
 */
export function tcpsCertTypeToProtocol(t: TcpsCertTypeOption): TcpsProtocol {
  return t === 'intl' ? 'tls' : 'gmssl';
}

/**
 * 反向推断证书类型（详情/编辑回填用）。
 *
 * - gmssl：看是否同时存在国际证书 → intl_gmssl / gmssl
 * - tls   ：intl
 * - pqc   ：使用国际 PQC 证书，归到 intl（表单不再支持单独选 pqc）
 * - tcp   ：无证书，按 intl 兜底（表单保存时会被归一为 tls）
 */
export function inferTcpsCertType(p: TcpsProtocol, hasIntl?: boolean): TcpsCertTypeOption {
  if (p === 'gmssl') return hasIntl ? 'intl_gmssl' : 'gmssl';
  return 'intl';
}

export const TCPS_PROTOCOL_OPTIONS: { label: string; value: TcpsProtocol }[] = [
  { label: 'TCP（明文）', value: 'tcp' },
  { label: 'TLS（国际）', value: 'tls' },
  { label: 'GMTLS（国密）', value: 'gmssl' },
  { label: 'PQC TLS（后量子）', value: 'pqc' },
];

export function tcpsProtocolNeedsSSL(p: TcpsProtocol): boolean {
  return p !== 'tcp';
}

export const TCPS_PROTOCOL_LABEL: Record<TcpsProtocol, string> = {
  tcp: 'TCP',
  tls: 'TLS',
  gmssl: 'GMTLS',
  pqc: 'PQC',
};

export const TCPS_STATUS_OPTIONS = [
  { label: '全部', value: '' },
  { label: '已停止', value: 'STOPPED' },
  { label: '启动中', value: 'STARTING' },
  { label: '运行中', value: 'RUNNING' },
  { label: '停止中', value: 'STOPPING' },
  { label: '失败', value: 'FAILED' },
];

export const TCPS_STATUS_TAG_TYPE: Record<string, 'default' | 'success' | 'warning' | 'error' | 'info'> = {
  STOPPED: 'default',
  STARTING: 'info',
  RUNNING: 'success',
  STOPPING: 'warning',
  FAILED: 'error',
};

export const TCPS_LISTEN_IP_OPTIONS = [
  { label: '0.0.0.0（所有网卡）', value: '0.0.0.0' },
  { label: '127.0.0.1（本机）', value: '127.0.0.1' },
];

export const TCPS_GM_SSL_PROTOCOL_VALUE = 'NTLSv1.1';

export const TCPS_SSL_PROTOCOL_OPTIONS = [
  { label: 'TLS 1.3', value: 'TLSv1.3' },
  { label: 'TLS 1.2', value: 'TLSv1.2' },
  { label: 'TLS 1.1', value: 'TLSv1.1' },
  { label: 'TLS 1.0', value: 'TLSv1' },
  { label: 'SSL 3.0', value: 'SSLv3' },
  { label: 'GMTLSv1.1（国密）', value: TCPS_GM_SSL_PROTOCOL_VALUE },
];

export const TCPS_DEFAULT_SSL_PROTOCOLS = 'TLSv1.2 TLSv1.3';
export const TCPS_DEFAULT_GMSSL_PROTOCOLS = 'NTLSv1.1';
export const TCPS_DEFAULT_PQC_PROTOCOLS = 'TLSv1.3';

export function tcpsDefaultProtocolsFor(p: TcpsProtocol): string {
  if (p === 'tcp') return '';
  if (p === 'pqc') return TCPS_DEFAULT_PQC_PROTOCOLS;
  if (p === 'gmssl') return TCPS_DEFAULT_GMSSL_PROTOCOLS;
  return TCPS_DEFAULT_SSL_PROTOCOLS;
}

export const TCPS_DEFAULT_SSL_CIPHERS =
  'HIGH:!RC4:!3DES:!aDSS:!aNULL:!kPSK:!kSRP:!MD5:@STRENGTH:+SHA1:+kRSA:ECC-SM4-SM3:ECDHE-SM4-SM3:ECC-SM4-GCM-SM3:ECDHE-SM4-GCM-SM3';

export function tcpsDefaultCiphers(): string {
  return TCPS_DEFAULT_SSL_CIPHERS;
}

export const TCPS_DEFAULT_PQC_CONF_CMD = 'Groups X25519MLKEM768:X25519';

export function tcpsDefaultConfCmdFor(p: TcpsProtocol): string {
  return p === 'pqc' ? TCPS_DEFAULT_PQC_CONF_CMD : '';
}

export function tcpsProtocolsToUI(raw: string | undefined): string[] {
  return (raw || '').split(/\s+/).map((s) => s.trim()).filter(Boolean);
}

export function tcpsProtocolsFromUI(selected: string[], p: TcpsProtocol): string {
  const list = (selected || []).filter(Boolean);
  if (list.length === 0) return tcpsDefaultProtocolsFor(p);
  return list.join(' ');
}

export function isTcpsRunning(s: TcpsStatus): boolean {
  return s === 'RUNNING' || s === 'STARTING';
}
