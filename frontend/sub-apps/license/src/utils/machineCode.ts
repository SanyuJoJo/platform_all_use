import FingerprintJS from '@fingerprintjs/fingerprintjs';
let cachedMachineCode: string | null = null;
export async function getMachineCode(forceRefresh = false): Promise<string> {
  if (!forceRefresh && cachedMachineCode) {
    return cachedMachineCode;
  }
  try {
    const fp = await FingerprintJS.load();
    const result = await fp.get();
    cachedMachineCode = result.visitorId;
    return cachedMachineCode;
  } catch (error) {
    console.warn('[MachineCode] 获取设备指纹失败，使用备用方案:', error);
    let fallback = localStorage.getItem('machineCodeFallback');
    if (!fallback) {
      fallback = 'fallback-' + Math.random().toString(36).substring(2, 15);
      localStorage.setItem('machineCodeFallback', fallback);
    }
    cachedMachineCode = fallback;
    return cachedMachineCode;
  }
}
export async function refreshMachineCode(): Promise<string> {
  cachedMachineCode = null;
  return getMachineCode(true);
}
