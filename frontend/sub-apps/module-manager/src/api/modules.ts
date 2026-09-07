import request from './index';
import type { Module, Config, ApiResponse, PaginatedResponse } from '@/types';
export const moduleApi = {
  // 获取模块列表
  getList(params: {
    page: number;
    page_size: number;
    status?: string;
    keyword?: string;
  }) {
    return request.get<ApiResponse<PaginatedResponse<Module>>>('/modules', { params });
  },
  // 安装模块
  install(data: { install_type: 'zip' | 'path'; file_path?: string; source_path?: string }) {
    return request.post<ApiResponse<Module>>('/modules', data);
  },
  // 卸载模块
  uninstall(moduleId: string, force: boolean = false) {
    return request.delete<ApiResponse<null>>(`/modules/${moduleId}`, { params: { force } });
  },
  // 启用模块
  enable(moduleId: string) {
    return request.post<ApiResponse<Module>>(`/modules/${moduleId}/enable`);
  },
  // 停用模块
  disable(moduleId: string) {
    return request.post<ApiResponse<Module>>(`/modules/${moduleId}/disable`);
  },
  // 获取模块配置
  getConfig(moduleId: string) {
    return request.get<ApiResponse<Config>>(`/modules/${moduleId}/config`);
  },
  // 更新模块配置
  updateConfig(moduleId: string, config: any) {
    return request.put<ApiResponse<Config>>(`/modules/${moduleId}/config`, { config });
  },
};
