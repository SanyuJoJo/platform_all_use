/** 后端登录响应中的角色可能是对象，也可能仅返回角色编码 */
type RoleLike = string | { name?: string; code?: string }

const ROLE_CODE_LABELS: Record<string, string> = {
  admin: '超级管理员',
  guest: '访客',
}

/** 把角色数组格式化为用于展示的文本 */
export function formatUserRoles(roles?: RoleLike[]): string {
  if (!Array.isArray(roles) || roles.length === 0) return '系统用户'

  const names = roles
    .map((role) => {
      if (typeof role === 'string') return ROLE_CODE_LABELS[role] || role
      return role?.name || role?.code || ''
    })
    .filter((name): name is string => !!name)

  return names.length ? names.join(' / ') : '系统用户'
}
