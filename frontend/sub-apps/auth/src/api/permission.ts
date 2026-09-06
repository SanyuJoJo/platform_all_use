import request from './index';
import type { Permission, ApiResponse } from '@/types';
export const permissionApi = {
  getList(params?: { module_id?: string; resource?: string }) {
    return request.get<ApiResponse<Permission[]>>('/auth/permissions', { params });
  },
  getByModule(moduleId: string) {
    return request.get<ApiResponse<Permission[]>>(`/auth/permissions/modules/${moduleId}`);
  },
};
