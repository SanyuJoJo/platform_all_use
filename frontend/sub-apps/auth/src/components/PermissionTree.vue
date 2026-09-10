<template>
  <n-tree
    :data="treeData"
    :checked-keys="checkedKeys"
    :default-expanded-keys="expandedKeys"
    checkable
    check-strategy="child"
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

const emit = defineEmits<{
  (e: 'update:checked', value: string[]): void;
}>();

const allPermissions = ref<Permission[]>([]);
const checkedKeys = ref<string[]>(props.checked || []);
const treeData = ref<any[]>([]);
const expandedKeys = ref<string[]>([]);

/**
 * 加载权限列表并构建权限树。
 */
async function loadPermissions() {
  try {
    const params = props.moduleFilter ? { module_id: props.moduleFilter } : {};
    const res = await permissionApi.getList(params);
    allPermissions.value = res.data;
    buildTree();
  } catch (error) {
    console.error('[PermissionTree] 加载权限失败', error);
  }
}

/**
 * 构建权限树。
 *
 * 关键说明：
 * - 模块节点 key 使用 `module-{moduleId}`，仅用于 UI 分组，不提交给后端；
 * - 叶子节点 key 使用真实权限编码 `{module}:{resource}:{action}`，可提交给后端；
 * - 结合 `check-strategy="child"`，n-tree 只会把叶子节点的 key 放入 checkedKeys。
 */
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

  // 默认展开所有模块
  expandedKeys.value = treeData.value.map(node => node.key);
}

/**
 * 过滤出合法权限编码（双保险）。
 *
 * 即使 Naive UI 在某些版本下仍返回父节点 key，也能被过滤掉：
 * - 过滤 `module-` 前缀；
 * - 过滤不含冒号的字符串（真实权限编码必须包含冒号）。
 */
function filterValidPermissionCodes(keys: string[]): string[] {
  if (!Array.isArray(keys)) return [];
  return keys.filter(
    k => typeof k === 'string' && !k.startsWith('module-') && k.includes(':')
  );
}

/**
 * n-tree 勾选事件处理。
 */
function handleCheck(keys: string[]) {
  const validCodes = filterValidPermissionCodes(keys);
  checkedKeys.value = validCodes;
  emit('update:checked', validCodes);
}

// 监听外部传入的 checked，同步到内部状态
watch(
  () => props.checked,
  (val) => {
    const valid = filterValidPermissionCodes(val || []);
    // 简单比较，避免死循环
    if (JSON.stringify(valid) !== JSON.stringify(checkedKeys.value)) {
      checkedKeys.value = valid;
    }
  },
  { deep: true }
);

onMounted(loadPermissions);
</script>
