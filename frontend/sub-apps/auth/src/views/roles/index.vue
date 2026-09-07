<template>
  <div>
    <h1>角色管理</h1>
    <n-space vertical size="large">
      <n-space>
        <n-input
          v-model:value="searchKeyword"
          placeholder="搜索角色名/编码"
          style="width: 300px"
          clearable
          @input="handleSearch"
        />
        <n-button type="primary" v-permission="'auth:role:create'" @click="handleCreate">新建角色</n-button>
        <n-button @click="resetSearch">重置</n-button>
      </n-space>
      <n-data-table
        :columns="columns"
        :data="roleList"
        :loading="loading"
        :pagination="pagination"
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
        size="small"
      />
    </n-space>
    <RoleFormModal
      v-model:visible="formModalVisible"
      :mode="formMode"
      :initial-data="editData"
      @success="fetchRoles"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue';
import { NDataTable, NSpace, NInput, NButton, useMessage, NTag, NPopconfirm } from 'naive-ui';
import { roleApi } from '@/api/role';
import RoleFormModal from './components/RoleFormModal.vue';
import { useUserStore } from '@/store/user';
import type { Role } from '@/types';

const message = useMessage();
const userStore = useUserStore();

const roleList = ref<Role[]>([]);
const loading = ref(false);
const total = ref(0);
const pagination = reactive({
  page: 1,
  pageSize: 20,
});
const searchKeyword = ref('');

const columns = [
  { title: 'ID', key: 'id' },
  { title: '角色名', key: 'name' },
  { title: '编码', key: 'code' },
  { title: '描述', key: 'description' },
  {
    title: '系统内置',
    key: 'is_system',
    render(row: Role) {
      return h(
        NTag,
        { type: row.is_system ? 'info' : 'default' },
        { default: () => row.is_system ? '是' : '否' }
      );
    },
  },
  {
    title: '权限数',
    key: 'permission_codes',
    render(row: Role) {
      return row.permission_codes.length;
    },
  },
  { title: '创建时间', key: 'created_at', render: (row: Role) => new Date(row.created_at).toLocaleString() },
  {
    title: '操作',
    key: 'actions',
    render(row: Role) {
      const canEdit = userStore.hasPermission('auth:role:edit');
      const canDelete = userStore.hasPermission('auth:role:delete');
      const buttons = [];
      if (canEdit) {
        buttons.push(
          h(NButton, { size: 'small', onClick: () => handleEdit(row) }, { default: () => '编辑' })
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
              default: () => '确认删除该角色？',
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', type: 'error', disabled: row.is_system },
                  { default: () => '删除' }
                ),
            }
          )
        );
      }
      return h(NSpace, null, { default: () => buttons });
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
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: searchKeyword.value || undefined,
    });
    roleList.value = res.data.items;
    total.value = res.data.total;
  } catch (error: any) {
    message.error(error.message || '加载角色列表失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.page = 1;
  fetchRoles();
}

function resetSearch() {
  searchKeyword.value = '';
  pagination.page = 1;
  fetchRoles();
}

function onPageChange(page: number) {
  pagination.page = page;
  fetchRoles();
}

function onPageSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.page = 1;
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
    message.success('删除成功');
    fetchRoles();
  } catch (error: any) {
    // 根据错误码给出更友好的提示
    if (error.code === 20003) {
      message.error('系统内置角色不可删除');
    } else if (error.code === 20005) {
      message.error('角色已被用户使用，请先解除关联');
    } else {
      message.error(error.message || '删除失败');
    }
  }
}

onMounted(fetchRoles);
</script>