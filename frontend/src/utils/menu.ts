import type { RouteNode } from '@/types'
import { isMainAppId } from '@/types'

/** 侧边栏菜单节点（TDesign Menu 结构） */
export interface MenuNode {
  value: string
  label: string
  icon?: string
  path?: string
  children?: MenuNode[]
}

/** 规范化路径：保证以 / 开头，且去掉重复斜杠 */
export function normalizePath(path?: string): string {
  let p = String(path ?? '').trim()
  if (!p) return '/'
  if (!p.startsWith('/')) p = '/' + p
  p = p.replace(/\/{2,}/g, '/')
  return p
}

/**
 * 计算节点在浏览器中的完整路径。
 * - main/platform：直接使用 path
 * - 其它模块：/{microApp}{path}
 */
export function fullPathOf(node: RouteNode): string {
  const p = normalizePath(node.path)
  if (isMainAppId(node.microApp)) return p
  if (p === '/') return `/${node.microApp}/`
  return `/${node.microApp}${p}`
}

/** 把后端路由树转换为侧边栏菜单树（过滤隐藏节点） */
export function buildMenuTree(nodes: RouteNode[]): MenuNode[] {
  const out: MenuNode[] = []
  for (const node of nodes) {
    if (node.meta?.hidden) continue
    const children = node.children?.length ? buildMenuTree(node.children) : undefined
    const isDirectory = node.type === 'directory' || (!node.component && !!children)

    if (isDirectory) {
      if (!children || children.length === 0) continue
      out.push({
        value: `dir:${node.id}`,
        label: node.meta?.title || node.name,
        icon: node.meta?.icon,
        children,
      })
      continue
    }

    const path = fullPathOf(node)
    out.push({
      value: path,
      label: node.meta?.title || node.name,
      icon: node.meta?.icon,
      path,
    })
  }
  return out
}

/** 找到第一个可访问的菜单页面路径（用于根路径重定向） */
export function findFirstMenuPath(nodes: RouteNode[]): string | null {
  for (const node of nodes) {
    if (node.meta?.hidden) continue
    if (node.type === 'menu' && node.component) return fullPathOf(node)
    if (node.children?.length) {
      const found = findFirstMenuPath(node.children)
      if (found) return found
    }
  }
  return null
}
