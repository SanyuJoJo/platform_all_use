import request from './index';
import type { User } from '@/store/user';
interface LoginParams {
  username: string;
  password: string;
}
interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}
export function login(data: LoginParams) {
  return request.post<LoginResponse>('/auth/login', data);
}
export function getMe() {
  return request.get<User>('/auth/me');
}
