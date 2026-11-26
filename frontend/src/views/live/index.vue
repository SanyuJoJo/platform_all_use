<template>
  <t-card :title="'直播管理'" :bordered="false">
    <div class="toolbar">
      <t-space>
        <t-button @click="loadRooms">刷新列表</t-button>
        <t-button theme="primary" @click="openCreate">新建直播间</t-button>
        <t-button @click="openManualPlay">任意拉流</t-button>
      </t-space>
      <div class="tips">提示：拉流可以直接播放任意 URL，或者点击下方卡片的「拉流」按钮</div>
    </div>

    <t-loading :loading="loading">
      <t-empty v-if="!loading && rooms.length === 0" description="暂无直播间" />
      <div v-else class="grid">
        <div v-for="room in rooms" :key="room.id" class="card">
          <div class="cover" :class="{ 'is-live': room.is_pushing }">
            <div class="cover-fallback">📺</div>
            <div class="badge" :class="room.is_pushing ? 'on' : 'off'">
              {{ room.is_pushing ? '● 直播中' : '未推流' }}
            </div>
          </div>
          <div class="card-body">
            <div class="title">{{ room.name }}</div>
            <div class="meta">ID: {{ room.stream_id }}</div>

            <div class="actions">
              <t-button size="small" theme="primary" @click="openPlay(room)">
                拉流
              </t-button>

              <t-button
                v-if="livePusher.activePushes.has(room.stream_id)"
                size="small"
                @click="openPushView(room)"
              >
                回到推流画面
              </t-button>

              <t-button size="small" theme="danger" variant="outline" @click="deleteRoom(room)">
                删除
              </t-button>
            </div>
          </div>
        </div>
      </div>
    </t-loading>

    <CreateLiveModal v-model:visible="showCreate" @created="onRoomCreated" />

    <PushStreamView
      v-model:visible="showPushView"
      :stream-id="currentPushRoom?.stream_id || ''"
      @stop="loadRooms"
    />

    <PlayStreamModal
      v-model:visible="showPlay"
      :stream="playRoom"
      :manual-mode="isManualPlay"
    />
  </t-card>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import { mediaApi } from '@/api/media';
import { livePusher } from '@/utils/media/live-pusher';
import CreateLiveModal from './components/CreateLiveModal.vue';
import PushStreamView from './components/PushStreamView.vue';
import PlayStreamModal from './components/PlayStreamModal.vue';
import type { LiveRoom } from '@/types';

const rooms = ref<LiveRoom[]>([]);
const loading = ref(false);

const showCreate = ref(false);
const showPushView = ref(false);
const currentPushRoom = ref<LiveRoom | null>(null);

const showPlay = ref(false);
const playRoom = ref<LiveRoom | null>(null);
const isManualPlay = ref(false);

let pollTimer: number | null = null;

async function loadRooms() {
  if (rooms.value.length === 0) loading.value = true;
  try {
    const res = await mediaApi.getRooms();
    rooms.value = res.data.list;
  } catch (e: any) {
    console.error('[live/index] 获取直播间失败', e);
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  showCreate.value = true;
}

function onRoomCreated(streamId: string) {
  loadRooms().then(() => {
    const room = rooms.value.find((r) => r.stream_id === streamId);
    if (room) {
      openPushView(room);
    }
  });
}

function openPushView(room: LiveRoom) {
  currentPushRoom.value = room;
  showPushView.value = true;
}

function openPlay(room: LiveRoom) {
  playRoom.value = room;
  isManualPlay.value = false;
  showPlay.value = true;
}

function openManualPlay() {
  playRoom.value = null;
  isManualPlay.value = true;
  showPlay.value = true;
}

async function deleteRoom(room: LiveRoom) {
  const dialog = DialogPlugin.confirm({
    header: '删除直播间',
    body: `确定删除「${room.name}」吗？`,
    confirmBtn: '删除',
    cancelBtn: '取消',
    theme: 'danger',
    onConfirm: async () => {
      try {
        await mediaApi.deleteRoom(room.id);
        if (livePusher.activePushes.has(room.stream_id)) {
          livePusher.stopPush(room.stream_id);
        }
        MessagePlugin.success('已删除');
        loadRooms();
      } catch (e: any) {
        MessagePlugin.error('删除失败');
      }
      dialog.destroy();
    },
    onClose: () => dialog.destroy(),
  });
}

onMounted(() => {
  loadRooms();
  pollTimer = window.setInterval(() => {
    loadRooms();
  }, 5000);
});

onBeforeUnmount(() => {
  if (pollTimer) window.clearInterval(pollTimer);
});
</script>

<style scoped>
.page {
  padding: 16px;
}
.toolbar {
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}
.tips {
  font-size: 12px;
  color: #888;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}
.card {
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
  transition: box-shadow 0.2s;
}
.card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.cover {
  position: relative;
  height: 140px;
  background: #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cover.is-live {
  background: #e8f5ee;
}
.cover-fallback {
  font-size: 48px;
  opacity: 0.5;
}
.badge {
  position: absolute;
  top: 8px;
  left: 8px;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  color: #fff;
}
.badge.on {
  background: #d03050;
}
.badge.off {
  background: #909399;
}
.card-body {
  padding: 12px;
}
.title {
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 4px;
}
.meta {
  font-size: 12px;
  color: #999;
  margin-bottom: 12px;
}
.actions {
  display: flex;
  gap: 8px;
}
</style>