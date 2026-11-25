<template>
  <n-modal
    v-model:show="visible"
    preset="dialog"
    :title="modalTitle"
    :z-index="3000"
    :auto-focus="false"
    :display-directive="'show'"
    style="width: 900px;"
    @after-leave="onAfterLeave"
  >
    <div ref="bodyRef" class="body">
      <!-- 手动输入模式 -->
      <div v-if="manualMode && !currentUrl" class="manual-row">
        <n-input 
          v-model:value="manualInput" 
          placeholder="输入 stream_id 或完整 URL (如 rtmp://... 或 http://...m3u8)" 
          @keyup.enter="playManual" 
        />
        <n-button type="primary" :loading="resolving" @click="playManual">拉流播放</n-button>
      </div>

      <!-- 错误提示 -->
      <n-alert v-if="errorMsg" type="error" :bordered="false" style="margin-bottom: 12px;">
        {{ errorMsg }}
      </n-alert>

      <!-- 播放区域 -->
      <div v-if="currentUrl && !errorMsg" class="player-section">
        <div class="format-switch">
          <n-radio-group v-model:value="format" @update:value="switchFormat">
            <n-radio-button value="flv">HTTP-FLV (低延迟)</n-radio-button>
            <n-radio-button value="hls">HLS (高兼容)</n-radio-button>
            <n-radio-button value="webrtc">WebRTC (超低延迟)</n-radio-button>
          </n-radio-group>
        </div>

        <div class="video-container">
          <video v-show="format === 'flv'" ref="videoElFlv" controls autoplay playsinline class="video-el" />
          <video v-show="format === 'hls'" ref="videoElHls" controls autoplay playsinline class="video-el" />
          <video v-show="format === 'webrtc'" ref="videoElWebrtc" controls autoplay playsinline class="video-el" />
        </div>
      </div>

      <!-- ============================================== -->
      <!-- 地址分发面板 (如果是真实的流，显示全部协议地址) -->
      <!-- ============================================== -->
      <div v-if="!manualMode && currentStreamId" class="address-panel">
        <n-tabs type="line" animated>
          
          <!-- 拉流地址：用于第三方服务器来拉取我们的流 -->
          <n-tab-pane name="pull" tab="拉流分发地址 (给第三方系统拉取)">
            <n-alert type="info" :bordered="false" style="margin-bottom: 12px;">
              将此地址填入第三方流媒体服务器、CDN 或 VLC 播放器，即可拉取当前流。
            </n-alert>
            
            <div class="addr-item">
              <span class="label">HTTP-FLV</span>
              <n-input :value="getAddr('flv')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" type="primary" @click="copy(getAddr('flv'), $event)">复制</n-button>
            </div>
            <div class="addr-item">
              <span class="label">HLS (m3u8)</span>
              <n-input :value="getAddr('hls')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" @click="copy(getAddr('hls'), $event)">复制</n-button>
            </div>
            <div class="addr-item">
              <span class="label">RTMP 拉流</span>
              <n-input :value="getAddr('rtmp')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" @click="copy(getAddr('rtmp'), $event)">复制</n-button>
            </div>
            <div class="addr-item">
              <span class="label">RTSP 拉流</span>
              <n-input :value="getAddr('rtsp')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" @click="copy(getAddr('rtsp'), $event)">复制</n-button>
            </div>
            <div class="addr-item">
              <span class="label">WebRTC 播放</span>
              <n-input :value="getAddr('webrtc')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" @click="copy(getAddr('webrtc'), $event)">复制</n-button>
            </div>
          </n-tab-pane>
          
          <!-- 推流地址：用于 OBS 等工具往这里推流 -->
          <n-tab-pane name="push" tab="推流地址 (给 OBS 等工具推流)">
            <n-alert type="warning" :bordered="false" style="margin-bottom: 12px;">
              如果当前流是由前端浏览器推送的，请先停止浏览器推流，否则 OBS 会因 Stream ID 冲突而推流失败。
            </n-alert>
            <div class="addr-item">
              <span class="label">RTMP 推流</span>
              <n-input :value="getAddr('rtmp')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" @click="copy(getAddr('rtmp'), $event)">复制</n-button>
            </div>
            <div class="addr-item">
              <span class="label">RTSP 推流</span>
              <n-input :value="getAddr('rtsp')" readonly @focus="selectAllOnFocus" />
              <n-button size="small" @click="copy(getAddr('rtsp'), $event)">复制</n-button>
            </div>
          </n-tab-pane>
        </n-tabs>
      </div>
    </div>
    
    <template #action>
      <n-button @click="visible = false">关闭</n-button>
      <n-button v-if="manualMode && currentUrl" @click="resetManual">重新输入</n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
import {
  NModal, NRadioGroup, NRadioButton, NInput, NButton, NAlert, NTabs, NTabPane, useMessage,
} from 'naive-ui';
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
  { stream: null, streamId: '', manualMode: false }
);

const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>();
const message = useMessage();

const visible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

type Format = 'hls' | 'flv' | 'webrtc';
const format = ref<Format>('webrtc');
const currentUrl = ref('');
// 仅 WebRTC 播放页使用 iframe，避免 FLV/HLS 地址被 iframe 导航后触发浏览器下载
const webrtcUrl = ref('');
const manualInput = ref('');
const resolving = ref(false);
const errorMsg = ref('');

const videoElHls = ref<HTMLVideoElement | null>(null);
const videoElFlv = ref<HTMLVideoElement | null>(null);
const videoElWebrtc = ref<HTMLVideoElement | null>(null);
// 弹窗内容容器：复制降级时临时 textarea 必须放在弹窗内部，否则会被 Modal 焦点陷阱抢走焦点
const bodyRef = ref<HTMLElement | null>(null);

let hls: Hls | null = null;
let flvPlayer: mpegts.Player | null = null;
let webrtcPc: RTCPeerConnection | null = null;
// 用于中止“弹窗已关闭/已切换格式”后仍然返回的异步拉流请求
let switchSeq = 0;

const modalTitle = computed(() => {
  if (props.manualMode) return '统一拉流控制台';
  return `拉流观看与地址分发 - ${props.stream?.name || props.streamId || ''}`;
});

const currentStreamId = computed(() => props.streamId || props.stream?.stream_id || '');

/* ==========================================
 * 地址生成器 (支持获取全协议地址)
 * ========================================== */
function getAddr(type: Format | 'rtmp' | 'rtsp'): string {
  const sid = currentStreamId.value;
  if (!sid) return '';
  
  // 获取当前访问的主机 IP/域名 (假设 ZLM 与前端同域部署或端口对等)
  const host = window.location.hostname;
  
  // 如果后端返回了真实地址，优先用后端的
  const playUrls = props.stream?.play_urls as any;
  if (playUrls && playUrls[type]) return playUrls[type];

  // 兜底拼接方案 (基于 ZLM 默认端口)
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

/* ==========================================
 * 播放核心逻辑
 * ========================================== */
function cleanup() {
  // 使正在进行中的 switchFormat 请求失效，避免弹窗关闭后仍启动播放器
  switchSeq++;
  if (hls) { hls.destroy(); hls = null; }
  if (flvPlayer) {
    flvPlayer.pause(); flvPlayer.unload(); flvPlayer.detachMediaElement(); flvPlayer.destroy(); flvPlayer = null;
  }
  if (webrtcPc) {
    try { webrtcPc.close(); } catch (_) { /* noop */ }
    webrtcPc = null;
  }
  videoElHls.value?.removeAttribute('src');
  videoElFlv.value?.removeAttribute('src');
  if (videoElWebrtc.value) { videoElWebrtc.value.srcObject = null; videoElWebrtc.value.removeAttribute('src'); }
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

/** WebRTC 播放：原生 RTCPeerConnection，信令经 /index/api/webrtc?type=play（同域 nginx 转发到 ZLM） */
async function playWebrtc(url: string) {
  const v = videoElWebrtc.value;
  if (!v) return;
  if (typeof RTCPeerConnection === 'undefined') {
    errorMsg.value = '浏览器不支持 WebRTC';
    return;
  }
  const sid = currentStreamId.value;
  if (!sid) return;
  const signalUrl = url && /index\/api\/webrtc/i.test(url)
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
    // 播放方向：声明接收音频/视频（recvonly），否则 offer 里没有 m= 行，ZLM 会报 media.empty
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
    try { pc.close(); } catch (_) {}
    if (webrtcPc === pc) webrtcPc = null;
  }
}

async function switchFormat(next?: Format | string | number) {
  // 直接使用事件传入的目标格式，避免 v-model 与 @update:value 执行顺序导致的错位
  const target = (next ?? format.value) as Format;
  format.value = target;
  cleanup();
  const seq = switchSeq;

  let url = getAddr(target);

  // 优先请求后端获取带 Token 的鉴权地址
  if (currentStreamId.value && !props.manualMode) {
    try {
      const res = await mediaApi.getPlayUrlOfStream(currentStreamId.value, target as any);
      // 请求期间弹窗被关闭或用户切换了格式，丢弃本次结果
      if (seq !== switchSeq || !props.visible) return;
      if (res.data.url) url = res.data.url;
    } catch (e) {
      console.warn('获取鉴权地址失败，使用默认拼装', e);
    }
  }
  if (seq !== switchSeq || !props.visible) return;

  // 规范化 URL 格式
  if (url && !/^(https?|rtmp|rtsp):\/\//i.test(url)) {
    url = `${window.location.origin}${url.startsWith('/') ? url : '/' + url}`;
  }

  currentUrl.value = withToken(url);
  // WebRTC 也走原生 RTCPeerConnection（不再用 iframe）；FLV/HLS 用各自播放器
  webrtcUrl.value = target === 'webrtc' ? currentUrl.value : '';

  void nextTick(() => {
    if (target === 'flv') playFlv(currentUrl.value);
    else if (target === 'hls') playHls(currentUrl.value);
    else if (target === 'webrtc') playWebrtc(currentUrl.value);
  });
}

/** 手动模式点击“重新输入”：销毁播放器并清空 iframe，避免后台继续请求流地址 */
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
      : /webrtc|type=play/i.test(input) ? 'webrtc' : 'flv';
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
    if (!v) { cleanup(); return; }
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

function onAfterLeave() { cleanup(); }
onBeforeUnmount(cleanup);

/** 点击只读输入框时自动全选，方便手动复制 */
function selectAllOnFocus(e: FocusEvent) {
  const el = e.target as HTMLInputElement | null;
  el?.select?.();
}

/**
 * HTTP 非安全上下文下的复制降级方案，返回是否真正复制成功。
 *
 * 关键点：Naive UI Modal 自带焦点陷阱，如果把临时 textarea 挂到 document.body，
 * 焦点会被立刻拉回弹窗内，导致 execCommand('copy') 复制不到任何内容
 * （表现为只有手动选中的第一个链接能复制，其他复制按钮全部失效）。
 * 因此优先选中当前行自己的只读输入框，兜底 textarea 也必须挂在弹窗内容内。
 */
function legacyCopy(text: string, btn?: HTMLElement | null): boolean {
  // 1) 优先选中当前地址行内的输入框（位于弹窗内部，不会被焦点陷阱抢走焦点）
  const input = btn?.closest('.addr-item')?.querySelector<HTMLInputElement>('input');
  if (input && input.value === text) {
    input.focus();
    input.select();
    input.setSelectionRange(0, input.value.length);
    if (document.activeElement === input) {
      try {
        if (document.execCommand('copy')) return true;
      } catch {
        // 继续走 textarea 兜底
      }
    }
  }

  // 2) 兜底：临时 textarea 放进弹窗内容容器内，而不是 document.body
  const host =
    bodyRef.value?.parentElement ??
    document.querySelector<HTMLElement>('.n-dialog__content') ??
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
  // 确认焦点没有被 Modal 焦点陷阱移走后再执行复制
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
  // 在 await 前同步取出按钮元素（原生的 event.currentTarget 在异步后会被置空）
  const btn = (event?.currentTarget as HTMLElement | null) ?? null;

  if (!text) {
    message.warning('暂无可复制的地址');
    return;
  }

  let ok = false;
  // 非 HTTPS / 非 localhost 环境下 navigator.clipboard 不可用，需要降级处理
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
    message.success('已复制到剪贴板');
  } else {
    message.error('复制失败，请点击地址输入框全选后手动复制');
  }
}
</script>

<style scoped>
.body { padding: 8px 0; }
.manual-row { display: flex; gap: 8px; margin-bottom: 12px; }
.player-section { margin-bottom: 16px; }
.format-switch { margin-bottom: 12px; }
.video-container { position: relative; width: 100%; aspect-ratio: 16/9; background: #000; border-radius: 6px; overflow: hidden; }
.video-el { width: 100%; height: 100%; object-fit: contain; border: none; }

.address-panel { background: #fafafa; padding: 12px; border-radius: 6px; border: 1px solid #eee; }
.addr-item { display: flex; align-items: center; gap: 12px; margin-bottom: 10px; }
.addr-item .label { width: 120px; font-size: 13px; color: #666; text-align: right; flex-shrink: 0; }
</style>
