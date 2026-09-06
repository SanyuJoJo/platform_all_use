<template>
  <div>
    <h1>用户管理</h1>
    <n-space vertical size="large">
      <!-- 搜索栏 -->
      <n-space>
        <n-input
          v-model:value="searchKeyword"
          placeholder="搜索用户名/昵称/邮箱"
          style="width: 300px"
          clearable
          @input="handleSearch"
        />
        <n-select
          v-model:value="searchStatus"
          placeholder="状态"
          clearable
          style="width: 150px"
          :options="statusOptions"
          @update:value="handleSearch"
        />
        <n-select
          v-model:value="searchRoleId"
          placeholder="角色"
          clearable
          style="width: 150px"
          :options="roleOptions"
          @update:value="handleSearch"
        />
        <n-button type="primary" v-permission="'auth:user:create'" @click="handleCreate">新建用户</n-button>
        <n-button @click="resetSearch">重置</n-button>
      </n-space>
      <!-- 表格 -->
      <n-data-table
        :columns="columns"
        :data="userList"
        :loading="loading"
        :pagination="pagination"
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
        size="small"
      />
    </n-space>
    <!-- 创建/编辑弹窗 -->
    <UserFormModal
      v-model:visible="formModalVisible"
      :mode="formMode"
      :initial-data="editData"
      @success="fetchUsers"
    />
    <!-- 重置密码弹窗 -->
    <ResetPasswordModal
      v-model:visible="resetPwdVisible"
      :user-id="resetUserId"
      @success="fetchUsers"
    />
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import { NDataTable, NSpace, NInput, NSelect, NButton, useMessage, NTag, NPopconfirm } from 'naive-ui';
import { userApi } from '@/api/user';
import { roleApi } from '@/api/role';
import UserFormModal from './components/UserFormModal.vue';
import ResetPasswordModal from './components/ResetPasswordModal.vue';
import { useUserStore } from '@/store/user';
import type { User } from '@/types';
const message = useMessage();
const userStore = useUserStore();
// 列表数据
const userList = ref<User[]>([]);
const loading = ref(false);
const total = ref(0);
const pagination = reactive({
  page: 1,
  pageSize: 20,
});
// 搜索条件
const searchKeyword = ref('');
const searchStatus = ref<number | null>(null);
const searchRoleId = ref<number | null>(null);
// 角色下拉选项
const roleOptions = ref<{ label: string; value: number }[]>([]);
// 表格列定义（操作列已根据权限控制）
const columns = [
  { title: 'ID', key: 'id' },
  { title: '用户名', key: 'username' },
  { title: '昵称', key: 'nickname' },
  { title: '邮箱', key: 'email' },
  {
    title: '状态',
    key: 'status',
    render(row: User) {
      return h(
        NTag,
        { type: row.status === 1 ? 'success' : 'error' },
        { default: () => row.status === 1 ? '启用' : '禁用' }
      );
    },
  },
  {
    title: '角色',
    key: 'roles',
    render(row: User) {
      return row.roles.map(r => r.name).join('、');
    },
  },
  { title: '创建时间', key: 'created_at', render: (row: User) => new Date(row.created_at).toLocaleString() },
  {
    title: '操作',
    key: 'actions',
    render(row: User) {
      const canEdit = userStore.hasPermission('auth:user:edit');
      const canDelete = userStore.hasPermission('auth:user:delete');
      const canReset = userStore.hasPermission('auth:user:edit');
      const canToggle = userStore.hasPermission('auth:user:edit');
      const buttons = [];
      if (canEdit) {
        buttons.push(
          h(NButton, { size: 'small', onClick: () => handleEdit(row) }, { default: () => '编辑' })
        );
      }
      if (canToggle) {
        buttons.push(
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleToggleStatus(row),
            },
            {
              default: () => `确认${row.status === 1 ? '禁用' : '启用'}该用户？`,
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', type: row.status === 1 ? 'warning' : 'success' },
                  { default: () => row.status === 1 ? '禁用' : '启用' }
                ),
            }
          )
        );
      }
      if (canReset) {
        buttons.push(
          h(NButton, { size: 'small', onClick: () => handleResetPassword(row) }, { default: () => '重置密码' })
        );
      }
      if (canDelete) {
        buttons.push(
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDelete(row),
            },
            {
              default: () => '确认删除该用户？',
              trigger: () =>
                h(NButton, { size: 'small', type: 'error' }, { default: () => '删除' }),
            }
          )
        );
      }
      return h(NSpace, null, { default: () => buttons });
    },
  },
];
// 弹窗控制
const formModalVisible = ref(false);
const formMode = ref<'create' | 'edit'>('create');
const editData = ref<User | null>(null);
const resetPwdVisible = ref(false);
const resetUserId = ref<number>(0);
// 获取用户列表
async function fetchUsers() {
  loading.value = true;
  try {
    const res = await userApi.getList({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchKeyword.value || undefined,
      status: searchStatus.value ?? undefined,
      role_id: searchRoleId.value ?? undefined,
    });
    userList.value = res.data.items;
    total.value = res.data.total;
  } catch (error: any) {
    message.error(error.message || '加载用户列表失败');
  } finally {
    loading.value = false;
  }
}
// 获取角色列表用于下拉
async function fetchRolesForSelect() {
  try {
    const res = await roleApi.getList({ page: 1, page_size: 100 });
    roleOptions.value = res.data.items.map(r => ({ label: r.name, value: r.id }));
  } catch (error) {
    // 忽略
  }
}
// 搜索与翻页
function handleSearch() {
  pagination.page = 1;
  fetchUsers();
}
function resetSearch() {
  searchKeyword.value = '';
  searchStatus.value = null;
  searchRoleId.value = null;
  pagination.page = 1;
  fetchUsers();
}
function onPageChange(page: number) {
  pagination.page = page;
  fetchUsers();
}
function onPageSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.page = 1;
  fetchUsers();
}
// 创建/编辑
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
// 启用/禁用
async function handleToggleStatus(row: User) {
  try {
    const newStatus = row.status === 1 ? 0 : 1;
    await userApi.setStatus(row.id, newStatus);
    message.success('状态更新成功');
    fetchUsers();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
}
// 重置密码
function handleResetPassword(row: User) {
  resetUserId.value = row.id;
  resetPwdVisible.value = true;
}
// 删除
async function handleDelete(row: User) {
  try {
    await userApi.delete(row.id);
    message.success('删除成功');
    fetchUsers();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
}
const statusOptions = [
  { label: '全部', value: null },
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
];
onMounted(() => {
  fetchRolesForSelect();
  fetchUsers();
});
</script>
