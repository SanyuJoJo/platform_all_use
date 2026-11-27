/* ============================================================================
 * 证书管理控制台（crypto-console）类型
 * ==========================================================================*/

export interface OperationOptions {
  timeout_ms?: number;
  dry_run?: boolean;
}

export interface OperationResponse<T = unknown> {
  task_id?: string | null;
  [key: string]: unknown;
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
  domain?: string;
  not_before?: string;
  not_after?: string;
}

export interface CreateDomainResponse {
  name: string;
}

export interface DomainListResponse {
  domains: string[];
}

export type CertTypeItem = 'signature' | 'encryption';

export type P10Format = 'pkcs10' | 'cfca';

/**
 * 摘要（散列）算法。
 *
 * - SM3：国密摘要，SM2 专用（SM2 密钥位数 256，仅支持 SM3）；
 * - SHA1 / SHA224 / SHA256 / SHA384 / SHA512：国际摘要，RSA / ECC 可用；
 * - 后量子（ML-DSA / SLH-DSA / Falcon）：摘要由子算法自带，不可另行指定，
 *   例如 slh-dsa-shake-256s 的摘要即为 "slh-dsa-shake-256s"。
 *
 * 不传（undefined）时由后端按密钥算法自动选择：SM2 → SM3，RSA/ECC → SHA256。
 */
export type DigestAlgorithm =
  | 'SM3'
  | 'SHA1'
  | 'SHA224'
  | 'SHA256'
  | 'SHA384'
  | 'SHA512';

export interface CSRCreateParams {
  algorithm?: string;
  subject: Record<string, string>;
  key_source?: 'generate';
  p10_format?: P10Format;
  san?: string[];
  /** ★ 摘要算法（签名算法），仅对 RSA / ECC / SM2 生效 */
  digest_algorithm?: DigestAlgorithm | string;
}

export interface CSRCreateResult {
  csr_id?: string;
  csr_pem?: string;
  tmp_key_pem?: string;
  sign_key_pem?: string;
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
  /** ★ 摘要算法（签名算法），仅对普通证书生效：SM2→SM3，RSA/ECC→SHA256 */
  digest_algorithm?: DigestAlgorithm | string;
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
  // 摘要算法（签名算法），CFCA 格式固定为 SM3
  digest_algorithm?: DigestAlgorithm | string;
  return_sign_key?: boolean;
  key_export_password?: string;
}

export interface CFCASignResult {
  sign_cert_pem: string;
  enc_cert_pem: string;
  encrypted_envelope: string;
  sign_cert_id?: string;
  enc_cert_id?: string;
  tmp_key_pem?: string;
  sign_key_pem?: string;
}

export interface CFCADecodeEnvelopeParams {
  encrypted_envelope: string;
  tmp_key_pem?: string;
}

export interface CFCADecodeEnvelopeResult {
  decrypted_key_pem: string;
}

export type TaskStatus =
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCESS'
  | 'FAILED'
  | 'TIMEOUT'
  | 'CANCELLED';

export interface TaskItem {
  task_id: string;
  operation_id: string;
  request_id: string;
  status: TaskStatus;
  progress?: number;
  result_ref?: string;
  error_code?: string;
  error_message?: string;
  created_at?: string;
  started_at?: string;
  finished_at?: string;
}