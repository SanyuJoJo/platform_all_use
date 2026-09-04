<template>
  <div id="app">
    <!-- 子应用容器 -->
    <div
      id="subapp-container"
      v-show="showSubAppContainer"
      style="position: absolute; top: 64px; left: 240px; right: 0; bottom: 0; padding: 20px; overflow: auto; background: #fff; z-index: 10;"
    ></div>

    <!-- 主布局 -->
    <n-layout style="height: 100vh;">
      <!-- 顶部栏 -->
      <n-layout-header
        class="main-header"
        bordered
        style="position: relative; z-index: 11; height:64px; background: #fff !important; color: #000 !important;"
      >
        <div style="display:flex; align-items:center; justify-content:space-between; padding:0 20px; height:64px;">
          <span style="font-weight:bold;font-size:18px;">多功能管理平台</span>
          <div>
            <span style="margin-right:20px;">{{ userStore.user?.nickname || '管理员' }}</span>
            <n-button @click="handleLogout">退出</n-button>
          </div>
        </div>
      </n-layout-header>

      <n-layout has-sider>
        <!-- 左侧菜单 -->
        <n-layout-sider
          class="main-sider"
          bordered
          style="width:240px; position: relative; z-index: 11; background: #fff !important; color: #000 !important;"
        >
          <n-menu
            :key="menuTreeKey"
            :options="menuStore.menuTree"
            @update:value="handleMenuSelect"
            :value="activeMenuKey"
            default-expand-all
          />
        </n-layout-sider>

        <!-- 内容区域 -->
        <n-layout-content style="padding:20px; position: relative;">
          <router-view v-if="!isSubAppRoute" />
        </n-layout-content>
      </n-layout>
    </n-layout>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, nextTick, onBeforeUnmount } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
  NLayout,
  NLayoutHeader,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  NButton,
} from 'naive-ui';
import { useUserStore } from '@/store/user';
import { useMenuStore } from '@/store/menu';
import { useModuleStore } from '@/store/module';
import { message } from '@/utils/naive';

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const menuStore = useMenuStore();
const moduleStore = useModuleStore();

const activeMenuKey = computed(() => route.name || '');
const menuTreeKey = ref(0);

const isSubAppRoute = computed(() => {
  const modules = moduleStore.modules;
  return modules.some(m => route.path.startsWith(`/${m.id}`));
});

const showSubAppContainer = computed(() => isSubAppRoute.value);

watch(
  () => menuStore.menuTree.length,
  () => {
    menuTreeKey.value += 1;
    console.log('[Layout] 菜单树更新，新 key:', menuTreeKey.value);
  },
  { immediate: true }
);

function fixMainAppStyles() {
  const header = document.querySelector('.main-header') as HTMLElement;
  const sider = document.querySelector('.main-sider') as HTMLElement;
  if (header) {
    header.style.setProperty('background', '#fff', 'important');
    header.style.setProperty('color', '#000', 'important');
  }
  if (sider) {
    sider.style.setProperty('background', '#fff', 'important');
    sider.style.setProperty('color', '#000', 'important');
  }
  document.querySelectorAll('.n-menu-item-content').forEach(el => {
    (el as HTMLElement).style.setProperty('color', '#000', 'important');
  });
  document.querySelectorAll('.main-header .n-button').forEach(el => {
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
  observer.observe(document.body, { attributes: true, attributeFilter: ['style', 'class'] });
}

onMounted(() => {
  console.log('[Layout] 布局组件挂载，当前路由:', route.path);
  const container = document.getElementById('subapp-container');
  console.log('[Layout] subapp-container 是否存在:', !!container);
  fixMainAppStyles();
  startStyleWatcher();
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

function handleMenuSelect(key: string, item: any) {
  if (item.path) {
    console.log(`[Layout] 点击菜单 ${key}，跳转到 ${item.path}`);
    router.push(item.path);
  }
}

async function handleLogout() {
  console.log('[Layout] 用户登出');
  userStore.logout();
  message.success('已退出');
  router.push('/login');
}
</script>

<style>
#app {
  height: 100vh;
  position: relative;
  background: #fff !important;
  color: #000 !important;
}
#subapp-container {
  background: #fff;
  z-index: 10;
}
#app .main-header,
#app .main-sider {
  background: #fff !important;
  color: #000 !important;
}
#app .n-menu-item-content {
  color: #000 !important;
}
#app .n-layout-header .n-button {
  color: #000 !important;
}
</style>