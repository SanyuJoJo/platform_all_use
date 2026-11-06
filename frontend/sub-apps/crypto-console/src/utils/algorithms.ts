// frontend/sub-apps/crypto-console/src/utils/algorithms.ts
//
// 与后端 algorithms.go 保持一致的算法 ID 与展示映射。
// - algorithm 提交请使用本文件的 ALGORITHM_IDS 中的规范 ID（SLH-DSA 全小写）；
// - 展示层通过 ALGORITHM_DISPLAY / PUBLIC_KEY_LABEL 做映射；
// - 实际可用算法请以后端 GET /crypto/algorithms 为准，此处仅为兜底。

export const ALGORITHM_IDS = [
  'SM2',
  'RSA1024', 'RSA2048', 'RSA4096',
  'MLDSA44', 'MLDSA65', 'MLDSA87',
  'falcon512', 'falcon1024',
  'slh-dsa-sha2-128s', 'slh-dsa-sha2-128f',
  'slh-dsa-sha2-192s', 'slh-dsa-sha2-192f',
  'slh-dsa-sha2-256s', 'slh-dsa-sha2-256f',
  'slh-dsa-shake-128s', 'slh-dsa-shake-128f',
  'slh-dsa-shake-192s', 'slh-dsa-shake-192f',
  'slh-dsa-shake-256s', 'slh-dsa-shake-256f',
] as const;

export type AlgorithmId = (typeof ALGORITHM_IDS)[number];

// ============================================================================
// 算法 ID（DB 规范 ID）→ 展示名
// ★ Falcon 统一用 "Falcon(512)" / "Falcon(1024)" 括号形态
// ============================================================================
export const ALGORITHM_DISPLAY: Record<string, string> = {
  SM2: 'SM2',
  RSA1024: 'RSA-1024',
  RSA2048: 'RSA-2048',
  RSA4096: 'RSA-4096',
  MLDSA44: 'ML-DSA-44',
  MLDSA65: 'ML-DSA-65',
  MLDSA87: 'ML-DSA-87',
  falcon512: 'Falcon(512)',
  falcon1024: 'Falcon(1024)',
  'slh-dsa-sha2-128s': 'SLH-DSA-SHA2-128s',
  'slh-dsa-sha2-128f': 'SLH-DSA-SHA2-128f',
  'slh-dsa-sha2-192s': 'SLH-DSA-SHA2-192s',
  'slh-dsa-sha2-192f': 'SLH-DSA-SHA2-192f',
  'slh-dsa-sha2-256s': 'SLH-DSA-SHA2-256s',
  'slh-dsa-sha2-256f': 'SLH-DSA-SHA2-256f',
  'slh-dsa-shake-128s': 'SLH-DSA-SHAKE-128s',
  'slh-dsa-shake-128f': 'SLH-DSA-SHAKE-128f',
  'slh-dsa-shake-192s': 'SLH-DSA-SHAKE-192s',
  'slh-dsa-shake-192f': 'SLH-DSA-SHAKE-192f',
  'slh-dsa-shake-256s': 'SLH-DSA-SHAKE-256s',
  'slh-dsa-shake-256f': 'SLH-DSA-SHAKE-256f',
};

// ============================================================================
// 证书 public_key_algorithm（解析结果）→ 展示名
//
// ★ 不再对 ECC 形态做改写：后端解析出的 "ECC(256)" / "ECC(384)" / "ECC(521)"
//   原样展示（通过 publicKeyLabel 的 `|| v` 兜底）。
//
// ★ Falcon 用括号形态：若后端解析结果已是 "Falcon(512)" / "Falcon(1024)"，
//   原样展示；若为 "Falcon-512" / "Falcon-1024"（连字符形态），
//   通过此映射统一成括号形态。
//
// ★ SM2 证书的展示依赖 certificate.algorithm（DB 规范 ID = "SM2"）。
// ============================================================================
export const PUBLIC_KEY_LABEL: Record<string, string> = {
  RSA: 'RSA',
  'ML-DSA-44': 'ML-DSA-44',
  'ML-DSA-65': 'ML-DSA-65',
  'ML-DSA-87': 'ML-DSA-87',
  // Falcon：连字符形态 → 括号形态
  'Falcon-512': 'Falcon(512)',
  'Falcon-1024': 'Falcon(1024)',
  // Falcon：若后端已是括号形态，原样保留（冗余声明便于自查）
  'Falcon(512)': 'Falcon(512)',
  'Falcon(1024)': 'Falcon(1024)',
  'SLH-DSA-SHA2-128s': 'SLH-DSA-SHA2-128s',
  'SLH-DSA-SHA2-128f': 'SLH-DSA-SHA2-128f',
  'SLH-DSA-SHA2-192s': 'SLH-DSA-SHA2-192s',
  'SLH-DSA-SHA2-192f': 'SLH-DSA-SHA2-192f',
  'SLH-DSA-SHA2-256s': 'SLH-DSA-SHA2-256s',
  'SLH-DSA-SHA2-256f': 'SLH-DSA-SHA2-256f',
  'SLH-DSA-SHAKE-128s': 'SLH-DSA-SHAKE-128s',
  'SLH-DSA-SHAKE-128f': 'SLH-DSA-SHAKE-128f',
  'SLH-DSA-SHAKE-192s': 'SLH-DSA-SHAKE-192s',
  'SLH-DSA-SHAKE-192f': 'SLH-DSA-SHAKE-192f',
  'SLH-DSA-SHAKE-256s': 'SLH-DSA-SHAKE-256s',
  'SLH-DSA-SHAKE-256f': 'SLH-DSA-SHAKE-256f',
};

export interface AlgorithmInfo {
  ID: string;
  DisplayName: string;
  Family?: string;
  Category?: string;
  KeyBits?: number;
  SupportsDual?: boolean;
  Notes?: string;
}

// 后端不可用时的兜底列表，仅在算法接口失败时使用
export const DEFAULT_ALGORITHMS: AlgorithmInfo[] = [
  { ID: 'SM2', DisplayName: 'SM2', Family: 'SM2', Category: '国密', KeyBits: 256, SupportsDual: true, Notes: '国密标准；支持数字信封、双证签发、PKCS#12 导出' },
  { ID: 'RSA1024', DisplayName: 'RSA-1024', Category: 'RSA', KeyBits: 1024, SupportsDual: false, Notes: '历史兼容，建议迁移到 RSA-2048+' },
  { ID: 'RSA2048', DisplayName: 'RSA-2048', Category: 'RSA', KeyBits: 2048, SupportsDual: false, Notes: '主流安全级别' },
  { ID: 'RSA4096', DisplayName: 'RSA-4096', Category: 'RSA', KeyBits: 4096, SupportsDual: false, Notes: '高安全级别，性能开销较大' },
  { ID: 'MLDSA44', DisplayName: 'ML-DSA-44', Category: '后量子', SupportsDual: false, Notes: 'FIPS 204，NIST Level 2' },
  { ID: 'MLDSA65', DisplayName: 'ML-DSA-65', Category: '后量子', SupportsDual: false, Notes: 'FIPS 204，NIST Level 3' },
  { ID: 'MLDSA87', DisplayName: 'ML-DSA-87', Category: '后量子', SupportsDual: false, Notes: 'FIPS 204，NIST Level 5' },
  { ID: 'falcon512', DisplayName: 'Falcon(512)', Category: '后量子', SupportsDual: false, Notes: 'NIST PQC 标准，Level 1' },
  { ID: 'falcon1024', DisplayName: 'Falcon(1024)', Category: '后量子', SupportsDual: false, Notes: 'NIST PQC 标准，Level 5' },
  { ID: 'slh-dsa-sha2-128s', DisplayName: 'SLH-DSA-SHA2-128s', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-sha2-128f', DisplayName: 'SLH-DSA-SHA2-128f', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-sha2-192s', DisplayName: 'SLH-DSA-SHA2-192s', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-sha2-192f', DisplayName: 'SLH-DSA-SHA2-192f', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-sha2-256s', DisplayName: 'SLH-DSA-SHA2-256s', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-sha2-256f', DisplayName: 'SLH-DSA-SHA2-256f', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-shake-128s', DisplayName: 'SLH-DSA-SHAKE-128s', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-shake-128f', DisplayName: 'SLH-DSA-SHAKE-128f', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-shake-192s', DisplayName: 'SLH-DSA-SHAKE-192s', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-shake-192f', DisplayName: 'SLH-DSA-SHAKE-192f', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-shake-256s', DisplayName: 'SLH-DSA-SHAKE-256s', Category: '后量子', SupportsDual: false },
  { ID: 'slh-dsa-shake-256f', DisplayName: 'SLH-DSA-SHAKE-256f', Category: '后量子', SupportsDual: false },
];

/** 依据算法 ID 判定是否为 PQC（仅做兜底判断，优先用 Category） */
export function isPqcAlgorithm(id?: string | null): boolean {
  if (!id) return false;
  return /^(MLDSA|falcon|slh-dsa-)/i.test(id);
}

/** 算法 ID → 展示名（RSA-2048、ML-DSA-65、Falcon(1024)、SLH-DSA-SHA2-128s…） */
export function algorithmDisplay(id?: string | null): string {
  if (!id) return '';
  return ALGORITHM_DISPLAY[id] || id;
}

/**
 * public_key_algorithm（解析结果）→ 展示名
 *   - 命中 PUBLIC_KEY_LABEL 时用映射值；
 *   - 未命中时原样返回（如 "ECC(256)" 会原样展示）。
 */
export function publicKeyLabel(v?: string | null): string {
  if (!v) return '';
  return PUBLIC_KEY_LABEL[v] || v;
}

/** 通用：兼容列表可能返回 ID 或归一化展示，取最合适展示名 */
export function algorithmLabel(v?: string | null): string {
  if (!v) return '';
  return ALGORITHM_DISPLAY[v] || PUBLIC_KEY_LABEL[v] || v;
}

// ============================================================================
// 把后端 snake_case 契约归一化为前端内部 PascalCase 结构。
// ============================================================================
export function normalizeAlgorithmFromApi(raw: any): AlgorithmInfo {
  if (!raw || typeof raw !== 'object') {
    return { ID: '', DisplayName: '' };
  }
  return {
    ID: raw.id ?? raw.ID ?? '',
    DisplayName: raw.display_name ?? raw.DisplayName ?? '',
    Family: raw.family ?? raw.Family,
    Category: raw.category ?? raw.Category,
    KeyBits: raw.key_bits ?? raw.KeyBits,
    SupportsDual: raw.supports_dual ?? raw.SupportsDual ?? false,
    Notes: raw.notes ?? raw.Notes,
  };
}

// ============================================================================
// 即使上游漏做归一化，也保证 label/value 不为 undefined。
// ============================================================================
export function buildAlgorithmOptions(items: AlgorithmInfo[]): any[] {
  const groups = new Map<string, Array<{ label: string; value: string }>>();
  for (const a of items) {
    if (!a || !a.ID) continue;

    const cat = a.Category || '其他';
    if (!groups.has(cat)) groups.set(cat, []);

    const label = a.DisplayName || algorithmDisplay(a.ID) || a.ID;
    const value = a.ID;

    groups.get(cat)!.push({ label, value });
  }
  return Array.from(groups.entries()).map(([label, children]) => ({
    type: 'group',
    label,
    key: label,
    children,
  }));
}
