import request from './index';
import type { Module, Config, PaginatedResponse } from '@/types';

export const moduleApi = {
  // 泛型只到业务数据结构，不再嵌套 ApiResponse
  getList(params: {
    page: number;
    page_size: number;
    status?: string;
    keyword?: string;
  }) {
    return request.get<PaginatedResponse<Module>>('/modules', { params });
  },
  install(data: {
    install_type: 'zip' | 'path';
    file_path?: string;
    source_path?: string;
  }) {
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
