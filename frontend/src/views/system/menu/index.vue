<template>
  <t-card :title="'菜单管理'" :bordered="false">
    <div class="toolbar">
      <t-button theme="primary" @click="handleAddRoot()">新增菜单</t-button>
      <t-button @click="loadTree" :loading="loading">刷新</t-button>
    </div>

    <t-table
      :columns="columns"
      :data="menuTree"
      :loading="loading"
      row-key="id"
      :pagination="false"
      children-key="children"
      default-expand-all
    />

    <MenuFormModal
      v-model:show="showForm"
      :editing="editingMenu"
      :default-menu-type="defaultMenuType"
      :lock-menu-type="lockMenuType"
      :all-menus="menuTree"
      @success="handleFormSuccess"
    />
  </t-card>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue';
import { Button, Tag, Popconfirm, Space, MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { getMenuTree, deleteMenu, type Menu } from '@/api/system-menu';
import MenuFormModal from './components/MenuFormModal.vue';

const loading = ref(false);
const menuTree = ref<Menu[]>([]);
const showForm = ref(false);
const editingMenu = ref<Menu | null>(null);
const defaultMenuType = ref<'M' | 'C' | 'F' | undefined>(undefined);
const lockMenuType = ref(false);

async function loadTree() {
  loading.value = true;
  try {
    const res: any = await getMenuTree();
    menuTree.value = (res?.data ?? []) as Menu[];
  } catch (e: any) {
    MessagePlugin.error(e?.message || '加载菜单失败');
  } finally {
    loading.value = false;
  }
}

function handleAddRoot() {
  editingMenu.value = null;
  defaultMenuType.value = 'C';
  lockMenuType.value = false;
  showForm.value = true;
}

function handleAddChildOfDirectory(row: Menu) {
  editingMenu.value = { ...emptyMenu(), parentId: row.id } as Menu;
  defaultMenuType.value = 'C';
  lockMenuType.value = false;
  showForm.value = true;
}

function handleAddButtonOfMenu(row: Menu) {
  editingMenu.value = { ...emptyMenu(), parentId: row.id } as Menu;
  defaultMenuType.value = 'F';
  lockMenuType.value = true;
  showForm.value = true;
}

function handleEdit(row: Menu) {
  editingMenu.value = { ...row };
  defaultMenuType.value = row.menuType;
  lockMenuType.value = false;
  showForm.value = true;
}

function handleFormSuccess() {
  showForm.value = false;
  loadTree();
}

async function handleDelete(row: Menu) {
  try {
    await deleteMenu(row.id);
    MessagePlugin.success('删除成功');
    loadTree();
  } catch (e: any) {
    MessagePlugin.error(e?.message || '删除失败');
  }
}

function emptyMenu(): Menu {
  return {
    id: 0,
    parentId: 0,
    menuName: '',
    name: '',
    microApp: 'main',
    orderNum: 0,
    path: '',
    component: '',
    query: '',
    isFrame: 1,
    isCache: 0,
    menuType: 'C',
    visible: 0,
    status: 0,
    perms: '',
    icon: '#',
    moduleId: '',
    remark: '',
  };
}

const menuTypeLabel: Record<string, string> = { M: '目录', C: '菜单', F: '按钮' };
const menuTypeTag: Record<string, 'primary' | 'success' | 'warning' | 'default'> = {
  M: 'primary',
  C: 'success',
  F: 'warning',
};

const columns: PrimaryTableCol<Menu>[] = [
  { colKey: 'menuName', title: '菜单名称', width: 200 },
  {
    colKey: 'menuType',
    title: '类型',
    width: 80,
    cell: (_h, { row }) =>
      h(
        Tag,
        { theme: menuTypeTag[row.menuType] || 'default', variant: 'light' },
        () => menuTypeLabel[row.menuType] || row.menuType,
      ),
  },
  { colKey: 'icon', title: '图标', width: 90 },
  { colKey: 'orderNum', title: '排序', width: 60 },
  {
    colKey: 'path',
    title: '路由地址',
    width: 180,
    cell: (_h, { row }) => (row.menuType === 'F' ? '-' : row.path || '-'),
  },
  {
    colKey: 'perms',
    title: '权限标识',
    width: 180,
    cell: (_h, { row }) => row.perms || '-',
  },
  {
    colKey: 'component',
    title: '组件路径',
    width: 180,
    cell: (_h, { row }) => (row.menuType === 'F' ? '-' : row.component || '-'),
  },
  { colKey: 'microApp', title: '微应用', width: 130 },
  {
    colKey: 'status',
    title: '状态',
    width: 80,
    cell: (_h, { row }) =>
      h(
        Tag,
        { theme: row.status === 0 ? 'success' : 'danger', variant: 'light' },
        () => (row.status === 0 ? '正常' : '停用'),
      ),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 280,
    fixed: 'right',
    cell: (_h, { row }) => {
      const actions: any[] = [];

      actions.push(
        h(
          Button,
          { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleEdit(row) },
          () => '编辑',
        ),
      );

      if (row.menuType === 'M') {
        actions.push(
          h(
            Button,
            { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleAddChildOfDirectory(row) },
            () => '新增子项',
          ),
        );
      }

      if (row.menuType === 'C') {
        actions.push(
          h(
            Button,
            { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleAddButtonOfMenu(row) },
            () => '新增按钮',
          ),
        );
      }

      actions.push(
        h(
          Popconfirm,
          { content: `确定删除「${row.menuName}」？`, onConfirm: () => handleDelete(row) },
          () => h(Button, { size: 'small', variant: 'text', theme: 'danger' }, () => '删除'),
        ),
      );

      return h(Space, { size: 4 }, () => actions);
    },
  },
];

onMounted(loadTree);
</script>

<style scoped>
.menu-manager {
  padding: 4px;
}
.toolbar {
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
}
</style>