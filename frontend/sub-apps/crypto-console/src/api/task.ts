import request from './index';
export const taskApi = {
  getTask(taskId: string) {
    return request.get(`/tasks/${taskId}`);
  },
  listTasks(params?: Record<string, unknown>) {
    return request.get('/tasks', { params });
  },
  cancelTask(taskId: string) {
    return request.post(`/tasks/${taskId}/cancel`);
  },
};
