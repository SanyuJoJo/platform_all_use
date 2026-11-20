package live

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/config"
	"backend-go/internal/exception"
)

// Service 直播业务。
type Service struct {
	db     *gorm.DB
	cfg    *config.Config
	engine *EngineManager

	mu        sync.RWMutex
	sessions  map[string]*pushSession
	proxyKeys map[string]string
}

type pushSession struct {
	StreamID  string
	Name      string
	App       string
	Kind      string
	ExpireAt  time.Time
	CreatedAt time.Time
}

// NewService 创建 Service。
func NewService(db *gorm.DB, cfg *config.Config, engine *EngineManager) *Service {
	s := &Service{
		db:        db,
		cfg:       cfg,
		engine:    engine,
		sessions:  make(map[string]*pushSession),
		proxyKeys: make(map[string]string),
	}
	go s.gcLoop()
	return s
}

// Engine 返回引擎管理器。
func (s *Service) Engine() *EngineManager { return s.engine }

// zlm 每次调用取 EngineManager 的最新 client。
func (s *Service) zlm() *zlmClient {
	if s.engine == nil {
		return newZLMClient(s.cfg.ZLMBaseURL, s.cfg.ZLMSecret)
	}
	return s.engine.Client()
}

// requireEngine 探活，失败时返回 503。
func (s *Service) requireEngine(action string) error {
	if s.engine == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.engine.Probe(ctx); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("%s失败：直播引擎未就绪（%s）", action, err.Error()),
			503, nil)
	}
	return nil
}

func (s *Service) gcLoop() {
	t := time.NewTicker(1 * time.Minute)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		s.mu.Lock()
		for k, v := range s.sessions {
			if now.After(v.ExpireAt) {
				delete(s.sessions, k)
				delete(s.proxyKeys, k)
			}
		}
		s.mu.Unlock()
	}
}

// ============================================================
// 直播间：rooms（核心）
// ============================================================

// CreateRoom 创建直播间。
func (s *Service) CreateRoom(userID uint, req *CreateRoomReq) (*Room, error) {
	// ★ v1.5：webrtc / obs / proxy 都要求 ZLM 引擎就绪。
	//   否则房间建出来也没有推流/播放能力，且前端"推流失败"后会在 DB 残留空房间。
	//   统一在这里前置校验，未就绪直接报错、不创建房间。
	if req.PushType == "webrtc" || req.PushType == "obs" || req.PushType == "proxy" {
		if err := s.requireEngine("创建直播间"); err != nil {
			return nil, err
		}
	}
	if req.PushType == "proxy" {
		if strings.TrimSpace(req.SourceURL) == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"push_type=proxy 时 source_url 必填", 400, nil)
		}
		if !strings.Contains(req.SourceURL, "://") {
			return nil, exception.New(exception.CodeParamInvalid,
				"source_url 必须包含协议前缀（rtmp:// / rtsp:// / http:// / https://）",
				400, nil)
		}
	}

	streamID := "s" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]

	room := &LiveRoom{
		StreamID: streamID,
		Name:     req.Name,
		PushType: req.PushType,
		UserID:   userID,
	}

	if req.PushType == "proxy" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		key, err := s.zlm().AddStreamProxy(ctx, "live", streamID, req.SourceURL)
		if err != nil {
			log.Error().Err(err).
				Str("src_url", req.SourceURL).
				Str("stream_id", streamID).
				Msg("创建直播间时触发 ZLM 拉流代理失败")
			return nil, exception.New(exception.CodeInternalError,
				"创建拉流代理失败："+err.Error(), 500, nil)
		}
		room.ProxyKey = key
	}

	if err := s.db.Create(room).Error; err != nil {
		if room.ProxyKey != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.zlm().DelStreamProxy(ctx, room.ProxyKey)
		}
		return nil, exception.New(exception.CodeInternalError,
			"创建直播间失败："+err.Error(), 500, nil)
	}

	log.Info().
		Str("stream_id", streamID).
		Str("name", req.Name).
		Str("push_type", req.PushType).
		Uint("user_id", userID).
		Msg("直播间已创建")

	return s.roomToResp(room, false, 0), nil
}

// ListRooms 直播间列表（DB + ZLM 实时状态合并）。
//
// 诊断增强：
//   - 打印 DB 里的 stream_id 列表；
//   - 打印 ZLM 返回的流列表；
//   - 打印每个 room 的匹配结果；
//   - GetMediaList 失败时记 warn 日志（返回值仍降级）。
func (s *Service) ListRooms() (*RoomListResp, error) {
	var rooms []LiveRoom
	if err := s.db.Order("created_at DESC").Find(&rooms).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询直播间失败", 500, nil)
	}

	// 调试：打印 DB 里的 stream_id
	dbStreamIDs := make([]string, 0, len(rooms))
	for i := range rooms {
		dbStreamIDs = append(dbStreamIDs, rooms[i].StreamID)
	}
	log.Debug().
		Int("room_count", len(rooms)).
		Strs("db_stream_ids", dbStreamIDs).
		Msg("ListRooms: DB 中的直播间")

	// 查 ZLM
	items, err := s.zlm().GetMediaList()
	if err != nil {
		log.Warn().Err(err).Msg("ZLM getMediaList 失败（降级：全部房间视为未推流）")
		items = nil
	}

	// 调试：打印 ZLM 返回的流
	zlmStreams := make([]string, 0, len(items))
	for i := range items {
		zlmStreams = append(zlmStreams, items[i].App+"/"+items[i].Stream+"/"+items[i].Schema)
	}
	log.Debug().
		Int("zlm_stream_count", len(items)).
		Strs("zlm_streams", zlmStreams).
		Msg("ListRooms: ZLM 活跃流")

	// 按 stream 归并（只统计 app=live）
	type agg struct {
		pushing bool
		viewers int
		bitrate int
	}
	active := make(map[string]*agg, len(items))
	for i := range items {
		it := &items[i]
		if it.App != "live" {
			continue
		}
		a, ok := active[it.Stream]
		if !ok {
			a = &agg{}
			active[it.Stream] = a
		}
		a.pushing = true
		if it.ReaderCount > a.viewers {
			a.viewers = it.ReaderCount
		}
		if sum := sumTrackBitrate(it.Tracks); sum > a.bitrate {
			a.bitrate = sum
		}
	}

	result := make([]*Room, 0, len(rooms))
	for i := range rooms {
		room := &rooms[i]
		a := active[room.StreamID]
		if a != nil {
			log.Debug().
				Str("stream_id", room.StreamID).
				Bool("is_pushing", a.pushing).
				Int("viewers", a.viewers).
				Msg("ListRooms: room 匹配到 ZLM 活跃流")
			result = append(result, s.roomToResp(room, a.pushing, a.viewers))
		} else {
			log.Debug().
				Str("stream_id", room.StreamID).
				Msg("ListRooms: room 未匹配到 ZLM 活跃流")
			result = append(result, s.roomToResp(room, false, 0))
		}
	}
	return &RoomListResp{List: result}, nil
}

// DeleteRoom 删除直播间（同时踢流 + 清理代理）。
func (s *Service) DeleteRoom(userID uint, roomID uint) error {
	var room LiveRoom
	if err := s.db.First(&room, roomID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return exception.New(exception.CodeNotFound, "直播间不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError, "查询直播间失败", 500, nil)
	}
	if userID != 0 && room.UserID != userID && !isAdminUser(userID) {
		return exception.New(exception.CodeAuthForbidden, "无权删除该直播间", 403, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.zlm().CloseStreams(ctx, "live", room.StreamID); err != nil {
		log.Warn().Err(err).Str("stream_id", room.StreamID).
			Msg("关闭 ZLM 流失败（忽略，继续删 DB）")
	}
	if room.ProxyKey != "" {
		if err := s.zlm().DelStreamProxy(ctx, room.ProxyKey); err != nil {
			log.Warn().Err(err).Str("key", room.ProxyKey).
				Msg("删除 ZLM 代理失败（忽略）")
		}
	}

	if err := s.db.Delete(&room).Error; err != nil {
		return exception.New(exception.CodeInternalError, "删除直播间失败", 500, nil)
	}

	log.Info().
		Uint("room_id", room.ID).
		Str("stream_id", room.StreamID).
		Msg("直播间已删除")
	return nil
}

func isAdminUser(userID uint) bool {
	return userID == 1
}

func (s *Service) roomToResp(r *LiveRoom, isPushing bool, viewers int) *Room {
	return &Room{
		ID:        strconv.FormatUint(uint64(r.ID), 10),
		StreamID:  r.StreamID,
		Name:      r.Name,
		PushType:  r.PushType,
		IsPushing: isPushing,
		Viewers:   viewers,
		CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		PlayURLs:  s.buildPlayURLs("live", r.StreamID),
		PushURLs:  s.buildPushURLs("live", r.StreamID),
	}
}

func (s *Service) buildPushURLs(app, stream string) PushURLs {
	return PushURLs{
		RTMP: fmt.Sprintf("rtmp://%s:%d/%s/%s",
			s.cfg.ZLMHTTPHost, s.cfg.ZLMRTMPPort, app, stream),
		WebRTC: fmt.Sprintf("http://%s:%d/index/api/webrtc?app=%s&stream=%s&type=push",
			s.cfg.ZLMHTTPHost, s.cfg.ZLMHTTPPort, app, url.QueryEscape(stream)),
	}
}

// ============================================================
// 老接口：push/start、push/stop、push/webrtc（保留）
// ============================================================

func (s *Service) StartPush(req *PushStartReq) (*PushSession, error) {
	if req.ExpireSeconds <= 0 {
		req.ExpireSeconds = 3600
	}
	streamID := "s" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	expireAt := time.Now().Add(time.Duration(req.ExpireSeconds) * time.Second)
	app := "live"

	s.mu.Lock()
	s.sessions[streamID] = &pushSession{
		StreamID: streamID, Name: req.Name, App: app, Kind: "push",
		ExpireAt: expireAt, CreatedAt: time.Now(),
	}
	s.mu.Unlock()

	urls := s.buildPushURLs(app, streamID)
	return &PushSession{
		StreamID:  streamID,
		RTMPURL:   urls.RTMP,
		WebRTCURL: urls.WebRTC,
		ExpireAt:  expireAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) StopPush(streamID string) error {
	s.mu.Lock()
	_, ok := s.sessions[streamID]
	delete(s.sessions, streamID)
	key, isProxy := s.proxyKeys[streamID]
	if isProxy {
		delete(s.proxyKeys, streamID)
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.zlm().CloseStreams(ctx, "live", streamID); err != nil {
		log.Warn().Err(err).Str("stream_id", streamID).Msg("关闭 ZLM 流失败（忽略）")
	}
	if isProxy && key != "" {
		_ = s.zlm().DelStreamProxy(ctx, key)
	}
	return nil
}

func (s *Service) PushWebRTC(streamID, offerSDP string) (*PushWebRTCResp, error) {
	if err := s.requireEngine("WebRTC 推流"); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	answerSDP, err := s.zlm().WebRTCPush(ctx, "live", streamID, offerSDP)
	if err != nil {
		log.Error().Err(err).Str("stream_id", streamID).Msg("ZLM WebRTC 推流协商失败")
		return nil, exception.New(exception.CodeInternalError,
			"WebRTC 推流协商失败："+err.Error(), 500, nil)
	}
	log.Info().Str("stream_id", streamID).Msg("WebRTC 推流协商成功")
	return &PushWebRTCResp{SDP: answerSDP, Type: "answer"}, nil
}

// ============================================================
// 老接口：proxy/add、proxy/remove（保留）
// ============================================================

func (s *Service) AddStreamProxy(req *AddStreamProxyReq) (*AddStreamProxyResp, error) {
	if err := s.requireEngine("创建拉流代理"); err != nil {
		return nil, err
	}
	if !strings.Contains(req.URL, "://") {
		return nil, exception.New(exception.CodeParamInvalid,
			"外部流地址必须包含协议前缀", 400, nil)
	}

	streamID := "p" + strings.ReplaceAll(uuid.NewString(), "-", "")[:15]
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	key, err := s.zlm().AddStreamProxy(ctx, "live", streamID, req.URL)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"创建拉流代理失败："+err.Error(), 500, nil)
	}

	s.mu.Lock()
	s.sessions[streamID] = &pushSession{
		StreamID: streamID, Name: req.Name, App: "live", Kind: "proxy",
		ExpireAt: time.Now().Add(24 * time.Hour), CreatedAt: time.Now(),
	}
	s.proxyKeys[streamID] = key
	s.mu.Unlock()

	return &AddStreamProxyResp{StreamID: streamID, Key: key}, nil
}

func (s *Service) RemoveStreamProxy(streamID string) error {
	s.mu.Lock()
	key, ok := s.proxyKeys[streamID]
	if ok {
		delete(s.proxyKeys, streamID)
		delete(s.sessions, streamID)
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.zlm().DelStreamProxy(ctx, key)
	return nil
}

// ============================================================
// 老接口：streams、play（保留）
// ============================================================

func (s *Service) ListStreams() (*StreamListResp, error) {
	items, err := s.zlm().GetMediaList()
	if err != nil {
		log.Warn().Err(err).Msg("ZLM getMediaList 失败（降级为空列表）")
		return &StreamListResp{List: []*LiveStream{}}, nil
	}

	seen := make(map[string]bool, len(items))
	result := make([]*LiveStream, 0, len(items))
	for i := range items {
		it := &items[i]
		if seen[it.Stream] {
			continue
		}
		seen[it.Stream] = true

		name := it.Stream
		s.mu.RLock()
		if sess, ok := s.sessions[it.Stream]; ok && sess.Name != "" {
			name = sess.Name
		}
		s.mu.RUnlock()

		bitrate := int(it.BytesSpeed * 8)
		if sum := sumTrackBitrate(it.Tracks); sum > 0 {
			bitrate = sum
		}

		result = append(result, &LiveStream{
			StreamID:  it.Stream,
			Name:      name,
			App:       it.App,
			IsPushing: true,
			Viewers:   it.ReaderCount,
			Bitrate:   bitrate,
			CreatedAt: time.Unix(it.CreateStamp, 0).UTC().Format(time.RFC3339),
			PlayURLs:  s.buildPlayURLs(it.App, it.Stream),
		})
	}
	return &StreamListResp{List: result}, nil
}

func sumTrackBitrate(tracks []track) int {
	sum := 0
	for _, t := range tracks {
		if t.BitRate > 0 {
			sum += t.BitRate
		}
	}
	return sum
}

func (s *Service) buildPlayURLs(app, stream string) StreamPlayURL {
	httpBase := fmt.Sprintf("http://%s:%d", s.cfg.ZLMHTTPHost, s.cfg.ZLMHTTPPort)
	return StreamPlayURL{
		HLS:  fmt.Sprintf("%s/%s/%s/hls.m3u8", httpBase, app, stream),
		FLV:  fmt.Sprintf("%s/%s/%s.live.flv", httpBase, app, stream),
		RTMP: fmt.Sprintf("rtmp://%s:%d/%s/%s", s.cfg.ZLMHTTPHost, s.cfg.ZLMRTMPPort, app, stream),
		WebRTC: fmt.Sprintf("%s/index/api/webrtc?app=%s&stream=%s&type=play",
			httpBase, app, url.QueryEscape(stream)),
	}
}

func (s *Service) GetPlayURL(streamID, format string) (*PlayStreamResp, error) {
	app := "live"
	if strings.Contains(streamID, "/") {
		parts := strings.SplitN(streamID, "/", 2)
		app, streamID = parts[0], parts[1]
	}
	urls := s.buildPlayURLs(app, streamID)
	switch strings.ToLower(format) {
	case "hls":
		return &PlayStreamResp{URL: urls.HLS, Type: "application/vnd.apple.mpegurl"}, nil
	case "flv":
		return &PlayStreamResp{URL: urls.FLV, Type: "video/x-flv"}, nil
	case "rtmp":
		return &PlayStreamResp{URL: urls.RTMP, Type: "rtmp/mp4"}, nil
	case "webrtc":
		return &PlayStreamResp{URL: urls.WebRTC, Type: "text/html"}, nil
	default:
		return nil, exception.New(exception.CodeParamInvalid,
			"不支持的 format，仅 hls/flv/rtmp/webrtc", 400, nil)
	}
}
