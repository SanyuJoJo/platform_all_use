package models

import (
	"time"

	"gorm.io/gorm"
)

// CA 根 CA 与中间 CA 元数据表。
type CA struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	CAID         string    `gorm:"size:64;not null;uniqueIndex" json:"ca_id"`
	ParentCAID   *string   `gorm:"size:64;index" json:"parent_ca_id"`
	SubjectCN    string    `gorm:"size:255;not null" json:"subject_cn"`
	SubjectO     *string   `gorm:"size:255" json:"subject_o"`
	Algorithm    string    `gorm:"size:32;not null" json:"algorithm"`
	KeyParams    JSONMap   `gorm:"type:json" json:"key_params,omitempty"`
	ValidityDays int       `gorm:"not null" json:"validity_days"`
	CertPath     string    `gorm:"size:255;not null" json:"cert_path"`
	KeyRef       string    `gorm:"size:64;not null" json:"key_ref"`
	ChainPath    *string   `gorm:"size:255" json:"chain_path,omitempty"`
	Serial       *string   `gorm:"size:128" json:"serial,omitempty"`
	Status       string    `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CA) TableName() string { return "platform_ca" }

// Certificate 终端证书元数据表。
type Certificate struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	CertID     string    `gorm:"size:64;not null;uniqueIndex" json:"cert_id"`
	CertType   string    `gorm:"size:32;not null" json:"cert_type"`
	Serial     string    `gorm:"size:128;not null;index" json:"serial"`
	SubjectCN  string    `gorm:"size:255;not null" json:"subject_cn"`
	IssuerCN   string    `gorm:"size:255" json:"issuer_cn"`
	CAID       string    `gorm:"size:64;not null;index" json:"ca_id"`
	Algorithm  string    `gorm:"size:32;not null" json:"algorithm"`
	NotBefore  time.Time `gorm:"not null" json:"not_before"`
	NotAfter   time.Time `gorm:"not null;index" json:"not_after"`
	CertPath   string    `gorm:"size:255;not null" json:"cert_path"`
	ChainPath  *string   `gorm:"size:255" json:"chain_path,omitempty"`
	Status     string    `gorm:"size:16;not null;default:VALID;index" json:"status"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Certificate) TableName() string { return "platform_certificate" }

// CSR CSR/P10 元数据表。
type CSR struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	CSRID     string    `gorm:"size:64;not null;uniqueIndex" json:"csr_id"`
	SubjectCN string    `gorm:"size:255;not null" json:"subject_cn"`
	Algorithm string    `gorm:"size:32;not null" json:"algorithm"`
	CSRPath   string    `gorm:"size:255;not null" json:"csr_path"`
	KeyRef    *string   `gorm:"size:64" json:"key_ref,omitempty"`
	Status    string    `gorm:"size:16;not null;default:NEW;index" json:"status"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (CSR) TableName() string { return "platform_csr" }

// CRL CRL 元数据表。
type CRL struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	CRLID          string    `gorm:"size:64;not null;uniqueIndex" json:"crl_id"`
	CAID           string    `gorm:"size:64;not null;index" json:"ca_id"`
	CRLPath        string    `gorm:"size:255;not null" json:"crl_path"`
	RevokedCount   int       `gorm:"not null;default:0" json:"revoked_count"`
	DigestAlgo     string    `gorm:"size:16;not null" json:"digest_algorithm"`
	NextUpdate     *time.Time `json:"next_update,omitempty"`
	Status         string    `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	CreatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (CRL) TableName() string { return "platform_crl" }

// KeyMeta 密钥元数据表。
type KeyMeta struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	KeyID         string    `gorm:"size:64;not null;uniqueIndex" json:"key_id"`
	KeyRef        string    `gorm:"size:64;not null;index" json:"key_ref"`
	Algorithm     string    `gorm:"size:32;not null" json:"algorithm"`
	KeyParams     JSONMap   `gorm:"type:json" json:"key_params,omitempty"`
	EncryptedPath string    `gorm:"size:255;not null" json:"encrypted_path"`
	Permission    string    `gorm:"size:8;not null;default:0600" json:"permission"`
	Usage         string    `gorm:"size:64" json:"usage"`
	State         string    `gorm:"size:16;not null;default:ACTIVE;index" json:"state"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (KeyMeta) TableName() string { return "platform_key_meta" }

// EnsureCryptoMetaTables 确保密码元数据表存在（仅供开发/测试使用）。
func EnsureCryptoMetaTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&CA{},
		&Certificate{},
		&CSR{},
		&CRL{},
		&KeyMeta{},
	)
}
