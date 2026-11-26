<template>
  <t-dialog
    v-model:visible="visible"
    :header="'本地推流画面'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="800"
    @close="onAfterLeave"
  >
    <div v-if="isPushing" class="video-wrapper">
      <video ref="videoRef" autoplay muted playsinline class="video-el" />
      <div v-if="!videoReady" class="loading-mask">正在连接本地视频画面...</div>
    </div>
    <t-empty v-else description="当前没有本地推流，或流已断开" />

    <template #footer>
      <t-space>
        <t-button @click="backgroundPush">收起（后台继续推流）</t-button>
        <t-button v-if="isPushing" theme="danger" @click="stopPush">停止推流</t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { livePusher } from '@/utils/media/live-pusher';

const props = defineProps<{
  visible: boolean;
  streamId: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
  (e: 'stop'): void;
}>();

const visible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const videoRef = ref<HTMLVideoElement | null>(null);
const videoReady = ref(false);

const isPushing = computed(() => livePusher.activePushes.has(props.streamId));

function bindStream() {
  if (!isPushing.value) return;
  const push = livePusher.activePushes.get(props.streamId);
  if (push && push.stream && videoRef.value) {
    videoRef.value.srcObject = push.stream;
    videoRef.value
      .play()
      .then(() => {
        videoReady.value = true;
      })
      .catch((e) => {
        console.warn('[PushView] 播放本地流被拦截', e);
        videoReady.value = true;
      });
  }
}

watch(
  () => props.visible,
  async (v) => {
    if (v) {
      videoReady.value = false;
      await nextTick();
      bindStream();
    }
  },
);

onMounted(() => {
  if (props.visible) {
    bindStream();
  }
});

function backgroundPush() {
  livePusher.backgroundKeep(props.streamId);
  MessagePlugin.success('已收起窗口，推流继续（可回到推流画面）');
  visible.value = false;
}

function stopPush() {
  livePusher.stopPush(props.streamId, true);
  MessagePlugin.success('已停止推流');
  emit('stop');
  visible.value = false;
}

function onAfterLeave() {
  if (videoRef.value) {
    videoRef.value.srcObject = null;
  }
  videoReady.value = false;
}
</script>

<style scoped>
.video-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000;
  border-radius: 6px;
  overflow: hidden;
}
.video-el {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.loading-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: rgba(0, 0, 0, 0.5);
}
</style>