/* ============================================================================
 * media 子应用类型定义
 * ==========================================================================*/

/* ========== 通用响应体 ========== */
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
  requestId: string;
}

/* ========== 动态路由 ========== */
export interface RouteMeta {
  title: string;
  icon?: string;
  permission?: string;
  hidden?: boolean;
  keepAlive?: boolean;
  [key: string]: unknown;
  [key: symbol]: unknown;
}
export interface RouteNode {
  id: string;
  parentId: string | null;
  name: string;
  path?: string;
  component?: string;
  microApp: string;
  type: 'directory' | 'menu';
  meta: RouteMeta;
  order?: number;
  children?: RouteNode[];
}

/* ========== 媒体元数据 ========== */
export type MediaType = 'video' | 'audio';
export type MediaStatus = 'processing' | 'ready' | 'error';

export interface MediaInfo {
  id: string;
  name: string;
  type: MediaType;
  format: string;
  size: number;
  duration: number;
  resolution: string;
  bitrate: number;
  thumbnail: string;
  play_url: string;
  download_url: string;
  status: MediaStatus;
  created_at: string;
  updated_at: string;
}

/* ========== 分片上传 ========== */
export interface InitUploadResp {
  upload_id: string;
  exist: MediaInfo | null;
}

/* ========== 播放地址集 ========== */
export interface PlayURLs {
  hls: string;
  flv: string;
  rtmp: string;
  webrtc: string;
}

/* ========== 直播流（ZLM 实时状态，内部使用） ========== */
export interface LiveStream {
  stream_id: string;
  name: string;
  app: string;
  is_pushing: boolean;
  viewers: number;
  bitrate: number;
  created_at: string;
  play_urls: PlayURLs;
}

/* ========== 直播间（持久化记录） ========== */
export interface LiveRoom {
  id: string;
  stream_id: string;
  name: string;
  push_type: 'webrtc' | 'obs' | 'proxy';
  is_pushing: boolean;
  play_urls?: PlayURLs;
  created_at?: string;
}

/* ========== 直播引擎 ========== */
export interface EngineStatus {
  running: boolean;
  managed: boolean;
  pid?: number;
  base_url: string;
  last_error?: string;
  auto_start: boolean;
  shutdown_on_exit: boolean;
}
