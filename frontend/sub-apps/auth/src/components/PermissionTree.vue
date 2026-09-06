<template>
  <n-tree
    :data="treeData"
    :checked-keys="checkedKeys"
    checkable
    cascade
    show-line
    @update:checked-keys="handleCheck"
  />
</template>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { NTree } from 'naive-ui';
import { permissionApi } from '@/api/permission';
import type { Permission } from '@/types';
const props = defineProps<{
  checked: string[];              // 当前选中的权限编码列表
  moduleFilter?: string;          // 可选，仅加载指定模块的权限
}>();
const emit = defineEmits(['update:checked']);
const allPermissions = ref<Permission[]>([]);
const checkedKeys = ref<string[]>(props.checked);
const treeData = ref<any[]>([]);
async function loadPermissions() {
  try {
    const params = props.moduleFilter ? { module_id: props.moduleFilter } : {};
    const res = await permissionApi.getList(params);
    allPermissions.value = res.data;
    buildTree();
  } catch (error) {
    // ignore
  }
}
function buildTree() {
  const groupMap: Record<string, Permission[]> = {};
  allPermissions.value.forEach(p => {
    if (!groupMap[p.module_id]) groupMap[p.module_id] = [];
    groupMap[p.module_id].push(p);
  });
  treeData.value = Object.keys(groupMap).map(moduleId => ({
    label: moduleId,
    key: `module-${moduleId}`,
    children: groupMap[moduleId].map(p => ({
      label: `${p.name} (${p.code})`,
      key: p.code,
    })),
  }));
}
function handleCheck(keys: string[]) {
  checkedKeys.value = keys;
  emit('update:checked', keys);
}
watch(
  () => props.checked,
  (val) => {
    checkedKeys.value = val;
  },
  { deep: true }
);
onMounted(loadPermissions);
</script>
