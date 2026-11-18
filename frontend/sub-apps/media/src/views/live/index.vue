<template>
  <div class="page">
    <div class="toolbar">
      <n-space>
        <n-button @click="loadRooms">刷新列表</n-button>
        <n-button type="primary" @click="openCreate">新建直播间</n-button>
        <n-button type="info" @click="openManualPlay">任意拉流</n-button>
      </n-space>
      <div class="tips">提示：拉流可以直接播放任意 URL，或者点击下方卡片的「拉流」按钮</div>
    </div>

    <n-spin :show="loading">
      <n-empty v-if="!loading && rooms.length === 0" description="暂无直播间" />
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
              <!-- 统一的拉流按钮 -->
              <n-button size="small" type="primary" @click="openPlay(room)">
                拉流
              </n-button>

              <!-- 只有本地推流时才显示回到推流界面 -->
              <n-button 
                v-if="livePusher.activePushes.has(room.stream_id)" 
                size="small" 
                type="info" 
                @click="openPushView(room)"
              >
                回到推流画面
              </n-button>
              
              <n-button size="small" type="error" ghost @click="deleteRoom(room)">
                删除
              </n-button>
            </div>
          </div>
        </div>
      </div>
    </n-spin>

    <!-- 新建弹窗 -->
    <CreateLiveModal 
      v-model:visible="showCreate" 
      @created="onRoomCreated" 
    />
    
    <!-- 推流本地画面弹窗 -->
    <PushStreamView
      v-model:visible="showPushView"
      :stream-id="currentPushRoom?.stream_id || ''"
      @stop="loadRooms"
    />

    <!-- 拉流弹窗 (点击卡片或顶部按钮都会打开它) -->
    <PlayStreamModal
      v-model:visible="showPlay"
      :stream="playRoom"
      :manual-mode="isManualPlay"
    />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { NSpace, NButton, NSpin, NEmpty, useMessage, useDialog } from 'naive-ui';
import { mediaApi } from '@/api/media';
import { livePusher } from '@/utils/live-pusher';
import CreateLiveModal from '@/components/CreateLiveModal.vue';
import PushStreamView from '@/components/PushStreamView.vue';
import PlayStreamModal from '@/components/PlayStreamModal.vue';
import type { LiveRoom } from '@/types';

const message = useMessage();
const dialog = useDialog();

const rooms = ref<LiveRoom[]>([]);
const loading = ref(false);

/* ---------- 弹窗状态 ---------- */
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

// 新建成功后，自动打开推流界面
function onRoomCreated(streamId: string) {
  loadRooms().then(() => {
    const room = rooms.value.find(r => r.stream_id === streamId);
    if (room) {
      openPushView(room);
    }
  });
}

function openPushView(room: LiveRoom) {
  currentPushRoom.value = room;
  showPushView.value = true;
}

// 统一拉流入口
function openPlay(room: LiveRoom) {
  playRoom.value = room;
  isManualPlay.value = false;
  showPlay.value = true;
}

// 任意拉流入口
function openManualPlay() {
  playRoom.value = null;
  isManualPlay.value = true;
  showPlay.value = true;
}

async function deleteRoom(room: LiveRoom) {
  dialog.warning({
    title: '删除直播间',
    content: `确定删除「${room.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await mediaApi.deleteRoom(room.id);
        if (livePusher.activePushes.has(room.stream_id)) {
          livePusher.stopPush(room.stream_id);
        }
        message.success('已删除');
        loadRooms();
      } catch (e: any) {
        message.error('删除失败');
      }
    }
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
.page { padding: 16px; }
.toolbar { margin-bottom: 16px; display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.tips { font-size: 12px; color: #888; }
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.card { border: 1px solid #eee; border-radius: 8px; overflow: hidden; background: #fff; transition: box-shadow 0.2s; }
.card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.cover { position: relative; height: 140px; background: #f0f2f5; display: flex; align-items: center; justify-content: center; }
.cover.is-live { background: #e8f5ee; }
.cover-fallback { font-size: 48px; opacity: 0.5; }
.badge { position: absolute; top: 8px; left: 8px; font-size: 12px; padding: 2px 8px; border-radius: 4px; color: #fff; }
.badge.on { background: #d03050; }
.badge.off { background: #909399; }
.card-body { padding: 12px; }
.title { font-weight: 600; font-size: 15px; margin-bottom: 4px; }
.meta { font-size: 12px; color: #999; margin-bottom: 12px; }
.actions { display: flex; gap: 8px; flex-wrap: wrap; }
</style>
