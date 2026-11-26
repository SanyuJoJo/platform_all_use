import { defineStore } from 'pinia';
import { cryptoApi } from '@/api/crypto';
import {
  DEFAULT_ALGORITHMS,
  buildAlgorithmOptions,
  algorithmDisplay,
  algorithmLabel,
  isPqcAlgorithm,
  normalizeAlgorithmFromApi,
  type AlgorithmInfo,
} from '@/utils/crypto-algorithms';

function extractAlgorithmItems(res: any): any[] {
  if (!res) return [];
  if (res.data && !Array.isArray(res.data)) {
    const biz = res.data;
    if (Array.isArray(biz.items)) return biz.items;
    if (biz.data && Array.isArray(biz.data.items)) return biz.data.items;
  }
  if (res.data && Array.isArray(res.data.items)) return res.data.items;
  if (Array.isArray(res.items)) return res.items;
  return [];
}

export const useAlgorithmStore = defineStore('algorithm', {
  state: () => ({
    items: [] as AlgorithmInfo[],
    loaded: false,
    loading: false,
  }),
  getters: {
    options(state): any[] {
      const list = state.items.length ? state.items : DEFAULT_ALGORITHMS;
      return buildAlgorithmOptions(list);
    },
    all(state): AlgorithmInfo[] {
      return state.items.length ? state.items : DEFAULT_ALGORITHMS;
    },
  },
  actions: {
    async ensureLoaded() {
      if (this.loaded || this.loading) return;
      this.loading = true;
      try {
        const res: any = await cryptoApi.listAlgorithms();
        const rawList = extractAlgorithmItems(res);

        const items: AlgorithmInfo[] = rawList
          .map(normalizeAlgorithmFromApi)
          .filter((a) => !!a.ID);

        if (items.length) {
          this.items = items;
        } else {
          console.warn(
            '[algorithm] 接口返回 0 条有效算法，使用内置兜底。rawList=',
            rawList,
          );
          this.items = DEFAULT_ALGORITHMS;
        }
      } catch (e) {
        console.warn('[algorithm] 拉取算法列表失败，使用内置兜底', e);
        this.items = DEFAULT_ALGORITHMS;
      } finally {
        this.loaded = true;
        this.loading = false;
      }
    },
    find(id?: string | null): AlgorithmInfo | null {
      if (!id) return null;
      const list = this.items.length ? this.items : DEFAULT_ALGORITHMS;
      return list.find((a) => a.ID === id) || null;
    },
    displayName(id?: string | null): string {
      return this.find(id)?.DisplayName || algorithmDisplay(id);
    },
    supportsDual(id?: string | null): boolean {
      return this.find(id)?.SupportsDual === true;
    },
    isPqc(id?: string | null): boolean {
      if (this.find(id)?.Category === '后量子') return true;
      return isPqcAlgorithm(id);
    },
    label(v?: string | null): string {
      const byId = this.displayName(v);
      if (byId && byId !== v) return byId;
      return algorithmLabel(v);
    },
  },
});