/* ============================================================================
 * 认证授权（系统管理）类型
 * ==========================================================================*/

export interface UserRole {
  id: number
  name: string
  code: string
}

export interface User {
  id: number
  username: string
  nickname: string
  email?: string
  avatar?: string
  status: number
  roles: UserRole[]
  permissions?: string[]
  created_at?: string
  updated_at?: string
}

export interface UserCreate {
  username: string
  password: string
  nickname: string
  email?: string
  role_ids?: number[]
  status?: number
}

export interface UserUpdate {
  nickname?: string
  email?: string
  role_ids?: number[]
  status?: number
}

export interface Role {
  id: number
  name: string
  code: string
  description?: string
  is_system: boolean
  /** 已勾选的菜单 id 列表（含 M/C/F） */
  menuIds?: number[]
  /** 菜单数量（列表接口可选返回的轻量统计） */
  menuCount?: number
  created_at?: string
  updated_at?: string
}

export interface RoleCreate {
  name: string
  code: string
  description?: string
  menuIds?: number[]
}

export interface RoleUpdate {
  name?: string
  description?: string
  menuIds?: number[]
}

export interface Permission {
  id: number
  code: string
  name: string
  module_id: string
  resource: string
  action: string
  created_at?: string
}
