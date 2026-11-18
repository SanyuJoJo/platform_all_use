<template>
  <n-modal
    v-model:show="visible"
    preset="dialog"
    title="上传媒体"
    :z-index="3000"
    :auto-focus="false"
    :mask-closable="true"
    :close-on-esc="true"
    :display-directive="'show'"
    style="width: 560px;"
  >
    <div class="upload-body">
      <n-upload
        :max="1"
        :show-file-list="false"
        :default-upload="false"
        accept="video/*,audio/*"
        @change="onFileChange"
      >
        <n-upload-dragger>
          <div style="font-size: 36px;">📁</div>
          <p>点击或拖拽音视频文件到此处</p>
          <p style="font-size: 12px; color: #999;">
            支持 mp4 / mkv / mov / mp3 / wav 等
          </p>
        </n-upload-dragger>
      </n-upload>

      <div v-if="file" class="file-info">
        <div class="file-name">{{ file.name }}</div>
        <div class="file-meta">
          <span>{{ formatFileSize(file.size) }}</span>
          <span>{{ mediaTypeLabel }}</span>
        </div>
      </div>

      <div v-if="uploading" class="progress-wrap">
        <n-progress
          type="line"
          :percentage="progress"
          :height="14"
          :border-radius="4"
        />
        <div class="progress-text">
          {{ currentChunk }} / {{ totalChunks }} 分片（{{ progress }}%）
        </div>
      </div>
    </div>

    <template #action>
      <n-button @click="visible = false">取消</n-button>
      <n-button
        type="primary"
        :loading="uploading"
        :disabled="!file"
        @click="startUpload"
      >
        {{ uploading ? '上传中...' : '开始上传' }}
      </n-button>
    </template>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  NModal,
  NUpload,
  NUploadDragger,
  NProgress,
  NButton,
  useMessage,
} from 'naive-ui';
import type { UploadFileInfo } from 'naive-ui';
import { mediaApi } from '@/api/media';
import {
  computeFileHash,
  detectMediaType,
  formatFileSize,
} from '@/utils/format';

const props = defineProps<{ visible: boolean }>();
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void;
  (e: 'success'): void;
}>();

const message = useMessage();

const visible = computed({
  get: () => props.visible,
  set: (v) => {
    console.log('[UploadModal] visible 变化 →', v);
    emit('update:visible', v);
  },
});

watch(
  () => props.visible,
  (v) => console.log('[UploadModal] props.visible =', v),
  { immediate: true },
);

const file = ref<File | null>(null);
const uploading = ref(false);
const progress = ref(0);
const currentChunk = ref(0);
const totalChunks = ref(0);
const CHUNK_SIZE = 5 * 1024 * 1024;

const mediaTypeLabel = computed(() => {
  if (!file.value) return '';
  return detectMediaType(file.value.type, file.value.name) === 'audio'
    ? '音频'
    : '视频';
});

function onFileChange(data: {
  file: UploadFileInfo;
  fileList: UploadFileInfo[];
}) {
  console.log('[UploadModal] onFileChange:', data.fileList);
  const info = data.fileList[0];
  const f = (info?.file ?? null) as File | null;
  file.value = f;
  progress.value = 0;
  currentChunk.value = 0;
  totalChunks.value = 0;
  if (!f) {
    message.warning('未获取到文件，请重新选择');
  }
}

async function startUpload() {
  const f = file.value;
  if (!f) {
    message.warning('请先选择文件');
    return;
  }
  uploading.value = true;
  try {
    const fileHash = await computeFileHash(f);
    const total = Math.ceil(f.size / CHUNK_SIZE);
    totalChunks.value = total;

    const initRes = await mediaApi.initUpload({
      file_name: f.name,
      file_size: f.size,
      file_hash: fileHash,
      chunk_size: CHUNK_SIZE,
      total_chunks: total,
      media_type: detectMediaType(f.type, f.name),
    });

    if (initRes.data.exist) {
      message.success('秒传成功');
      emit('success');
      visible.value = false;
      return;
    }

    const uploadId = initRes.data.upload_id;
    for (let i = 0; i < total; i++) {
      const chunk = f.slice(
        i * CHUNK_SIZE,
        Math.min((i + 1) * CHUNK_SIZE, f.size),
      );
      await mediaApi.uploadChunk(uploadId, i, chunk);
      currentChunk.value = i + 1;
      progress.value = Math.round(((i + 1) / total) * 100);
    }

    await mediaApi.completeUpload(uploadId);
    message.success('上传成功');
    emit('success');
    visible.value = false;
  } catch (e: any) {
    console.error('[UploadModal] 上传失败:', e);
    message.error(e?.message || '上传失败');
  } finally {
    uploading.value = false;
    currentChunk.value = 0;
    totalChunks.value = 0;
  }
}
</script>

<style scoped>
.upload-body {
  padding: 8px 0;
}
.file-info {
  margin-top: 16px;
  padding: 12px;
  background: #f7f7f7;
  border-radius: 6px;
}
.file-name {
  font-weight: 600;
  margin-bottom: 4px;
  word-break: break-all;
}
.file-meta {
  font-size: 12px;
  color: #666;
  display: flex;
  gap: 12px;
}
.progress-wrap {
  margin-top: 16px;
}
.progress-text {
  margin-top: 8px;
  font-size: 12px;
  color: #666;
  text-align: center;
}
</style>
