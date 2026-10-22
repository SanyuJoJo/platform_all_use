<template>
  <div id="app" class="app-root">
    <n-layout class="layout-root" :native-scrollbar="false">
      <n-layout-header class="main-header" bordered>
        <div class="header-inner">
          <span class="app-title">多功能管理平台</span>
          <div class="header-right">
            <span class="nickname">
              {{ userStore.user?.nickname || '管理员' }}
            </span>
            <n-button size="small" @click="handleLogout">退出</n-button>
          </div>
        </div>
      </n-layout-header>

      <n-layout has-sider class="layout-body" :native-scrollbar="false">
        <n-layout-sider
          class="main-sider"
          bordered
          :width="240"
          :native-scrollbar="false"
        >
          <div class="sider-scroll">
            <n-menu
              :key="menuTreeKey"
              :options="menuStore.menuTree"
              :value="activeMenuKey"
              :indent="18"
              @update:value="handleMenuSelect"
              default-expand-all
            />
          </div>
        </n-layout-sider>

        <n-layout-content class="main-content" :native-scrollbar="false">
          <div v-if="!isSubAppRoute" class="content-scroll">
            <router-view />
          </div>
        </n-layout-content>
      </n-layout>
    </n-layout>

    <!-- 子应用容器：position: fixed，不受父级滚动影响 -->
    <div
      id="subapp-container"
      v-show="isSubAppRoute"
      class="subapp-container"
    ></div>
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  ref,
  watch,
  onMounted,
  onBeforeUnmount,
  nextTick,
} from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
  NLayout,
  NLayoutHeader,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NButton,
} from 'naive-ui';
import type { MenuOption } from 'naive-ui';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
import { usePermissionStore } from '@/store/permission';
import { message } from '@/utils/naive';
import { registerModules } from '@/micro-frontend/registry';
import { resetRouter } from '@/router';

interface MenuLookupNode {
  key: string | number;
  path?: string;
  children?: MenuLookupNode[];
}

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const menuStore = useMenuStore();
const moduleStore = useModuleStore();
const permissionStore = usePermissionStore();

const menuTreeKey = ref(0);

const isSubAppRoute = computed(() =>
  moduleStore.modules.some((m) => route.path.startsWith(`/${m.id}`))
);

const activeMenuKey = computed<string>(() => {
  const currentPath = route.path;
  const tree = menuStore.menuTree as unknown as MenuLookupNode[];

  const find = (nodes: MenuLookupNode[]): string | null => {
    for (const n of nodes) {
      if (n.path === currentPath) return String(n.key);
      if (n.children && n.children.length > 0) {
        const r = find(n.children);
        if (r) return r;
      }
    }
    return null;
  };

  return find(tree) ?? '';
});

watch(
  () => menuStore.menuTree,
  (v) => {
    console.log('[Layout] menuTree 变化, length =', v?.length ?? 0);
    menuTreeKey.value += 1;
  },
  { immediate: true, deep: false }
);

function fixMainAppStyles() {
  const header = document.querySelector('.main-header') as HTMLElement | null;
  const sider = document.querySelector('.main-sider') as HTMLElement | null;
  if (header) {
    header.style.setProperty('background', '#fff', 'important');
    header.style.setProperty('color', '#000', 'important');
  }
  if (sider) {
    sider.style.setProperty('background', '#fff', 'important');
    sider.style.setProperty('color', '#000', 'important');
  }
  document.querySelectorAll('.n-menu-item-content').forEach((el) => {
    (el as HTMLElement).style.setProperty('color', '#000', 'important');
  });
  document.querySelectorAll('.main-header .n-button').forEach((el) => {
    (el as HTMLElement).style.setProperty('color', '#000', 'important');
  });
}

let observer: MutationObserver | null = null;
function startStyleWatcher() {
  if (observer) observer.disconnect();
  observer = new MutationObserver(() => {
    const container = document.getElementById('subapp-container');
    if (container && container.style.display !== 'none') {
      fixMainAppStyles();
    }
  });
  observer.observe(document.head, { childList: true, subtree: true });
  observer.observe(document.body, {
    attributes: true,
    attributeFilter: ['style', 'class'],
  });
}

onMounted(async () => {
  fixMainAppStyles();
  startStyleWatcher();

  if (!moduleStore.loaded) {
    try {
      await moduleStore.fetchModules();
    } catch (e) {
      console.warn('[Layout] 模块加载失败', e);
    }
  }
  if (moduleStore.loaded && moduleStore.modules.length > 0) {
    registerModules(moduleStore.modules, router);
  }

  if (
    (!menuStore.menuLoaded || menuStore.menuTree.length === 0) &&
    permissionStore.routeTree.length > 0
  ) {
    console.warn('[Layout] 菜单未构建，兜底 buildMenus');
    menuStore.buildMenus(permissionStore.routeTree);
  }
});

onBeforeUnmount(() => {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
});

watch(
  () => isSubAppRoute.value,
  (newVal) => {
    if (newVal) {
      nextTick(() => fixMainAppStyles());
      setTimeout(fixMainAppStyles, 500);
    }
  },
  { immediate: true }
);

function handleMenuSelect(_key: string | number, item: MenuOption): void {
  const path = (item as unknown as { path?: string }).path;
  if (path) router.push(path);
}

async function handleLogout() {
  userStore.logout();
  resetRouter();
  menuStore.reset();
  message.success('已退出');
  
  // ★ 核心修改：使用 window.location.href 强制刷新页面
  // 这样不仅清空了状态，也让 qiankun 重新初始化，避免切换角色后的 404
  setTimeout(() => {
    window.location.href = '/login';
  }, 300);
}
</script>

<style>
html,
body {
  margin: 0;
  padding: 0;
  height: 100%;
  overflow: hidden;
}

#app.app-root {
  height: 100vh;
  width: 100%;
  position: relative;
  overflow: hidden;
  background: #fff;
  color: #000;
}

.layout-root {
  height: 100vh !important;
}

.main-header {
  height: 64px !important;
  background: #fff !important;
  color: #000 !important;
  flex-shrink: 0;
}

.header-inner {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-sizing: border-box;
}

.app-title {
  font-weight: bold;
  font-size: 18px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.nickname {
  font-size: 14px;
}

.layout-body {
  height: calc(100vh - 64px) !important;
  overflow: hidden;
}

.main-sider {
  height: 100% !important;
  background: #fff !important;
  color: #000 !important;
  overflow: hidden;
}

.sider-scroll {
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0;
  box-sizing: border-box;
}

.main-content {
  height: 100% !important;
  overflow: hidden;
  background: #fff;
}

.content-scroll {
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 20px;
  box-sizing: border-box;
}

.subapp-container {
  position: fixed;
  top: 64px;
  left: 240px;
  right: 0;
  bottom: 0;
  overflow: auto;
  padding: 20px;
  background: #fff;
  z-index: 10;
  box-sizing: border-box;
}

#app .n-menu-item-content {
  color: #000 !important;
}
#app .n-layout-header .n-button {
  color: #000 !important;
}

.sider-scroll::-webkit-scrollbar,
.content-scroll::-webkit-scrollbar,
.subapp-container::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.sider-scroll::-webkit-scrollbar-thumb,
.content-scroll::-webkit-scrollbar-thumb,
.subapp-container::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}
.sider-scroll::-webkit-scrollbar-thumb:hover,
.content-scroll::-webkit-scrollbar-thumb:hover,
.subapp-container::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.35);
}
</style>
