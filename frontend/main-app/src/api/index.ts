import axios from 'axios';
import { useUserStore } from '@/store/user';

// 从环境变量获取后端基础地址，并拼接 API 前缀
const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1';

const instance = axios.create({
  baseURL,
  timeout: 10000,
  withCredentials: true,
});

instance.interceptors.request.use((config) => {
  const userStore = useUserStore();
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`;
  }
  return config;
});

instance.interceptors.response.use(
  (res) => {
    const { code, message, data } = res.data;
    if (code !== 0) {
      return Promise.reject({ code, message, data });
    }
    return res.data;
  },
  (err) => {
    const { response } = err;
    if (response) {
      const { status, data } = response;
      if (status === 401) {
        const userStore = useUserStore();
        userStore.logout();
        window.location.href = '/login';
        return Promise.reject({ code: 10001, message: '登录已过期，请重新登录' });
      }
      const msg = data?.message || '请求失败';
      return Promise.reject({ code: status, message: msg });
    }
    return Promise.reject({ code: -1, message: '网络异常' });
  }
);

export default instance;
