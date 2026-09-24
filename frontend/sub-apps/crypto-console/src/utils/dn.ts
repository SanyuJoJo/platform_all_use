// frontend/sub-apps/crypto-console/src/utils/dn.ts

/** DN 支持的项（与后端 ValidateSubject 对齐） */
export const DN_ALLOWED_KEYS = [
  'C',
  'CN',
  'O',
  'OU',
  'ST',
  'E',
  'L',
  'SERIALNUMBER',
  'SURNAME',
  'GIVENNAME',
] as const;

export type DnKey = (typeof DN_ALLOWED_KEYS)[number];

/** 分隔符类型：用户自由输入的字符串 */
export type DnSeparator = string;

/** 默认分隔符：逗号 */
export const DEFAULT_DN_SEPARATOR: DnSeparator = ',';

/**
 * 归一化用户输入的分隔符：
 *   - 空字符串 → 默认逗号
 *   - 支持 \n \t \r 转义（用户可直接输入字面 "\n" 表示换行）
 *   - 其他情况按原样返回（例如 ";"、"|"、"|"、"，"）
 */
export function normalizeSeparator(sep: string): string {
  const raw = sep ?? '';
  if (raw === '') return DEFAULT_DN_SEPARATOR;
  return raw
    .replace(/\\n/g, '\n')
    .replace(/\\t/g, '\t')
    .replace(/\\r/g, '\r');
}

export interface DnParseResult {
  /** 解析后的 subject 对象，直接用于 params.subject */
  subject: Record<string, string>;
  /** 解析过程中的错误（字段级别），为空表示无错误 */
  errors: string[];
  /** 识别到的原始 key（保留用户输入的大小写，便于提示） */
  rawKeys: string[];
}

/**
 * 输入预处理：兼容多种 DN 书写风格。
 *
 * 处理内容：
 *   1. 去首尾空白
 *   2. 剥离前导 "/"（OpenSSL -subj 风格，如 "/C=CN/CN=TEST"）
 *   3. 若整体以 "/" 作分隔符，先归一为逗号再走标准拆分
 *   4. 若用户配置的分隔符是 ASCII 逗号，则同时接受全角逗号 "，"
 */
function preprocessDN(input: string, sep: string): string {
  let text = (input ?? '').trim();
  if (!text) return '';

  // 2. 剥离前导 "/"
  if (text.startsWith('/')) {
    text = text.slice(1);
  }

  // 3. "/" 作为分隔符的情况，归一为逗号
  //    注意：仅当 text 中不存在 "=" 之前已经用了逗号时才做整体替换，
  //    避免误伤值中本身含 "/" 的合法情况（如 URL）。DN 值中 "/" 不常见，
  //    这里采取保守策略：仅当 "/" 出现在 key=value 之间（即 "=xxx/" 或 "/" 紧跟 key）时才替换。
  if (text.includes('/')) {
    text = text.replace(/\//g, ',');
  }

  // 4. 全角逗号兼容
  if (sep === ',') {
    text = text.replace(/，/g, ',');
  }

  return text;
}

/**
 * 解析 DN 字符串为 subject 对象。
 *
 * 支持的输入示例：
 *   - "C=CN,CN=TEST"
 *   - "C=CN，CN=TEST"        （全角逗号，sep=',' 时自动兼容）
 *   - "/C=CN/CN=TEST"        （OpenSSL -subj 风格）
 *   - "C=CN;CN=TEST"         （sep=';'）
 *   - "C=CN\nCN=TEST"        （sep='\n'）
 */
export function parseDN(input: string, separator: string): DnParseResult {
  const errors: string[] = [];
  const subject: Record<string, string> = {};
  const rawKeys: string[] = [];

  const sep = normalizeSeparator(separator);
  const working = preprocessDN(input, sep);

  if (!working) {
    return { subject, errors, rawKeys };
  }

  const allowedSet = new Set<string>(DN_ALLOWED_KEYS);
  const parts = working.split(sep);

  for (const raw of parts) {
    const item = raw.trim();
    if (!item) continue;

    const idx = item.indexOf('=');
    if (idx <= 0) {
      errors.push(`无效项：「${item}」缺少 “=” 或键为空`);
      continue;
    }
    const rawKey = item.slice(0, idx).trim();
    const value = item.slice(idx + 1).trim();
    const key = rawKey.toUpperCase();

    if (!allowedSet.has(key)) {
      errors.push(`不支持的 DN 项：「${rawKey}」`);
      continue;
    }
    if (!value) {
      errors.push(`DN 项「${key}」的值为空`);
      continue;
    }
    // C 字段必须是 2 位国家码（ISO 3166-1 alpha-2）
    if (key === 'C' && value.length !== 2) {
      errors.push(`DN 项 C 必须是 2 位国家码，当前为「${value}」`);
      continue;
    }
    if (value.length > 256) {
      errors.push(`DN 项「${key}」的值过长（>256）`);
      continue;
    }
    if (subject[key]) {
      errors.push(`DN 项「${key}」重复出现`);
      continue;
    }
    subject[key] = value;
    rawKeys.push(rawKey);
  }

  return { subject, errors, rawKeys };
}

/**
 * 校验 DN：仅要求非空且语法正确，不强制 CN。
 * 例如只输入 "C=CN" 也是合法的。
 */
export function validateDN(input: string, separator: string): DnParseResult {
  const result = parseDN(input, separator);
  if (!(input ?? '').trim()) {
    result.errors.push('DN 不能为空');
  }
  return result;
}

/** 帮助文案（放在 ? 弹层里，与文档一致） */
export const DN_HELP_KEYS =
  'C、CN、O、OU、ST、E、L、SERIALNUMBER、SURNAME、GIVENNAME';

export const DN_HELP_EXAMPLE =
  'C=CN,CN=Test,O=test,OU=test,ST=BJ,E=email,L=local,SERIALNUMBER=0123,SURNAME=Smith,GIVENNAME=John';
