import type { ProxyMode, ProxyProtocol, BackendScheme } from '@/types';

export const MODE_OPTIONS: { label: string; value: ProxyMode }[] = [
  { label: '客户端代理（nchttp）', value: 'client' },
  { label: '服务端代理（nshttp）', value: 'server' },
];

export type CertTypeOption = 'intl' | 'gmssl' | 'intl_gmssl';

export const CERT_TYPE_OPTIONS: { label: string; value: CertTypeOption }[] = [
  { label: '国际', value: 'intl' },
  { label: '国密', value: 'gmssl' },
  { label: '国际+国密', value: 'intl_gmssl' },
];

export function certTypeToProtocol(t: CertTypeOption): ProxyProtocol {
  return t === 'intl' ? 'http' : 'gmssl';
}

export function inferCertType(protocol: ProxyProtocol, hasIntl?: boolean): CertTypeOption {
  if (protocol === 'gmssl' || protocol === 'pqc') {
    return hasIntl ? 'intl_gmssl' : 'gmssl';
  }
  return 'intl';
}

export const BACKEND_SCHEME_OPTIONS: { label: string; value: BackendScheme }[] = [
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
];

export const LISTEN_IP_OPTIONS = [
  { label: '0.0.0.0（所有网卡）', value: '0.0.0.0' },
  { label: '127.0.0.1（本机）', value: '127.0.0.1' },
];

export const STATUS_OPTIONS = [
  { label: '全部', value: '' },
  { label: '已停止', value: 'STOPPED' },
  { label: '启动中', value: 'STARTING' },
  { label: '运行中', value: 'RUNNING' },
  { label: '停止中', value: 'STOPPING' },
  { label: '失败', value: 'FAILED' },
];

export const STATUS_TAG_TYPE: Record<
  string,
  'default' | 'success' | 'warning' | 'danger' | 'primary'
> = {
  STOPPED: 'default',
  STARTING: 'primary',
  RUNNING: 'success',
  STOPPING: 'warning',
  FAILED: 'danger',
};

export const PROTOCOL_LABEL: Record<ProxyProtocol, string> = {
  http: '国际',
  pqc: '国际',
  gmssl: '国密',
};

export const GM_SSL_PROTOCOL_VALUE = 'NTLSv1.1';

export const SSL_PROTOCOL_OPTIONS = [
  { label: 'TLS 1.3', value: 'TLSv1.3' },
  { label: 'TLS 1.2', value: 'TLSv1.2' },
  { label: 'TLS 1.1', value: 'TLSv1.1' },
  { label: 'TLS 1.0', value: 'TLSv1' },
  { label: 'SSL 3.0', value: 'SSLv3' },
  { label: 'GMTLSv1.1（国密）', value: GM_SSL_PROTOCOL_VALUE },
];

export const DEFAULT_SSL_PROTOCOLS = 'TLSv1.2 TLSv1.3';
export const DEFAULT_GMSSL_PROTOCOLS = 'NTLSv1.1';
export const DEFAULT_PQC_PROTOCOLS = 'TLSv1.3';

export function defaultProtocolsFor(protocol: ProxyProtocol): string {
  if (protocol === 'pqc') return DEFAULT_PQC_PROTOCOLS;
  if (protocol === 'gmssl') return DEFAULT_GMSSL_PROTOCOLS;
  return DEFAULT_SSL_PROTOCOLS;
}

export function protocolsToUI(raw: string | undefined): string[] {
  return (raw || '')
    .split(/\s+/)
    .map((s) => s.trim())
    .filter(Boolean);
}

export function protocolsFromUI(selected: string[], protocol: ProxyProtocol): string {
  const list = (selected || []).filter(Boolean);
  if (list.length === 0) {
    return defaultProtocolsFor(protocol);
  }
  return list.join(' ');
}

export const DEFAULT_SSL_CIPHERS =
  'HIGH:!RC4:!3DES:!aDSS:!aNULL:!kPSK:!kSRP:!MD5:@STRENGTH:+SHA1:+kRSA:ECC-SM4-SM3:ECDHE-SM4-SM3:ECC-SM4-GCM-SM3:ECDHE-SM4-GCM-SM3';

export const CLIENT_DEFAULT_CIPHERS = DEFAULT_SSL_CIPHERS;
export const SERVER_DEFAULT_CIPHERS = DEFAULT_SSL_CIPHERS;

export function defaultCiphersForMode(_mode: ProxyMode): string {
  return DEFAULT_SSL_CIPHERS;
}

export function defaultCiphersFor(_protocol: ProxyProtocol): string {
  return DEFAULT_SSL_CIPHERS;
}

export interface CipherOption {
  label: string;
  value: string;
  group: 'tls13' | 'tls12' | 'gm';
}

export const CIPHERS_OPTIONS: CipherOption[] = [
  { label: 'TLS_AES_256_GCM_SHA384', value: 'TLS_AES_256_GCM_SHA384', group: 'tls13' },
  { label: 'TLS_CHACHA20_POLY1305_SHA256', value: 'TLS_CHACHA20_POLY1305_SHA256', group: 'tls13' },
  { label: 'TLS_AES_128_GCM_SHA256', value: 'TLS_AES_128_GCM_SHA256', group: 'tls13' },
  { label: 'TLS_SM4_GCM_SM3（国密）', value: 'TLS_SM4_GCM_SM3', group: 'tls13' },
  { label: 'TLS_SM4_CCM_SM3（国密）', value: 'TLS_SM4_CCM_SM3', group: 'tls13' },
  { label: 'ECDHE-ECDSA-AES256-GCM-SHA384', value: 'ECDHE-ECDSA-AES256-GCM-SHA384', group: 'tls12' },
  { label: 'ECDHE-RSA-AES256-GCM-SHA384', value: 'ECDHE-RSA-AES256-GCM-SHA384', group: 'tls12' },
  { label: 'DHE-RSA-AES256-GCM-SHA384', value: 'DHE-RSA-AES256-GCM-SHA384', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES128-GCM-SHA256', value: 'ECDHE-ECDSA-AES128-GCM-SHA256', group: 'tls12' },
  { label: 'ECDHE-RSA-AES128-GCM-SHA256', value: 'ECDHE-RSA-AES128-GCM-SHA256', group: 'tls12' },
  { label: 'DHE-RSA-AES128-GCM-SHA256', value: 'DHE-RSA-AES128-GCM-SHA256', group: 'tls12' },
  { label: 'AES256-GCM-SHA384', value: 'AES256-GCM-SHA384', group: 'tls12' },
  { label: 'AES128-GCM-SHA256', value: 'AES128-GCM-SHA256', group: 'tls12' },
  { label: 'ECDHE-ECDSA-CHACHA20-POLY1305', value: 'ECDHE-ECDSA-CHACHA20-POLY1305', group: 'tls12' },
  { label: 'ECDHE-RSA-CHACHA20-POLY1305', value: 'ECDHE-RSA-CHACHA20-POLY1305', group: 'tls12' },
  { label: 'DHE-RSA-CHACHA20-POLY1305', value: 'DHE-RSA-CHACHA20-POLY1305', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES256-SHA', value: 'ECDHE-ECDSA-AES256-SHA', group: 'tls12' },
  { label: 'ECDHE-RSA-AES256-SHA', value: 'ECDHE-RSA-AES256-SHA', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES128-SHA', value: 'ECDHE-ECDSA-AES128-SHA', group: 'tls12' },
  { label: 'ECDHE-RSA-AES128-SHA', value: 'ECDHE-RSA-AES128-SHA', group: 'tls12' },
  { label: 'ECC-SM4-GCM-SM3（国密）', value: 'ECC-SM4-GCM-SM3', group: 'gm' },
  { label: 'ECDHE-SM4-GCM-SM3（国密）', value: 'ECDHE-SM4-GCM-SM3', group: 'gm' },
  { label: 'ECC-SM4-SM3（国密）', value: 'ECC-SM4-SM3', group: 'gm' },
  { label: 'ECDHE-SM4-SM3（国密）', value: 'ECDHE-SM4-SM3', group: 'gm' },
  { label: 'RSA-SM4-GCM-SHA256（国密）', value: 'RSA-SM4-GCM-SHA256', group: 'gm' },
  { label: 'RSA-SM4-GCM-SM3（国密）', value: 'RSA-SM4-GCM-SM3', group: 'gm' },
  { label: 'RSA-SM4-CBC-SHA256（国密）', value: 'RSA-SM4-CBC-SHA256', group: 'gm' },
  { label: 'RSA-SM4-CBC-SM3（国密）', value: 'RSA-SM4-CBC-SM3', group: 'gm' },
];

export const CLIENT_CIPHERS_PLACEHOLDER = DEFAULT_SSL_CIPHERS;
export const SERVER_CIPHERS_PLACEHOLDER = DEFAULT_SSL_CIPHERS;

export function ciphersPlaceholderFor(_mode: ProxyMode): string {
  return DEFAULT_SSL_CIPHERS;
}

export const DEFAULT_PQC_CONF_CMD = 'Groups X25519MLKEM768:X25519';

export function defaultConfCmdFor(protocol: ProxyProtocol): string {
  return protocol === 'pqc' ? DEFAULT_PQC_CONF_CMD : '';
}