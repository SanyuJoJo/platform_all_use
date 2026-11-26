import type { RouteRecordRaw } from 'vue-router'
import type { RouteNode } from '@/types'
import { fullPathOf } from '@/utils/menu'

/**
 * component 字符串 → views 目录下的实际文件。
 *
 * 后端约定 component 为「相对 src/views/ 的路径，不带 views/ 前缀」，
 * 例如 system/user/index、list/index、tcps/list/index。
 */
const COMPONENT_ALIAS: Record<string, string> = {
  // 历史兼容：旧 ssl-proxy 子应用使用 ssl-proxy/index
  'ssl-proxy/index': 'list/index',
}

const viewModules = import.meta.glob('../views/**/*.vue')

export function resolveView(component?: string): (() => Promise<unknown>) | null {
  if (!component) return null
  let normalized = String(component).trim()
  normalized = normalized.replace(/^\/+/, '').replace(/^views\//, '').replace(/\.vue$/, '')
  normalized = COMPONENT_ALIAS[normalized] ?? normalized

  const candidates = [`../views/${normalized}.vue`, `/src/views/${normalized}.vue`]
  for (const key of candidates) {
    if (viewModules[key]) return viewModules[key] as () => Promise<unknown>
  }
  console.warn(`[router] 未找到视图组件：${component}（尝试过 ${candidates.join(' / ')}）`)
  return null
}

/** 后端路由树 → vue-router 路由（注册到 Layout 下） */
export function buildRoutes(nodes: RouteNode[]): RouteRecordRaw[] {
  const out: RouteRecordRaw[] = []

  const walk = (list: RouteNode[]) => {
    for (const node of list) {
      if (node.meta?.hidden) continue

      if (node.type === 'directory') {
        if (node.children?.length) walk(node.children)
        continue
      }

      const view = resolveView(node.component)
      if (!view) {
        if (node.children?.length) walk(node.children)
        continue
      }

      out.push({
        path: fullPathOf(node),
        name: node.name,
        component: view,
        meta: {
          ...(node.meta || {}),
          title: node.meta?.title ?? node.name,
        },
      })
    }
  }

  walk(nodes)
  return out
}
