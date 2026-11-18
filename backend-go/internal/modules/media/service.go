package media

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/config"
	"backend-go/internal/exception"
)

// Service 媒体库业务逻辑。
type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewService 创建 Service。
//
// 启动时调用 ensureDirs 保证 <MediaRoot>/{video,audio,video/thumbnails,tmp/uploads}
// 目录存在；即使初始化失败也仅记日志，由后续 InitUpload / CompleteUpload
// 中的兜底逻辑再次尝试创建。
func NewService(db *gorm.DB, cfg *config.Config) *Service {
	s := &Service{db: db, cfg: cfg}
	if err := s.ensureDirs(); err != nil {
		log.Error().Err(err).Str("media_root", cfg.MediaRoot).
			Msg("媒体根目录初始化失败（将在上传时重试）")
	} else {
		log.Info().Str("media_root", cfg.MediaRoot).Msg("媒体根目录已就绪")
	}
	return s
}

// ---------------------------------------------------------------------------
// 工具
// ---------------------------------------------------------------------------

func (s *Service) mediaDir(typ string) string {
	return filepath.Join(s.cfg.MediaRoot, typ)
}

func (s *Service) thumbnailDir() string {
	return filepath.Join(s.cfg.MediaRoot, "video", "thumbnails")
}

func (s *Service) tmpUploadDir(uploadID string) string {
	return filepath.Join(s.cfg.MediaRoot, "tmp", "uploads", uploadID)
}

func (s *Service) metaFile(uploadID string) string {
	return filepath.Join(s.tmpUploadDir(uploadID), "meta.json")
}

// ensureDirs 确保存储目录存在（幂等）。
func (s *Service) ensureDirs() error {
	dirs := []string{
		s.mediaDir("video"),
		s.mediaDir("audio"),
		s.thumbnailDir(),
		filepath.Join(s.cfg.MediaRoot, "tmp", "uploads"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}
	return nil
}

// publicURL 拼接完整对外 URL。
func (s *Service) publicURL(path string) string {
	base := strings.TrimRight(s.cfg.PublicBaseURL, "/")
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

// extOf 从文件名取扩展名（小写、不含点）。
func extOf(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// detectFormatFromProbe 根据 ffprobe 输出推断 format。
func detectFormatFromProbe(formatName, ext string) string {
	if ext != "" {
		return ext
	}
	if i := strings.Index(formatName, ","); i > 0 {
		return formatName[:i]
	}
	return formatName
}

// toMediaInfo 将模型转成前端契约对象。
func (s *Service) toMediaInfo(m *Media) *MediaInfo {
	if m == nil {
		return nil
	}
	info := &MediaInfo{
		ID:         m.ID,
		Name:       m.Name,
		Type:       m.Type,
		Format:     m.Format,
		Size:       m.Size,
		Duration:   m.Duration,
		Resolution: m.Resolution,
		Bitrate:    m.Bitrate,
		Status:     m.Status,
		CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  m.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if m.Thumbnail != "" {
		// 缩略图通过鉴权直链 /media/:id/thumbnail 提供（<img> 无法带 Authorization 头，
		// 前端会用 withToken 附加 ?token=）。不再直接指向文件系统路径。
		info.Thumbnail = s.publicURL("/api/v1/media/" + m.ID + "/thumbnail")
	}
	info.PlayURL = s.publicURL("/api/v1/media/" + m.ID + "/raw")
	info.DownloadURL = s.publicURL("/api/v1/media/" + m.ID + "/download")
	return info
}

// ---------------------------------------------------------------------------
// ffprobe / ffmpeg
// ---------------------------------------------------------------------------

type probeResult struct {
	Duration   float64
	Bitrate    int
	Resolution string
	Format     string
}

func (s *Service) probeMedia(ctx context.Context, filePath string) (*probeResult, error) {
	cmd := exec.CommandContext(ctx,
		s.cfg.MediaFFprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe 执行失败: %w", err)
	}

	var raw struct {
		Format struct {
			Duration   string `json:"duration"`
			BitRate    string `json:"bit_rate"`
			FormatName string `json:"format_name"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			BitRate   string `json:"bit_rate"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("ffprobe 输出解析失败: %w", err)
	}

	res := &probeResult{}
	res.Duration, _ = strconv.ParseFloat(raw.Format.Duration, 64)
	res.Bitrate, _ = strconv.Atoi(raw.Format.BitRate)
	res.Format = detectFormatFromProbe(raw.Format.FormatName, "")

	for _, st := range raw.Streams {
		if st.CodecType == "video" && st.Width > 0 && st.Height > 0 {
			res.Resolution = fmt.Sprintf("%dx%d", st.Width, st.Height)
			break
		}
	}
	return res, nil
}

func (s *Service) generateThumbnail(ctx context.Context, mediaID, srcPath string) (string, error) {
	rel := filepath.Join("video", "thumbnails", mediaID+".jpg")
	abs := filepath.Join(s.cfg.MediaRoot, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx,
		s.cfg.MediaFFmpegPath,
		"-y", "-ss", "1", "-i", srcPath,
		"-vframes", "1",
		"-vf", "scale=320:-1",
		"-q:v", "3",
		abs,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Warn().Err(err).Str("out", string(out)).Msg("ffmpeg 缩略图生成失败")
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return rel, nil
}

// ---------------------------------------------------------------------------
// 上传：init / chunk / complete
// ---------------------------------------------------------------------------

// InitUpload 初始化上传；若同 hash+size 的媒体已存在，返回秒传结果。
func (s *Service) InitUpload(req *InitUploadReq, userID uint) (*InitUploadResp, error) {
	// 兜底：确保目录存在（NewService 已创建，这里幂等无开销）
	if err := s.ensureDirs(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"媒体目录初始化失败："+err.Error(), 500, nil)
	}

	// 秒传：hash + size 精确匹配
	var exist Media
	err := s.db.Where("file_hash = ? AND size = ? AND status = ?", req.FileHash, req.FileSize, "ready").
		First(&exist).Error
	if err == nil {
		return &InitUploadResp{UploadID: "", Exist: s.toMediaInfo(&exist)}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, exception.New(exception.CodeInternalError, "查重失败", 500, nil)
	}

	uploadID := uuid.NewString()
	dir := s.tmpUploadDir(uploadID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			"创建上传目录失败："+err.Error(), 500, nil)
	}

	meta := &uploadSessionMeta{
		UploadID:    uploadID,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		FileHash:    req.FileHash,
		ChunkSize:   req.ChunkSize,
		TotalChunks: req.TotalChunks,
		MediaType:   req.MediaType,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}
	if err := s.saveMeta(meta); err != nil {
		return nil, err
	}

	go s.cleanupExpiredSessions()

	return &InitUploadResp{UploadID: uploadID, Exist: nil}, nil
}

func (s *Service) saveMeta(meta *uploadSessionMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return exception.New(exception.CodeInternalError, "序列化 meta 失败", 500, nil)
	}
	return os.WriteFile(s.metaFile(meta.UploadID), data, 0o644)
}

func (s *Service) loadMeta(uploadID string) (*uploadSessionMeta, error) {
	data, err := os.ReadFile(s.metaFile(uploadID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, exception.New(exception.CodeNotFound, "上传会话不存在或已过期", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "读取上传会话失败", 500, nil)
	}
	var meta uploadSessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, exception.New(exception.CodeInternalError, "上传会话损坏", 500, nil)
	}
	return &meta, nil
}

// SaveChunk 保存单个分片。
func (s *Service) SaveChunk(uploadID string, chunkIndex int, src io.Reader, userID uint) (int, error) {
	meta, err := s.loadMeta(uploadID)
	if err != nil {
		return 0, err
	}
	if meta.UserID != userID {
		return 0, exception.New(exception.CodeAuthForbidden, "越权访问上传会话", 403, nil)
	}
	if chunkIndex < 0 || chunkIndex >= meta.TotalChunks {
		return 0, exception.New(exception.CodeParamInvalid, "分片序号越界", 400, nil)
	}

	chunkPath := filepath.Join(s.tmpUploadDir(uploadID), fmt.Sprintf("chunk_%d", chunkIndex))
	f, err := os.Create(chunkPath)
	if err != nil {
		return 0, exception.New(exception.CodeInternalError,
			"创建分片失败："+err.Error(), 500, nil)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if _, err := io.Copy(w, src); err != nil {
		return 0, exception.New(exception.CodeInternalError, "写入分片失败", 500, nil)
	}
	if err := w.Flush(); err != nil {
		return 0, exception.New(exception.CodeInternalError, "刷新分片失败", 500, nil)
	}

	count := 0
	for i := 0; i < meta.TotalChunks; i++ {
		if _, err := os.Stat(filepath.Join(s.tmpUploadDir(uploadID), fmt.Sprintf("chunk_%d", i))); err == nil {
			count++
		}
	}
	return count, nil
}

// CompleteUpload 合并分片、ffprobe、缩略图、落库。
func (s *Service) CompleteUpload(uploadID string, userID uint) (*MediaInfo, error) {
	meta, err := s.loadMeta(uploadID)
	if err != nil {
		return nil, err
	}
	if meta.UserID != userID {
		return nil, exception.New(exception.CodeAuthForbidden, "越权合并上传", 403, nil)
	}

	// 1. 校验完整性
	for i := 0; i < meta.TotalChunks; i++ {
		p := filepath.Join(s.tmpUploadDir(uploadID), fmt.Sprintf("chunk_%d", i))
		if _, err := os.Stat(p); err != nil {
			return nil, exception.New(exception.CodeParamInvalid,
				fmt.Sprintf("分片缺失：%d", i), 400, nil)
		}
	}

	// 2. 目标路径 + ★ 确保父目录存在（最终防线）
	mediaID := uuid.NewString()
	ext := extOf(meta.FileName)
	if ext == "" {
		ext = "bin"
	}
	relDir := meta.MediaType
	finalName := mediaID + "." + ext
	relPath := filepath.Join(relDir, finalName)
	absPath := filepath.Join(s.cfg.MediaRoot, relPath)

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建目标目录失败（%s）：%v", filepath.Dir(absPath), err),
			500, nil)
	}

	// 3. 合并
	out, err := os.Create(absPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建目标文件失败（%s）：%v", absPath, err),
			500, nil)
	}
	for i := 0; i < meta.TotalChunks; i++ {
		p := filepath.Join(s.tmpUploadDir(uploadID), fmt.Sprintf("chunk_%d", i))
		in, err := os.Open(p)
		if err != nil {
			out.Close()
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("打开分片失败（%s）：%v", p, err), 500, nil)
		}
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("合并分片失败（%s）：%v", p, err), 500, nil)
		}
		in.Close()
	}
	if err := out.Close(); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("关闭目标文件失败：%v", err), 500, nil)
	}

	// 4. ffprobe（失败不阻塞落库）
	probe, probeErr := s.probeMedia(context.Background(), absPath)
	if probeErr != nil {
		log.Warn().Err(probeErr).Str("file", absPath).Msg("ffprobe 失败，元数据置空")
		probe = &probeResult{Format: ext}
	}

	// 5. 缩略图（仅视频且分辨率有效）
	thumbRel := ""
	if meta.MediaType == "video" && probe.Resolution != "" {
		if rel, err := s.generateThumbnail(context.Background(), mediaID, absPath); err == nil {
			thumbRel = rel
		}
	}

	// 6. 落库
	now := time.Now().UTC()
	m := &Media{
		ID:          mediaID,
		Name:        meta.FileName,
		Type:        meta.MediaType,
		Format:      probe.Format,
		Size:        meta.FileSize,
		Duration:    probe.Duration,
		Resolution:  probe.Resolution,
		Bitrate:     probe.Bitrate,
		Thumbnail:   thumbRel,
		FilePath:    relPath,
		FileHash:    meta.FileHash,
		Status:      "ready",
		Description: "",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.db.Create(m).Error; err != nil {
		_ = os.Remove(absPath)
		return nil, exception.New(exception.CodeInternalError,
			"保存媒体记录失败："+err.Error(), 500, nil)
	}

	_ = os.RemoveAll(s.tmpUploadDir(uploadID))

	return s.toMediaInfo(m), nil
}

// ---------------------------------------------------------------------------
// 列表 / 详情 / 删除 / 下载 / 播放
// ---------------------------------------------------------------------------

type ListQuery struct {
	Page     int
	PageSize int
	Type     string
	Keyword  string
}

func (s *Service) List(q *ListQuery) (*ListResp, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 200 {
		q.PageSize = 20
	}

	tx := s.db.Model(&Media{})
	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}
	if q.Keyword != "" {
		tx = tx.Where("name LIKE ?", "%"+q.Keyword+"%")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "统计失败", 500, nil)
	}

	var rows []Media
	if err := tx.Order("created_at DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&rows).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}

	list := make([]*MediaInfo, 0, len(rows))
	for i := range rows {
		list = append(list, s.toMediaInfo(&rows[i]))
	}
	return &ListResp{Total: int(total), List: list}, nil
}

func (s *Service) Detail(id string) (*MediaInfo, error) {
	var m Media
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "媒体不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	return s.toMediaInfo(&m), nil
}

func (s *Service) GetRaw(id string) (*Media, error) {
	var m Media
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.New(exception.CodeNotFound, "媒体不存在", 404, nil)
		}
		return nil, exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	return &m, nil
}

func (s *Service) Remove(id string) error {
	var m Media
	if err := s.db.Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.New(exception.CodeNotFound, "媒体不存在", 404, nil)
		}
		return exception.New(exception.CodeInternalError, "查询失败", 500, nil)
	}
	if err := s.db.Delete(&m).Error; err != nil {
		return exception.New(exception.CodeInternalError, "删除失败", 500, nil)
	}
	if m.FilePath != "" {
		_ = os.Remove(filepath.Join(s.cfg.MediaRoot, m.FilePath))
	}
	if m.Thumbnail != "" {
		_ = os.Remove(filepath.Join(s.cfg.MediaRoot, m.Thumbnail))
	}
	return nil
}

func (s *Service) Play(id string) (*PlayResp, error) {
	m, err := s.GetRaw(id)
	if err != nil {
		return nil, err
	}
	if m.Status != "ready" {
		return nil, exception.New(exception.CodeParamInvalid, "媒体未就绪", 400, nil)
	}
	typ := "video/" + m.Format
	if m.Type == "audio" {
		typ = "audio/" + m.Format
	}
	return &PlayResp{
		URL:  s.publicURL("/api/v1/media/" + m.ID + "/raw"),
		Type: typ,
	}, nil
}

func (s *Service) AbsPath(m *Media) string {
	return filepath.Join(s.cfg.MediaRoot, m.FilePath)
}

// ---------------------------------------------------------------------------
// 清理
// ---------------------------------------------------------------------------

func (s *Service) cleanupExpiredSessions() {
	root := filepath.Join(s.cfg.MediaRoot, "tmp", "uploads")
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-s.cfg.MediaSessionTTL)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metaPath := filepath.Join(root, e.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta uploadSessionMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		if meta.CreatedAt.Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(root, e.Name()))
			log.Info().Str("upload_id", meta.UploadID).Msg("清理过期上传会话")
		}
	}
}

// randomHex 生成 n 字节随机 hex（预留）。
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
