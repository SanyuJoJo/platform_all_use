package media
import (
	"time"
	"gorm.io/gorm"
)
// Media 媒体元数据（platform_media）。
type Media struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Type        string    `gorm:"size:16;not null;index" json:"type"` // video|audio
	Format      string    `gorm:"size:32;not null" json:"format"`
	Size        int64     `gorm:"not null" json:"size"`
	Duration    float64   `gorm:"not null;default:0" json:"duration"`
	Resolution  string    `gorm:"size:32" json:"resolution"`
	Bitrate     int       `gorm:"not null;default:0" json:"bitrate"`
	Thumbnail   string    `gorm:"size:512" json:"thumbnail"`
	FilePath    string    `gorm:"size:1024;not null" json:"-"`
	FileHash    string    `gorm:"size:128;not null;index" json:"-"`
	Status      string    `gorm:"size:16;not null;default:ready;index" json:"status"`
	Description string    `gorm:"size:1024" json:"-"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}
// TableName 指定表名。
func (Media) TableName() string { return "platform_media" }
// UploadSession 上传会话（不持久化到业务库之外的表，独立管理）。
//
// 这里**不使用 GORM**，改用文件系统记录：tmp/uploads/{upload_id}/meta.json。
// 原因：上传会话生命周期短（< 2h），无需数据库；文件存储亦可避免垃圾表清理。
type uploadSessionMeta struct {
	UploadID    string    `json:"upload_id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	FileHash    string    `json:"file_hash"`
	ChunkSize   int64     `json:"chunk_size"`
	TotalChunks int       `json:"total_chunks"`
	MediaType   string    `json:"media_type"`
	UserID      uint      `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}
// EnsureTable 开发环境确保媒体表存在。
func EnsureTable(db *gorm.DB) error {
	return db.AutoMigrate(&Media{})
}
