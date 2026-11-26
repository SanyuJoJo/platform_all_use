/**
 * 全局共享类型。
 *
 * 各业务模块的类型放在 ./<module>.ts 中，这里统一导出，调用方
 * 直接 `import type { Xxx } from '@/types'` 即可。
 */

/** 后端统一响应体 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  timestamp?: string
  requestId?: string
}

/** 通用分页数据（嵌套在 ApiResponse.data 中） */
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages?: number
}

export * from './permission'
export * from './media'
export * from './auth'
export * from './module'
export * from './audit-log'
export * from './license'
export * from './proxy'
export * from './tcps'
export * from './crypto'
