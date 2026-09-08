import request from './index';
import axios from 'axios';
import type { AuditLog, ApiResponse, PaginatedResponse } from '@/types';
// 创建一个不经过通用拦截器的 axios 实例，用于导出（处理 blob）
const exportInstance = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1',
  timeout: 30000,
  withCredentials: true,
});
// 请求拦截器：注入 token
exportInstance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
// ✅ 响应拦截器：检测 content-type，如果是 JSON 则解析为错误并 reject
exportInstance.interceptors.response.use(
  (response) => {
    const contentType = response.headers['content-type'] || '';
    // 如果返回的是 JSON（业务错误），则解析并 reject
    if (contentType.includes('application/json')) {
      // 注意：response.data 是字符串，需要解析
      const data = typeof response.data === 'string' ? JSON.parse(response.data) : response.data;
      if (data.code !== 0) {
        return Promise.reject({ code: data.code, message: data.message || '导出失败', data: data.data });
      }
      // 如果 code 为 0 但 content-type 是 JSON，理论上不会发生，但为了安全
      return response;
    }
    // 否则（如 text/csv）直接返回
    return response;
  },
  (error) => {
    // 网络错误或 HTTP 错误，直接抛出
    return Promise.reject(error);
  }
);
export const auditLogApi = {
  // 获取日志列表
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
    return request.get<ApiResponse<PaginatedResponse<AuditLog>>>('/audit-logs', { params });
  },
  // 获取日志详情
  getDetail(logId: number) {
    return request.get<ApiResponse<AuditLog>>(`/audit-logs/${logId}`);
  },
  // 导出日志（使用独立实例，含错误处理）
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