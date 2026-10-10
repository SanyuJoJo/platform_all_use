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

	// KeyRef 旧模式：core keystore 私钥引用。新模式下为空字符串。
	KeyRef string `gorm:"size:64;not null" json:"key_ref,omitempty"`

	// KeyPath 新模式：白盒私钥路径（<pubkey_sm3>.key.pem）。
	// 与 KeyRef 二选一：
	//   - KeyRef 非空  → 旧模式，签发时走 core local 分支
	//   - KeyPath 非空 → 新模式，签发时解密白盒私钥后走 core manual 分支
	KeyPath string `gorm:"size:255" json:"key_path,omitempty"`

	ChainPath *string   `gorm:"size:255" json:"chain_path,omitempty"` // 保留兼容，新数据不写
	Serial    *string   `gorm:"size:128" json:"serial,omitempty"`
	Status    string    `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CA) TableName() string { return "platform_ca" }

// Certificate 终端证书元数据表。
type Certificate struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	CertID             string    `gorm:"size:64;not null;uniqueIndex" json:"cert_id"`
	CertType           string    `gorm:"size:32;not null;index" json:"cert_type"`
	Serial             string    `gorm:"size:128;not null;index" json:"serial"`
	Subject            string    `gorm:"size:512" json:"subject"`
	SubjectCN          string    `gorm:"size:255;not null" json:"subject_cn"`
	Issuer             string    `gorm:"size:512" json:"issuer"`
	IssuerCN           string    `gorm:"size:255" json:"issuer_cn"`
	CAID               string    `gorm:"size:64;not null;index" json:"ca_id"`
	Algorithm          string    `gorm:"size:32;not null" json:"algorithm"`
	Fingerprint        string    `gorm:"size:128" json:"fingerprint"`
	PublicKeyAlgorithm string    `gorm:"size:64" json:"public_key_algorithm"`
	SignatureAlgorithm string    `gorm:"size:64" json:"signature_algorithm"`
	NotBefore          time.Time `gorm:"not null" json:"not_before"`
	NotAfter           time.Time `gorm:"not null;index" json:"not_after"`
	CertPath           string    `gorm:"size:255;not null" json:"cert_path"`

	// KeyPath 白盒私钥路径（<pubkey_sm3>.key.pem）。可为空（证书不带私钥）。
	KeyPath *string `gorm:"size:255" json:"key_path,omitempty"`

	// KeyRef 旧模式：core keystore 私钥引用。兼容保留。
	KeyRef *string `gorm:"size:64" json:"key_ref,omitempty"`

	// ChainPath 旧的证书链路径，已废弃。
	// 保留字段用于清理历史数据，新数据不写。
	ChainPath *string `gorm:"size:255" json:"chain_path,omitempty"`

	Status    string    `gorm:"size:16;not null;default:VALID;index" json:"status"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Certificate) TableName() string { return "platform_certificate" }

// CSR P10 元数据表。
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
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"-"`
	CRLID        string     `gorm:"size:64;not null;uniqueIndex" json:"crl_id"`
	CAID         string     `gorm:"size:64;not null;index" json:"ca_id"`
	CRLPath      string     `gorm:"size:255;not null" json:"crl_path"`
	RevokedCount int        `gorm:"not null;default:0" json:"revoked_count"`
	DigestAlgo   string     `gorm:"size:16;not null" json:"digest_algorithm"`
	NextUpdate   *time.Time `json:"next_update,omitempty"`
	Status       string     `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
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

// EnvelopedKeyRecord 数字信封记录表。
type EnvelopedKeyRecord struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	SignCertID          string    `gorm:"size:64;not null;index" json:"sign_cert_id"`
	EncCertID           string    `gorm:"size:64;index" json:"enc_cert_id"`
	Format              string    `gorm:"size:16;not null;default:pkcs10;index" json:"format"`
	Algorithm           string    `gorm:"size:32;not null" json:"algorithm"`
	SignAlg             string    `gorm:"size:32;not null" json:"sign_alg"`
	EncAlg              string    `gorm:"size:32;not null" json:"enc_alg"`
	SymmetricKeyCipher  string    `gorm:"type:text;not null" json:"symmetric_key_cipher"`
	IV                  string    `gorm:"size:64;not null" json:"iv"`
	EncryptedPrivateKey string    `gorm:"type:text;not null" json:"encrypted_private_key"`
	CreatedAt           time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (EnvelopedKeyRecord) TableName() string { return "platform_enveloped_key" }

// EnsureCryptoMetaTables 确保密码元数据表存在（仅供开发/测试使用）。
func EnsureCryptoMetaTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&CA{},
		&Certificate{},
		&CSR{},
		&CRL{},
		&KeyMeta{},
		&EnvelopedKeyRecord{},
	)
}
