<template>
  <div class="page">
    <!-- 工具栏 -->
    <div class="toolbar">
      <n-space>
        <n-input
          v-model:value="keyword"
          placeholder="按名称搜索"
          clearable
          style="width: 240px;"
          @keyup.enter="reload"
        />
        <n-select
          v-model:value="type"
          :options="typeOptions"
          placeholder="全部类型"
          clearable
          style="width: 140px;"
          @update:value="reload"
        />
        <n-button @click="reload">刷新</n-button>
        <n-button type="primary" @click="onOpenUpload">上传媒体</n-button>
      </n-space>
    </div>

    <!-- 媒体网格 -->
    <n-spin :show="loading">
      <div v-if="list.length === 0 && !loading" class="empty">
        <n-empty description="暂无媒体，请先上传" />
      </div>
      <div v-else class="grid">
        <div
          v-for="item in list"
          :key="item.id"
          class="card"
          @click="openPlayer(item)"
        >
          <div class="cover">
            <img
              v-if="item.thumbnail"
              :src="withToken(item.thumbnail)"
              alt=""
              @error="onThumbError"
            />
            <div v-else class="cover-fallback">
              {{ item.type === 'audio' ? '🎵' : '🎬' }}
            </div>
            <div class="duration">{{ formatDuration(item.duration) }}</div>
            <div class="badge" :class="item.type">
              {{ item.type === 'audio' ? '音频' : '视频' }}
            </div>
          </div>
          <div class="card-body">
            <div class="title" :title="item.name">{{ item.name }}</div>
            <div class="meta">
              <span>{{ formatFileSize(item.size) }}</span>
              <span>{{ item.resolution || item.format }}</span>
            </div>
            <div class="actions" @click.stop>
              <n-button size="tiny" @click="openPlayer(item)">播放</n-button>
              <n-button size="tiny" @click="download(item)">下载</n-button>
              <n-button size="tiny" type="error" ghost @click="remove(item)">
                删除
              </n-button>
            </div>
          </div>
        </div>
      </div>
    </n-spin>

    <!-- 分页 -->
    <div class="pager">
      <n-pagination
        v-model:page="page"
        :page-size="pageSize"
        :item-count="total"
        @update:page="reload"
      />
    </div>

    <!-- 上传弹窗 -->
    <UploadModal v-model:visible="showUpload" @success="reload" />

    <!-- 播放弹窗 -->
    <n-modal
      v-model:show="showPlayer"
      preset="card"
      :title="current?.name"
      :display-directive="'show'"
      :auto-focus="false"
      :z-index="3000"
      style="width: 900px;"
      @after-leave="current = null"
    >
      <MediaPlayer
        v-if="current"
        :src="withToken(current.play_url)"
        :poster="withToken(current.thumbnail)"
        :is-audio="current.type === 'audio'"
        autoplay
      />
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import {
  NSpace,
  NInput,
  NSelect,
  NButton,
  NSpin,
  NEmpty,
  NPagination,
  NModal,
  useMessage,
} from 'naive-ui';
import { mediaApi, downloadMedia, withToken } from '@/api/media';
import { formatDuration, formatFileSize } from '@/utils/format';
import UploadModal from '@/components/UploadModal.vue';
import MediaPlayer from '@/components/MediaPlayer.vue';
import type { MediaInfo } from '@/types';

const message = useMessage();

/* ---------- 状态 ---------- */
const list = ref<MediaInfo[]>([]);
const loading = ref(false);
const keyword = ref('');
const type = ref<'video' | 'audio' | ''>('');
const page = ref(1);
const pageSize = ref(12);
const total = ref(0);

/* ---------- 弹窗 ---------- */
const showUpload = ref(false);
const showPlayer = ref(false);
const current = ref<MediaInfo | null>(null);

const typeOptions = [
  { label: '视频', value: 'video' },
  { label: '音频', value: 'audio' },
];

/* ---------- 列表 ---------- */
async function reload() {
  loading.value = true;
  try {
    const res = await mediaApi.list({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      type: (type.value || undefined) as any,
    });
    list.value = res.data.list;
    total.value = res.data.total;
  } catch (e: any) {
    message.error(e?.message || '加载失败');
  } finally {
    loading.value = false;
  }
}

/* ---------- 上传 ---------- */
function onOpenUpload() {
  console.log('[media/index] 点击上传按钮，showUpload 变为 true');
  showUpload.value = true;
}

/* ---------- 播放 / 下载 / 删除 ---------- */
function openPlayer(item: MediaInfo) {
  current.value = item;
  showPlayer.value = true;
}

function download(item: MediaInfo) {
  downloadMedia(item);
}

async function remove(item: MediaInfo) {
  try {
    await mediaApi.remove(item.id);
    message.success('已删除');
    reload();
  } catch (e: any) {
    message.error(e?.message || '删除失败');
  }
}

function onThumbError(e: Event) {
  // 缩略图加载失败（比如后端没生成缩略图）→ 静默忽略，隐藏 img
  (e.target as HTMLImageElement).style.display = 'none';
}

onMounted(reload);
</script>

<style scoped>
.page {
  padding: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}
.card {
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s;
  background: #fff;
}
.card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.cover {
  position: relative;
  height: 130px;
  background: #f0f0f0;
  overflow: hidden;
}
.cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.cover-fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 48px;
  opacity: 0.5;
}
.duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 3px;
}
.badge {
  position: absolute;
  top: 6px;
  left: 6px;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 3px;
  color: #fff;
}
.badge.video { background: #18a058; }
.badge.audio { background: #2080f0; }

.card-body {
  padding: 10px;
}
.title {
  font-weight: 600;
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.meta {
  font-size: 12px;
  color: #999;
  margin: 4px 0 8px;
  display: flex;
  gap: 12px;
}
.actions {
  display: flex;
  gap: 6px;
}

.pager {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}
.empty {
  padding: 60px 0;
}
</style>
