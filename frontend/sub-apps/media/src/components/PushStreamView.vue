<template>
  <n-modal
    v-model:show="visible"
    preset="dialog"
    title="本地推流画面"
    :z-index="3100"
    style="width: 800px;"
    @after-leave="onAfterLeave"
  >
    <div v-if="isPushing" class="video-wrapper">
      <!-- 关键：绑定从 livePusher 取出的 MediaStream -->
      <video ref="videoRef" autoplay muted playsinline class="video-el" />
      <div v-if="!videoReady" class="loading-mask">正在连接本地视频画面...</div>
    </div>
    <n-empty v-else description="当前没有本地推流，或流已断开" />

    <template #action>
      <n-button @click="visible = false">关闭窗口 (后台推流)</n-button>
      <n-button v-if="isPushing" type="error" @click="stopPush">停止推流</n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { NModal, NButton, NEmpty, useMessage } from 'naive-ui';
import { livePusher } from '@/utils/live-pusher';

const props = defineProps<{
  visible: boolean;
  streamId: string;
}>();

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
  (e: 'stop'): void;
}>();

const message = useMessage();
const visible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const videoRef = ref<HTMLVideoElement | null>(null);
const videoReady = ref(false);

// 判断当前有没有在推流
const isPushing = computed(() => livePusher.activePushes.has(props.streamId));

// 把 MediaStream 绑定到 video 标签
function bindStream() {
  if (!isPushing.value) return;
  const push = livePusher.activePushes.get(props.streamId);
  if (push && push.stream && videoRef.value) {
    videoRef.value.srcObject = push.stream;
    videoRef.value.play()
      .then(() => { videoReady.value = true; })
      .catch((e) => {
        console.warn('[PushView] 播放本地流被拦截', e);
        videoReady.value = true; // 尝试显示
      });
  }
}

watch(() => props.visible, async (v) => {
  if (v) {
    videoReady.value = false;
    await nextTick();
    bindStream();
  }
});

onMounted(() => {
  if (props.visible) {
    bindStream();
  }
});

function stopPush() {
  livePusher.stopPush(props.streamId);
  message.success('已停止推流');
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
