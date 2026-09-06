import request from './index';
import type { User, UserCreate, UserUpdate, PaginatedResponse, ApiResponse } from '@/types';
export const userApi = {
  getList(params: {
    page: number;
    page_size: number;
    keyword?: string;
    status?: number;
    role_id?: number;
  }) {
    return request.get<ApiResponse<PaginatedResponse<User>>>('/auth/users', { params });
  },
  create(data: UserCreate) {
    return request.post<ApiResponse<User>>('/auth/users', data);
  },
  getDetail(id: number) {
    return request.get<ApiResponse<User>>(`/auth/users/${id}`);
  },
  update(id: number, data: UserUpdate) {
    return request.put<ApiResponse<User>>(`/auth/users/${id}`, data);
  },
  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/auth/users/${id}`);
  },
  setStatus(id: number, status: number) {
    return request.patch<ApiResponse<User>>(`/auth/users/${id}/status`, { status });
  },
  resetPassword(id: number, new_password: string) {
    return request.patch<ApiResponse<null>>(`/auth/users/${id}/password`, { new_password });
  },
};
