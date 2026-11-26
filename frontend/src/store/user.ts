import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getMe } from '@/api/auth'
import type { User } from '@/types'

function readUser(): User | null {
  try {
    const raw = localStorage.getItem('user')
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

function readPermissions(): string[] {
  try {
    const parsed = JSON.parse(localStorage.getItem('permissions') || '[]')
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<User | null>(readUser())
  const permissions = ref<string[]>(readPermissions())

  const username = computed(() => user.value?.nickname || user.value?.username || '')
  const isLoggedIn = computed(() => !!token.value)

  function setToken(value: string) {
    token.value = value
    localStorage.setItem('token', value)
  }

  function setUser(value: User) {
    user.value = value
    permissions.value = Array.isArray(value.permissions) ? value.permissions : []
    localStorage.setItem('user', JSON.stringify(value))
    localStorage.setItem('permissions', JSON.stringify(permissions.value))
  }

  function hasPermission(code?: string): boolean {
    if (!code) return true
    return permissions.value.includes(code)
  }

  async function login(userName: string, password: string) {
    const res = await apiLogin({ username: userName, password })
    const data = res?.data
    if (!data?.access_token) throw new Error('登录响应缺少 access_token')
    setToken(data.access_token)
    if (data.user) setUser(data.user)
    return data
  }

  async function fetchMe() {
    const res = await getMe()
    if (res?.data) setUser(res.data)
    return res?.data
  }

  function logout() {
    token.value = ''
    user.value = null
    permissions.value = []
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    localStorage.removeItem('permissions')
  }

  return {
    token,
    user,
    permissions,
    username,
    isLoggedIn,
    setToken,
    setUser,
    hasPermission,
    login,
    fetchMe,
    logout,
  }
})
