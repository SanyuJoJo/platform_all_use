// src/api/index.ts
import axios from 'axios';
import { useUserStore } from '@/store/user';
import router from '@/router';

const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1';
console.log('[DEBUG] api/index.ts: baseURL =', baseURL);

const instance = axios.create({
  baseURL,
  timeout: 10000,
  withCredentials: true,
});

instance.interceptors.request.use((config) => {
  const userStore = useUserStore();
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`;
    console.log(`[DEBUG] api: 请求 ${config.url} 注入 token`);
  } else {
    console.log(`[DEBUG] api: 请求 ${config.url} 无 token`);
  }
  return config;
});

instance.interceptors.response.use(
  (res) => {
    console.log(`[DEBUG] api: 响应 ${res.config.url} 成功`, res.data);
    const { code, message, data } = res.data;
    if (code !== 0) {
      console.warn(`[DEBUG] api: 业务错误 code=${code}, message=${message}`);
      return Promise.reject({ code, message, data });
    }
    return res.data;
  },
  (err) => {
    const { response } = err;
    if (response) {
      const { status, data } = response;
      console.error(`[DEBUG] api: HTTP 错误 ${status}`, data);
      if (status === 401) {
        const userStore = useUserStore();
        userStore.logout();
        router.push('/login');
        return Promise.reject({ code: 10001, message: '登录已过期，请重新登录' });
      }
      const msg = data?.message || '请求失败';
      return Promise.reject({ code: status, message: msg });
    }
    console.error('[DEBUG] api: 网络异常', err);
    return Promise.reject({ code: -1, message: '网络异常' });
  }
);

export default instance;