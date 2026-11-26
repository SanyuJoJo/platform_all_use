<template>
  <t-dialog
    v-model:visible="visible"
    :header="modalTitle"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="900"
    @close="onAfterLeave"
  >
    <div ref="bodyRef" class="body">
      <div v-if="manualMode && !currentUrl" class="manual-row">
        <t-input
          v-model="manualInput"
          placeholder="输入 stream_id 或完整 URL (如 rtmp://... 或 http://...m3u8)"
          @keyup.enter="playManual"
        />
        <t-button theme="primary" :loading="resolving" @click="playManual">拉流播放</t-button>
      </div>

      <t-alert v-if="errorMsg" theme="error" :message="errorMsg" style="margin-bottom: 12px" />

      <div v-if="currentUrl && !errorMsg" class="player-section">
        <div class="format-switch">
          <t-radio-group v-model="format" @change="switchFormat">
            <t-radio-button value="flv">HTTP-FLV (低延迟)</t-radio-button>
            <t-radio-button value="hls">HLS (高兼容)</t-radio-button>
            <t-radio-button value="webrtc">WebRTC (超低延迟)</t-radio-button>
          </t-radio-group>
        </div>

        <div class="video-container">
          <video v-show="format === 'flv'" ref="videoElFlv" controls autoplay playsinline class="video-el" />
          <video v-show="format === 'hls'" ref="videoElHls" controls autoplay playsinline class="video-el" />
          <video v-show="format === 'webrtc'" ref="videoElWebrtc" controls autoplay playsinline class="video-el" />
        </div>
      </div>

      <div v-if="!manualMode && currentStreamId" class="address-panel">
        <t-tabs :default-value="'pull'">
          <t-tab-panel value="pull" label="拉流分发地址 (给第三方系统拉取)">
            <t-alert theme="info" :message="'将此地址填入第三方流媒体服务器、CDN 或 VLC 播放器，即可拉取当前流。'" style="margin-bottom: 12px" />

            <div class="addr-item">
              <span class="label">HTTP-FLV</span>
              <t-input :value="getAddr('flv')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" theme="primary" @click="copy(getAddr('flv'), $event)">复制</t-button>
            </div>
            <div class="addr-item">
              <span class="label">HLS (m3u8)</span>
              <t-input :value="getAddr('hls')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" @click="copy(getAddr('hls'), $event)">复制</t-button>
            </div>
            <div class="addr-item">
              <span class="label">RTMP 拉流</span>
              <t-input :value="getAddr('rtmp')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" @click="copy(getAddr('rtmp'), $event)">复制</t-button>
            </div>
            <div class="addr-item">
              <span class="label">RTSP 拉流</span>
              <t-input :value="getAddr('rtsp')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" @click="copy(getAddr('rtsp'), $event)">复制</t-button>
            </div>
            <div class="addr-item">
              <span class="label">WebRTC 播放</span>
              <t-input :value="getAddr('webrtc')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" @click="copy(getAddr('webrtc'), $event)">复制</t-button>
            </div>
          </t-tab-panel>

          <t-tab-panel value="push" label="推流地址 (给 OBS 等工具推流)">
            <t-alert theme="warning" :message="'如果当前流是由前端浏览器推送的，请先停止浏览器推流，否则 OBS 会因 Stream ID 冲突而推流失败。'" style="margin-bottom: 12px" />
            <div class="addr-item">
              <span class="label">RTMP 推流</span>
              <t-input :value="getAddr('rtmp')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" @click="copy(getAddr('rtmp'), $event)">复制</t-button>
            </div>
            <div class="addr-item">
              <span class="label">RTSP 推流</span>
              <t-input :value="getAddr('rtsp')" readonly @focus="selectAllOnFocus" />
              <t-button size="small" @click="copy(getAddr('rtsp'), $event)">复制</t-button>
            </div>
          </t-tab-panel>
        </t-tabs>
      </div>
    </div>

    <template #footer>
      <t-space>
        <t-button @click="visible = false">关闭</t-button>
        <t-button v-if="manualMode && currentUrl" @click="resetManual">重新输入</t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import Hls from 'hls.js';
import mpegts from 'mpegts.js';
import { mediaApi, withToken } from '@/api/media';
import type { LiveStream, LiveRoom } from '@/types';

const props = withDefaults(
  defineProps<{
    visible: boolean;
    stream?: LiveStream | LiveRoom | null;
    streamId?: string;
    manualMode?: boolean;
  }>(),
  { stream: null, streamId: '', manualMode: false },
);

const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>();

const visible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

type Format = 'hls' | 'flv' | 'webrtc';
const format = ref<Format>('webrtc');
const currentUrl = ref('');
const webrtcUrl = ref('');
const manualInput = ref('');
const resolving = ref(false);
const errorMsg = ref('');

const videoElHls = ref<HTMLVideoElement | null>(null);
const videoElFlv = ref<HTMLVideoElement | null>(null);
const videoElWebrtc = ref<HTMLVideoElement | null>(null);
const bodyRef = ref<HTMLElement | null>(null);

let hls: Hls | null = null;
let flvPlayer: mpegts.Player | null = null;
let webrtcPc: RTCPeerConnection | null = null;
let switchSeq = 0;

const modalTitle = computed(() => {
  if (props.manualMode) return '统一拉流控制台';
  return `拉流观看与地址分发 - ${props.stream?.name || props.streamId || ''}`;
});

const currentStreamId = computed(() => props.streamId || props.stream?.stream_id || '');

function getAddr(type: Format | 'rtmp' | 'rtsp'): string {
  const sid = currentStreamId.value;
  if (!sid) return '';

  const host = window.location.hostname;

  const playUrls = props.stream?.play_urls as any;
  if (playUrls && playUrls[type]) return playUrls[type];

  switch (type) {
    case 'rtmp':
      return `rtmp://${host}:1935/live/${sid}`;
    case 'rtsp':
      return `rtsp://${host}:554/live/${sid}`;
    case 'flv':
      return `http://${host}/live/${sid}.live.flv`;
    case 'hls':
      return `http://${host}/live/${sid}/hls.m3u8`;
    case 'webrtc':
      return `https://${host}:8443/index/api/webrtc?app=live&stream=${sid}&type=play`;
    default:
      return '';
  }
}

function cleanup() {
  switchSeq++;
  if (hls) {
    hls.destroy();
    hls = null;
  }
  if (flvPlayer) {
    flvPlayer.pause();
    flvPlayer.unload();
    flvPlayer.detachMediaElement();
    flvPlayer.destroy();
    flvPlayer = null;
  }
  if (webrtcPc) {
    try {
      webrtcPc.close();
    } catch (_) {
      /* noop */
    }
    webrtcPc = null;
  }
  videoElHls.value?.removeAttribute('src');
  videoElFlv.value?.removeAttribute('src');
  if (videoElWebrtc.value) {
    videoElWebrtc.value.srcObject = null;
    videoElWebrtc.value.removeAttribute('src');
  }
  webrtcUrl.value = '';
  errorMsg.value = '';
}

function playHls(url: string) {
  if (!videoElHls.value || !url) return;
  if (Hls.isSupported()) {
    hls = new Hls({ enableWorker: true, lowLatencyMode: true });
    hls.loadSource(url);
    hls.attachMedia(videoElHls.value);
    hls.on(Hls.Events.ERROR, (_e, data) => {
      if (data.fatal) errorMsg.value = 'HLS 播放失败，请检查流状态或稍后重试';
    });
  } else if (videoElHls.value.canPlayType('application/vnd.apple.mpegurl')) {
    videoElHls.value.src = url;
  } else {
    errorMsg.value = '浏览器不支持 HLS';
  }
}

function playFlv(url: string) {
  if (!videoElFlv.value || !url) return;
  if (mpegts.isSupported()) {
    flvPlayer = mpegts.createPlayer({ type: 'flv', isLive: true, url });
    flvPlayer.attachMediaElement(videoElFlv.value);
    flvPlayer.load();
    (flvPlayer.play() as any)?.catch?.(() => {});
    flvPlayer.on(mpegts.Events.ERROR, () => {
      errorMsg.value = 'FLV 播放失败，可能是音频编码不支持，请尝试切换 HLS 格式';
    });
  } else {
    errorMsg.value = '浏览器不支持 FLV';
  }
}

async function playWebrtc(url: string) {
  const v = videoElWebrtc.value;
  if (!v) return;
  if (typeof RTCPeerConnection === 'undefined') {
    errorMsg.value = '浏览器不支持 WebRTC';
    return;
  }
  const sid = currentStreamId.value;
  if (!sid) return;
  const signalUrl =
    url && /index\/api\/webrtc/i.test(url)
      ? url
      : `https://${window.location.hostname}:8443/index/api/webrtc?app=live&stream=${sid}&type=play`;

  const pc = new RTCPeerConnection({ iceServers: [] });
  webrtcPc = pc;
  pc.ontrack = (e) => {
    const [stream] = e.streams;
    if (stream && videoElWebrtc.value) {
      videoElWebrtc.value.srcObject = stream;
      videoElWebrtc.value.play().catch(() => {});
    }
  };
  pc.onconnectionstatechange = () => {
    if (pc.connectionState === 'failed' || pc.connectionState === 'closed') {
      errorMsg.value = 'WebRTC 播放连接失败';
    }
  };
  try {
    pc.addTransceiver('audio', { direction: 'recvonly' });
    pc.addTransceiver('video', { direction: 'recvonly' });
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    const resp = await fetch(signalUrl, { method: 'POST', body: pc.localDescription!.sdp });
    const json = await resp.json();
    if (!json || !json.sdp) {
      throw new Error(json?.msg || 'WebRTC 信令响应异常');
    }
    await pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: json.sdp }));
  } catch (e: any) {
    console.error('[WebRTC播放]', e);
    errorMsg.value = 'WebRTC 播放失败: ' + (e?.message || String(e));
    try {
      pc.close();
    } catch (_) {}
    if (webrtcPc === pc) webrtcPc = null;
  }
}

async function switchFormat(next?: Format | string | number) {
  const target = (next ?? format.value) as Format;
  format.value = target;
  cleanup();
  const seq = switchSeq;

  let url = getAddr(target);

  if (currentStreamId.value && !props.manualMode) {
    try {
      const res = await mediaApi.getPlayUrlOfStream(currentStreamId.value, target as any);
      if (seq !== switchSeq || !props.visible) return;
      if (res.data.url) url = res.data.url;
    } catch (e) {
      console.warn('获取鉴权地址失败，使用默认拼装', e);
    }
  }
  if (seq !== switchSeq || !props.visible) return;

  if (url && !/^(https?|rtmp|rtsp):\/\//i.test(url)) {
    url = `${window.location.origin}${url.startsWith('/') ? url : '/' + url}`;
  }

  currentUrl.value = withToken(url);
  webrtcUrl.value = target === 'webrtc' ? currentUrl.value : '';

  void nextTick(() => {
    if (target === 'flv') playFlv(currentUrl.value);
    else if (target === 'hls') playHls(currentUrl.value);
    else if (target === 'webrtc') playWebrtc(currentUrl.value);
  });
}

function resetManual() {
  cleanup();
  currentUrl.value = '';
  webrtcUrl.value = '';
}

async function playManual() {
  const input = manualInput.value.trim();
  if (!input) return;
  errorMsg.value = '';

  if (/^https?:\/\//i.test(input)) {
    format.value = input.includes('.m3u8')
      ? 'hls'
      : /webrtc|type=play/i.test(input)
        ? 'webrtc'
        : 'flv';
    currentUrl.value = input;
    webrtcUrl.value = format.value === 'webrtc' ? input : '';
    void nextTick(() => {
      if (format.value === 'flv') playFlv(currentUrl.value);
      else if (format.value === 'hls') playHls(currentUrl.value);
      else if (format.value === 'webrtc') playWebrtc(currentUrl.value);
    });
  } else {
    resolving.value = true;
    try {
      await switchFormat();
      if (!currentUrl.value) errorMsg.value = '该流暂无播放地址';
    } finally {
      resolving.value = false;
    }
  }
}

watch(
  () => props.visible,
  async (v) => {
    if (!v) {
      cleanup();
      return;
    }
    errorMsg.value = '';
    currentUrl.value = '';
    webrtcUrl.value = '';
    manualInput.value = '';
    if (!props.manualMode) {
      format.value = 'flv';
      await switchFormat();
    }
  },
);

function onAfterLeave() {
  cleanup();
}
onBeforeUnmount(cleanup);

function selectAllOnFocus(e: FocusEvent) {
  const el = e.target as HTMLInputElement | null;
  el?.select?.();
}

function legacyCopy(text: string, btn?: HTMLElement | null): boolean {
  const input = btn?.closest('.addr-item')?.querySelector<HTMLInputElement>('input');
  if (input && input.value === text) {
    input.focus();
    input.select();
    input.setSelectionRange(0, input.value.length);
    if (document.activeElement === input) {
      try {
        if (document.execCommand('copy')) return true;
      } catch {
        // fallthrough
      }
    }
  }

  const host =
    bodyRef.value?.parentElement ??
    document.querySelector<HTMLElement>('.t-dialog__body') ??
    document.body;
  const ta = document.createElement('textarea');
  ta.value = text;
  ta.setAttribute('readonly', '');
  ta.style.position = 'fixed';
  ta.style.top = '0';
  ta.style.left = '0';
  ta.style.opacity = '0';
  ta.style.pointerEvents = 'none';
  host.appendChild(ta);

  ta.focus();
  ta.select();
  ta.setSelectionRange(0, text.length);

  let ok = false;
  if (document.activeElement === ta) {
    try {
      ok = document.execCommand('copy');
    } catch {
      ok = false;
    }
  }
  host.removeChild(ta);
  return ok;
}

async function copy(text: string, event?: MouseEvent) {
  const btn = (event?.currentTarget as HTMLElement | null) ?? null;

  if (!text) {
    MessagePlugin.warning('暂无可复制的地址');
    return;
  }

  let ok = false;
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      ok = true;
    }
  } catch {
    ok = false;
  }

  if (!ok) ok = legacyCopy(text, btn);

  if (ok) {
    MessagePlugin.success('已复制到剪贴板');
  } else {
    MessagePlugin.error('复制失败，请点击地址输入框全选后手动复制');
  }
}
</script>

<style scoped>
.body {
  padding: 8px 0;
}
.manual-row {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.player-section {
  margin-bottom: 16px;
}
.format-switch {
  margin-bottom: 12px;
}
.video-container {
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
  border: none;
}
.address-panel {
  background: #fafafa;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #eee;
}
.addr-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}
.addr-item .label {
  width: 120px;
  font-size: 13px;
  color: #666;
  text-align: right;
  flex-shrink: 0;
}
</style>