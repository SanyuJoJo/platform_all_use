import axios, { type AxiosRequestConfig } from 'axios';
import type { ApiResponse } from '@/types';
import { errorMessage } from '@/constants/errorCodes';
const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1';
const instance = axios.create({
  baseURL,
  timeout: 15000,
  withCredentials: true,
});
instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
instance.interceptors.response.use(
  (res) => {
    const { code, message, data } = res.data;
    if (code !== 0 && code !== 'OK') {
      return Promise.reject({
        code,
        message: errorMessage(code, message),
        data,
      });
    }
    return res.data;
  },
  (err) => {
    const { response } = err;
    if (response) {
      const { status, data } = response;
      if (status === 401) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('permissions');
        if (window.__POWERED_BY_QIANKUN__) {
          window.dispatchEvent(new CustomEvent('platform:auth-expired'));
        } else {
          window.location.href = '/login';
        }
        return Promise.reject({
          code: 'AUTH_UNAUTHORIZED',
          message: '登录已过期，请重新登录',
        });
      }
      const code = data?.code || data?.error?.code || `HTTP_${status}`;
      const msg = data?.message || data?.error?.message || '请求失败';
      return Promise.reject({
        code,
        message: errorMessage(code, msg),
        data,
      });
    }
    return Promise.reject({ code: 'NETWORK_ERROR', message: '网络异常' });
  }
);
interface ApiClient {
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>;
  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>>;
  put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>>;
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>;
}
export default instance as unknown as ApiClient;
