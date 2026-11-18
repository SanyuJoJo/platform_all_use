package live

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// zlmClient ZLMediaKit HTTP API 客户端。
type zlmClient struct {
	baseURL string
	secret  string
	http    *http.Client
}

func newZLMClient(baseURL, secret string) *zlmClient {
	return &zlmClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		secret:  secret,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// mediaItem ZLM getMediaList 单条返回。
//
// 注意：不同 ZLM 版本字段略有差异，用宽松解析；缺失字段用默认零值。
type mediaItem struct {
	App           string  `json:"app"`
	Stream        string  `json:"stream"`
	Schema        string  `json:"schema"`
	Vhost         string  `json:"vhost"`
	ReaderCount   int     `json:"readerCount"`
	TotalReader   int     `json:"totalReaderCount"`
	BytesSpeed    int64   `json:"bytesSpeed"`
	AliveSecond   int64   `json:"aliveSecond"`
	CreateStamp   int64   `json:"createStamp"`
	OriginTypeStr string  `json:"originTypeStr"`
	Tracks        []track `json:"tracks"`
}

type track struct {
	CodecID     int    `json:"codec_id"`
	CodecType   int    `json:"codec_type"` // 0=video 1=audio
	BitRate     int    `json:"bit_rate"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	CodecIDName string `json:"codec_id_name"`
}

// GetMediaList 拉取活跃流列表。
//
// 说明：
//   - 显式传 vhost=__defaultVhost__（某些 ZLM 版本默认不返回完整列表）；
//   - 返回空 data（无 data 字段）视为空列表，不是错误；
//   - 打印详细 debug 日志，方便排障。
func (c *zlmClient) GetMediaList() ([]mediaItem, error) {
	params := url.Values{}
	params.Set("secret", c.secret)
	params.Set("vhost", "__defaultVhost__")

	u := fmt.Sprintf("%s/index/api/getMediaList?%s", c.baseURL, params.Encode())
	resp, err := c.http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("zlm getMediaList 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("zlm getMediaList 读取响应失败: %w", err)
	}

	log.Debug().
		Str("url", u).
		Int("resp_size", len(body)).
		Str("resp_preview", previewBody(body, 300)).
		Msg("zlm getMediaList")

	var raw struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("zlm getMediaList 解析失败: %w, body=%s", err, string(body))
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("zlm getMediaList 错误: code=%d msg=%s", raw.Code, raw.Msg)
	}

	// data 可能不存在（无流），或 null，或 []
	if len(raw.Data) == 0 || string(raw.Data) == "null" {
		return []mediaItem{}, nil
	}

	var items []mediaItem
	if err := json.Unmarshal(raw.Data, &items); err != nil {
		return nil, fmt.Errorf("zlm getMediaList data 解析失败: %w, data=%s", err, string(raw.Data))
	}

	log.Debug().
		Int("count", len(items)).
		Msg("zlm getMediaList 返回流数量")

	return items, nil
}

// previewBody 截断字符串用于日志。
func previewBody(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

// CloseStreams 强制关闭指定 app/stream 上的所有会话。
func (c *zlmClient) CloseStreams(ctx context.Context, app, stream string) error {
	params := url.Values{}
	params.Set("secret", c.secret)
	params.Set("vhost", "__defaultVhost__")
	params.Set("app", app)
	params.Set("stream", stream)
	params.Set("force", "1")

	u := fmt.Sprintf("%s/index/api/close_streams?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("zlm close_streams 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var raw struct {
		Code        int    `json:"code"`
		Msg         string `json:"msg"`
		CountHit    int    `json:"count_hit"`
		CountClosed int    `json:"count_closed"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil // 兼容老版本
	}
	if raw.Code != 0 {
		return fmt.Errorf("zlm close_streams 错误: code=%d msg=%s", raw.Code, raw.Msg)
	}
	return nil
}

// WebRTCPush 转发 SDP Offer 到 ZLM，返回 Answer SDP。
func (c *zlmClient) WebRTCPush(ctx context.Context, app, stream, offerSDP string) (string, error) {
	u := fmt.Sprintf("%s/index/api/webrtc?secret=%s&app=%s&stream=%s&type=push",
		c.baseURL, url.QueryEscape(c.secret),
		url.QueryEscape(app), url.QueryEscape(stream),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(offerSDP))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("zlm webrtc 请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("zlm webrtc 读取响应失败: %w", err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("ZLM 返回空响应：当前 MediaServer 可能未编译 WebRTC 支持")
	}

	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		SDP  string `json:"sdp"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("zlm webrtc 解析失败: %w, body=%s", err, string(body))
	}
	if raw.Code != 0 {
		return "", fmt.Errorf("zlm webrtc 错误: code=%d msg=%s", raw.Code, raw.Msg)
	}
	if raw.SDP == "" {
		return "", fmt.Errorf("zlm webrtc 未返回 sdp")
	}
	return raw.SDP, nil
}

// AddStreamProxy 拉流代理。
func (c *zlmClient) AddStreamProxy(ctx context.Context, app, stream, srcURL string) (string, error) {
	params := url.Values{}
	params.Set("secret", c.secret)
	params.Set("vhost", "__defaultVhost__")
	params.Set("app", app)
	params.Set("stream", stream)
	params.Set("url", srcURL)
	params.Set("enable_hls", "1")
	params.Set("enable_mp4", "0")
	params.Set("enable_rtmp", "1")
	params.Set("enable_rtsp", "1")

	u := fmt.Sprintf("%s/index/api/addStreamProxy?%s", c.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("zlm addStreamProxy 请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Key string `json:"key"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("zlm addStreamProxy 解析失败: %w", err)
	}
	if raw.Code != 0 {
		return "", fmt.Errorf("zlm addStreamProxy 错误: code=%d msg=%s", raw.Code, raw.Msg)
	}
	if raw.Data.Key == "" {
		return "", fmt.Errorf("zlm addStreamProxy 未返回 key")
	}
	return raw.Data.Key, nil
}

// DelStreamProxy 删除拉流代理。
func (c *zlmClient) DelStreamProxy(ctx context.Context, key string) error {
	u := fmt.Sprintf("%s/index/api/delStreamProxy?secret=%s&key=%s",
		c.baseURL, url.QueryEscape(c.secret), url.QueryEscape(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("zlm delStreamProxy 请求失败: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var raw struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil
	}
	if raw.Code != 0 {
		return fmt.Errorf("zlm delStreamProxy 错误: code=%d msg=%s", raw.Code, raw.Msg)
	}
	return nil
}
