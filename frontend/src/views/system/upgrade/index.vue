<template>
  <t-card title="系统升级" :bordered="false">
    <!-- 当前状态 -->
    <t-descriptions title="当前版本" :colon="true" bordered>
      <t-descriptions-item label="当前版本">{{ status?.current_version || 'unknown' }}</t-descriptions-item>
      <t-descriptions-item label="安装目录">{{ status?.install_root || '-' }}</t-descriptions-item>
      <t-descriptions-item label="最近升级">{{ status?.last_upgrade_at || '-' }}</t-descriptions-item>
      <t-descriptions-item label="最近结果">{{ resultLabel(status?.last_upgrade_result) }}</t-descriptions-item>
      <t-descriptions-item label="可用更新">
        <t-tag :theme="status?.update_available ? 'warning' : 'success'" variant="light">
          {{ status?.update_available ? '有可用更新' : '已是最新' }}
        </t-tag>
      </t-descriptions-item>
    </t-descriptions>

    <!-- 可用升级提醒 -->
    <template v-if="status?.update_available && status?.pending?.length">
      <t-alert
        theme="warning"
        message="检测到可用升级版本，请在维护窗口执行升级。升级仅重启平台后端（约 1-2 秒），不会关机。"
        :close="false"
        style="margin-top: 16px"
      />
    </template>

    <!-- 操作按钮 -->
    <div class="toolbar" style="margin-top: 16px">
      <t-button theme="primary" @click="refresh">刷新</t-button>
    </div>

    <!-- 升级包列表 -->
    <t-table
      :columns="columns"
      :data="packages"
      :loading="loading"
      row-key="id"
      :pagination="false"
      style="margin-top: 16px"
    >
      <template #ops="{ row }">
        <t-button size="small" theme="primary" :disabled="!row.verified" @click="confirmUpgrade(row)">
          升级
        </t-button>
      </template>
    </t-table>

    <!-- 回退入口 -->
    <t-divider style="margin: 24px 0 8px" />
    <t-space>
      <t-input v-model="rollbackId" placeholder="输入回退 ID（见 apply 输出）" style="width: 280px" />
      <t-button theme="danger" variant="outline" :disabled="!rollbackId" @click="confirmRollback">
        回退
      </t-button>
    </t-space>

    <!-- 二次确认对话框 -->
    <t-dialog
      v-model:visible="confirmVisible"
      :header="confirmTitle"
      :body="confirmBody"
      confirm-btn="确认执行"
      cancel-btn="取消"
      @confirm="doExecute"
    />
  </t-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { upgradeApi, type UpgradeStatus, type UpgradePackage } from '@/api/upgrade';

const status = ref<UpgradeStatus | null>(null);
const packages = ref<UpgradePackage[]>([]);
const loading = ref(false);
const confirmVisible = ref(false);
const confirmTitle = ref('');
const confirmBody = ref('');
const rollbackId = ref('');
const pendingAction = ref<'upgrade' | 'rollback'>('upgrade');
const pendingPackageId = ref('');

const columns = [
  { colKey: 'to_version', title: '目标版本', width: 120 },
  { colKey: 'from_version', title: '源版本', width: 120 },
  { colKey: 'channel', title: '通道', width: 90 },
  { colKey: 'requires', title: '依赖', width: 90 },
  { colKey: 'created_at', title: '生成时间', width: 200 },
  { colKey: 'verified', title: '已校验', width: 90 },
  { colKey: 'ops', title: '操作', width: 90 },
];

function resultLabel(r?: string) {
  if (!r) return '-';
  const map: Record<string, string> = {
    ok: '✅ 成功',
    rolled_back: '↩️ 已回退',
    failed: '❌ 失败',
  };
  return map[r] || r;
}

async function refresh() {
  loading.value = true;
  try {
    const [st, pkg] = await Promise.all([upgradeApi.getStatus(), upgradeApi.listPackages()]);
    status.value = st.data;
    packages.value = pkg.data?.items || [];
  } catch (e: any) {
    MessagePlugin.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

function confirmUpgrade(row: UpgradePackage) {
  pendingAction.value = 'upgrade';
  pendingPackageId.value = row.id;
  confirmTitle.value = `升级到 ${row.to_version}？`;
  confirmBody.value = `将应用升级包 ${row.id}（${row.from_version} → ${row.to_version}）。` +
    '仅重启平台后端，不会关机。请确保已备份关键数据。';
  confirmVisible.value = true;
}

function confirmRollback() {
  pendingAction.value = 'rollback';
  pendingPackageId.value = rollbackId.value;
  confirmTitle.value = `回退到回退槽 ${rollbackId.value}？`;
  confirmBody.value = '将恢复上一槽位并重启平台后端。回退不删除数据。';
  confirmVisible.value = true;
}

async function doExecute() {
  confirmVisible.value = false;
  loading.value = true;
  try {
    if (pendingAction.value === 'upgrade') {
      await upgradeApi.run(pendingPackageId.value);
      MessagePlugin.success('升级已提交');
    } else {
      await upgradeApi.rollback(pendingPackageId.value);
      MessagePlugin.success('回退已提交');
    }
    rollbackId.value = '';
    await refresh();
  } catch (e: any) {
    MessagePlugin.error(e?.message || '操作失败');
  } finally {
    loading.value = false;
  }
}

onMounted(refresh);
</script>