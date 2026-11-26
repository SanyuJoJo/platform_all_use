import request from '@/utils/request';
import type { Module, Config, InstallPayload } from '@/types';

export const moduleApi = {
  getList(params: {
    page: number;
    page_size: number;
    status?: string;
    keyword?: string;
  }) {
    return request.get<{ items: Module[]; total: number }>('/modules', { params });
  },
  install(data: InstallPayload) {
    return request.post<Module>('/modules', data);
  },
  uninstall(moduleId: string, force = false) {
    return request.delete<null>(`/modules/${moduleId}`, {
      params: { force },
    });
  },
  enable(moduleId: string) {
    return request.post<Module>(`/modules/${moduleId}/enable`);
  },
  disable(moduleId: string) {
    return request.post<Module>(`/modules/${moduleId}/disable`);
  },
  getConfig(moduleId: string) {
    return request.get<Config>(`/modules/${moduleId}/config`);
  },
  updateConfig(moduleId: string, config: Config) {
    return request.put<Config>(`/modules/${moduleId}/config`, { config });
  },
};