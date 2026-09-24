import { useMessage } from 'naive-ui';
import { cryptoApi } from '@/api/crypto';
import { useTaskStore } from '@/store/task';
export function useOperation() {
  const message = useMessage();
  const taskStore = useTaskStore();
  async function run<T = unknown>(
    operationId: string,
    params: Record<string, unknown>,
    options?: { timeout_ms?: number; dry_run?: boolean }
  ): Promise<T | { taskId: string }> {
    try {
      const res = await cryptoApi.execute<T>(operationId, params, options);
      const taskId =
        (res as any)?.task_id ||
        (res as any)?.data?.task_id ||
        (res as any)?.data?.taskId;
      if (taskId) {
        taskStore.watch(taskId);
        message.info('任务已提交，正在处理');
        return { taskId };
      }
      message.success('操作成功');
      return res.data as T;
    } catch (e: any) {
      message.error(e.message || '操作失败');
      throw e;
    }
  }
  return { run };
}
