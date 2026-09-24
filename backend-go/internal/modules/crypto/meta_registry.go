package crypto

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"backend-go/internal/config"
)

// MetaRegistry 集中管理所有元数据类 Service / Handler 的构造与注册。
//
// 装配顺序：
//   coreRoot → CoreAdapter → FileStore → CoreCaller → CertParser → KeyCrypto
//   → 各领域 Service → 各领域 Handler
type MetaRegistry struct {
	CA   *CAService
	Cert *CertService
	CSR  *CSRService
	CRL  *CRLService
	Key  *KeyService
}

// NewMetaRegistry 装配所有元数据类 Service。
func NewMetaRegistry(db *gorm.DB, cfg *config.Config) *MetaRegistry {
	coreRoot := coreRootOf(cfg.CoreDispatchPath)

	adapter := NewCoreAdapter(
		cfg.CoreDispatchPath,
		cfg.CoreTimeoutMs,
		cfg.CoreMaxConcurrency,
	)
	files := NewFileStore(coreRoot)
	caller := NewCoreCaller(adapter, coreRoot, files)
	parser := NewCertParser(coreRoot)
	keys := NewKeyCrypto()

	return &MetaRegistry{
		CA:   NewCAService(db, caller, files, parser, keys),
		Cert: NewCertService(db, caller, files, parser, keys),
		CSR:  NewCSRService(db),
		CRL:  NewCRLService(db),
		Key:  NewKeyService(db),
	}
}

// Register 注册所有元数据类路由。
func (r *MetaRegistry) Register(g *gin.RouterGroup) {
	NewCAHandler(r.CA).RegisterRoutes(g)
	NewCertHandler(r.Cert).RegisterRoutes(g)
	NewCSRHandler(r.CSR).RegisterRoutes(g)
	NewCRLHandler(r.CRL).RegisterRoutes(g)
	NewKeyHandler(r.Key).RegisterRoutes(g)
}
