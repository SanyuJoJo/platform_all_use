<template>
  <t-layout class="app-layout">
    <div v-if="isMobile && mobileOpen" class="aside-backdrop" @click="mobileOpen = false"></div>

    <t-aside
      class="app-aside"
      :class="{ 'is-open': mobileOpen, 'is-collapsed': collapsed && !isMobile }"
      :width="asideWidth"
    >
      <div class="brand" @click="goHome">
        <BrandMark :size="36" />
        <div v-show="!collapsed || isMobile" class="brand-text">
          <span class="brand-cn">多功能管理平台</span>
          <span class="brand-en">ENTERPRISE CONSOLE</span>
        </div>
      </div>

      <div class="aside-scroll">
        <t-menu
          :value="activeKey"
          :collapsed="collapsed && !isMobile"
          :expanded="expandedKeys"
          theme="dark"
          class="app-menu"
          @change="handleMenuChange"
          @expand="handleExpand"
        >
          <AppMenu :items="menuStore.menuTree" />
        </t-menu>
      </div>

      <div class="aside-footer">
        <button type="button" class="aside-toggle" @click="toggleSidebar">
          <t-icon :name="collapsed && !isMobile ? 'chevron-right' : 'chevron-left'" />
          <span v-show="!collapsed || isMobile">收起菜单</span>
        </button>
        <div v-show="!collapsed || isMobile" class="aside-version">
          <span class="status-dot"></span>
          控制台已就绪
        </div>
      </div>
    </t-aside>

    <t-layout class="app-main">
      <t-header class="app-header">
        <div class="header-left">
          <t-button variant="text" shape="square" class="icon-btn" @click="toggleSidebar">
            <t-icon
              :name="isMobile ? 'menu' : collapsed ? 'chevron-right' : 'chevron-left'"
              size="20px"
            />
          </t-button>

          <t-breadcrumb class="breadcrumb">
            <t-breadcrumb-item>
              <span class="crumb-home" @click="goHome">
                <t-icon name="home" />
              </span>
            </t-breadcrumb-item>
            <t-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="index">
              {{ item }}
            </t-breadcrumb-item>
          </t-breadcrumb>
        </div>

        <div class="header-right">
          <div v-if="!isMobile" class="header-search">
            <t-input
              v-model="searchKeyword"
              class="search-input"
              placeholder="搜索菜单"
              clearable
              @focus="searchOpen = true"
              @blur="handleSearchBlur"
              @keydown="handleSearchKeydown"
              @keyup="handleSearchKeyup"
            >
              <template #prefix-icon>
                <t-icon name="search" />
              </template>
            </t-input>

            <transition name="fade">
              <div v-if="searchOpen && searchKeyword.trim()" class="search-panel">
                <button
                  v-for="item in searchResults"
                  :key="item.path"
                  type="button"
                  class="search-item"
                  @mousedown.prevent="goSearch(item.path)"
                >
                  <span class="search-item-icon">
                    <t-icon :name="tdesignIcon(item.icon)" />
                  </span>
                  <span class="search-item-main">
                    <strong>{{ item.label }}</strong>
                    <small v-if="item.group">{{ item.group }}</small>
                  </span>
                  <t-icon name="chevron-right" class="search-item-arrow" />
                </button>

                <div v-if="!searchResults.length" class="search-empty">
                  <t-icon name="search" />
                  <span>未找到相关菜单</span>
                </div>
              </div>
            </transition>
          </div>

          <t-tooltip content="全屏 / 退出全屏" placement="bottom">
            <t-button variant="text" shape="square" class="icon-btn" @click="toggleFullscreen">
              <t-icon :name="isFullscreen ? 'fullscreen-exit' : 'fullscreen'" size="19px" />
            </t-button>
          </t-tooltip>

          <t-popup trigger="click" placement="bottom-right" :show-arrow="false">
            <t-button variant="text" shape="square" class="icon-btn">
              <t-icon name="notification" size="19px" />
            </t-button>
            <template #content>
              <div class="notice-panel">
                <div class="notice-head">
                  <strong>通知中心</strong>
                  <span>实时消息</span>
                </div>
                <div class="notice-empty">
                  <span class="notice-empty-icon">
                    <t-icon name="notification" />
                  </span>
                  <strong>暂无新的系统通知</strong>
                  <span>有新的审计或告警消息时会在这里提醒您</span>
                </div>
              </div>
            </template>
          </t-popup>

          <span class="header-divider"></span>

          <t-dropdown trigger="click" :options="userOptions" @click="handleUserAction">
            <button type="button" class="user-trigger">
              <t-avatar
                :image="userStore.user?.avatar"
                size="34px"
                class="user-avatar"
              >
                {{ initial }}
              </t-avatar>
              <span class="user-meta">
                <strong>{{ displayName }}</strong>
                <small>{{ roleLabel }}</small>
              </span>
              <t-icon name="chevron-down" class="user-caret" />
            </button>
          </t-dropdown>
        </div>
      </t-header>

      <t-content class="app-content">
        <div class="content-inner">
          <router-view v-slot="{ Component }">
            <transition name="page" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </t-content>
    </t-layout>

    <t-dialog
      v-model:visible="profileVisible"
      header="个人信息"
      :footer="false"
      width="420px"
    >
      <div class="profile-body">
        <div class="profile-hero">
          <t-avatar :image="userStore.user?.avatar" size="56px" class="profile-avatar">
            {{ initial }}
          </t-avatar>
          <div class="profile-identity">
            <strong>{{ displayName }}</strong>
            <span>@{{ userStore.user?.username || '-' }}</span>
          </div>
        </div>

        <div class="profile-list">
          <div><span>所属角色</span><strong>{{ roleLabel }}</strong></div>
          <div><span>邮箱</span><strong>{{ userStore.user?.email || '未设置' }}</strong></div>
          <div><span>权限项</span><strong>{{ permissionCount }} 项</strong></div>
          <div><span>账号状态</span><strong class="profile-ok">正常</strong></div>
        </div>
      </div>
    </t-dialog>
  </t-layout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/store/user'
import { useMenuStore } from '@/store/menu'
import { usePermissionStore } from '@/store/permission'
import AppMenu from '@/components/AppMenu.vue'
import BrandMark from '@/components/BrandMark.vue'
import { tdesignIcon } from '@/utils/icon-map'
import { flattenMenuLeaves } from '@/utils/menu'
import type { MenuNode } from '@/utils/menu'
import { formatUserRoles } from '@/utils/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const menuStore = useMenuStore()
const permissionStore = usePermissionStore()

const collapsed = ref(false)
const mobileOpen = ref(false)
const isMobile = ref(false)
const isFullscreen = ref(false)
const profileVisible = ref(false)
const searchKeyword = ref('')
const searchOpen = ref(false)
const expandedKeys = ref<Array<string | number>>([])

const asideWidth = computed(() =>
  isMobile.value ? '248px' : collapsed.value ? '76px' : '248px',
)

const activeKey = computed(() => route.path)

const displayName = computed(
  () => userStore.user?.nickname || userStore.user?.username || '管理员',
)

const initial = computed(() =>
  (userStore.user?.nickname || userStore.user?.username || 'A').slice(0, 1).toUpperCase(),
)

const roleLabel = computed(() => formatUserRoles(userStore.user?.roles))

const permissionCount = computed(() => userStore.permissions?.length || 0)

const breadcrumbs = computed(() => {
  const titles = route.matched
    .map((record) => record.meta?.title as string | undefined)
    .filter((title): title is string => !!title)
  return titles.length ? titles : ['控制台']
})

const flatMenus = computed(() => flattenMenuLeaves(menuStore.menuTree))

const searchResults = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return []

  return flatMenus.value
    .filter(
      (item) =>
        item.label.toLowerCase().includes(keyword) ||
        item.path.toLowerCase().includes(keyword) ||
        (item.group || '').toLowerCase().includes(keyword),
    )
    .sort((a, b) => {
      const aHit = a.label.toLowerCase().includes(keyword) ? 0 : 1
      const bHit = b.label.toLowerCase().includes(keyword) ? 0 : 1
      if (aHit !== bHit) return aHit - bHit
      return a.label.localeCompare(b.label, 'zh-CN')
    })
    .slice(0, 8)
})

const userOptions = [
  { content: '个人信息', value: 'profile' },
  { content: '退出登录', value: 'logout' },
]

function handleSearchBlur() {
  window.setTimeout(() => {
    searchOpen.value = false
  }, 120)
}

function handleSearchKeydown(_value: string, context: { e: KeyboardEvent }) {
  if (context?.e?.key === 'Escape') {
    searchOpen.value = false
  }
}

function handleSearchKeyup(_value: string, context: { e: KeyboardEvent }) {
  if (context?.e?.key === 'Enter') {
    goSearch(searchResults.value[0]?.path)
  }
}

function goSearch(path?: string) {
  if (!path) return
  searchKeyword.value = ''
  searchOpen.value = false
  mobileOpen.value = false
  router.push(path)
}

function handleMenuChange(value: unknown) {
  if (typeof value === 'string' && value.startsWith('/')) {
    mobileOpen.value = false
    router.push(value)
  }
}

function findActiveAncestorKeys(
  nodes: MenuNode[],
  targetPath: string,
  trail: string[] = [],
): string[] | null {
  for (const node of nodes) {
    if (node.children?.length) {
      const found = findActiveAncestorKeys(node.children, targetPath, [...trail, node.value])
      if (found) return found
    } else if (node.path === targetPath) {
      return trail
    }
  }
  return null
}

function syncExpandedKeys() {
  const ancestors = findActiveAncestorKeys(menuStore.menuTree, route.path)
  if (!ancestors?.length) return

  const next = new Set<string | number>(expandedKeys.value)
  ancestors.forEach((key) => next.add(key))
  expandedKeys.value = [...next]
}

function handleExpand(value: Array<string | number>) {
  expandedKeys.value = value
}

function toggleSidebar() {
  if (isMobile.value) {
    mobileOpen.value = !mobileOpen.value
    return
  }
  collapsed.value = !collapsed.value
}

function goHome() {
  mobileOpen.value = false
  router.push('/')
}

async function toggleFullscreen() {
  try {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen()
    } else {
      await document.exitFullscreen()
    }
    isFullscreen.value = !!document.fullscreenElement
  } catch {
    MessagePlugin.warning('当前浏览器不支持全屏显示')
  }
}

function handleUserAction(option: { value?: string }) {
  if (option?.value === 'profile') {
    profileVisible.value = true
    return
  }

  if (option?.value === 'logout') {
    userStore.logout()
    permissionStore.reset()
    menuStore.reset()
    MessagePlugin.success('已退出登录')
    router.push('/login')
  }
}

function syncViewport() {
  const mobile = window.innerWidth <= 900
  isMobile.value = mobile
  if (!mobile) mobileOpen.value = false
}

function syncFullscreen() {
  isFullscreen.value = !!document.fullscreenElement
}

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
    searchOpen.value = false
  },
)

watch(() => route.path, syncExpandedKeys, { immediate: true })

watch(
  () => menuStore.menuTree,
  () => {
    expandedKeys.value = []
    syncExpandedKeys()
  },
)

watch(
  () => route.meta.title,
  (title) => {
    document.title = title ? `${title} - 多功能管理平台` : '多功能管理平台'
  },
  { immediate: true },
)

onMounted(() => {
  syncViewport()
  syncFullscreen()
  window.addEventListener('resize', syncViewport)
  document.addEventListener('fullscreenchange', syncFullscreen)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', syncViewport)
  document.removeEventListener('fullscreenchange', syncFullscreen)
})
</script>

<style scoped>
.app-layout {
  height: 100vh;
  overflow: hidden;
  background: #f4f6fa;
}

/* ============================ 侧边栏 ============================ */
.app-aside {
  --td-gray-color-13: #0b1220;
  position: relative;
  z-index: 30;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
  border-right: 1px solid rgba(255, 255, 255, 0.04);
  background: linear-gradient(180deg, #0b1220 0%, #0c1628 58%, #0b1220 100%);
  transition: width 0.22s cubic-bezier(0.4, 0, 0.2, 1);
}

.app-aside::before {
  content: '';
  position: absolute;
  top: -140px;
  left: -120px;
  width: 380px;
  height: 380px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(37, 99, 235, 0.24), transparent 68%);
  pointer-events: none;
}

.brand {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  height: 64px;
  padding: 0 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  cursor: pointer;
}

.brand-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  white-space: nowrap;
}

.brand-cn {
  font-size: 14.5px;
  font-weight: 600;
  letter-spacing: 0.03em;
  color: #f8fafc;
}

.brand-en {
  font-size: 8.5px;
  letter-spacing: 0.17em;
  color: rgba(148, 163, 184, 0.55);
}

.aside-scroll {
  position: relative;
  z-index: 1;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 10px 0;
}

.app-menu {
  width: 100%;
  border-right: none !important;
  background: transparent !important;
}

.app-aside :deep(.t-menu__item) {
  height: 42px;
  line-height: 42px;
  margin: 2px 10px;
  border-radius: 9px;
  color: rgba(226, 232, 240, 0.62);
  font-size: 13.5px;
  transition: background 0.16s ease, color 0.16s ease;
}

.app-aside :deep(.t-menu__item .t-icon) {
  font-size: 18px;
}

.app-aside :deep(.t-menu__item:hover:not(.t-is-active):not(.t-is-opened):not(.t-is-disabled)) {
  background: rgba(255, 255, 255, 0.06);
  color: #f8fafc;
}

.app-aside :deep(.t-menu__item.t-is-active) {
  background: linear-gradient(92deg, #2563eb 0%, #0891b2 100%);
  color: #fff;
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.32);
}

.app-aside :deep(.t-menu__item.t-is-active .t-icon) {
  color: #fff;
}

/* 子菜单展开时，父级仅做轻量高亮，避免与当前页面项形成“双选中” */
.app-aside :deep(.t-submenu > .t-menu__item.t-is-active) {
  background: rgba(255, 255, 255, 0.06);
  box-shadow: none;
}

.app-aside :deep(.t-menu__item.t-is-opened) {
  color: #f8fafc;
}

.app-aside :deep(.t-submenu.t-is-active > .t-menu__item) {
  color: #f8fafc;
}

.app-aside :deep(.t-menu__sub) {
  background: transparent !important;
}

.aside-footer {
  position: relative;
  z-index: 1;
  flex-shrink: 0;
  padding: 12px 14px 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.aside-toggle {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  height: 36px;
  padding: 0 10px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: rgba(226, 232, 240, 0.55);
  font-size: 12.5px;
  cursor: pointer;
  transition: background 0.16s ease, color 0.16s ease;
}

.aside-toggle:hover {
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
}

.app-aside.is-collapsed .aside-toggle {
  justify-content: center;
  padding: 0;
}

.aside-version {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-top: 10px;
  padding-left: 10px;
  color: rgba(148, 163, 184, 0.5);
  font-size: 11px;
  white-space: nowrap;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #34d399;
  box-shadow: 0 0 0 3px rgba(52, 211, 153, 0.16);
}

/* ============================ 顶部栏 ============================ */
.app-main {
  min-width: 0;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  position: relative;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-shrink: 0;
  height: 64px;
  padding: 0 22px;
  border-bottom: 1px solid #eaeef5;
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(10px);
}

.header-left,
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.header-left {
  flex: 1;
}

.header-right {
  flex-shrink: 0;
}

.icon-btn {
  color: #64748b;
  border-radius: 8px;
}

.icon-btn:hover {
  color: #2563eb;
  background: #eff6ff;
}

.breadcrumb {
  margin-left: 4px;
  font-size: 13px;
  white-space: nowrap;
}

.crumb-home {
  display: inline-flex;
  align-items: center;
  color: #94a3b8;
  cursor: pointer;
  transition: color 0.15s ease;
}

.crumb-home:hover {
  color: #2563eb;
}

/* 菜单搜索 */
.header-search {
  position: relative;
  width: 232px;
  margin-right: 4px;
}

.search-input :deep(.t-input) {
  height: 36px;
  border-color: transparent;
  border-radius: 8px;
  background: #f5f7fb;
  transition: background 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
}

.search-input :deep(.t-input:hover) {
  background: #fff;
  border-color: #dbe3ef;
}

.search-input :deep(.t-input.t-is-focused) {
  background: #fff;
  border-color: #93c5fd;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.08);
}

.search-panel {
  position: absolute;
  top: 44px;
  left: 0;
  right: 0;
  z-index: 60;
  max-height: 340px;
  overflow: auto;
  padding: 6px;
  border: 1px solid #e8eef7;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.14);
}

.search-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 10px;
  border: none;
  border-radius: 8px;
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition: background 0.14s ease;
}

.search-item:hover {
  background: #f1f6ff;
}

.search-item-icon {
  display: grid;
  place-items: center;
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border-radius: 7px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 15px;
}

.search-item-main {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.search-item-main strong {
  overflow: hidden;
  color: #1e293b;
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-item-main small {
  overflow: hidden;
  color: #94a3b8;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-item-arrow {
  color: #cbd5e1;
  font-size: 14px;
}

.search-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 22px 0;
  color: #94a3b8;
  font-size: 12px;
}

.search-empty .t-icon {
  font-size: 20px;
}

.header-divider {
  width: 1px;
  height: 22px;
  margin: 0 4px;
  background: #e8edf5;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 6px 4px 4px;
  border: none;
  border-radius: 10px;
  background: transparent;
  cursor: pointer;
  transition: background 0.16s ease;
}

.user-trigger:hover {
  background: #f5f7fb;
}

.user-avatar {
  background: linear-gradient(135deg, #2563eb 0%, #0891b2 100%);
  color: #fff;
  font-weight: 600;
}

.user-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.25;
}

.user-meta strong {
  color: #1e293b;
  font-size: 13px;
  font-weight: 600;
}

.user-meta small {
  color: #94a3b8;
  font-size: 11px;
}

.user-caret {
  color: #94a3b8;
  font-size: 14px;
}

/* ============================ 内容区 ============================ */
.app-content {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: #f4f6fa;
}

.content-inner {
  height: 100%;
  overflow: auto;
  padding: 22px 24px 28px;
  box-sizing: border-box;
}

.content-inner > * {
  max-width: 1680px;
  margin-right: auto;
  margin-left: auto;
}

/* 页面切换动画 */
.page-enter-active,
.page-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(6px);
}

.page-leave-to {
  opacity: 0;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ============================ 响应式 ============================ */
@media (max-width: 1100px) {
  .header-search {
    width: 180px;
  }
}

@media (max-width: 900px) {
  .aside-backdrop {
    position: fixed;
    inset: 0;
    z-index: 25;
    background: rgba(15, 23, 42, 0.45);
    backdrop-filter: blur(2px);
  }

  .app-aside {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 248px !important;
    transform: translateX(-100%);
    box-shadow: 0 0 40px rgba(2, 6, 23, 0.35);
    transition: transform 0.24s ease;
  }

  .app-aside.is-open {
    transform: translateX(0);
  }

  .app-header {
    padding: 0 12px;
  }

  .header-search,
  .user-meta,
  .user-caret,
  .header-divider,
  .breadcrumb {
    display: none;
  }

  .content-inner {
    padding: 14px 12px 20px;
  }
}
</style>

<style>
/* 弹出层挂载在 body 下，需使用全局样式 */
.t-menu__popup {
  border: 1px solid rgba(255, 255, 255, 0.08) !important;
  border-radius: 10px !important;
  background: #0b1220 !important;
  box-shadow: 0 18px 40px rgba(2, 6, 23, 0.4) !important;
}

.t-menu__popup .t-menu__item {
  margin: 2px 6px;
  border-radius: 8px;
  color: rgba(226, 232, 240, 0.7);
}

.t-menu__popup .t-menu__item.t-is-active {
  background: #2563eb;
  color: #fff;
}

.t-menu__popup .t-menu__item:hover:not(.t-is-active) {
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
}

.notice-panel {
  width: 300px;
  padding: 18px;
}

.notice-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  padding-bottom: 14px;
  border-bottom: 1px solid #eef2f7;
}

.notice-head strong {
  color: #0f172a;
  font-size: 14px;
  font-weight: 600;
}

.notice-head span {
  color: #94a3b8;
  font-size: 11px;
}

.notice-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
  padding: 30px 10px 20px;
  text-align: center;
}

.notice-empty-icon {
  display: grid;
  place-items: center;
  width: 46px;
  height: 46px;
  margin-bottom: 4px;
  border-radius: 50%;
  background: #f1f6ff;
  color: #93c5fd;
  font-size: 22px;
}

.notice-empty strong {
  color: #334155;
  font-size: 13px;
  font-weight: 600;
}

.notice-empty span {
  max-width: 220px;
  color: #94a3b8;
  font-size: 11.5px;
  line-height: 1.6;
}

.profile-body {
  padding-top: 4px;
}

.profile-hero {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px;
  border-radius: 12px;
  background: linear-gradient(135deg, #eff6ff 0%, #f0fdff 100%);
}

.profile-avatar {
  background: linear-gradient(135deg, #2563eb 0%, #0891b2 100%);
  color: #fff;
  font-weight: 600;
}

.profile-identity {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.profile-identity strong {
  color: #0f172a;
  font-size: 16px;
  font-weight: 600;
}

.profile-identity span {
  color: #64748b;
  font-size: 12px;
}

.profile-list {
  display: grid;
  gap: 2px;
  margin-top: 18px;
}

.profile-list > div {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 4px;
  border-bottom: 1px solid #f1f5f9;
}

.profile-list > div:last-child {
  border-bottom: none;
}

.profile-list span {
  color: #64748b;
  font-size: 13px;
}

.profile-list strong {
  max-width: 60%;
  overflow: hidden;
  color: #1e293b;
  font-size: 13px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-ok {
  color: #059669 !important;
}
</style>
