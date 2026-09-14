import request from './index';
import type {
  Role,
  RoleCreate,
  RoleUpdate,
  PaginatedResponse,
} from '@/types';

export const roleApi = {
  getList(params: { page: number; page_size: number; keyword?: string }) {
    return request.get<PaginatedResponse<Role>>('/auth/roles', { params });
  },
  create(data: RoleCreate) {
    return request.post<Role>('/auth/roles', data);
  },
  getDetail(id: number) {
    return request.get<Role>(`/auth/roles/${id}`);
  },
  update(id: number, data: RoleUpdate) {
    return request.put<Role>(`/auth/roles/${id}`, data);
  },
  delete(id: number) {
    return request.delete<null>(`/auth/roles/${id}`);
  },
};
