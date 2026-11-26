export const ALGORITHM_IDS = [
  'SM2',
  'RSA1024',
  'RSA2048',
  'RSA4096',
  'MLDSA44',
  'MLDSA65',
  'MLDSA87',
  'falcon512',
  'falcon1024',
  'slh-dsa-sha2-128s',
  'slh-dsa-sha2-128f',
  'slh-dsa-sha2-192s',
  'slh-dsa-sha2-192f',
  'slh-dsa-sha2-256s',
  'slh-dsa-sha2-256f',
  'slh-dsa-shake-128s',
  'slh-dsa-shake-128f',
  'slh-dsa-shake-192s',
  'slh-dsa-shake-192f',
  'slh-dsa-shake-256s',
  'slh-dsa-shake-256f',
] as const;

export type AlgorithmId = (typeof ALGORITHM_IDS)[number];

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

export const PUBLIC_KEY_LABEL: Record<string, string> = {
  RSA: 'RSA',
  'ML-DSA-44': 'ML-DSA-44',
  'ML-DSA-65': 'ML-DSA-65',
  'ML-DSA-87': 'ML-DSA-87',
  'Falcon-512': 'Falcon(512)',
  'Falcon-1024': 'Falcon(1024)',
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

export function isPqcAlgorithm(id?: string | null): boolean {
  if (!id) return false;
  return /^(MLDSA|falcon|slh-dsa-)/i.test(id);
}

export function algorithmDisplay(id?: string | null): string {
  if (!id) return '';
  return ALGORITHM_DISPLAY[id] || id;
}

export function publicKeyLabel(v?: string | null): string {
  if (!v) return '';
  return PUBLIC_KEY_LABEL[v] || v;
}

export function algorithmLabel(v?: string | null): string {
  if (!v) return '';
  return ALGORITHM_DISPLAY[v] || PUBLIC_KEY_LABEL[v] || v;
}

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