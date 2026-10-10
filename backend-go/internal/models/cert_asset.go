package models

import "time"

// CertAsset 证书资产。
type CertAsset struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	AssetID        string    `gorm:"size:96;not null;uniqueIndex" json:"asset_id"`
	Kind           string    `gorm:"size:32;not null;index" json:"kind"` // ca | server | csr | dual_sign | dual_enc
	Domain         string    `gorm:"size:64;index" json:"domain,omitempty"`
	DirNo          string    `gorm:"size:8;index" json:"dir_no,omitempty"`
	PubkeySM3      string    `gorm:"size:64;not null;index" json:"pubkey_sm3"`
	SubjectCN      string    `gorm:"size:255" json:"subject_cn"`
	Algorithm      string    `gorm:"size:32" json:"algorithm"`
	CertPath       string    `gorm:"size:512" json:"cert_path,omitempty"`
	KeyPath        string    `gorm:"size:512" json:"key_path,omitempty"`
	PassPath       string    `gorm:"size:512" json:"pass_path,omitempty"`
	CSRPath        string    `gorm:"size:512" json:"csr_path,omitempty"`
	RehashPath     string    `gorm:"size:512" json:"rehash_path,omitempty"`
	HasPrivateKey  bool      `gorm:"not null;default:false" json:"has_private_key"`
	PrivateKeyMode string    `gorm:"size:32;not null;default:whitebox" json:"private_key_mode"`
	Status         string    `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	CreatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名。
func (CertAsset) TableName() string { return "platform_cert_asset" }
