<template>
  <div class="menu-manager">
    <div class="toolbar">
      <n-button type="primary" @click="handleAddRoot()">新增菜单</n-button>
      <n-button @click="loadTree" :loading="loading">刷新</n-button>
    </div>

    <n-data-table
      :columns="columns"
      :data="menuTree"
      :loading="loading"
      :row-key="(row: Menu) => row.id"
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
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted } from 'vue';
import {
  NButton,
  NDataTable,
  NTag,
  NPopconfirm,
  NSpace,
  useMessage,
  type DataTableColumns,
} from 'naive-ui';
import { getMenuTree, deleteMenu, type Menu } from '../../../api/menu';
import MenuFormModal from './components/MenuFormModal.vue';

const message = useMessage();
const loading = ref(false);
const menuTree = ref<Menu[]>([]);
const showForm = ref(false);
const editingMenu = ref<Menu | null>(null);
/** 新建时的默认 menuType（M 下新增子项默认 C，C 下新增按钮默认 F） */
const defaultMenuType = ref<'M' | 'C' | 'F' | undefined>(undefined);
/** 是否锁定 menuType 不可改（C 下新增按钮锁定为 F） */
const lockMenuType = ref(false);

async function loadTree() {
  loading.value = true;
  try {
    const res: any = await getMenuTree();
    menuTree.value = (res?.data ?? []) as Menu[];
  } catch (e: any) {
    message.error(e?.message || '加载菜单失败');
  } finally {
    loading.value = false;
  }
}

/** 顶层新增：不锁定类型，默认 C */
function handleAddRoot() {
  editingMenu.value = null;
  defaultMenuType.value = 'C';
  lockMenuType.value = false;
  showForm.value = true;
}

/**
 * 目录（M）下新增子项：可选 M/C/F，默认 C
 */
function handleAddChildOfDirectory(row: Menu) {
  editingMenu.value = { ...emptyMenu(), parentId: row.id } as Menu;
  defaultMenuType.value = 'C';
  lockMenuType.value = false;
  showForm.value = true;
}

/**
 * 菜单（C）下新增按钮：强制 F，锁定不可改
 */
function handleAddButtonOfMenu(row: Menu) {
  editingMenu.value = { ...emptyMenu(), parentId: row.id } as Menu;
  defaultMenuType.value = 'F';
  lockMenuType.value = true;
  showForm.value = true;
}

function handleEdit(row: Menu) {
  editingMenu.value = { ...row };
  // 编辑时保留原始类型；不锁定，但要受父节点约束校验
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
    message.success('删除成功');
    loadTree();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
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
const menuTypeTag: Record<string, 'info' | 'success' | 'warning'> = {
  M: 'info',
  C: 'success',
  F: 'warning',
};

const columns: DataTableColumns<Menu> = [
  {
    title: '菜单名称',
    key: 'menuName',
    width: 200,
    ellipsis: { tooltip: true },
  },
  {
    title: '类型',
    key: 'menuType',
    width: 80,
    render(row) {
      return h(
        NTag,
        { type: menuTypeTag[row.menuType] || 'default', size: 'small' },
        { default: () => menuTypeLabel[row.menuType] || row.menuType }
      );
    },
  },
  { title: '图标', key: 'icon', width: 90, ellipsis: { tooltip: true } },
  { title: '排序', key: 'orderNum', width: 60 },
  {
    title: '路由地址',
    key: 'path',
    width: 180,
    ellipsis: { tooltip: true },
    render: (row) => (row.menuType === 'F' ? '-' : row.path || '-'),
  },
  {
    title: '权限标识',
    key: 'perms',
    width: 180,
    ellipsis: { tooltip: true },
    render: (row) => row.perms || '-',
  },
  {
    title: '组件路径',
    key: 'component',
    width: 180,
    ellipsis: { tooltip: true },
    render: (row) => (row.menuType === 'F' ? '-' : row.component || '-'),
  },
  { title: '微应用', key: 'microApp', width: 130, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row) {
      return h(
        NTag,
        { type: row.status === 0 ? 'success' : 'error', size: 'small' },
        { default: () => (row.status === 0 ? '正常' : '停用') }
      );
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right',
    render(row) {
      const actions: any[] = [];

      // 编辑
      actions.push(
        h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleEdit(row),
          },
          { default: () => '编辑' }
        )
      );

      // 目录 → 新增子项（M/C/F 可选）
      if (row.menuType === 'M') {
        actions.push(
          h(
            NButton,
            {
              size: 'small',
              type: 'info',
              text: true,
              onClick: () => handleAddChildOfDirectory(row),
            },
            { default: () => '新增子项' }
          )
        );
      }

      // 菜单 → 新增按钮（强制 F）
      if (row.menuType === 'C') {
        actions.push(
          h(
            NButton,
            {
              size: 'small',
              type: 'info',
              text: true,
              onClick: () => handleAddButtonOfMenu(row),
            },
            { default: () => '新增按钮' }
          )
        );
      }

      // 按钮（F）不显示新增，只能编辑/删除

      // 删除
      actions.push(
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', type: 'error', text: true },
                { default: () => '删除' }
              ),
            default: () => `确定删除「${row.menuName}」？`,
          }
        )
      );

      return h(NSpace, { size: 4 }, { default: () => actions });
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
