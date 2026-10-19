import type { ProxyMode, ProxyProtocol, BackendScheme } from '@/types';

// ============================================================
// 模式 / 证书类型 / 后端协议 / 监听地址 / 状态
// ============================================================
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

export function inferCertType(
  protocol: ProxyProtocol,
  hasIntl?: boolean
): CertTypeOption {
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
  'default' | 'success' | 'warning' | 'error' | 'info'
> = {
  STOPPED: 'default',
  STARTING: 'info',
  RUNNING: 'success',
  STOPPING: 'warning',
  FAILED: 'error',
};

export const PROTOCOL_LABEL: Record<ProxyProtocol, string> = {
  http: '国际',
  pqc: '国际',
  gmssl: '国密',
};

// ============================================================
// SSL 协议选项
//
// ★ NTLSv1.1 是后端约定的国密通道标识值，仅用于 UI 勾选与持久化。
//   nginx ssl_protocols 不支持此值；后端在渲染 nginx.conf 时会通过
//   sanitizeSSLProtocols() 剔除，并靠 `enable_ntls on;` 开启国密通道。
//
//   前端只负责：
//     - 显示名：NTLSv1.1 → GMTLSv1.1（对用户更直观）
//     - value：保持 NTLSv1.1，不做任何转换
// ============================================================
export const GM_SSL_PROTOCOL_VALUE = 'NTLSv1.1';

export const SSL_PROTOCOL_OPTIONS = [
  { label: 'TLS 1.3', value: 'TLSv1.3' },
  { label: 'TLS 1.2', value: 'TLSv1.2' },
  { label: 'TLS 1.1', value: 'TLSv1.1' },
  { label: 'TLS 1.0', value: 'TLSv1' },
  { label: 'SSL 3.0', value: 'SSLv3' },
  // ★ 显示名 GMTLSv1.1，value 保持 NTLSv1.1（与后端 DefaultGMSSLProtocols 一致）
  { label: 'GMTLSv1.1（国密）', value: GM_SSL_PROTOCOL_VALUE },
];

// ============================================================
// SSL 默认协议（与后端 constants.go 对齐）
//   - 国际：TLSv1.2 TLSv1.3
//   - 国密：NTLSv1.1
//   - PQC：TLSv1.3
// ============================================================
export const DEFAULT_SSL_PROTOCOLS = 'TLSv1.2 TLSv1.3';
export const DEFAULT_GMSSL_PROTOCOLS = 'NTLSv1.1';
export const DEFAULT_PQC_PROTOCOLS = 'TLSv1.3';

export function defaultProtocolsFor(protocol: ProxyProtocol): string {
  if (protocol === 'pqc') return DEFAULT_PQC_PROTOCOLS;
  if (protocol === 'gmssl') return DEFAULT_GMSSL_PROTOCOLS;
  return DEFAULT_SSL_PROTOCOLS;
}

// ============================================================
// 协议字符串 ↔ UI 数组
//
// ★ 纯 split / join，不做任何剔除或补全；后端 sanitizeSSLProtocols
//   会在渲染 nginx.conf 时过滤非法值，前端保持"所见即所存"。
// ============================================================
export function protocolsToUI(raw: string | undefined): string[] {
  return (raw || '').split(/\s+/).map((s) => s.trim()).filter(Boolean);
}

export function protocolsFromUI(selected: string[], protocol: ProxyProtocol): string {
  const list = (selected || []).filter(Boolean);
  if (list.length === 0) {
    // 空 → 按 protocol 回落默认
    return defaultProtocolsFor(protocol);
  }
  return list.join(' ');
}

// ============================================================
// SSL 默认加密套件
// ============================================================
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

// ============================================================
// 自定义套件选项
// ============================================================
export interface CipherOption {
  label: string;
  value: string;
  group: 'tls13' | 'tls12' | 'gm';
}

export const CIPHERS_OPTIONS: CipherOption[] = [
  // ---- TLS 1.3 国际 ----
  { label: 'TLS_AES_256_GCM_SHA384', value: 'TLS_AES_256_GCM_SHA384', group: 'tls13' },
  { label: 'TLS_CHACHA20_POLY1305_SHA256', value: 'TLS_CHACHA20_POLY1305_SHA256', group: 'tls13' },
  { label: 'TLS_AES_128_GCM_SHA256', value: 'TLS_AES_128_GCM_SHA256', group: 'tls13' },
  { label: 'TLS_SM4_GCM_SM3（国密）', value: 'TLS_SM4_GCM_SM3', group: 'tls13' },
  { label: 'TLS_SM4_CCM_SM3（国密）', value: 'TLS_SM4_CCM_SM3', group: 'tls13' },

  // ---- TLS 1.2 AES-GCM ----
  { label: 'ECDHE-ECDSA-AES256-GCM-SHA384', value: 'ECDHE-ECDSA-AES256-GCM-SHA384', group: 'tls12' },
  { label: 'ECDHE-RSA-AES256-GCM-SHA384', value: 'ECDHE-RSA-AES256-GCM-SHA384', group: 'tls12' },
  { label: 'DHE-RSA-AES256-GCM-SHA384', value: 'DHE-RSA-AES256-GCM-SHA384', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES128-GCM-SHA256', value: 'ECDHE-ECDSA-AES128-GCM-SHA256', group: 'tls12' },
  { label: 'ECDHE-RSA-AES128-GCM-SHA256', value: 'ECDHE-RSA-AES128-GCM-SHA256', group: 'tls12' },
  { label: 'DHE-RSA-AES128-GCM-SHA256', value: 'DHE-RSA-AES128-GCM-SHA256', group: 'tls12' },
  { label: 'AES256-GCM-SHA384', value: 'AES256-GCM-SHA384', group: 'tls12' },
  { label: 'AES128-GCM-SHA256', value: 'AES128-GCM-SHA256', group: 'tls12' },

  // ---- ChaCha20 ----
  { label: 'ECDHE-ECDSA-CHACHA20-POLY1305', value: 'ECDHE-ECDSA-CHACHA20-POLY1305', group: 'tls12' },
  { label: 'ECDHE-RSA-CHACHA20-POLY1305', value: 'ECDHE-RSA-CHACHA20-POLY1305', group: 'tls12' },
  { label: 'DHE-RSA-CHACHA20-POLY1305', value: 'DHE-RSA-CHACHA20-POLY1305', group: 'tls12' },

  // ---- AES-CBC ----
  { label: 'ECDHE-ECDSA-AES256-SHA384', value: 'ECDHE-ECDSA-AES256-SHA384', group: 'tls12' },
  { label: 'ECDHE-RSA-AES256-SHA384', value: 'ECDHE-RSA-AES256-SHA384', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES128-SHA256', value: 'ECDHE-ECDSA-AES128-SHA256', group: 'tls12' },
  { label: 'ECDHE-RSA-AES128-SHA256', value: 'ECDHE-RSA-AES128-SHA256', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES256-SHA', value: 'ECDHE-ECDSA-AES256-SHA', group: 'tls12' },
  { label: 'ECDHE-RSA-AES256-SHA', value: 'ECDHE-RSA-AES256-SHA', group: 'tls12' },
  { label: 'ECDHE-ECDSA-AES128-SHA', value: 'ECDHE-ECDSA-AES128-SHA', group: 'tls12' },
  { label: 'ECDHE-RSA-AES128-SHA', value: 'ECDHE-RSA-AES128-SHA', group: 'tls12' },

  // ---- 国密 NTLS ----
  { label: 'ECC-SM4-GCM-SM3（国密）', value: 'ECC-SM4-GCM-SM3', group: 'gm' },
  { label: 'ECDHE-SM4-GCM-SM3（国密）', value: 'ECDHE-SM4-GCM-SM3', group: 'gm' },
  { label: 'ECC-SM4-SM3（国密）', value: 'ECC-SM4-SM3', group: 'gm' },
  { label: 'ECDHE-SM4-SM3（国密）', value: 'ECDHE-SM4-SM3', group: 'gm' },
  { label: 'RSA-SM4-GCM-SHA256（国密）', value: 'RSA-SM4-GCM-SHA256', group: 'gm' },
  { label: 'RSA-SM4-GCM-SM3（国密）', value: 'RSA-SM4-GCM-SM3', group: 'gm' },
  { label: 'RSA-SM4-CBC-SHA256（国密）', value: 'RSA-SM4-CBC-SHA256', group: 'gm' },
  { label: 'RSA-SM4-CBC-SM3（国密）', value: 'RSA-SM4-CBC-SM3', group: 'gm' },
];

// ============================================================
// SSL 输入套件占位提示
// ============================================================
export const CLIENT_CIPHERS_PLACEHOLDER = DEFAULT_SSL_CIPHERS;
export const SERVER_CIPHERS_PLACEHOLDER = DEFAULT_SSL_CIPHERS;

export function ciphersPlaceholderFor(_mode: ProxyMode): string {
  return DEFAULT_SSL_CIPHERS;
}

// ============================================================
// PQC 相关
// ============================================================
export const DEFAULT_PQC_CONF_CMD = 'Groups X25519MLKEM768:X25519';

export function defaultConfCmdFor(protocol: ProxyProtocol): string {
  return protocol === 'pqc' ? DEFAULT_PQC_CONF_CMD : '';
}
