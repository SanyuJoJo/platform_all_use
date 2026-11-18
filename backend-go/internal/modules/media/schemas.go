package media
// MediaInfo 前端契约（严格对齐前端 types/index.ts）。
type MediaInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Format      string  `json:"format"`
	Size        int64   `json:"size"`
	Duration    float64 `json:"duration"`
	Resolution  string  `json:"resolution"`
	Bitrate     int     `json:"bitrate"`
	Thumbnail   string  `json:"thumbnail"`
	PlayURL     string  `json:"play_url"`
	DownloadURL string  `json:"download_url"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
// InitUploadReq POST /media/upload/init 请求体。
type InitUploadReq struct {
	FileName    string `json:"file_name" binding:"required,min=1,max=255"`
	FileSize    int64  `json:"file_size" binding:"required,gt=0"`
	FileHash    string `json:"file_hash" binding:"required,min=1,max=128"`
	ChunkSize   int64  `json:"chunk_size" binding:"required,gt=0"`
	TotalChunks int    `json:"total_chunks" binding:"required,gt=0"`
	MediaType   string `json:"media_type" binding:"required,oneof=video audio"`
	Description string `json:"description"`
}
// InitUploadResp POST /media/upload/init 响应体。
type InitUploadResp struct {
	UploadID string     `json:"upload_id"`
	Exist    *MediaInfo `json:"exist"`
}
// CompleteUploadReq POST /media/upload/complete 请求体。
type CompleteUploadReq struct {
	UploadID string `json:"upload_id" binding:"required,min=1,max=64"`
}
// CompleteUploadResp POST /media/upload/complete 响应体。
type CompleteUploadResp struct {
	Media MediaInfo `json:"media"`
}
// ListResp GET /media/list 响应体。
type ListResp struct {
	Total int          `json:"total"`
	List  []*MediaInfo `json:"list"`
}
// PlayResp GET /media/:id/play 响应体。
type PlayResp struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}
