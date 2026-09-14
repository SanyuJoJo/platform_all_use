import request from './index';
import axios from 'axios';
import type { AuditLog, PaginatedResponse } from '@/types';

/**
 * 导出用的独立 Axios 实例（responseType: 'blob'）。
 * 单独处理，避免被通用拦截器将 blob 当作业务 JSON 拆包。
 */
const exportInstance = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1',
  timeout: 30000,
  withCredentials: true,
});

exportInstance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

exportInstance.interceptors.response.use(
  (response) => {
    // response.headers 类型为 AxiosHeaders | ...，用 String() 归一
    const contentType = String(response.headers['content-type'] || '');

    if (contentType.includes('application/json')) {
      const data =
        typeof response.data === 'string'
          ? JSON.parse(response.data)
          : response.data;
      if (data?.code !== 0) {
        return Promise.reject({
          code: data.code,
          message: data.message || '导出失败',
          data: data.data,
        });
      }
    }
    return response;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export const auditLogApi = {
  // 注意：泛型只到业务数据结构，不再嵌套 ApiResponse
  getList(params: {
    page: number;
    page_size: number;
    module_id?: string;
    user_id?: number;
    action?: string;
    start_time?: string;
    end_time?: string;
    keyword?: string;
  }) {
    return request.get<PaginatedResponse<AuditLog>>('/audit-logs', { params });
  },

  getDetail(logId: number) {
    return request.get<AuditLog>(`/audit-logs/${logId}`);
  },

  exportLogs(params: {
    module_id?: string;
    user_id?: number;
    action?: string;
    start_time?: string;
    end_time?: string;
    keyword?: string;
  }) {
    return exportInstance.get('/audit-logs/export', {
      params,
      responseType: 'blob',
    });
  },
};
