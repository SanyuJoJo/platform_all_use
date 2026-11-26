import request from '@/utils/request';
import type {
  User,
  UserCreate,
  UserUpdate,
  PaginatedResponse,
} from '@/types';

export const userApi = {
  getList(params: {
    page: number;
    page_size: number;
    keyword?: string;
    status?: number;
    role_id?: number;
  }) {
    return request.get<PaginatedResponse<User>>('/auth/users', { params });
  },
  create(data: UserCreate) {
    return request.post<User>('/auth/users', data);
  },
  getDetail(id: number) {
    return request.get<User>(`/auth/users/${id}`);
  },
  update(id: number, data: UserUpdate) {
    return request.put<User>(`/auth/users/${id}`, data);
  },
  delete(id: number) {
    return request.delete<null>(`/auth/users/${id}`);
  },
  setStatus(id: number, status: number) {
    return request.patch<User>(`/auth/users/${id}/status`, { status });
  },
  resetPassword(id: number, new_password: string) {
    return request.patch<null>(`/auth/users/${id}/password`, { new_password });
  },
};