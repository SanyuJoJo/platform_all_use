import { ref } from 'vue';
import { useMessage } from 'naive-ui';
import proxyApi from '@/api/proxy';
import type { ProxyStatus } from '@/types';

export type ProxyAction = 'start' | 'stop' | 'restart';

export function useProxyAction() {
  const message = useMessage();

  // service_id → 当前正在执行的动作（undefined 表示无操作）
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
    op: () => Promise<unknown>
  ): Promise<boolean> {
    if (isPending(serviceId)) {
      message.warning(`正在${opName}中，请稍候…`);
      return false;
    }
    setPending(serviceId, action);
    try {
      await op();
      // ★ 后端 start/stop/restart 均为同步：返回时状态已切换完成
      message.success(`${opName}成功`);
      return true;
    } catch (e: any) {
      const code = e?.code;
      if (code === 40902) {
        message.warning(e?.message || '实例正在处理中，请稍候');
      } else if (code === 40901) {
        message.error(e?.message || '端口冲突');
      } else {
        message.error(e?.message || `${opName}失败`);
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
