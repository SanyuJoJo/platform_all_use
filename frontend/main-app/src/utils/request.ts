// main-app/src/utils/request.ts
import axios, { type AxiosInstance, type AxiosError } from 'axios';
import { useUserStore } from '@/store/user';
import { handleError } from './error';
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
});
// 请求拦截器：自动注入 token
request.interceptors.request.use((config) => {
  const userStore = useUserStore();
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`;
  }
  return config;
});
// 响应拦截器：统一错误处理
request.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    const userStore = useUserStore();
    // 401 未授权 → 跳转登录
    if (error.response?.status === 401) {
      userStore.logout();
      window.location.href = '/login';
    }
    return Promise.reject(handleError(error));
  }
);
export default request;
