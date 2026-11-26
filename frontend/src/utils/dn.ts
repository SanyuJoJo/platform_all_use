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

export type DnSeparator = string;

export const DEFAULT_DN_SEPARATOR: DnSeparator = ',';

export function normalizeSeparator(sep: string): string {
  const raw = sep ?? '';
  if (raw === '') return DEFAULT_DN_SEPARATOR;
  return raw
    .replace(/\\n/g, '\n')
    .replace(/\\t/g, '\t')
    .replace(/\\r/g, '\r');
}

export interface DnParseResult {
  subject: Record<string, string>;
  errors: string[];
  rawKeys: string[];
}

function preprocessDN(input: string, sep: string): string {
  let text = (input ?? '').trim();
  if (!text) return '';

  if (text.startsWith('/')) {
    text = text.slice(1);
  }

  if (text.includes('/')) {
    text = text.replace(/\//g, ',');
  }

  if (sep === ',') {
    text = text.replace(/，/g, ',');
  }

  return text;
}

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

export function validateDN(input: string, separator: string): DnParseResult {
  const result = parseDN(input, separator);
  if (!(input ?? '').trim()) {
    result.errors.push('DN 不能为空');
  }
  return result;
}

export const DN_HELP_KEYS =
  'C、CN、O、OU、ST、E、L、SERIALNUMBER、SURNAME、GIVENNAME';

export const DN_HELP_EXAMPLE =
  'C=CN,CN=Test,O=test,OU=test,ST=BJ,E=email,L=local,SERIALNUMBER=0123,SURNAME=Smith,GIVENNAME=John';