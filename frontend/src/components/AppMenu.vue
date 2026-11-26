<template>
  <template v-for="item in items" :key="item.value">
    <t-submenu v-if="item.children && item.children.length" :value="item.value" :title="item.label">
      <template #icon>
        <t-icon :name="tdesignIcon(item.icon)" />
      </template>
      <AppMenu :items="item.children" />
    </t-submenu>
    <t-menu-item v-else :value="item.value">
      <template #icon>
        <t-icon :name="tdesignIcon(item.icon)" />
      </template>
      {{ item.label }}
    </t-menu-item>
  </template>
</template>

<script setup lang="ts">
import { tdesignIcon } from '@/utils/icon-map'
import type { MenuNode } from '@/utils/menu'

defineProps<{ items: MenuNode[] }>()
</script>
