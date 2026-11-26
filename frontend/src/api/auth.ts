import request from '@/utils/request'
import type { User } from '@/types'

export interface LoginResp {
  access_token: string
  refresh_token?: string
  token_type?: string
  expires_in?: number
  user: User
}

export function login(data: { username: string; password: string }) {
  return request.post<LoginResp>('/auth/login', data)
}

export function getMe() {
  return request.get<User>('/auth/me')
}
