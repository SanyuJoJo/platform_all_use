package live

// LiveStream 裸流对象（保留，供 /live/streams 老接口使用）。
type LiveStream struct {
	StreamID  string        `json:"stream_id"`
	Name      string        `json:"name"`
	App       string        `json:"app"`
	IsPushing bool          `json:"is_pushing"`
	Viewers   int           `json:"viewers"`
	Bitrate   int           `json:"bitrate"`
	CreatedAt string        `json:"created_at"`
	PlayURLs  StreamPlayURL `json:"play_urls"`
}

// StreamPlayURL 各协议播放地址。
type StreamPlayURL struct {
	HLS    string `json:"hls"`
	FLV    string `json:"flv"`
	RTMP   string `json:"rtmp"`
	WebRTC string `json:"webrtc"`
}

// PushURLs 推流地址（供 OBS / WebRTC 推流端使用）。
//
// OBS 用 rtmp；前端摄像头推流用 webrtc（不是 http URL，是 SDP 交换的 endpoint，
// 前端应直接拿 PushWebRTC 接口，无需 webRTC URL；此处提供 URL 是为了让
// 前端在 OBS 模式或调试时能看到）
type PushURLs struct {
	RTMP   string `json:"rtmp,omitempty"`
	WebRTC string `json:"webrtc,omitempty"`
}

// ---------------------------------------------------------------------------
// 老接口 schemas（保留兼容）
// ---------------------------------------------------------------------------

// PushStartReq POST /live/push/start 请求体。
type PushStartReq struct {
	Name          string `json:"name" binding:"required,min=1,max=255"`
	ExpireSeconds int    `json:"expire_seconds"`
}

// PushSession POST /live/push/start 响应体。
type PushSession struct {
	StreamID  string `json:"stream_id"`
	RTMPURL   string `json:"rtmp_url"`
	WebRTCURL string `json:"webrtc_url"`
	ExpireAt  string `json:"expire_at"`
}

// PushStopReq POST /live/push/stop 请求体。
type PushStopReq struct {
	StreamID string `json:"stream_id" binding:"required,min=1,max=128"`
}

// PushWebRTCReq POST /live/push/webrtc 请求体。
type PushWebRTCReq struct {
	StreamID string `json:"stream_id" binding:"required,min=1,max=128"`
	SDP      string `json:"sdp" binding:"required,min=10"`
}

// PushWebRTCResp POST /live/push/webrtc 响应体。
type PushWebRTCResp struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

// AddStreamProxyReq POST /live/proxy/add 请求体。
type AddStreamProxyReq struct {
	URL  string `json:"url" binding:"required,min=1,max=2048"`
	Name string `json:"name" binding:"max=255"`
}

// AddStreamProxyResp POST /live/proxy/add 响应体。
type AddStreamProxyResp struct {
	StreamID string `json:"stream_id"`
	Key      string `json:"key"`
}

// RemoveStreamProxyReq POST /live/proxy/remove 请求体。
type RemoveStreamProxyReq struct {
	StreamID string `json:"stream_id" binding:"required,min=1,max=128"`
}

// StreamListResp GET /live/streams 响应体。
type StreamListResp struct {
	List []*LiveStream `json:"list"`
}

// PlayStreamResp GET /live/play/:streamId 响应体。
type PlayStreamResp struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

// ---------------------------------------------------------------------------
// 直播间 schemas（新增）
// ---------------------------------------------------------------------------

// CreateRoomReq POST /live/rooms 请求体。
//
// source_url：仅 push_type=proxy 时必填，其余忽略。
//   - webrtc / obs：后端只生成 stream_id，等待浏览器/OBS 主动推流；
//   - proxy：后端调 ZLM addStreamProxy 主动拉取 source_url。
type CreateRoomReq struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	PushType  string `json:"push_type" binding:"required,oneof=webrtc obs proxy"`
	SourceURL string `json:"source_url" binding:"max=2048"`
}

// Room 直播间响应对象（对齐前端规范）。
type Room struct {
	ID        string        `json:"id"`         // 字符串，符合前端规范 "id": "1"
	StreamID  string        `json:"stream_id"`
	Name      string        `json:"name"`
	PushType  string        `json:"push_type"`
	IsPushing bool          `json:"is_pushing"`
	Viewers   int           `json:"viewers"`     // 额外字段：观看人数（有则返回，无则 0）
	CreatedAt string        `json:"created_at"`
	PlayURLs  StreamPlayURL `json:"play_urls"`
	PushURLs  PushURLs      `json:"push_urls"`   // 额外字段：OBS 模式必用
}

// RoomListResp GET /live/rooms 响应体。
type RoomListResp struct {
	List []*Room `json:"list"`
}
