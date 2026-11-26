<template>
  <t-card :title="'模块管理'" :bordered="false">
    <t-space direction="vertical" size="large" style="width: 100%">
      <t-space>
        <t-input
          v-model="searchKeyword"
          placeholder="搜索模块 ID / 名称"
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
        <t-button theme="primary" v-if="canCreate" @click="showInstallModal = true">
          安装模块
        </t-button>
        <t-button @click="resetSearch">重置</t-button>
      </t-space>

      <t-table
        row-key="id"
        :columns="columns"
        :data="moduleList"
        :loading="loading"
        :pagination="pagination"
        size="small"
        @page-change="onPageChange"
      />
    </t-space>

    <InstallModal v-model:visible="showInstallModal" @success="fetchModules" />

    <ConfigModal
      v-model:visible="showConfigModal"
      :module-id="configModuleId"
      :initial-config="configData"
      @success="fetchModules"
    />

    <t-dialog
      v-model:visible="showUninstallModal"
      :header="'卸载模块'"
      :confirm-btn="null"
      :cancel-btn="null"
      @close="showUninstallModal = false"
    >
      <div>
        <p>
          确认卸载模块
          <strong>{{ uninstallModule?.name }}</strong>
          （ID: {{ uninstallModule?.id }}）？
        </p>
        <p v-if="uninstallModule?.dependencies?.length" style="color: #e67e22">
          注意：该模块被其他模块依赖（{{ uninstallModule.dependencies.join(', ') }}），
          强制卸载可能导致依赖方异常。
        </p>
        <t-checkbox v-model="forceUninstallChecked">强制卸载（忽略依赖检查）</t-checkbox>
      </div>
      <template #footer>
        <t-space>
          <t-button @click="showUninstallModal = false">取消</t-button>
          <t-button theme="danger" :loading="uninstallLoading" @click="confirmUninstall">
            确认卸载
          </t-button>
        </t-space>
      </template>
    </t-dialog>
  </t-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, computed } from 'vue';
import type { VNode } from 'vue';
import { Tag, Space, Button, MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { moduleApi } from '@/api/modules';
import InstallModal from './components/InstallModal.vue';
import ConfigModal from './components/ConfigModal.vue';
import { useUserStore } from '@/store/user';
import type { Module } from '@/types';

const userStore = useUserStore();

const canCreate = computed(() => userStore.hasPermission('module_manager:module:create'));
const canEdit = computed(() => userStore.hasPermission('module_manager:module:edit'));
const canDelete = computed(() => userStore.hasPermission('module_manager:module:delete'));

const moduleList = ref<Module[]>([]);
const loading = ref(false);
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const searchKeyword = ref('');
const searchStatus = ref<string | null>(null);

const statusOptions = [
  { label: '全部', value: '' },
  { label: '已启用', value: 'active' },
  { label: '已停用', value: 'inactive' },
];

const showUninstallModal = ref(false);
const uninstallModule = ref<Module | null>(null);
const forceUninstallChecked = ref(false);
const uninstallLoading = ref(false);

const columns: PrimaryTableCol<any>[] = [
  { colKey: 'id', title: 'ID', width: 150 },
  { colKey: 'name', title: '名称', width: 150 },
  { colKey: 'version', title: '版本', width: 100 },
  {
    colKey: 'status',
    title: '状态',
    width: 100,
    cell: (_h, { row }) =>
      h(Tag, { theme: row.status === 'active' ? 'success' : 'default', variant: 'light' }, () =>
        row.status === 'active' ? '已启用' : '已停用',
      ),
  },
  {
    colKey: 'dependencies',
    title: '依赖',
    cell: (_h, { row }) => {
      if (!row.dependencies || row.dependencies.length === 0) {
        return h('span', '无');
      }
      return h(
        'div',
        { style: 'display:flex; flex-wrap:wrap; gap:4px;' },
        row.dependencies.map((dep: string) => h(Tag, { size: 'small', variant: 'light' }, () => dep)),
      );
    },
  },
  {
    colKey: 'installed_at',
    title: '安装时间',
    cell: (_h, { row }) => (row.installed_at ? new Date(row.installed_at).toLocaleString() : '-'),
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 220,
    cell: (_h, { row }) => {
      const buttons: VNode[] = [];

      if (canEdit.value) {
        const action = row.status === 'active' ? '停用' : '启用';
        buttons.push(
          h(
            Button,
            {
              size: 'small',
              variant: 'text',
              theme: row.status === 'active' ? 'warning' : 'success',
              onClick: () => handleToggle(row),
            },
            () => action,
          ),
        );
        buttons.push(
          h(Button, { size: 'small', variant: 'text', theme: 'primary', onClick: () => handleConfig(row) }, () => '配置'),
        );
      }

      if (canDelete.value) {
        buttons.push(
          h(Button, { size: 'small', variant: 'text', theme: 'danger', onClick: () => openUninstallDialog(row) }, () => '卸载'),
        );
      }

      return h(Space, { size: 4 }, () => buttons);
    },
  },
];

const showInstallModal = ref(false);
const showConfigModal = ref(false);
const configModuleId = ref('');
const configData = ref<Record<string, unknown>>({});

async function handleToggle(row: Module) {
  try {
    if (row.status === 'active') {
      await moduleApi.disable(row.id);
      MessagePlugin.success('停用成功');
    } else {
      await moduleApi.enable(row.id);
      MessagePlugin.success('启用成功');
    }
    fetchModules();
  } catch (err: any) {
    MessagePlugin.error(err.message || '操作失败');
  }
}

function openUninstallDialog(row: Module) {
  uninstallModule.value = row;
  forceUninstallChecked.value = false;
  showUninstallModal.value = true;
}

async function confirmUninstall() {
  if (!uninstallModule.value) return;
  uninstallLoading.value = true;
  try {
    await moduleApi.uninstall(uninstallModule.value.id, forceUninstallChecked.value);
    MessagePlugin.success('卸载成功');
    showUninstallModal.value = false;
    fetchModules();
  } catch (err: any) {
    MessagePlugin.error(err.message || '卸载失败');
  } finally {
    uninstallLoading.value = false;
  }
}

async function fetchModules() {
  loading.value = true;
  try {
    const res = await moduleApi.getList({
      page: pagination.current,
      page_size: pagination.pageSize,
      status: searchStatus.value || undefined,
      keyword: searchKeyword.value || undefined,
    });
    moduleList.value = res.data.items;
    pagination.total = res.data.total;
  } catch (err: any) {
    MessagePlugin.error(err.message || '加载模块列表失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.current = 1;
  fetchModules();
}

function resetSearch() {
  searchKeyword.value = '';
  searchStatus.value = null;
  pagination.current = 1;
  fetchModules();
}

function onPageChange(info: PageInfo) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchModules();
}

async function handleConfig(row: Module) {
  try {
    const res = await moduleApi.getConfig(row.id);
    configData.value = res.data as Record<string, unknown>;
    configModuleId.value = row.id;
    showConfigModal.value = true;
  } catch (err: any) {
    MessagePlugin.error(err.message || '获取配置失败');
  }
}

onMounted(() => {
  fetchModules();
});
</script>