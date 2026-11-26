import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Router } from 'vue-router'
import type { RouteNode } from '@/types'
import { getRouters } from '@/api/menu'
import { buildRoutes } from '@/router/dynamic'

export const usePermissionStore = defineStore('permission', () => {
  const routeTree = ref<RouteNode[]>([])
  const loaded = ref(false)
  const loading = ref(false)
  let removeRouteFns: Array<() => void> = []

  async function fetchRouters(force = false): Promise<RouteNode[]> {
    if (loaded.value && !force && routeTree.value.length > 0) return routeTree.value
    if (loading.value) {
      while (loading.value) await new Promise((r) => setTimeout(r, 50))
      return routeTree.value
    }
    loading.value = true
    try {
      routeTree.value = await getRouters()
      loaded.value = true
      return routeTree.value
    } finally {
      loading.value = false
    }
  }

  function registerRoutes(router: Router) {
    removeRouteFns.forEach((fn) => fn())
    removeRouteFns = []

    const routes = buildRoutes(routeTree.value)
    routes.forEach((route) => {
      if (route.name && router.hasRoute(route.name)) return
      removeRouteFns.push(router.addRoute('Layout', route))
    })
    return routes.length
  }

  function reset() {
    removeRouteFns.forEach((fn) => fn())
    removeRouteFns = []
    routeTree.value = []
    loaded.value = false
  }

  return { routeTree, loaded, loading, fetchRouters, registerRoutes, reset }
})
