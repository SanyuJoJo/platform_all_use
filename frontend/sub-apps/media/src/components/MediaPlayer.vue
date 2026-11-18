<template>
  <div class="media-player" :class="{ 'is-audio': isAudio }" @mousemove="showControls">
    <video
      v-if="!isAudio"
      ref="mediaEl"
      :src="src"
      :poster="poster"
      class="media-el"
      preload="metadata"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @progress="onProgress"
      @play="onPlay"
      @pause="onPause"
      @ended="onEnded"
      @waiting="onWaiting"
      @canplay="onCanPlay"
      @click="togglePlay"
    />
    <audio
      v-else
      ref="mediaEl"
      :src="src"
      class="media-el"
      preload="metadata"
      @loadedmetadata="onLoadedMetadata"
      @timeupdate="onTimeUpdate"
      @play="onPlay"
      @pause="onPause"
      @ended="onEnded"
    />
    <div class="ctrl-bar" :class="{ hidden: !controlsVisible }">
      <!-- 进度条 -->
      <div class="progress-wrap" @click="seekByClick">
        <div class="progress-bg"></div>
        <div class="progress-buffer" :style="{ width: bufferedPercent + '%' }"></div>
        <div class="progress-played" :style="{ width: playedPercent + '%' }"></div>
        <div class="progress-thumb" :style="{ left: playedPercent + '%' }"></div>
      </div>
      <!-- 操作按钮 -->
      <div class="ctrl-row">
        <button class="btn" @click="togglePlay">{{ playing ? '⏸' : '▶' }}</button>
        <button class="btn" title="后退 10s" @click="seekBy(-10)">⏪10s</button>
        <button class="btn" title="快进 10s" @click="seekBy(10)">10s⏩</button>
        <span class="time">{{ formatDuration(currentTime) }} / {{ formatDuration(duration) }}</span>
        <div class="spacer"></div>
        <!-- 音量 -->
        <div class="vol-wrap">
          <button class="btn" @click="toggleMute">{{ muted || volume === 0 ? '🔇' : '🔊' }}</button>
          <input type="range" min="0" max="1" step="0.05" v-model.number="volume" class="vol-slider" />
        </div>
        <!-- 倍速 -->
        <select class="speed" v-model.number="playbackRate" @change="applyRate">
          <option v-for="r in rates" :key="r" :value="r">{{ r }}x</option>
        </select>
        <!-- 画中画 -->
        <button v-if="!isAudio" class="btn" title="画中画" @click="togglePiP">⧉</button>
        <!-- 全屏 -->
        <button v-if="!isAudio" class="btn" title="全屏" @click="toggleFullscreen">⛶</button>
      </div>
    </div>
    <div v-if="loading" class="loading-mask">加载中…</div>
    <div v-if="error" class="error-mask">{{ error }}</div>
  </div>
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { formatDuration } from '@/utils/format';
const props = withDefaults(defineProps<{
  src: string;
  poster?: string;
  isAudio?: boolean;
  autoplay?: boolean;
}>(), { isAudio: false, autoplay: false });
const emit = defineEmits<{
  (e: 'ended'): void;
  (e: 'error', msg: string): void;
}>();
const mediaEl = ref<HTMLVideoElement | HTMLAudioElement | null>(null);
const playing = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const bufferedPercent = ref(0);
const volume = ref(1);
const muted = ref(false);
const playbackRate = ref(1);
const loading = ref(false);
const error = ref('');
const controlsVisible = ref(true);
const rates = [0.5, 0.75, 1, 1.25, 1.5, 2];
let hideTimer: number | null = null;
const playedPercent = computed(() => (duration.value ? (currentTime.value / duration.value) * 100 : 0));
function showControls() {
  controlsVisible.value = true;
  if (hideTimer) window.clearTimeout(hideTimer);
  hideTimer = window.setTimeout(() => {
    if (playing.value) controlsVisible.value = false;
  }, 3000);
}
function togglePlay() {
  const el = mediaEl.value; if (!el) return;
  if (el.paused) el.play().catch((e) => { error.value = '播放失败: ' + e.message; });
  else el.pause();
}
function seekBy(sec: number) {
  const el = mediaEl.value; if (!el) return;
  el.currentTime = Math.max(0, Math.min(el.duration || 0, el.currentTime + sec));
}
function seekByClick(e: MouseEvent) {
  const el = mediaEl.value; if (!el || !el.duration) return;
  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
  const ratio = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
  el.currentTime = ratio * el.duration;
}
function onLoadedMetadata() {
  const el = mediaEl.value; if (!el) return;
  duration.value = el.duration || 0;
  if (props.autoplay) el.play().catch(() => {});
}
function onTimeUpdate() { currentTime.value = mediaEl.value?.currentTime || 0; }
function onProgress() {
  const el = mediaEl.value; if (!el || !el.buffered.length || !el.duration) return;
  bufferedPercent.value = (el.buffered.end(el.buffered.length - 1) / el.duration) * 100;
}
function onPlay() { playing.value = true; showControls(); }
function onPause() { playing.value = false; controlsVisible.value = true; }
function onEnded() { playing.value = false; emit('ended'); }
function onWaiting() { loading.value = true; }
function onCanPlay() { loading.value = false; error.value = ''; }
function toggleMute() {
  const el = mediaEl.value; if (!el) return;
  muted.value = !muted.value;
  el.muted = muted.value;
}
function applyRate() { if (mediaEl.value) mediaEl.value.playbackRate = playbackRate.value; }
async function togglePiP() {
  const el = mediaEl.value as HTMLVideoElement | null; if (!el) return;
  try {
    if ((document as any).pictureInPictureElement) await (document as any).exitPictureInPicture();
    else await (el as any).requestPictureInPicture();
  } catch (e) { console.warn('画中画失败', e); }
}
async function toggleFullscreen() {
  const wrap = mediaEl.value?.parentElement; if (!wrap) return;
  if (document.fullscreenElement) await document.exitFullscreen();
  else await wrap.requestFullscreen();
}
watch(volume, (v) => { if (mediaEl.value) { mediaEl.value.volume = v; mediaEl.value.muted = false; muted.value = false; } });
onBeforeUnmount(() => { if (hideTimer) window.clearTimeout(hideTimer); });
</script>
<style scoped>
.media-player { position: relative; width: 100%; background: #000; border-radius: 8px; overflow: hidden; }
.media-player.is-audio { background: #1a1a1a; padding: 20px 0; }
.media-el { width: 100%; max-height: 70vh; display: block; background: #000; }
.media-player.is-audio .media-el { display: none; }
.ctrl-bar {
  position: absolute; left: 0; right: 0; bottom: 0;
  background: linear-gradient(to top, rgba(0,0,0,.85), transparent);
  padding: 8px 12px 12px; color: #fff; transition: opacity .2s;
}
.ctrl-bar.hidden { opacity: 0; pointer-events: none; }
.media-player.is-audio .ctrl-bar { position: static; background: none; }
.progress-wrap { position: relative; height: 8px; margin-bottom: 8px; cursor: pointer; }
.progress-bg, .progress-buffer, .progress-played {
  position: absolute; left: 0; top: 0; height: 100%; border-radius: 4px;
}
.progress-bg { width: 100%; background: rgba(255,255,255,.25); }
.progress-buffer { background: rgba(255,255,255,.4); }
.progress-played { background: #18a058; }
.progress-thumb {
  position: absolute; top: 50%; width: 14px; height: 14px; border-radius: 50%;
  background: #fff; transform: translate(-50%, -50%);
}
.ctrl-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.btn { background: transparent; border: none; color: #fff; cursor: pointer; font-size: 16px; padding: 4px 6px; border-radius: 4px; }
.btn:hover { background: rgba(255,255,255,.15); }
.time { font-size: 13px; opacity: .85; margin-left: 6px; }
.spacer { flex: 1; }
.vol-wrap { display: flex; align-items: center; gap: 4px; }
.vol-slider { width: 80px; }
.speed { background: transparent; color: #fff; border: 1px solid rgba(255,255,255,.4); border-radius: 4px; padding: 2px 4px; }
.loading-mask, .error-mask { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; color: #fff; background: rgba(0,0,0,.4); }
.error-mask { color: #ff7875; }
</style>
