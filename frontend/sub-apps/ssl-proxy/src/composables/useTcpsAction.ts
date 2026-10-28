// src/composables/useTcpsAction.ts
import { ref } from 'vue';
import { useMessage } from 'naive-ui';
import tcpsApi from '@/api/tcps';
import type { TcpsStatus } from '@/types/tcps';
export type TcpsAction = 'start' | 'stop' | 'restart';
export function useTcpsAction() {
  const message = useMessage();
  const pending = ref<Record<string, TcpsAction | undefined>>({});
  function isPending(id: string): boolean {
    return !!pending.value[id];
  }
  function pendingAction(id: string): TcpsAction | null {
    return pending.value[id] ?? null;
  }
  function setPending(id: string, action: TcpsAction | null) {
    const next = { ...pending.value };
    if (action) next[id] = action;
    else delete next[id];
    pending.value = next;
  }
  async function runOp(
    id: string,
    action: TcpsAction,
    label: string,
    op: () => Promise<unknown>,
  ): Promise<boolean> {
    if (isPending(id)) {
      message.warning(`正在${label}中，请稍候…`);
      return false;
    }
    setPending(id, action);
    try {
      await op();
      message.success(`${label}成功`);
      return true;
    } catch (e: any) {
      const code = e?.code;
      if (code === 40902) message.warning(e?.message || '实例正在处理中，请稍候');
      else if (code === 40901) message.error(e?.message || '端口冲突');
      else message.error(e?.message || `${label}失败`);
      return false;
    } finally {
      setPending(id, null);
    }
  }
  return {
    pending,
    isPending,
    pendingAction,
    start: (id: string) => runOp(id, 'start', '启动', () => tcpsApi.start(id)),
    stop: (id: string) => runOp(id, 'stop', '停止', () => tcpsApi.stop(id)),
    restart: (id: string) => runOp(id, 'restart', '重启', () => tcpsApi.restart(id)),
    isRunning: (s: TcpsStatus) => s === 'RUNNING' || s === 'STARTING',
    isTransitioning: (s: TcpsStatus) => s === 'STARTING' || s === 'STOPPING',
  };
}
