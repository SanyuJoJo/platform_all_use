<template>
  <t-card :title="'用户管理'" :bordered="false">
    <t-space direction="vertical" size="large" style="width: 100%">
      <t-space>
        <t-input
          v-model="searchKeyword"
          placeholder="搜索用户名/昵称/邮箱"
          style="width: 300px"
          clearable
          @input="handleSearch"
        />
        <t-select
          v-model="searchStatus"
          placeholder="状态"
          clearable
          style="width: 150px"
          :options="statusOptions"
          @change="handleSearch"
        />
        <t-select
          v-model="searchRoleId"
          placeholder="角色"
          clearable
          style="width: 150px"
          :options="roleOptions"
          @change="handleSearch"
        />
        <t-button theme="primary" v-permission="'auth:user:create'" @click="handleCreate">新建用户</t-button>
        <t-button @click="resetSearch">重置</t-button>
      </t-space>
      <t-table
        row-key="id"
        :columns="columns"
        :data="userList"
        :loading="loading"
        :pagination="pagination"
        size="small"
        @page-change="onPageChange"
      />
    </t-space>
    <UserFormModal
      v-model:visible="formModalVisible"
      :mode="formMode"
      :initial-data="editData"
      @success="fetchUsers"
    />
    <ResetPasswordModal
      v-model:visible="resetPwdVisible"
      :user-id="resetUserId"
      @success="fetchUsers"
    />
  </t-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import type { VNode } from 'vue';
import { Tag, Space, Button, Popconfirm } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { userApi } from '@/api/user';
import { roleApi } from '@/api/role';
import UserFormModal from './components/UserFormModal.vue';
import ResetPasswordModal from './components/ResetPasswordModal.vue';
import { useUserStore } from '@/store/user';
import type { User } from '@/types';

const userStore = useUserStore();

const userList = ref<User[]>([]);
const loading = ref(false);
const total = ref(0);
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const searchKeyword = ref('');
const searchStatus = ref<number | null>(null);
const searchRoleId = ref<number | null>(null);

const roleOptions = ref<{ label: string; value: number }[]>([]);

const statusOptions = [
  { label: '全部', value: -1 },
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
];

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'id', title: 'ID' },
  { colKey: 'username', title: '用户名' },
  { colKey: 'nickname', title: '昵称' },
  { colKey: 'email', title: '邮箱' },
  {
    colKey: 'status',
    title: '状态',
    cell: (_h, { row }) =>
      h(
        Tag,
        { theme: row.status === 1 ? 'success' : 'danger', variant: 'light' },
        () => (row.status === 1 ? '启用' : '禁用'),
      ),
  },
  {
    colKey: 'roles',
    title: '角色',
    cell: (_h, { row }) => row.roles.map((r: any) => r.name).join('、'),
  },
  {
    colKey: 'created_at',
    title: '创建时间',
    cell: (_h, { row }) => new Date(row.created_at).toLocaleString(),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 240,
    cell: (_h, { row }) => {
      const canEdit = userStore.hasPermission('auth:user:edit');
      const canDelete = userStore.hasPermission('auth:user:delete');
      const canReset = userStore.hasPermission('auth:user:edit');
      const canToggle = userStore.hasPermission('auth:user:edit');
      const buttons: VNode[] = [];

      if (canEdit) {
        buttons.push(
          h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleEdit(row) }, () => '编辑'),
        );
      }
      if (canToggle) {
        buttons.push(
          h(
            Popconfirm,
            { content: `确认${row.status === 1 ? '禁用' : '启用'}该用户？`, onConfirm: () => handleToggleStatus(row) },
            () =>
              h(
                Button,
                {
                  size: 'small',
                  variant: 'text',
                  theme: row.status === 1 ? 'warning' : 'success',
                },
                () => (row.status === 1 ? '禁用' : '启用'),
              ),
          ),
        );
      }
      if (canReset) {
        buttons.push(
          h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleResetPassword(row) }, () => '重置密码'),
        );
      }
      if (canDelete) {
        buttons.push(
          h(
            Popconfirm,
            { content: '确认删除该用户？', onConfirm: () => handleDelete(row) },
            () => h(Button, { size: 'small', variant: 'text', theme: 'danger' }, () => '删除'),
          ),
        );
      }
      return h(Space, { size: 4 }, () => buttons);
    },
  },
];

const formModalVisible = ref(false);
const formMode = ref<'create' | 'edit'>('create');
const editData = ref<User | null>(null);
const resetPwdVisible = ref(false);
const resetUserId = ref<number>(0);

async function fetchUsers() {
  loading.value = true;
  try {
    const res = await userApi.getList({
      page: pagination.current,
      page_size: pagination.pageSize,
      keyword: searchKeyword.value || undefined,
      status: searchStatus.value != null && searchStatus.value !== -1 ? searchStatus.value : undefined,
      role_id: searchRoleId.value ?? undefined,
    });
    userList.value = res.data.items;
    total.value = res.data.total;
    pagination.total = res.data.total;
  } catch (error: any) {
    // handled by global request; no-op
  } finally {
    loading.value = false;
  }
}

async function fetchRolesForSelect() {
  try {
    const res = await roleApi.getList({ page: 1, page_size: 100 });
    roleOptions.value = res.data.items.map((r) => ({
      label: r.name,
      value: r.id,
    }));
  } catch {
    // ignore
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchUsers();
}

function resetSearch() {
  searchKeyword.value = '';
  searchStatus.value = null;
  searchRoleId.value = null;
  pagination.current = 1;
  fetchUsers();
}

function onPageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchUsers();
}

function handleCreate() {
  formMode.value = 'create';
  editData.value = null;
  formModalVisible.value = true;
}

function handleEdit(row: User) {
  formMode.value = 'edit';
  editData.value = row;
  formModalVisible.value = true;
}

async function handleToggleStatus(row: User) {
  try {
    const newStatus = row.status === 1 ? 0 : 1;
    await userApi.setStatus(row.id, newStatus);
    fetchUsers();
  } catch (error: any) {
    // no-op
  }
}

function handleResetPassword(row: User) {
  resetUserId.value = row.id;
  resetPwdVisible.value = true;
}

async function handleDelete(row: User) {
  try {
    await userApi.delete(row.id);
    fetchUsers();
  } catch (error: any) {
    // no-op
  }
}

onMounted(() => {
  fetchRolesForSelect();
  fetchUsers();
});
</script>