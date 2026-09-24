import { defineStore } from 'pinia';
import { taskApi } from '@/api/task';
export type TaskStatus =
  | 'PENDING'
  | 'RUNNING'
  | 'SUCCESS'
  | 'FAILED'
  | 'TIMEOUT'
  | 'CANCELLED';
export interface TaskItem {
  task_id: string;
  operation_id: string;
  request_id: string;
  status: TaskStatus;
  progress?: number;
  result_ref?: string;
  error_code?: string;
  error_message?: string;
  created_at?: string;
  started_at?: string;
  finished_at?: string;
}
const TERMINAL: TaskStatus[] = [
  'SUCCESS',
  'FAILED',
  'TIMEOUT',
  'CANCELLED',
];
export const useTaskStore = defineStore('task', {
  state: () => ({
    tasks: [] as TaskItem[],
    timers: {} as Record<string, number>,
    pollInterval: Number(
      import.meta.env.VITE_TASK_POLL_INTERVAL_MS || 1000
    ),
  }),
  actions: {
    upsertTask(task: TaskItem) {
      const idx = this.tasks.findIndex((t) => t.task_id === task.task_id);
      if (idx >= 0) {
        this.tasks[idx] = { ...this.tasks[idx], ...task };
      } else {
        this.tasks.unshift(task);
      }
    },
    watch(taskId: string) {
      if (this.timers[taskId]) return;
      this.timers[taskId] = window.setInterval(async () => {
        try {
          const res = await taskApi.getTask(taskId);
          const task = res.data as unknown as TaskItem;
          this.upsertTask(task);
          if (TERMINAL.includes(task.status)) {
            this.stop(taskId);
          }
        } catch (e) {
          console.warn('[CryptoConsole] 轮询任务失败', taskId, e);
          this.stop(taskId);
        }
      }, this.pollInterval);
    },
    stop(taskId: string) {
      if (this.timers[taskId]) {
        window.clearInterval(this.timers[taskId]);
        delete this.timers[taskId];
      }
    },
    stopAll() {
      Object.keys(this.timers).forEach((id) => this.stop(id));
    },
  },
});
