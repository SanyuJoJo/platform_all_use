<template>
  <t-layout class="app-layout">
    <t-header class="app-header">
      <div class="header-left">
        <t-button shape="square" variant="text" @click="collapsed = !collapsed">
          <t-icon :name="collapsed ? 'chevron-right' : 'chevron-left'" size="20px" />
        </t-button>
        <span class="logo">多功能管理平台</span>
        <t-breadcrumb class="breadcrumb">
          <t-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="index">
            {{ item }}
          </t-breadcrumb-item>
        </t-breadcrumb>
      </div>
      <div class="header-right">
        <t-dropdown :options="userOptions" @click="handleUserAction">
          <t-button variant="text">
            <template #icon><t-icon name="user-circle" /></template>
            {{ userStore.username || 'admin' }}
          </t-button>
        </t-dropdown>
      </div>
    </t-header>

    <t-layout class="app-body">
      <t-aside :width="collapsed ? '64px' : '232px'" class="app-aside">
        <t-menu
          :value="activeKey"
          :collapsed="collapsed"
          theme="light"
          class="app-menu"
          @change="handleMenuChange"
        >
          <AppMenu :items="menuStore.menuTree" />
        </t-menu>
      </t-aside>

      <t-content class="app-content">
        <div class="content-inner">
          <router-view />
        </div>
      </t-content>
    </t-layout>
  </t-layout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/store/user'
import { useMenuStore } from '@/store/menu'
import { usePermissionStore } from '@/store/permission'
import AppMenu from '@/components/AppMenu.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const menuStore = useMenuStore()
const permissionStore = usePermissionStore()

const collapsed = ref(false)
const activeKey = computed(() => route.path)

const breadcrumbs = computed(() =>
  route.matched
    .map((r) => r.meta?.title as string | undefined)
    .filter((title): title is string => !!title),
)

const userOptions = [{ content: '退出登录', value: 'logout' }]

function handleMenuChange(value: unknown) {
  if (typeof value === 'string' && value.startsWith('/')) {
    router.push(value)
  }
}

function handleUserAction(option: { value?: string }) {
  if (option?.value === 'logout') {
    userStore.logout()
    permissionStore.reset()
    menuStore.reset()
    MessagePlugin.success('已退出登录')
    router.push('/login')
  }
}
</script>

<style scoped>
.app-layout {
  height: 100vh;
  background: #f5f7fa;
}
.app-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 10;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}
.logo {
  font-size: 18px;
  font-weight: 700;
  color: #0052d9;
  white-space: nowrap;
}
.breadcrumb {
  margin-left: 8px;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.app-body {
  height: calc(100vh - 56px);
}
.app-aside {
  background: #fff;
  border-right: 1px solid #e7e7e7;
  transition: width 0.2s;
  overflow: hidden;
  flex-shrink: 0;
}
.app-menu {
  height: 100%;
  overflow-y: auto;
  border-right: none;
}
.app-content {
  overflow: hidden;
  background: #f5f7fa;
  min-width: 0;
}
.content-inner {
  height: 100%;
  overflow: auto;
  padding: 16px;
  box-sizing: border-box;
}
</style>
