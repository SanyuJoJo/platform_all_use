package live

import "time"

// LiveRoom 直播间（media_live_room）。
//
// 设计要点：
//   - 直播间是持久化记录，刷新 / 重启不丢；
//   - StreamID 创建后固定，推流 / 拉流都以它为锚点；
//   - 是否在推流（is_pushing）不落库，实时由 ZLM getMediaList 判定；
//   - ProxyKey 仅 push_type=proxy 时使用，保存 ZLM addStreamProxy 返回的 key，
//     删除直播间时需要它来通知 ZLM 停止拉流。
type LiveRoom struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	StreamID  string    `gorm:"size:64;not null;uniqueIndex:uk_stream_id" json:"stream_id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	PushType  string    `gorm:"size:20;not null" json:"push_type"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ProxyKey  string    `gorm:"size:255;not null;default:''" json:"-"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
}

// TableName 指定表名。
func (LiveRoom) TableName() string { return "media_live_room" }
