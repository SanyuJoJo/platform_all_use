import { ref } from 'vue';
import proxyApi from '@/api/proxy';

const certMap = ref<Record<string, string>>({});
let loaded = false;
let loadingPromise: Promise<void> | null = null;

export function useCertMap() {
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
              if (!map[c.cert_id]) {
                map[c.cert_id] = c.subject_cn || c.cert_id;
              }
            }
          }
        }
        certMap.value = map;
        loaded = true;
      } catch {
        // silent
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