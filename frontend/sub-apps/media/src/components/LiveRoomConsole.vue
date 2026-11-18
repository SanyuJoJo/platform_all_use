<template>
  <n-modal
    v-model:show="visible"
    preset="dialog"
    :title="`直播间控制台 - ${room?.name || streamId}`"
    :z-index="3000"
    :auto-focus="false"
    :display-directive="'show'"
    style="width: 1000px;"
    @after-leave="onAfterLeave"
  >
    <div class="console-body">
      <n-alert v-if="isPushing" type="warning" :bordered="false" style="margin-bottom: 12px;">
        ⚠️ 本地推流正在进行中！刷新当前页面或关闭浏览器将导致推流中断，请谨慎操作。
      </n-alert>
      <n-alert v-else-if="!isPushing" type="info" :bordered="false" style="margin-bottom: 12px;">
        当前没有本地推流（可能是 OBS 推流或流已断开），您只能在右侧观看拉流画面。
      </n-alert>

      <div class="video-grid">
        <!-- 左侧：本地推流预览 -->
        <div v-if="isPushing" class="video-box">
          <div class="box-title">本地推流画面 (未编码)</div>
          <video ref="localVideoRef" autoplay muted playsinline class="video-el" />
        </div>
        
        <!-- 右侧：拉流预览 -->
        <div class="video-box">
          <div class="box-title">拉流观看 (HTTP-FLV)</div>
          <video ref="flvVideoRef" controls autoplay playsinline class="video-el" />
          <div v-if="loadingUrl" class="loading-mask">正在获取流地址...</div>
          <div v-if="errorMsg" class="error-mask">{{ errorMsg }}</div>
        </div>
      </div>
    </div>
    
    <template #action>
      <n-button @click="visible = false">关闭窗口 (后台运行)</n-button>
      <n-button v-if="isPushing" type="error" @click="stopLocalPush">停止本地推流</n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { NModal, NButton, NAlert, useMessage } from 'naive-ui';
import mpegts from 'mpegts.js';
import { livePusher } from '@/utils/live-pusher';
import { mediaApi, withToken } from '@/api/media';
import type { LiveRoom } from '@/types';

const props = defineProps<{
  visible: boolean;
  streamId: string;
  room?: LiveRoom | null;
}>();

const emit = defineEmits<{ (e: 'update:visible', v: boolean): void; (e: 'stop'): void; }>();

const message = useMessage();
const visible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const localVideoRef = ref<HTMLVideoElement | null>(null);
const flvVideoRef = ref<HTMLVideoElement | null>(null);
const loadingUrl = ref(false);
const errorMsg = ref('');
let flvPlayer: mpegts.Player | null = null;

// 判断当前是否有本地推流
const isPushing = computed(() => livePusher.activePushes.has(props.streamId));

// 监听刷新事件
function handleBeforeUnload(e: BeforeUnloadEvent) {
  if (isPushing.value) {
    e.preventDefault();
    e.returnValue = '刷新页面将中断推流，确定要离开吗？';
    return e.returnValue;
  }
}

onMounted(() => {
  window.addEventListener('beforeunload', handleBeforeUnload);
});

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', handleBeforeUnload);
  destroyPlayer();
});

watch(() => props.visible, async (v) => {
  if (v) {
    await nextTick();
    bindLocalStream();
    await playRemoteStream();
  } else {
    destroyPlayer();
  }
});

// 绑定本地推流画面
function bindLocalStream() {
  if (isPushing.value && localVideoRef.value) {
    const push = livePusher.activePushes.get(props.streamId);
    if (push && push.stream) {
      localVideoRef.value.srcObject = push.stream;
      localVideoRef.value.play().catch(() => {});
    }
  }
}

// 拉流播放
async function playRemoteStream() {
  if (!flvVideoRef.value) return;
  loadingUrl.value = true;
  errorMsg.value = '';
  destroyPlayer();

  try {
    const res = await mediaApi.getPlayUrlOfStream(props.streamId, 'flv');
    let url = res.data.url;
    
    // 规范 URL
    if (!/^https?:\/\//i.test(url)) {
      url = `${window.location.origin}${url.startsWith('/') ? url : '/' + url}`;
    }
    url = url.replace('/hls.m3u8', '.live.flv');
    url = withToken(url);

    if (mpegts.isSupported()) {
      flvPlayer = mpegts.createPlayer({ type: 'flv', isLive: true, url });
      flvPlayer.attachMediaElement(flvVideoRef.value);
      flvPlayer.load();
      const p = flvPlayer.play() as any;
      if (p && typeof p.catch === 'function') {
        p.catch(() => { errorMsg.value = '点击视频播放'; });
      }
      flvPlayer.on(mpegts.Events.ERROR, () => {
        errorMsg.value = '拉流失败或流已断开';
      });
    } else {
      errorMsg.value = '浏览器不支持 FLV';
    }
  } catch (e: any) {
    errorMsg.value = e?.message || '获取拉流地址失败';
  } finally {
    loadingUrl.value = false;
  }
}

function destroyPlayer() {
  if (flvPlayer) {
    flvPlayer.pause();
    flvPlayer.unload();
    flvPlayer.detachMediaElement();
    flvPlayer.destroy();
    flvPlayer = null;
  }
  if (flvVideoRef.value) flvVideoRef.value.removeAttribute('src');
  if (localVideoRef.value) localVideoRef.value.srcObject = null;
}

// 停止本地推流
function stopLocalPush() {
  livePusher.stopPush(props.streamId);
  message.success('本地推流已停止');
  emit('stop');
  visible.value = false;
}

function onAfterLeave() {
  destroyPlayer();
}
</script>

<style scoped>
.console-body { padding: 8px 0; }
.video-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.video-box {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000;
  border-radius: 6px;
  overflow: hidden;
}
.box-title {
  position: absolute;
  top: 0; left: 0; right: 0;
  background: rgba(0,0,0,0.6);
  color: #fff;
  font-size: 12px;
  padding: 4px 8px;
  z-index: 10;
}
.video-el {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.loading-mask, .error-mask {
  position: absolute; inset: 0;
  display: flex; align-items: center; justify-content: center;
  color: #fff; background: rgba(0,0,0,0.5);
}
.error-mask { color: #ff7875; }
</style>
