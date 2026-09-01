<template>
  <n-layout style="height: 100vh;">
    <n-layout-header bordered>
      <div style="display:flex; align-items:center; justify-content:space-between; padding:0 20px; height:64px;">
        <span style="font-weight:bold;font-size:18px;">多功能管理平台</span>
        <div>
          <span style="margin-right:20px;">{{ userStore.user?.nickname || '管理员' }}</span>
          <n-button @click="handleLogout">退出</n-button>
        </div>
      </div>
    </n-layout-header>
    <n-layout has-sider>
      <n-layout-sider bordered style="width:240px;">
        <!-- ✅ 删除以下调试标签 -->
        <!-- <pre style="..."> {{ menuStore.menuTree }} </pre> -->
        
        <n-menu
          :key="menuTreeKey"
          :options="menuStore.menuTree"
          @update:value="handleMenuSelect"
          :value="activeMenuKey"
          default-expand-all
        />
      </n-layout-sider>
      <n-layout-content style="padding:20px;">
        <div id="subapp-container" style="height:100%;" />
        <router-view v-if="!$route.meta.isMicroApp" />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import type { MenuOption } from 'naive-ui';
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
import { unloadCurrentApp } from '@/micro-frontend/registry';

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const menuStore = useMenuStore();

const activeMenuKey = computed(() => route.name || '');
const menuTreeKey = ref(0);

watch(
  () => menuStore.menuTree.length,
  () => {
    menuTreeKey.value += 1;
  },
  { immediate: true }
);

function handleMenuSelect(key: string, item: MenuOption) {
  const path = (item as any).path;
  if (path) {
    router.push(path);
  }
}

async function handleLogout() {
  await unloadCurrentApp();
  userStore.logout();
  router.push('/login');
}
</script>

<style scoped>
#subapp-container {
  min-height: 300px;
}
</style>
