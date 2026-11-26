/* ============================================================================
 * 通用工具函数
 * 路径：frontend/sub-apps/media/src/utils/format.ts
 * ==========================================================================*/

/* ---------- 时长 ---------- */
export function formatDuration(seconds: number): string {
  if (!seconds || seconds < 0) return '00:00';
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  const pad = (n: number) => String(n).padStart(2, '0');
  return h > 0 ? `${pad(h)}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`;
}

/* ---------- 文件大小 ---------- */
export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const kb = bytes / 1024;
  if (kb < 1024) return `${kb.toFixed(1)} KB`;
  const mb = kb / 1024;
  if (mb < 1024) return `${mb.toFixed(1)} MB`;
  return `${(mb / 1024).toFixed(2)} GB`;
}

/* ---------- 读取 Blob 片段 ---------- */
function readSlice(file: Blob, start: number, end: number): Promise<ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const r = new FileReader();
    r.onload = () => resolve(r.result as ArrayBuffer);
    r.onerror = () => reject(r.error);
    r.readAsArrayBuffer(file.slice(start, end));
  });
}

/* ---------- 判断是否安全上下文（https / localhost） ---------- */
function hasWebCrypto(): boolean {
  return typeof crypto !== 'undefined' && !!crypto.subtle;
}

/* ---------- FNV-1a 32 位（纯 JS，无需 crypto） ---------- */
function fnv1a(bytes: Uint8Array, seed: number): string {
  let h = seed >>> 0;
  for (let i = 0; i < bytes.length; i++) {
    h ^= bytes[i];
    h = Math.imul(h, 0x01000193) >>> 0;
  }
  return h.toString(16).padStart(8, '0');
}

/**
 * 计算文件"指纹"，用于秒传判断。
 *
 * 策略：
 *   1) 安全上下文（https / localhost）：头尾各 1MB 的 SHA-256
 *   2) 非安全上下文（http 局域网）：降级为 FNV-1a 双种子 + 文件元数据
 *
 * 降级返回形如：`fnv1a-<name>::<size>::<mtime>::<h1h2>`
 * 后端只需当作普通字符串存储/比较，无需感知前缀。
 */
export async function computeFileHash(file: File): Promise<string> {
  const CHUNK = 1024 * 1024;
  const headBuf = await readSlice(file, 0, Math.min(CHUNK, file.size));
  const tailStart = Math.max(CHUNK, file.size - CHUNK);
  const tailBuf =
    file.size > CHUNK ? await readSlice(file, tailStart, file.size) : new ArrayBuffer(0);

  const combined = new Uint8Array(headBuf.byteLength + tailBuf.byteLength);
  combined.set(new Uint8Array(headBuf), 0);
  combined.set(new Uint8Array(tailBuf), headBuf.byteLength);

  /* ---------- 首选 SHA-256 ---------- */
  if (hasWebCrypto()) {
    try {
      const buf = await crypto.subtle.digest('SHA-256', combined);
      return Array.from(new Uint8Array(buf))
        .map((b) => b.toString(16).padStart(2, '0'))
        .join('');
    } catch (e) {
      console.warn('[computeFileHash] crypto.subtle.digest 失败，降级 FNV-1a', e);
    }
  } else {
    console.warn(
      '[computeFileHash] 当前非安全上下文（非 https / 非 localhost），' +
        'crypto.subtle 不可用，降级为 FNV-1a 指纹',
    );
  }

  /* ---------- 降级：FNV-1a 双种子 + 元数据 ---------- */
  const h1 = fnv1a(combined, 0x811c9dc5);
  const h2 = fnv1a(combined, 0x01000193);
  const meta = `${file.name}::${file.size}::${file.lastModified ?? 0}`;
  return `fnv1a-${meta}::${h1}${h2}`;
}

/* ---------- 根据 MIME / 扩展名识别媒体类型 ---------- */
export function detectMediaType(mime: string, name: string): 'video' | 'audio' {
  if (mime.startsWith('audio/')) return 'audio';
  if (mime.startsWith('video/')) return 'video';
  const ext = name.split('.').pop()?.toLowerCase() ?? '';
  return ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a'].includes(ext)
    ? 'audio'
    : 'video';
}
