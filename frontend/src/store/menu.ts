import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RouteNode } from '@/types'
import { buildMenuTree, type MenuNode } from '@/utils/menu'

export const useMenuStore = defineStore('menu', () => {
  const menuTree = ref<MenuNode[]>([])
  const menuLoaded = ref(false)

  function buildMenus(tree: RouteNode[]) {
    menuTree.value = buildMenuTree(tree)
    menuLoaded.value = true
  }

  function reset() {
    menuTree.value = []
    menuLoaded.value = false
  }

  return { menuTree, menuLoaded, buildMenus, reset }
})
