/* ============================================================================
 * media 子应用 API
 * 路径：frontend/sub-apps/media/src/api/media.ts
 * ==========================================================================*/
import request from './index';
import type {
  EngineStatus,
  InitUploadResp,
  LiveStream,
  MediaInfo,
  LiveRoom,
} from '@/types';

/* ---------- 媒体库列表查询参数 ---------- */
export interface MediaListQuery {
  page?: number;
  page_size?: number;
  type?: 'video' | 'audio';
  keyword?: string;
}

export const mediaApi = {
  /* ========================================================================
   * 1. 分片上传 (媒体库)
   * ======================================================================*/
  initUpload(payload: {
    file_name: string;
    file_size: number;
    file_hash: string;
    chunk_size: number;
    total_chunks: number;
    media_type: 'video' | 'audio';
    description?: string;
  }) {
    return request.post<InitUploadResp>('/media/upload/init', payload);
  },

  uploadChunk(uploadId: string, chunkIndex: number, chunk: Blob) {
    const fd = new FormData();
    fd.append('upload_id', uploadId);
    fd.append('chunk_index', String(chunkIndex));
    fd.append('chunk_file', chunk, `chunk_${chunkIndex}`);
    return request.post<{ received: number }>('/media/upload/chunk', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 120000,
    });
  },

  completeUpload(uploadId: string) {
    return request.post<{ media: MediaInfo }>('/media/upload/complete', {
      upload_id: uploadId,
    });
  },

  /* ========================================================================
   * 2. 媒体库管理
   * ======================================================================*/
  list(query: MediaListQuery = {}) {
    return request.get<{ total: number; list: MediaInfo[] }>('/media/list', {
      params: query,
    });
  },

  detail(id: string) {
    return request.get<MediaInfo>(`/media/${id}`);
  },

  remove(id: string) {
    return request.delete(`/media/${id}`);
  },

  getPlayUrl(id: string) {
    return request.get<{ url: string; type: string }>(`/media/${id}/play`);
  },

  /* ========================================================================
   * 3. 直播间管理 (核心：持久化记录)
   * ======================================================================*/
  /** 创建直播间 */
  createRoom(payload: {
    name: string;
    push_type: 'webrtc' | 'obs' | 'proxy';
    source_url?: string;
  }) {
    return request.post<LiveRoom>('/live/rooms', payload);
  },

  /** 获取直播间列表（包含实时 is_pushing 状态和 play_urls） */
  getRooms() {
    return request.get<{ list: LiveRoom[] }>('/live/rooms');
  },

  /** 删除直播间 */
  deleteRoom(id: string) {
    return request.delete(`/live/rooms/${id}`);
  },

  /* ========================================================================
   * 4. WebRTC 信令推流
   * ======================================================================*/
  /** 将浏览器的 SDP Offer 发给后端换取 ZLM 的 SDP Answer */
  pushWebRTC(streamId: string, sdp: string) {
    return request.post<{ sdp: string; type: string }>('/live/push/webrtc', {
      stream_id: streamId,
      sdp,
    });
  },

  /** 主动停止推流 (通常用于 OBS 等外部推流，前端 WebRTC 直接 close 即可) */
  stopPush(streamId: string) {
    return request.post('/live/push/stop', { stream_id: streamId });
  },

  /* ========================================================================
   * 5. 外部流拉取代理
   * ======================================================================*/
  /** 告诉后端去拉取一个外部流，转换成本地流 */
  addStreamProxy(payload: { url: string; name?: string }) {
    return request.post<{ stream_id: string }>('/live/proxy/start', payload);
  },

  /** 停止并删除拉流代理 */
  removeStreamProxy(streamId: string) {
    return request.delete(`/live/proxy/${streamId}`);
  },

  /* ========================================================================
   * 6. 实时流与拉流地址 (ZLM 层级)
   * ======================================================================*/
  /** 获取当前 ZLM 中所有活跃的裸流（通常用于调试，不对外展示） */
  listStreams() {
    return request.get<{ list: LiveStream[] }>('/live/streams');
  },

  /** 按 stream_id 换任意格式的播放地址 */
  getPlayUrlOfStream(
    streamId: string,
    format: 'hls' | 'flv' | 'rtmp' | 'webrtc',
  ) {
    return request.get<{ url: string; type: string }>(
      `/live/play/${streamId}`,
      { params: { format } },
    );
  },

  /* ========================================================================
   * 7. 直播引擎 (ZLMediaKit 进程管理)
   * ======================================================================*/
  getEngineStatus() {
    return request.get<EngineStatus>('/live/engine/status');
  },

  startEngine() {
    return request.post<EngineStatus>('/live/engine/start');
  },

  stopEngine() {
    return request.post<EngineStatus>('/live/engine/stop');
  },
};

/* ============================================================================
 * 直链辅助工具
 * ==========================================================================*/

/**
 * 给需要鉴权的直链 URL 附加 token 查询参数。
 * <video>/<audio>/<img>/<a> 无法带 Authorization 头，必须走 query。
 */
export function withToken(url: string | undefined | null): string {
  if (!url) return '';
  // 如果已经带了 token，直接返回
  if (/(?:^|[?&])token=/.test(url)) return url;
  
  const token = localStorage.getItem('token') || '';
  if (!token) return url;
  
  const sep = url.includes('?') ? '&' : '?';
  return `${url}${sep}token=${encodeURIComponent(token)}`;
}

/**
 * 触发浏览器下载媒体文件。
 */
export function downloadMedia(media: MediaInfo) {
  const base = import.meta.env.VITE_API_BASE_URL || '';
  const url = `${base}/api/v1/media/${media.id}/download`;
  const href = withToken(url);

  const a = document.createElement('a');
  a.href = href;
  a.download = media.name;
  a.target = '_blank';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}
