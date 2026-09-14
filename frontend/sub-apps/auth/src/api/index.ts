import axios, { type AxiosRequestConfig } from 'axios';
import type { ApiResponse } from '@/types';

const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1';

const instance = axios.create({
  baseURL,
  timeout: 10000,
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
    if (code !== 0) {
      return Promise.reject({ code, message, data });
    }
    // 拦截器拆包：直接返回业务响应体
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
          const loginUrl = import.meta.env.VITE_LOGIN_URL || '/login';
          window.location.href = loginUrl;
        }

        return Promise.reject({ code: 10001, message: '登录已过期，请重新登录' });
      }

      const msg = data?.message || '请求失败';
      return Promise.reject({ code: status, message: msg });
    }
    return Promise.reject({ code: -1, message: '网络异常' });
  }
);

/**
 * 拦截器已把 AxiosResponse<T> 拆成 ApiResponse<T>，
 * 覆盖默认类型让调用方直接得到 Promise<ApiResponse<T>>。
 */
interface ApiClient {
  get<T = unknown>(
    url: string,
    config?: AxiosRequestConfig
  ): Promise<ApiResponse<T>>;
  post<T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<ApiResponse<T>>;
  put<T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<ApiResponse<T>>;
  patch<T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): Promise<ApiResponse<T>>;
  delete<T = unknown>(
    url: string,
    config?: AxiosRequestConfig
  ): Promise<ApiResponse<T>>;
}

export default instance as unknown as ApiClient;
