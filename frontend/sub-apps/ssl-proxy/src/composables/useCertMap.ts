// src/composables/useCertMap.ts
import { ref } from 'vue';
import proxyApi from '@/api/proxy';

// 全局单例：证书 ID → 使用者（subject_cn）
const certMap = ref<Record<string, string>>({});
let loaded = false;
let loadingPromise: Promise<void> | null = null;

export function useCertMap() {
  /**
   * 加载全部证书（国际+国密签名+国密加密），合并成一张映射表。
   * 幂等：多次调用只发一次请求。
   */
  async function load(force = false): Promise<void> {
    if (loaded && !force) return;
    if (loadingPromise && !force) return loadingPromise;

    loadingPromise = (async () => {
      const map: Record<string, string> = {};
      try {
        const results = await Promise.allSettled([
          proxyApi.listAvailableCerts({ purpose: 'sign', protocol: 'http' }),
          proxyApi.listAvailableCerts({ purpose: 'sign', protocol: 'gmssl' }),
          proxyApi.listAvailableCerts({ purpose: 'enc', protocol: 'gmssl' }),
        ]);
        for (const r of results) {
          if (r.status === 'fulfilled') {
            for (const c of r.value.data.items || []) {
              // 已存在时不覆盖（同名 cert_id 一般不会重复，保留先拿到的）
              if (!map[c.cert_id]) {
                map[c.cert_id] = c.subject_cn || c.cert_id;
              }
            }
          }
        }
        certMap.value = map;
        loaded = true;
      } catch {
        // 静默失败，展示回退到 cert_id
      } finally {
        loadingPromise = null;
      }
    })();

    return loadingPromise;
  }

  function certName(certId?: string): string {
    if (!certId) return '';
    return certMap.value[certId] || certId;
  }

  return { certMap, load, certName };
}
