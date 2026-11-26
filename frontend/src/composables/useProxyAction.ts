import { ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import proxyApi from '@/api/proxy';
import type { ProxyStatus } from '@/types';

export type ProxyAction = 'start' | 'stop' | 'restart';

export function useProxyAction() {
  const pending = ref<Record<string, ProxyAction | undefined>>({});

  function isPending(serviceId: string): boolean {
    return !!pending.value[serviceId];
  }

  function pendingAction(serviceId: string): ProxyAction | null {
    return pending.value[serviceId] ?? null;
  }

  function setPending(serviceId: string, action: ProxyAction | null) {
    const next = { ...pending.value };
    if (action) {
      next[serviceId] = action;
    } else {
      delete next[serviceId];
    }
    pending.value = next;
  }

  async function runOp(
    serviceId: string,
    action: ProxyAction,
    opName: string,
    op: () => Promise<unknown>,
  ): Promise<boolean> {
    if (isPending(serviceId)) {
      MessagePlugin.warning(`正在${opName}中，请稍候…`);
      return false;
    }
    setPending(serviceId, action);
    try {
      await op();
      MessagePlugin.success(`${opName}成功`);
      return true;
    } catch (e: any) {
      const code = e?.code;
      if (code === 40902) {
        MessagePlugin.warning(e?.message || '实例正在处理中，请稍候');
      } else if (code === 40901) {
        MessagePlugin.error(e?.message || '端口冲突');
      } else {
        MessagePlugin.error(e?.message || `${opName}失败`);
      }
      return false;
    } finally {
      setPending(serviceId, null);
    }
  }

  function start(serviceId: string) {
    return runOp(serviceId, 'start', '启动', () => proxyApi.start(serviceId));
  }
  function stop(serviceId: string) {
    return runOp(serviceId, 'stop', '停止', () => proxyApi.stop(serviceId));
  }
  function restart(serviceId: string) {
    return runOp(serviceId, 'restart', '重启', () => proxyApi.restart(serviceId));
  }

  function isRunning(status: ProxyStatus): boolean {
    return status === 'RUNNING' || status === 'STARTING';
  }
  function isTransitioning(status: ProxyStatus): boolean {
    return status === 'STARTING' || status === 'STOPPING';
  }

  return {
    start,
    stop,
    restart,
    isRunning,
    isTransitioning,
    isPending,
    pendingAction,
    pending,
  };
}