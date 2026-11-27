/* ============================================================================
 * 摘要（散列）算法选择工具
 *
 * 规则（按公钥算法约束可选摘要）：
 *   - RSA（1024/2048/3072/4096）：SHA1 / SHA224 / SHA256 / SHA384 / SHA512
 *   - ECC：SHA1 / SHA224 / SHA256 / SHA384 / SHA512
 *   - SM2：密钥位数 256，仅支持 SM3（默认；不允许失败再选非国密摘要）
 *   - PQC（ML-DSA / SLH-DSA / Falcon）：摘要由子算法自带，例如
 *     slh-dsa-shake-256s 的摘要就是 "slh-dsa-shake-256s" 本身，不可另选。
 *
 * 未指定摘要时由后端按密钥算法给出安全默认值（RSA/ECC -> SHA256，SM2 -> SM3）。
 * ==========================================================================*/

import type { AlgorithmInfo } from './crypto-algorithms';

export type DigestId = 'SM3' | 'SHA1' | 'SHA224' | 'SHA256' | 'SHA384' | 'SHA512';

/**
 * 摘要下拉项的值。
 *
 * - SM2 / RSA / ECC：为摘要 ID（SM3 / SHA1 … SHA512）；
 * - PQC：为算法 ID 本身（如 slh-dsa-shake-256s），不可另行选择摘要。
 */
export type DigestValue = DigestId | (string & {});

/** 算法族：决定可选摘要集合 */
export type AlgorithmFamily = 'SM2' | 'RSA' | 'ECC' | 'PQC' | '';

/** 摘要算法展示名 */
export const DIGEST_LABELS: Record<string, string> = {
  SM3: 'SM3',
  SHA1: 'SHA1',
  SHA224: 'SHA224',
  SHA256: 'SHA256',
  SHA384: 'SHA384',
  SHA512: 'SHA512',
};

/**
 * 把平台算法 ID 归并为摘要选择所需的算法族。
 *
 * 平台算法 ID 形如 SM2 / RSA1024 / RSA2048 / RSA4096 / ECC /
 * MLDSA65 / falcon512 / slh-dsa-sha2-128s / slh-dsa-shake-256s。
 * AlgorithmInfo.Family 若由后端提供则优先使用。
 */
export function algorithmFamily(
  algorithm?: string | null,
  spec?: AlgorithmInfo | null,
): AlgorithmFamily {
  const raw = (spec?.Family || spec?.ID || algorithm || '').toString().trim();
  if (!raw) return '';
  const id = raw.toUpperCase();
  switch (id) {
    case 'SM2':
      return 'SM2';
    case 'RSA':
      return 'RSA';
    case 'ECC':
    case 'EC':
      return 'ECC';
    case 'ML-DSA':
    case 'MLDSA':
    case 'FALCON':
    case 'SLH-DSA':
    case 'SLHDSA':
      return 'PQC';
  }
  if (id.startsWith('RSA')) return 'RSA';
  if (id.startsWith('MLDSA') || id.startsWith('ML-DSA')) return 'PQC';
  if (id.startsWith('FALCON')) return 'PQC';
  if (id.startsWith('SLH-DSA') || id.startsWith('SLHDSA')) return 'PQC';
  return '';
}

/** 该算法是否为 PQC（后量子）算法族 */
export function isPqcFamily(family: AlgorithmFamily): boolean {
  return family === 'PQC';
}

/** 返回算法 ID 的有效展示 ID（处理 slh-dsa 全小写 / falcon 小写） */
export function normalizeAlgorithmId(algorithm?: string | null): string {
  return (algorithm || '').toString().trim();
}

/**
 * 某算法族可选的摘要下拉项。
 *
 * - SM2：仅 SM3（密钥位数 256，国密曲线）；
 * - RSA / ECC：SHA1 / SHA224 / SHA256 / SHA384 / SHA512；
 * - PQC：仅“后量子算法本身”一项，label/value 均为算法 ID，
 *   例如 slh-dsa-shake-256s 的摘要即其本身，不可另行选择。
 */
export function digestOptionsFor(
  family: AlgorithmFamily,
  algorithm?: string | null,
  spec?: AlgorithmInfo | null,
): { label: string; value: DigestValue }[] {
  switch (family) {
    case 'SM2':
      return [{ label: 'SM3', value: 'SM3' }];
    case 'RSA':
    case 'ECC':
      return [
        { label: 'SHA1', value: 'SHA1' },
        { label: 'SHA224', value: 'SHA224' },
        { label: 'SHA256', value: 'SHA256' },
        { label: 'SHA384', value: 'SHA384' },
        { label: 'SHA512', value: 'SHA512' },
      ];
    case 'PQC': {
      const id = normalizeAlgorithmId(spec?.ID || algorithm);
      return id ? [{ label: id, value: id }] : [];
    }
    default:
      return [];
  }
}

/** 某算法族对应的默认摘要 */
export function defaultDigestFor(family: AlgorithmFamily): DigestValue {
  switch (family) {
    case 'SM2':
      return 'SM3';
    case 'RSA':
    case 'ECC':
      return 'SHA256';
    default:
      return '';
  }
}

/**
 * 某算法的默认摘要（含 PQC）。
 *
 * PQC 直接返回算法 ID 本身（如 slh-dsa-shake-256s），
 * 其它算法族返回摘要 ID，未知返回空串。
 */
export function defaultDigestForAlgorithm(
  algorithm: string | null | undefined,
  spec?: AlgorithmInfo | null,
): DigestValue {
  const family = algorithmFamily(algorithm, spec);
  if (family === 'PQC') {
    return normalizeAlgorithmId(spec?.ID || algorithm) || '';
  }
  return defaultDigestFor(family);
}

/**
 * 摘要选择提示文案。
 *
 * 产品要求界面不展示任何额外说明文字，统一返回空串；
 * 保留该函数仅为兼容既有调用方（各页面的 hint 变量）。
 */
export function digestHintFor(_family: AlgorithmFamily): string {
  return '';
}

/**
 * PQC 算法“摘要”展示值。
 *
 * 后量子算法（ML-DSA / SLH-DSA / Falcon）的签名内部已经使用子算法指定的散列，
 * 因此其摘要算法就是算法标识本身（例如 slh-dsa-shake-256s / MLDSA65 / falcon512）。
 * 返回空串表示不是 PQC。
 */
export function pqcDigestLabel(algorithm?: string | null, spec?: AlgorithmInfo | null): string {
  const family = algorithmFamily(algorithm, spec);
  if (family !== 'PQC') return '';
  const id = normalizeAlgorithmId(spec?.ID || algorithm);
  return id || '';
}

/**
 * 计算某算法当前应使用的摘要展示 ID。
 *
 * - PQC：返回算法子算法 ID（如 slh-dsa-shake-256s）；
 * - SM2 / RSA / ECC：返回用户选择的摘要，未选则返回该族默认值。
 */
export function resolveDigestForAlgorithm(
  algorithm: string | null | undefined,
  chosen: DigestId | '' | null | undefined,
  spec?: AlgorithmInfo | null,
): string {
  const family = algorithmFamily(algorithm, spec);
  if (family === 'PQC') return pqcDigestLabel(algorithm, spec);
  return chosen || defaultDigestFor(family);
}

/** 摘要展示名 */
export function digestLabel(id?: string | null): string {
  if (!id) return '';
  return DIGEST_LABELS[id.toUpperCase()] || id;
}
