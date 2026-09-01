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
        <n-menu
          :options="menuStore.menuTree"
	  @update:value="handleMenuSelect"
          :value="activeMenuKey"
        />
      </n-layout-sider>
      <n-layout-content style="padding:20px;">
        <!-- qiankun 子应用挂载点 -->
        <div id="subapp-container" style="height:100%;" />
        <!-- 主应用路由视图（非子应用路由） -->
        <router-view v-if="!$route.meta.isMicroApp" />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { computed } from 'vue';
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

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const menuStore = useMenuStore();

const activeMenuKey = computed(() => route.name || '');

function handleMenuSelect(key: string, item: any) {
  if (item.path) {
    router.push(item.path);
  }
}

function handleLogout() {
  userStore.logout();
  router.push('/login');
}
</script>

<style scoped>
#subapp-container {
  min-height: 300px;
}
</style>
