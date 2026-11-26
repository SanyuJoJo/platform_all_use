<template>
  <t-dialog
    v-model:visible="visible"
    :header="'新建直播间'"
    :confirm-btn="null"
    :cancel-btn="null"
    :width="680"
    :mask-closable="false"
    @close="close"
  >
    <div class="body">
      <t-form label-align="left" :label-width="100">
        <t-form-item label="直播间名称">
          <t-input v-model="name" placeholder="如：会议室直播" />
        </t-form-item>

        <t-form-item label="推流方式">
          <div class="source-type-selector">
            <div
              v-for="opt in sourceOptions"
              :key="opt.value"
              class="source-card"
              :class="{ active: sourceType === opt.value }"
              @click="sourceType = opt.value"
            >
              <div class="source-icon">{{ opt.icon }}</div>
              <div class="source-label">{{ opt.label }}</div>
              <div class="source-desc">{{ opt.desc }}</div>
            </div>
          </div>
        </t-form-item>

        <template v-if="sourceType === 'camera'">
          <div class="media-preview">
            <video ref="previewRef" autoplay muted playsinline class="preview-video" />
            <div v-if="!previewActive" class="preview-mask">
              <t-button theme="primary" @click="startCameraPreview">开启摄像头</t-button>
            </div>
          </div>
        </template>

        <template v-if="sourceType === 'file'">
          <div class="file-upload-wrap">
            <t-upload
              v-if="!localFile"
              :auto-upload="false"
              :show-upload-list="false"
              accept="video/*"
              :files="uploadFiles"
              @change="onFileChange"
            >
              <t-button variant="outline" theme="primary">
                <t-icon name="upload" />
                选择本地视频文件
              </t-button>
            </t-upload>
            <div v-else>
              <div class="media-preview">
                <video ref="previewRef" autoplay muted playsinline class="preview-video" :src="localFileUrl" />
                <div class="file-reset">
                  <t-button size="small" @click="resetFile">重新选择</t-button>
                </div>
              </div>
              <div class="file-meta-row">
                <span class="file-name">{{ localFile.name }}</span>
                <span class="file-size">{{ formatFileSize(localFile.size) }}</span>
              </div>
            </div>
          </div>
        </template>

        <template v-if="sourceType === 'proxy'">
          <t-form-item label="外部流地址">
            <t-input v-model="proxyUrl" placeholder="rtmp://... 或 http://...m3u8" />
          </t-form-item>
        </template>
      </t-form>
    </div>
    <template #footer>
      <t-space>
        <t-button @click="visible = false">取消</t-button>
        <t-button theme="primary" :loading="loading" :disabled="!canSubmit" @click="handleSubmit">
          创建并推流
        </t-button>
      </t-space>
    </template>
  </t-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { MessagePlugin, type UploadFile } from 'tdesign-vue-next';
import { mediaApi } from '@/api/media';
import { livePusher } from '@/utils/media/live-pusher';
import { formatFileSize } from '@/utils/media/format';

const props = defineProps<{
  visible: boolean;
  defaultType?: 'camera' | 'proxy';
}>();

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
  (e: 'created', streamId: string): void;
}>();

const visible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v),
});

const name = ref('直播间-' + Date.now());
const sourceType = ref<'camera' | 'file' | 'proxy' | 'obs'>('camera');
const loading = ref(false);
const proxyUrl = ref('');
const uploadFiles = ref<UploadFile[]>([]);

watch(
  () => props.visible,
  (v) => {
    if (v && props.defaultType) {
      sourceType.value = props.defaultType;
      name.value = (props.defaultType === 'proxy' ? '代理流-' : '直播间-') + Date.now();
    }
  },
);

const sourceOptions = [
  { value: 'camera', label: '本机摄像头', icon: '📹', desc: '本地推流' },
  { value: 'file', label: '本地视频', icon: '📁', desc: '推流文件' },
  { value: 'obs', label: 'OBS 工具', icon: '🖥️', desc: '生成地址' },
  { value: 'proxy', label: '拉取外流', icon: '🔗', desc: '代理转播' },
] as const;

const previewRef = ref<HTMLVideoElement | null>(null);
const previewActive = ref(false);
const localFile = ref<File | null>(null);
const localFileUrl = ref('');
let localStream: MediaStream | null = null;

const canSubmit = computed(() => {
  if (!name.value.trim()) return false;
  if (sourceType.value === 'camera') return previewActive.value;
  if (sourceType.value === 'file') return !!localFile.value;
  if (sourceType.value === 'proxy') return !!proxyUrl.value.trim();
  return true;
});

async function startCameraPreview() {
  if (!navigator.mediaDevices?.getUserMedia) {
    MessagePlugin.error('当前环境禁用摄像头');
    return;
  }
  try {
    localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
    if (previewRef.value) previewRef.value.srcObject = localStream;
    previewActive.value = true;
  } catch (e: any) {
    MessagePlugin.error('无法获取摄像头：' + e.message);
  }
}

function onFileChange(files: UploadFile[]) {
  const f = files?.[0]?.raw as File;
  if (!f) return;
  localFile.value = f;
  uploadFiles.value = files;
  localFileUrl.value = URL.createObjectURL(f);
  nextTick(() => {
    if (previewRef.value) previewRef.value.play();
    previewActive.value = true;
  });
}

function resetFile() {
  if (localFileUrl.value) URL.revokeObjectURL(localFileUrl.value);
  localFile.value = null;
  localFileUrl.value = '';
  uploadFiles.value = [];
  previewActive.value = false;
}

async function handleSubmit() {
  if (loading.value || !canSubmit.value) return;
  loading.value = true;
  try {
    const pushType =
      sourceType.value === 'camera' || sourceType.value === 'file' ? 'webrtc' : sourceType.value;
    const res = await mediaApi.createRoom({
      name: name.value,
      push_type: pushType as any,
      source_url: pushType === 'proxy' ? proxyUrl.value.trim() : undefined,
    });
    const room = res.data;

    if (sourceType.value === 'camera' || sourceType.value === 'file') {
      let mediaStream: MediaStream;
      let hiddenVideo: HTMLVideoElement | undefined;

      if (sourceType.value === 'camera') {
        mediaStream = localStream!;
      } else {
        hiddenVideo = document.createElement('video');
        hiddenVideo.src = localFileUrl.value;
        hiddenVideo.muted = false;
        hiddenVideo.volume = 1;
        hiddenVideo.loop = true;
        await hiddenVideo.play();
        const hv = hiddenVideo!;
        await new Promise<void>((resolve) => {
          const onReady = () => {
            hv.removeEventListener('loadedmetadata', onReady);
            hv.removeEventListener('canplay', onReady);
            resolve();
          };
          if (hv.readyState >= 1) {
            resolve();
            return;
          }
          hv.addEventListener('loadedmetadata', onReady);
          hv.addEventListener('canplay', onReady);
        });
        await new Promise((r) => setTimeout(r, 300));
        mediaStream = (hiddenVideo as any).captureStream
          ? (hiddenVideo as any).captureStream()
          : (hiddenVideo as any).mozCaptureStream();
      }

      await livePusher.startPush(room.stream_id, room.name, sourceType.value, mediaStream, hiddenVideo);
      MessagePlugin.success('推流已启动');
    } else if (sourceType.value === 'proxy') {
      MessagePlugin.success('代理拉流已启动');
    } else {
      MessagePlugin.success('创建成功');
    }

    emit('created', room.stream_id);
    close();
  } catch (e: any) {
    MessagePlugin.error(e?.message || '操作失败');
  } finally {
    loading.value = false;
  }
}

function close() {
  visible.value = false;
  setTimeout(() => {
    if (localStream && livePusher.activePushes.size === 0) {
      localStream.getTracks().forEach((t) => t.stop());
      localStream = null;
    }
    resetFile();
    name.value = '直播间-' + Date.now();
    sourceType.value = 'camera';
  }, 300);
}
</script>

<style scoped>
.body {
  padding: 8px 0;
}
.source-type-selector {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  width: 100%;
}
.source-card {
  border: 1px solid #e0e0e6;
  border-radius: 8px;
  padding: 12px 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fafafa;
}
.source-card:hover {
  border-color: #36ad6a;
}
.source-card.active {
  border-color: #18a058;
  background: #e8f5ee;
  box-shadow: 0 0 0 1px #18a058 inset;
}
.source-icon {
  font-size: 24px;
  margin-bottom: 6px;
}
.source-label {
  font-weight: 600;
  font-size: 13px;
  color: #333;
}
.source-desc {
  font-size: 11px;
  color: #999;
  margin-top: 4px;
}
.media-preview {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000;
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 12px;
}
.preview-video {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.preview-mask {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
}
.file-reset {
  position: absolute;
  top: 8px;
  right: 8px;
}
.file-upload-wrap {
  width: 100%;
}
.file-upload-wrap .t-upload {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 28px 0;
  border: 1px dashed #c0c4cc;
  border-radius: 8px;
  background: #fafafa;
  transition: border-color 0.2s;
}
.file-upload-wrap .t-upload:hover {
  border-color: #0052d9;
}
.file-meta-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}
.file-meta-row .file-name {
  font-weight: 600;
  font-size: 13px;
  color: #333;
  word-break: break-all;
  flex: 1;
  min-width: 0;
}
.file-meta-row .file-size {
  font-size: 12px;
  color: #999;
  flex-shrink: 0;
}
</style>