import axios, { type AxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types'

/**
 * 统一请求实例。
 *
 * 约定：后端响应体为 { code, message, data }，code === 0 表示成功。
 * 拦截器成功后返回整个响应体（调用方使用 res.data）；
 * 失败时 reject 一个普通对象 { code, message, data }。
 */
const baseURL = (import.meta.env.VITE_API_BASE_URL || '') + '/api/v1'

const instance = axios.create({
  baseURL,
  timeout: 20000,
  withCredentials: true,
})

instance.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

instance.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && typeof body.code === 'number' && body.code !== 0) {
      return Promise.reject({ code: body.code, message: body.message, data: body.data })
    }
    return body
  },
  (err) => {
    const response = err?.response
    if (response) {
      const { status, data } = response
      if (status === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        localStorage.removeItem('permissions')
        if (!window.location.pathname.startsWith('/login')) {
          window.location.href = '/login'
        }
        return Promise.reject({ code: 10001, message: '登录已过期，请重新登录' })
      }
      return Promise.reject({
        code: status,
        message: data?.message || `请求失败（${status}）`,
        data: data?.data,
      })
    }
    return Promise.reject({ code: -1, message: '网络异常，请检查网络连接' })
  },
)

interface ApiClient {
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>>
}

export default instance as unknown as ApiClient
