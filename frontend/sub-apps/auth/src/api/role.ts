import request from './index';
import type { Role, RoleCreate, RoleUpdate, PaginatedResponse, ApiResponse } from '@/types';
export const roleApi = {
  getList(params: { page: number; page_size: number; keyword?: string }) {
    return request.get<ApiResponse<PaginatedResponse<Role>>>('/auth/roles', { params });
  },
  create(data: RoleCreate) {
    return request.post<ApiResponse<Role>>('/auth/roles', data);
  },
  getDetail(id: number) {
    return request.get<ApiResponse<Role>>(`/auth/roles/${id}`);
  },
  update(id: number, data: RoleUpdate) {
    return request.put<ApiResponse<Role>>(`/auth/roles/${id}`, data);
  },
  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/auth/roles/${id}`);
  },
};
