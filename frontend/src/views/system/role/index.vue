<template>
  <t-card :title="'角色管理'" :bordered="false">
    <t-space direction="vertical" size="large" style="width: 100%">
      <t-space>
        <t-input
          v-model="searchKeyword"
          placeholder="搜索角色名/编码"
          style="width: 300px"
          clearable
          @input="handleSearch"
        />
        <t-button theme="primary" v-permission="'auth:role:create'" @click="handleCreate">新建角色</t-button>
        <t-button @click="resetSearch">重置</t-button>
      </t-space>
      <t-table
        row-key="id"
        :columns="columns"
        :data="roleList"
        :loading="loading"
        :pagination="pagination"
        size="small"
        @page-change="onPageChange"
      />
    </t-space>
    <RoleFormModal
      v-model:visible="formModalVisible"
      :mode="formMode"
      :initial-data="editData"
      @success="fetchRoles"
    />
  </t-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import type { VNode } from 'vue';
import { Tag, Space, Button, Popconfirm, MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { roleApi } from '@/api/role';
import RoleFormModal from './components/RoleFormModal.vue';
import { useUserStore } from '@/store/user';
import type { Role } from '@/types';

const userStore = useUserStore();

const roleList = ref<Role[]>([]);
const loading = ref(false);
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});
const searchKeyword = ref('');

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'id', title: 'ID' },
  { colKey: 'name', title: '角色名' },
  { colKey: 'code', title: '编码' },
  { colKey: 'description', title: '描述' },
  {
    colKey: 'is_system',
    title: '系统内置',
    cell: (_h, { row }) =>
      h(Tag, { theme: row.is_system ? 'primary' : 'default', variant: 'light' }, () =>
        row.is_system ? '是' : '否',
      ),
  },
  {
    colKey: 'menuCount',
    title: '菜单数',
    cell: (_h, { row }) => row.menuCount ?? row.menuIds?.length ?? 0,
  },
  {
    colKey: 'created_at',
    title: '创建时间',
    cell: (_h, { row }) => new Date(row.created_at).toLocaleString(),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 160,
    cell: (_h, { row }) => {
      const canEdit = userStore.hasPermission('auth:role:edit');
      const canDelete = userStore.hasPermission('auth:role:delete');
      const buttons: VNode[] = [];

      if (canEdit) {
        buttons.push(
          h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleEdit(row) }, () => '编辑'),
        );
      }
      if (canDelete) {
        buttons.push(
          h(
            Popconfirm,
            { content: '确认删除该角色？', onConfirm: () => handleDelete(row) },
            () =>
              h(
                Button,
                { size: 'small', variant: 'text', theme: 'danger', disabled: Boolean(row.is_system) },
                () => '删除',
              ),
          ),
        );
      }
      return h(Space, { size: 4 }, () => buttons);
    },
  },
];

const formModalVisible = ref(false);
const formMode = ref<'create' | 'edit'>('create');
const editData = ref<Role | null>(null);

async function fetchRoles() {
  loading.value = true;
  try {
    const res = await roleApi.getList({
      page: pagination.current,
      page_size: pagination.pageSize,
      keyword: searchKeyword.value || undefined,
    });
    roleList.value = res.data.items;
    pagination.total = res.data.total;
  } catch (error: any) {
    // no-op
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchRoles();
}

function resetSearch() {
  searchKeyword.value = '';
  pagination.current = 1;
  fetchRoles();
}

function onPageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchRoles();
}

function handleCreate() {
  formMode.value = 'create';
  editData.value = null;
  formModalVisible.value = true;
}

function handleEdit(row: Role) {
  formMode.value = 'edit';
  editData.value = row;
  formModalVisible.value = true;
}

async function handleDelete(row: Role) {
  try {
    await roleApi.delete(row.id);
    MessagePlugin.success('删除成功');
    fetchRoles();
  } catch (error: any) {
    if (error.code === 20003) {
      MessagePlugin.error('系统内置角色不可删除');
    } else if (error.code === 20005) {
      MessagePlugin.error('角色已被用户使用，请先解除关联');
    } else {
      MessagePlugin.error(error.message || '删除失败');
    }
  }
}

onMounted(fetchRoles);
</script>