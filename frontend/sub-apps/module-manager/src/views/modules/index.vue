<template>
  <div>
    <h1>模块管理</h1>
    <n-space vertical size="large">
      <!-- 搜索栏 -->
      <n-space>
        <n-input
          v-model:value="searchKeyword"
          placeholder="搜索模块 ID / 名称"
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
        <n-button type="primary" v-if="canCreate" @click="showInstallModal = true">
          安装模块
        </n-button>
        <n-button @click="resetSearch">重置</n-button>
      </n-space>

      <!-- 表格 -->
      <n-data-table
        :columns="columns"
        :data="moduleList"
        :loading="loading"
        :pagination="pagination"
        @update:page="onPageChange"
        @update:page-size="onPageSizeChange"
        size="small"
      />
    </n-space>

    <!-- 安装弹窗 -->
    <InstallModal v-model:visible="showInstallModal" @success="fetchModules" />

    <!-- 配置弹窗 -->
    <ConfigModal
      v-model:visible="showConfigModal"
      :module-id="configModuleId"
      :initial-config="configData"
      @success="fetchModules"
    />

    <!-- 卸载确认模态框 -->
    <n-modal v-model:show="showUninstallModal" title="卸载模块" preset="dialog" @close="showUninstallModal = false">
      <div>
        <p>确认卸载模块 <strong>{{ uninstallModule?.name }}</strong>（ID: {{ uninstallModule?.id }}）？</p>
        <p v-if="uninstallModule?.dependencies?.length" style="color: #e67e22;">
          注意：该模块被其他模块依赖（{{ uninstallModule.dependencies.join(', ') }}），强制卸载可能导致依赖方异常。
        </p>
        <n-checkbox v-model:checked="forceUninstallChecked">强制卸载（忽略依赖检查）</n-checkbox>
      </div>
      <template #action>
        <n-button @click="showUninstallModal = false">取消</n-button>
        <n-button type="error" :loading="uninstallLoading" @click="confirmUninstall">
          确认卸载
        </n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, computed } from 'vue';
import {
  NDataTable,
  NSpace,
  NInput,
  NSelect,
  NButton,
  useMessage,
  NTag,
  NModal,
  NCheckbox,
} from 'naive-ui';
import { moduleApi } from '@/api/modules';
import InstallModal from './components/InstallModal.vue';
import ConfigModal from './components/ConfigModal.vue';
import { useModuleStore } from '@/store/module';
import type { Module } from '@/types';

const message = useMessage();
const moduleStore = useModuleStore();

const canCreate = computed(() => moduleStore.hasPermission('module_manager:module:create'));
const canEdit = computed(() => moduleStore.hasPermission('module_manager:module:edit'));
const canDelete = computed(() => moduleStore.hasPermission('module_manager:module:delete'));

const moduleList = ref<Module[]>([]);
const loading = ref(false);
const total = ref(0);
const pagination = reactive({
  page: 1,
  pageSize: 20,
});

const searchKeyword = ref('');
const searchStatus = ref<string | null>(null);

const statusOptions = [
  { label: '全部', value: null },
  { label: '已启用', value: 'active' },
  { label: '已停用', value: 'inactive' },
];

// 卸载弹窗状态
const showUninstallModal = ref(false);
const uninstallModule = ref<Module | null>(null);
const forceUninstallChecked = ref(false);
const uninstallLoading = ref(false);

// 表格列定义
const columns = [
  { title: 'ID', key: 'id', width: 150 },
  { title: '名称', key: 'name', width: 150 },
  { title: '版本', key: 'version', width: 100 },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render(row: Module) {
      return h(
        NTag,
        { type: row.status === 'active' ? 'success' : 'default' },
        { default: () => row.status === 'active' ? '已启用' : '已停用' }
      );
    },
  },
  {
    title: '依赖',
    key: 'dependencies',
    render(row: Module) {
      if (!row.dependencies || row.dependencies.length === 0) {
        return h('span', '无');
      }
      return h('div', { style: 'display:flex; flex-wrap:wrap; gap:4px;' },
        row.dependencies.map(dep =>
          h(NTag, { size: 'small' }, { default: () => dep })
        )
      );
    },
  },
  { title: '安装时间', key: 'installed_at', render: (row: Module) => new Date(row.installed_at).toLocaleString() },
  {
    title: '操作',
    key: 'actions',
    width: 320,
    render(row: Module) {
      const buttons = [];

      if (canEdit.value) {
        const action = row.status === 'active' ? '停用' : '启用';
        const handleToggle = async () => {
          try {
            if (row.status === 'active') {
              await moduleApi.disable(row.id);
              message.success('停用成功');
            } else {
              await moduleApi.enable(row.id);
              message.success('启用成功');
            }
            fetchModules();
          } catch (err: any) {
            message.error(err.message || '操作失败');
          }
        };
        buttons.push(
          h(
            NButton,
            { size: 'small', type: row.status === 'active' ? 'warning' : 'primary', onClick: handleToggle },
            { default: () => action }
          )
        );
        buttons.push(
          h(
            NButton,
            { size: 'small', onClick: () => handleConfig(row) },
            { default: () => '配置' }
          )
        );
      }

      if (canDelete.value) {
        buttons.push(
          h(
            NButton,
            { size: 'small', type: 'error', onClick: () => openUninstallDialog(row) },
            { default: () => '卸载' }
          )
        );
      }

      return h(NSpace, null, { default: () => buttons });
    },
  },
];

// 弹窗控制
const showInstallModal = ref(false);
const showConfigModal = ref(false);
const configModuleId = ref('');
const configData = ref<any>({});

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
    message.success('卸载成功');
    showUninstallModal.value = false;
    fetchModules();
  } catch (err: any) {
    message.error(err.message || '卸载失败');
  } finally {
    uninstallLoading.value = false;
  }
}

async function fetchModules() {
  loading.value = true;
  try {
    const res = await moduleApi.getList({
      page: pagination.page,
      page_size: pagination.pageSize,
      status: searchStatus.value || undefined,
      keyword: searchKeyword.value || undefined,
    });
    moduleList.value = res.data.items;
    total.value = res.data.total;
  } catch (err: any) {
    message.error(err.message || '加载模块列表失败');
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  pagination.page = 1;
  fetchModules();
}

function resetSearch() {
  searchKeyword.value = '';
  searchStatus.value = null;
  pagination.page = 1;
  fetchModules();
}

function onPageChange(page: number) {
  pagination.page = page;
  fetchModules();
}

function onPageSizeChange(size: number) {
  pagination.pageSize = size;
  pagination.page = 1;
  fetchModules();
}

async function handleConfig(row: Module) {
  try {
    const res = await moduleApi.getConfig(row.id);
    configData.value = res.data;
    configModuleId.value = row.id;
    showConfigModal.value = true;
  } catch (err: any) {
    message.error(err.message || '获取配置失败');
  }
}

onMounted(() => {
  fetchModules();
});
</script>
