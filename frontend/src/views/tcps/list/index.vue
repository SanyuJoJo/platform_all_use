<template>
  <t-card :title="''" :bordered="false">
    <template #title>
      <t-space align="center">
        <t-radio-group v-model="activeMode" variant="default-filled" @change="handleModeChange">
          <t-radio-button value="client">客户端</t-radio-button>
          <t-radio-button value="server">服务端</t-radio-button>
        </t-radio-group>
        <span class="current-mode-hint">
          当前：{{ activeMode === 'client' ? '客户端管理' : '服务端管理' }}
        </span>
      </t-space>
    </template>
    <template #actions>
      <t-space>
        <t-button
          v-if="hasPerm('ssl_proxy:tcps:create')"
          theme="primary"
          @click="openCreate"
        >
          新建配置
        </t-button>
        <t-button
          :disabled="checkedRowKeys.length === 0"
          :loading="batchStarting"
          @click="handleBatchStart"
        >
          批量启动
        </t-button>
        <t-button
          :disabled="checkedRowKeys.length === 0"
          :loading="batchStopping"
          @click="handleBatchStop"
        >
          批量停止
        </t-button>
        <t-button
          :disabled="checkedRowKeys.length === 0"
          :loading="batchRestarting"
          @click="handleBatchRestart"
        >
          批量重载
        </t-button>
        <t-button @click="fetchList">刷新</t-button>
      </t-space>
    </template>

    <t-table
      :columns="columns"
      :data="list"
      :loading="loading"
      row-key="service_id"
      :pagination="pagination"
      :scroll="{ x: 1200 }"
      v-model:selected-row-keys="checkedRowKeys"
      @page-change="handlePageChange"
    />

    <TcpsFormModal
      v-model:show="showForm"
      :service-id="editingId"
      :default-mode="activeMode"
      @success="handleFormSuccess"
    />

    <TcpsDetailDrawer
      v-model:show="showDetail"
      :service-id="detailId"
      @saved="handleDrawerSaved"
    />

    <TcpsBatchResultModal
      v-model:show="showBatchResult"
      :title="batchTitle"
      :result="batchResult"
      @retry-failed="handleRetryFailed"
    />
  </t-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import tcpsApi from '@/api/tcps';
import { useTcpsAction } from '@/composables/useTcpsAction';
import { useTcpsCertMap } from '@/composables/useTcpsCertMap';
import { buildTcpsColumns } from './TcpsColumns';
import TcpsBatchResultModal from './TcpsBatchResultModal.vue';
import TcpsFormModal from './TcpsFormModal.vue';
import TcpsDetailDrawer from '../detail/TcpsDetailDrawer.vue';
import { useUserStore } from '@/store/user';
import type {
  TcpsBatchItemResult,
  TcpsBatchResp,
  TcpsListItem,
  TcpsMode,
  TcpsStatus,
} from '@/types';

const userStore = useUserStore();
const hasPerm = (code: string) => userStore.hasPermission(code);

const { start, stop, isRunning, isTransitioning, isPending, pendingAction } =
  useTcpsAction();
const { load: loadCertMap, certName } = useTcpsCertMap();

const activeMode = ref<TcpsMode>('client');
const loading = ref(false);
const list = ref<TcpsListItem[]>([]);
const checkedRowKeys = ref<string[]>([]);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
});

const showForm = ref(false);
const editingId = ref<string | undefined>(undefined);

function openCreate() {
  editingId.value = undefined;
  showForm.value = true;
}

function handleFormSuccess() {
  loadCertMap(true);
  fetchList();
}

const showDetail = ref(false);
const detailId = ref<string | undefined>(undefined);

function openDetail(row: TcpsListItem) {
  detailId.value = row.service_id;
  showDetail.value = true;
}

function handleDrawerSaved() {
  fetchList();
}

const batchStarting = ref(false);
const batchStopping = ref(false);
const batchRestarting = ref(false);
const showBatchResult = ref(false);
const batchResult = ref<TcpsBatchResp | null>(null);
const batchTitle = ref('批量操作结果');
const lastBatchOp = ref<'start' | 'stop' | 'restart'>('start');

function pickIdsByStatus(
  predicate: (s: TcpsStatus) => boolean,
): { matched: string[]; skipped: string[] } {
  const matched: string[] = [];
  const skipped: string[] = [];
  const map = new Map(list.value.map((r) => [r.service_id, r]));

  for (const id of checkedRowKeys.value) {
    const row = map.get(id);
    if (!row) {
      skipped.push(id);
      continue;
    }
    if (predicate(row.status)) matched.push(id);
    else skipped.push(id);
  }
  return { matched, skipped };
}

function isStartable(s: TcpsStatus): boolean {
  return s === 'STOPPED' || s === 'FAILED';
}
function isStoppable(s: TcpsStatus): boolean {
  return s === 'RUNNING' || s === 'STARTING';
}
function isRestartable(s: TcpsStatus): boolean {
  return s === 'RUNNING';
}

function delayedRefresh(delayMs = 250) {
  const t = setTimeout(() => {
    fetchList();
  }, delayMs);
  return () => clearTimeout(t);
}

const columns = computed(() =>
  buildTcpsColumns({
    onDetail: (row) => openDetail(row),
    onStart: async (row) => {
      const cancel = delayedRefresh();
      const ok = await start(row.service_id);
      cancel();
      if (ok) fetchList();
    },
    onStop: async (row) => {
      const cancel = delayedRefresh();
      const ok = await stop(row.service_id);
      cancel();
      if (ok) fetchList();
    },
    onDelete: async (row) => {
      try {
        await tcpsApi.remove(row.service_id);
        MessagePlugin.success('删除成功');
        fetchList();
      } catch (e: any) {
        MessagePlugin.error(e.message || '删除失败');
      }
    },
    isRunning,
    isTransitioning,
    isPending,
    pendingAction,
    certName,
    hasPerm,
  }),
);

function handleModeChange() {
  pagination.current = 1;
  checkedRowKeys.value = [];
  fetchList();
}

async function fetchList() {
  loading.value = true;
  try {
    const res = await tcpsApi.list({
      mode: activeMode.value,
      page: pagination.current,
      size: pagination.pageSize,
    });
    const items = res.data.items || [];
    list.value = items;
    pagination.total = res.data.total || items.length;
  } catch (e: any) {
    MessagePlugin.error(e.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function handlePageChange(info: any) {
  pagination.current = info.current;
  pagination.pageSize = info.pageSize;
  fetchList();
}

async function handleBatchStart() {
  if (checkedRowKeys.value.length === 0) {
    MessagePlugin.warning('请先勾选要启动的服务');
    return;
  }
  const { matched, skipped } = pickIdsByStatus(isStartable);
  if (matched.length === 0) {
    MessagePlugin.info(
      `选中的 ${checkedRowKeys.value.length} 个服务均已启动或正在启动/停止中，无需启动`,
    );
    return;
  }
  if (skipped.length > 0) {
    MessagePlugin.info(`已跳过 ${skipped.length} 个非停止状态的服务，仅启动 ${matched.length} 个`);
  }

  batchStarting.value = true;
  lastBatchOp.value = 'start';
  try {
    const res = await tcpsApi.batchStart({ service_ids: matched });
    batchResult.value = res.data;
    batchTitle.value =
      skipped.length > 0
        ? `批量启动结果（成功 ${res.data.succeeded} / 失败 ${res.data.failed} / 跳过 ${skipped.length}）`
        : `批量启动结果（成功 ${res.data.succeeded} / 失败 ${res.data.failed}）`;
    showBatchResult.value = true;

    if (res.data.failed === 0) {
      MessagePlugin.success(
        `批量启动完成：成功 ${res.data.succeeded} 个${skipped.length > 0 ? `，跳过 ${skipped.length} 个` : ''}`,
      );
    } else if (res.data.succeeded === 0) {
      MessagePlugin.error(`批量启动失败：共 ${res.data.total} 个，全部失败`);
    } else {
      MessagePlugin.warning(
        `批量启动完成：成功 ${res.data.succeeded} 个，失败 ${res.data.failed} 个${skipped.length > 0 ? `，跳过 ${skipped.length} 个` : ''}`,
      );
    }

    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '批量启动失败');
  } finally {
    batchStarting.value = false;
  }
}

async function handleBatchStop() {
  if (checkedRowKeys.value.length === 0) {
    MessagePlugin.warning('请先勾选要停止的服务');
    return;
  }
  const { matched, skipped } = pickIdsByStatus(isStoppable);
  if (matched.length === 0) {
    MessagePlugin.info(
      `选中的 ${checkedRowKeys.value.length} 个服务均已停止或正在停止中，无需停止`,
    );
    return;
  }
  if (skipped.length > 0) {
    MessagePlugin.info(`已跳过 ${skipped.length} 个非运行状态的服务，仅停止 ${matched.length} 个`);
  }

  batchStopping.value = true;
  lastBatchOp.value = 'stop';
  try {
    const res = await tcpsApi.batchStop({ service_ids: matched });
    batchResult.value = res.data;
    batchTitle.value =
      skipped.length > 0
        ? `批量停止结果（成功 ${res.data.succeeded} / 失败 ${res.data.failed} / 跳过 ${skipped.length}）`
        : `批量停止结果（成功 ${res.data.succeeded} / 失败 ${res.data.failed}）`;
    showBatchResult.value = true;

    if (res.data.failed === 0) {
      MessagePlugin.success(
        `批量停止完成：成功 ${res.data.succeeded} 个${skipped.length > 0 ? `，跳过 ${skipped.length} 个` : ''}`,
      );
    } else if (res.data.succeeded === 0) {
      MessagePlugin.error(`批量停止失败：共 ${res.data.total} 个，全部失败`);
    } else {
      MessagePlugin.warning(
        `批量停止完成：成功 ${res.data.succeeded} 个，失败 ${res.data.failed} 个${skipped.length > 0 ? `，跳过 ${skipped.length} 个` : ''}`,
      );
    }

    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '批量停止失败');
  } finally {
    batchStopping.value = false;
  }
}

async function restartSafe(sid: string): Promise<TcpsBatchItemResult> {
  try {
    await tcpsApi.restart(sid);
    return { service_id: sid, success: true, code: 0, message: 'ok' };
  } catch (e: any) {
    if (e?.code === 40902) {
      for (let i = 0; i < 15; i++) {
        await new Promise((r) => setTimeout(r, 300));
        try {
          const res = await tcpsApi.get(sid);
          const st = res.data.status;
          if (st === 'STOPPED' || st === 'FAILED') {
            try {
              await tcpsApi.start(sid);
              return { service_id: sid, success: true, code: 0, message: 'ok' };
            } catch (e2: any) {
              return {
                service_id: sid,
                success: false,
                code: e2?.code ?? 50000,
                message: e2?.message || '重启失败（补发启动失败）',
              };
            }
          }
          if (st === 'RUNNING') {
            return { service_id: sid, success: true, code: 0, message: 'ok' };
          }
        } catch {
          // ignore
        }
      }
      return {
        service_id: sid,
        success: false,
        code: 40902,
        message: '重启超时（后端状态未收敛）',
      };
    }
    return {
      service_id: sid,
      success: false,
      code: e?.code ?? 50000,
      message: e?.message || '重启失败',
    };
  }
}

async function runRestartMany(ids: string[]): Promise<TcpsBatchResp> {
  const results: TcpsBatchItemResult[] = [];
  let succeeded = 0;
  let failed = 0;
  for (const sid of ids) {
    const r = await restartSafe(sid);
    results.push(r);
    if (r.success) succeeded++;
    else failed++;
  }
  return { total: ids.length, succeeded, failed, results };
}

async function handleBatchRestart() {
  if (checkedRowKeys.value.length === 0) {
    MessagePlugin.warning('请先勾选要重载的服务');
    return;
  }
  const { matched, skipped } = pickIdsByStatus(isRestartable);
  if (matched.length === 0) {
    MessagePlugin.info(`选中的 ${checkedRowKeys.value.length} 个服务都未在运行中，无法重载`);
    return;
  }
  if (skipped.length > 0) {
    MessagePlugin.info(`已跳过 ${skipped.length} 个非运行状态的服务，仅重载 ${matched.length} 个`);
  }

  batchRestarting.value = true;
  lastBatchOp.value = 'restart';
  try {
    const resp = await runRestartMany(matched);
    batchResult.value = resp;
    batchTitle.value =
      skipped.length > 0
        ? `批量重载结果（成功 ${resp.succeeded} / 失败 ${resp.failed} / 跳过 ${skipped.length}）`
        : `批量重载结果（成功 ${resp.succeeded} / 失败 ${resp.failed}）`;
    showBatchResult.value = true;

    if (resp.failed === 0) {
      MessagePlugin.success(
        `批量重载完成：成功 ${resp.succeeded} 个${skipped.length > 0 ? `，跳过 ${skipped.length} 个` : ''}`,
      );
    } else if (resp.succeeded === 0) {
      MessagePlugin.error(`批量重载失败：共 ${resp.total} 个，全部失败`);
    } else {
      MessagePlugin.warning(
        `批量重载完成：成功 ${resp.succeeded} 个，失败 ${resp.failed} 个${skipped.length > 0 ? `，跳过 ${skipped.length} 个` : ''}`,
      );
    }

    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '批量重载失败');
  } finally {
    batchRestarting.value = false;
  }
}

async function handleRetryFailed() {
  if (!batchResult.value) return;
  const failedIds = batchResult.value.results
    .filter((r) => !r.success)
    .map((r) => r.service_id);
  if (failedIds.length === 0) {
    MessagePlugin.info('没有失败项');
    return;
  }

  try {
    let resp: TcpsBatchResp;
    if (lastBatchOp.value === 'start') {
      const res = await tcpsApi.batchStart({ service_ids: failedIds });
      resp = res.data;
    } else if (lastBatchOp.value === 'stop') {
      const res = await tcpsApi.batchStop({ service_ids: failedIds });
      resp = res.data;
    } else {
      resp = await runRestartMany(failedIds);
    }
    batchResult.value = resp;
    const opLabel =
      lastBatchOp.value === 'start'
        ? '启动'
        : lastBatchOp.value === 'stop'
          ? '停止'
          : '重载';
    batchTitle.value = `重试失败项结果 · ${opLabel}（成功 ${resp.succeeded} / 失败 ${resp.failed}）`;

    if (resp.failed === 0) {
      MessagePlugin.success(`重试完成：成功 ${resp.succeeded} 个`);
    } else if (resp.succeeded === 0) {
      MessagePlugin.error(`重试失败：全部 ${resp.failed} 个仍失败`);
    } else {
      MessagePlugin.warning(`重试完成：成功 ${resp.succeeded} 个，失败 ${resp.failed} 个`);
    }

    await fetchList();
  } catch (e: any) {
    MessagePlugin.error(e.message || '重试失败');
  }
}

onMounted(async () => {
  await loadCertMap();
  await fetchList();
});
</script>

<style scoped>
.current-mode-hint {
  color: #6b7280;
  font-size: 12px;
  margin-left: 4px;
}
:deep(.mono) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}
</style>