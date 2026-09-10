import axios from 'axios';

const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1';

const instance = axios.create({
  baseURL,
  timeout: 10000,
  withCredentials: true,
});

// 请求拦截器：注入 token
instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器：统一处理错误码和 401
instance.interceptors.response.use(
  (res) => {
    const { code, message, data } = res.data;
    if (code !== 0) {
      return Promise.reject({ code, message, data });
    }
    // ★ 修复：返回原始 res.data，由调用方通过泛型断言类型
    return res.data;
  },
  (err) => {
    const { response } = err;
    if (response) {
      const { status, data } = response;

      if (status === 401) {
        // 清理本地存储
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('permissions');

        // ★ 修复：微前端下不直接跳转，通过事件通知主应用
        if ((window as any).__POWERED_BY_QIANKUN__) {
          window.dispatchEvent(new CustomEvent('platform:auth-expired'));
        } else {
          const loginUrl = (import.meta.env.VITE_LOGIN_URL as string) || '/login';
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

export default instance;
