<template>
  <n-modal
    v-model:show="visible"
    preset="dialog"
    title="新建直播间"
    :z-index="3000"
    :auto-focus="false"
    :display-directive="'show'"
    :mask-closable="false"
    style="width: 680px;"
  >
    <div class="body">
      <n-form label-placement="left" label-width="100">
        <n-form-item label="直播间名称">
          <n-input v-model:value="name" placeholder="如：会议室直播" />
        </n-form-item>

        <n-form-item label="推流方式" :show-feedback="false">
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
        </n-form-item>

        <template v-if="sourceType === 'camera'">
          <div class="media-preview">
            <video ref="previewRef" autoplay muted playsinline class="preview-video" />
            <div v-if="!previewActive" class="preview-mask">
              <n-button type="primary" @click="startCameraPreview">开启摄像头</n-button>
            </div>
          </div>
        </template>

        <template v-if="sourceType === 'file'">
          <n-upload v-if="!localFile" :max="1" :show-file-list="false" accept="video/*" :default-upload="false" @change="onFileChange">
            <n-upload-dragger>
              <div style="font-size: 32px;">📁</div>
              <p>点击选择本地视频文件</p>
            </n-upload-dragger>
          </n-upload>
          <div v-else class="media-preview">
            <video ref="previewRef" autoplay muted playsinline class="preview-video" :src="localFileUrl" />
            <div class="file-reset"><n-button size="small" @click="resetFile">重新选择</n-button></div>
          </div>
        </template>

        <template v-if="sourceType === 'proxy'">
          <n-form-item label="外部流地址">
            <n-input v-model:value="proxyUrl" placeholder="rtmp://... 或 http://...m3u8" />
          </n-form-item>
        </template>
      </n-form>
    </div>
    <template #action>
      <n-button @click="visible = false">取消</n-button>
      <n-button type="primary" :loading="loading" :disabled="!canSubmit" @click="handleSubmit">
        创建并推流
      </n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import {
  NModal, NForm, NFormItem, NInput, NButton, NUpload, NUploadDragger, useMessage,
} from 'naive-ui';
import type { UploadFileInfo } from 'naive-ui';
import { mediaApi } from '@/api/media';
import { livePusher } from '@/utils/live-pusher';

const props = defineProps<{ 
  visible: boolean;
  defaultType?: 'camera' | 'proxy';
}>();

const emit = defineEmits<{ 
  (e: 'update:visible', v: boolean): void; 
  (e: 'created', streamId: string): void; 
}>();

const message = useMessage();
const visible = computed({ get: () => props.visible, set: (v) => emit('update:visible', v) });

const name = ref('直播间-' + Date.now());
const sourceType = ref<'camera' | 'file' | 'proxy' | 'obs'>('camera');
const loading = ref(false);
const proxyUrl = ref('');

// 监听外部传来的默认类型
watch(() => props.visible, (v) => {
  if (v && props.defaultType) {
    sourceType.value = props.defaultType;
    name.value = (props.defaultType === 'proxy' ? '代理流-' : '直播间-') + Date.now();
  }
});

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
    message.error('当前环境禁用摄像头');
    return;
  }
  try {
    localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
    if (previewRef.value) previewRef.value.srcObject = localStream;
    previewActive.value = true;
  } catch (e: any) {
    message.error('无法获取摄像头：' + e.message);
  }
}

function onFileChange(data: { file: UploadFileInfo; fileList: UploadFileInfo[] }) {
  const f = data.fileList[0]?.file as File;
  if (!f) return;
  localFile.value = f;
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
  previewActive.value = false;
}

async function handleSubmit() {
  if (loading.value || !canSubmit.value) return;
  loading.value = true;
  try {
    const pushType = (sourceType.value === 'camera' || sourceType.value === 'file') ? 'webrtc' : sourceType.value;
    // ★ proxy 模式必须传 source_url，后端 CreateRoom 会同步触发 ZLM 拉流代理；
    //   不再需要（也不能）额外调用 addStreamProxy，否则会产生一个未被持久化的孤儿代理流。
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
        hiddenVideo.muted = true;
        hiddenVideo.loop = true;
        await hiddenVideo.play();
        await new Promise(r => setTimeout(r, 500));
        mediaStream = (hiddenVideo as any).captureStream ? (hiddenVideo as any).captureStream() : (hiddenVideo as any).mozCaptureStream();
      }

      await livePusher.startPush(room.stream_id, room.name, sourceType.value, mediaStream, hiddenVideo);
      message.success('推流已启动');
    } else if (sourceType.value === 'proxy') {
      // 代理拉流已由 createRoom 在后端同步触发
      message.success('代理拉流已启动');
    } else {
      message.success('创建成功');
    }
    
    emit('created', room.stream_id);
    close();
  } catch (e: any) {
    message.error(e?.message || '操作失败');
  } finally {
    loading.value = false;
  }
}

function close() {
  visible.value = false;
  setTimeout(() => {
    if (localStream && livePusher.activePushes.size === 0) {
       localStream.getTracks().forEach(t => t.stop());
       localStream = null;
    }
    resetFile();
    name.value = '直播间-' + Date.now();
    sourceType.value = 'camera';
  }, 300);
}
</script>

<style scoped>
.body { padding: 8px 0; }
.source-type-selector { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; width: 100%; }
.source-card { border: 1px solid #e0e0e6; border-radius: 8px; padding: 12px 8px; text-align: center; cursor: pointer; transition: all 0.2s ease; background: #fafafa; }
.source-card:hover { border-color: #36ad6a; }
.source-card.active { border-color: #18a058; background: #e8f5ee; box-shadow: 0 0 0 1px #18a058 inset; }
.source-icon { font-size: 24px; margin-bottom: 6px; }
.source-label { font-weight: 600; font-size: 13px; color: #333; }
.source-desc { font-size: 11px; color: #999; margin-top: 4px; }
.media-preview { position: relative; width: 100%; aspect-ratio: 16/9; background: #000; border-radius: 6px; overflow: hidden; margin-bottom: 12px; }
.preview-video { width: 100%; height: 100%; object-fit: contain; }
.preview-mask { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); }
.file-reset { position: absolute; top: 8px; right: 8px; }
</style>
