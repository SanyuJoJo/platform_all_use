import axios from 'axios';
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
