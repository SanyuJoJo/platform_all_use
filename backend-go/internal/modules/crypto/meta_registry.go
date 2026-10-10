package crypto

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"backend-go/internal/config"
)

// MetaRegistry 集中管理所有元数据类 Service / Handler 的构造与注册。
type MetaRegistry struct {
	CA   *CAService
	Cert *CertService
	CSR  *CSRService
	CRL  *CRLService
	Key  *KeyService
}

// NewMetaRegistry 装配所有元数据类 Service。
func NewMetaRegistry(db *gorm.DB, cfg *config.Config) *MetaRegistry {
	// ★ 启动期打印路径配置
	log.Info().
		Str("install_root", cfg.InstallRoot).
		Str("core_go_bin", cfg.CoreGoBin).
		Str("core_dispatch_path", cfg.CoreDispatchPath).
		Str("core_root", cfg.CoreRoot).
		Str("whitebox_bin", cfg.CoreGoWhiteboxBin).
		Str("openssl_bin", cfg.CoreGoOpensslBin).
		Str("cert_root", cfg.CertRoot).
		Msg("crypto: 路径配置")

	// core 二进制路径兜底
	coreBin := strings.TrimSpace(cfg.CoreGoBin)
	if coreBin == "" {
		coreBin = strings.TrimSpace(cfg.CoreDispatchPath)
	}
	if coreBin == "" && cfg.InstallRoot != "" {
		coreBin = filepath.Join(cfg.InstallRoot, "core", "bin", "core")
		log.Warn().Str("fallback", coreBin).
			Msg("crypto: CoreGoBin 为空，已从 InstallRoot 派生")
	}

	coreRoot := strings.TrimSpace(cfg.CoreRoot)
	if coreRoot == "" && cfg.InstallRoot != "" {
		coreRoot = filepath.Join(cfg.InstallRoot, "core")
		log.Warn().Str("fallback", coreRoot).
			Msg("crypto: CoreRoot 为空，已从 InstallRoot 派生")
	}

	// ★ 证书路径改造依赖：先构造 layout 与 whitebox
	layout := NewPathLayout(cfg)
	_ = layout.EnsureRoots()
	whitebox := NewWhiteboxClient(cfg.CoreGoWhiteboxBin)

	// CoreAdapter + CoreCaller
	adapter := NewCoreAdapter(coreBin, cfg.CoreTimeoutMs, cfg.CoreMaxConcurrency)
	files := NewFileStore(coreRoot)
	caller := NewCoreCaller(adapter, coreRoot, files, layout)
	parser := NewCertParser(coreRoot)
	keys := NewKeyCrypto()

	return &MetaRegistry{
		CA:   NewCAService(db, caller, files, parser, keys, layout, whitebox), // ★ 加 whitebox
		Cert: NewCertService(db, caller, files, parser, keys, layout, whitebox),
		CSR:  NewCSRService(db, caller, files, parser, keys, layout, whitebox),
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
