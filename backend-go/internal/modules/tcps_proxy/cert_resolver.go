// backend-go/internal/modules/tcps_proxy/cert_resolver.go
package tcps_proxy

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// CertProvider 证书模块抽象接口。
type CertProvider interface {
	GetCertPath(certID string) (string, error)
	GetCACertPath(caCertID string) (string, error)
	GetPrivateKeyPEM(certID, keyRef string) ([]byte, error)
	GetGMPair(signCertID, encCertID string) (
		signCertPEM, signKeyPEM, encCertPEM, encKeyPEM []byte,
		err error,
	)
	SealPrivateKey(path string) error
}

// ResolvedCertPaths 解析后的证书路径集合。
type ResolvedCertPaths struct {
	ServerCertPath string
	ServerKeyPath  string
	CAPath         string
	ClientCertPath string
	ClientKeyPath  string

	ServerSignCert string
	ServerSignKey  string
	ServerEncCert  string
	ServerEncKey   string
	ClientSignCert string
	ClientSignKey  string
	ClientEncCert  string
	ClientEncKey   string
}

// CertResolver 证书解析器接口。
type CertResolver interface {
	Resolve(p *PlatformTCPSProxy, confDir string) (*ResolvedCertPaths, error)
	Seal(p *PlatformTCPSProxy, confDir string) error
	Cleanup(p *PlatformTCPSProxy, confDir string) error
}

// DefaultCertResolver 默认实现。
type DefaultCertResolver struct {
	Provider CertProvider
}

// NewCertResolver 构造。
func NewCertResolver(provider CertProvider) *DefaultCertResolver {
	return &DefaultCertResolver{Provider: provider}
}

// keyFileNames 所有可能的私钥文件名。
var keyFileNames = []string{
	"server.key", "server_sign.key", "server_enc.key",
	"client.key", "client_sign.key", "client_enc.key",
}

// Resolve 按 mode / protocol 解析证书并落盘到 confDir。
//
// ★ 本版修复（与 HTTP 同步）：
//   - 服务端 / 客户端统一调用 resolveCAPath，客户端也处理 ca_domain；
//   - 此前客户端分支只处理 ca_cert_id，导致 ca_domain=safe 时 CAPath 为空，
//     模板渲染出 `proxy_ssl_trusted_certificate ;`（nginx -t 必失败）。
func (r *DefaultCertResolver) Resolve(p *PlatformTCPSProxy, confDir string) (*ResolvedCertPaths, error) {
	out := &ResolvedCertPaths{}

	if err := os.MkdirAll(confDir, 0700); err != nil {
		return nil, newErr(ErrCertResolve, "创建配置目录失败", err.Error())
	}

	// 统一解析 CA（服务端 / 客户端共用）
	caPath, err := r.resolveCAPath(p)
	if err != nil {
		return nil, err
	}
	out.CAPath = caPath

	if p.Mode == ModeServer {
		if err := r.resolveServer(p, out, confDir); err != nil {
			return nil, err
		}
	} else {
		if err := r.resolveClient(p, out, confDir); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// resolveServer 服务端模式证书解析。
func (r *DefaultCertResolver) resolveServer(p *PlatformTCPSProxy, out *ResolvedCertPaths, confDir string) error {
	if !p.IsListenSSL() {
		return nil
	}

	switch p.Protocol {
	case ProtocolTLS, ProtocolPQC:
		if p.CertID == "" {
			return newErr(ErrCertResolve, "服务端证书缺失", "cert_id 必填")
		}
		certPath, err := r.Provider.GetCertPath(p.CertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取服务器证书路径失败", err.Error())
		}
		keyPEM, err := r.Provider.GetPrivateKeyPEM(p.CertID, "")
		if err != nil {
			return newErr(ErrCertResolve, "获取服务器私钥失败", err.Error())
		}
		kp, err := writeKey(confDir, "server", keyPEM)
		if err != nil {
			return err
		}
		out.ServerCertPath = certPath
		out.ServerKeyPath = kp

	case ProtocolGMSSL:
		if p.CertID == "" {
			return newErr(ErrCertResolve, "国密签名证书缺失", "cert_id 必填")
		}
		if p.EncCertID == "" {
			return newErr(ErrCertResolve, "国密加密证书缺失", "enc_cert_id 必填")
		}
		signCertPath, err := r.Provider.GetCertPath(p.CertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取国密签名证书路径失败", err.Error())
		}
		encCertPath, err := r.Provider.GetCertPath(p.EncCertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取国密加密证书路径失败", err.Error())
		}
		_, signKeyPEM, _, encKeyPEM, err := r.Provider.GetGMPair(p.CertID, p.EncCertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取国密双证书失败", err.Error())
		}
		sk, err := writeKey(confDir, "server_sign", signKeyPEM)
		if err != nil {
			return err
		}
		ek, err := writeKey(confDir, "server_enc", encKeyPEM)
		if err != nil {
			return err
		}
		out.ServerSignCert, out.ServerSignKey = signCertPath, sk
		out.ServerEncCert, out.ServerEncKey = encCertPath, ek

		// 国际 + 国密双栈
		if p.IntlCertID != "" {
			intlCertPath, err := r.Provider.GetCertPath(p.IntlCertID)
			if err != nil {
				return newErr(ErrCertResolve, "获取国际服务器证书路径失败", err.Error())
			}
			intlKeyPEM, err := r.Provider.GetPrivateKeyPEM(p.IntlCertID, "")
			if err != nil {
				return newErr(ErrCertResolve, "获取国际服务器私钥失败", err.Error())
			}
			intlKP, err := writeKey(confDir, "server", intlKeyPEM)
			if err != nil {
				return err
			}
			out.ServerCertPath, out.ServerKeyPath = intlCertPath, intlKP
		}
	}
	return nil
}

// resolveClient 客户端模式证书解析。
func (r *DefaultCertResolver) resolveClient(p *PlatformTCPSProxy, out *ResolvedCertPaths, confDir string) error {
	if !p.IsBackendTLS() {
		return nil
	}

	switch p.Protocol {
	case ProtocolTLS, ProtocolPQC:
		if p.ClientCertID != "" {
			certPath, err := r.Provider.GetCertPath(p.ClientCertID)
			if err != nil {
				return newErr(ErrCertResolve, "获取客户端证书路径失败", err.Error())
			}
			keyPEM, err := r.Provider.GetPrivateKeyPEM(p.ClientCertID, p.ClientKeyRef)
			if err != nil {
				return newErr(ErrCertResolve, "获取客户端私钥失败", err.Error())
			}
			kp, err := writeKey(confDir, "client", keyPEM)
			if err != nil {
				return err
			}
			out.ClientCertPath, out.ClientKeyPath = certPath, kp
		}

	case ProtocolGMSSL:
		if p.ClientCertID == "" || p.ClientEncCertID == "" {
			return newErr(ErrCertResolve,
				"国密客户端双证书缺失", "client_cert_id 与 client_enc_cert_id 必填")
		}
		signCertPath, err := r.Provider.GetCertPath(p.ClientCertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取国密客户端签名证书路径失败", err.Error())
		}
		encCertPath, err := r.Provider.GetCertPath(p.ClientEncCertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取国密客户端加密证书路径失败", err.Error())
		}
		_, signKeyPEM, _, encKeyPEM, err := r.Provider.GetGMPair(p.ClientCertID, p.ClientEncCertID)
		if err != nil {
			return newErr(ErrCertResolve, "获取国密客户端双证书失败", err.Error())
		}
		sk, err := writeKey(confDir, "client_sign", signKeyPEM)
		if err != nil {
			return err
		}
		ek, err := writeKey(confDir, "client_enc", encKeyPEM)
		if err != nil {
			return err
		}
		out.ClientSignCert, out.ClientSignKey = signCertPath, sk
		out.ClientEncCert, out.ClientEncKey = encCertPath, ek
	}
	return nil
}

// resolveCAPath 统一解析 CA 路径（服务端 / 客户端共用）。
//
// 优先级：
//  1. ca_domain（空值兜底 "safe"）→ <CertRoot>/ca/<domain>/bundle.crt
//  2. ca_cert_id → 单 CA 文件路径
//  3. bundle 不存在 / 无 ca_cert_id → 返回空串（降级，不阻断）
//
// ★ 本函数是本次修复的核心：之前客户端分支完全不处理 ca_domain，
// 导致 ca_domain=safe 时 CAPath 空，模板生成 `proxy_ssl_trusted_certificate ;`。
func (r *DefaultCertResolver) resolveCAPath(p *PlatformTCPSProxy) (string, error) {
	domain := strings.TrimSpace(p.CADomain)
	if domain == "" {
		domain = "safe"
	}

	bundlePath := filepath.Join(CertRoot(), "ca", domain, "bundle.crt")
	st, err := os.Stat(bundlePath)
	if err == nil && !st.IsDir() && st.Size() > 0 {
		return bundlePath, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", newErr(ErrCertResolve,
			"访问 CA 信任域 bundle 失败", bundlePath+": "+err.Error())
	}
	if err == nil && (st.IsDir() || st.Size() == 0) {
		return "", newErr(ErrCertResolve,
			"CA 信任域 bundle 为空或非法", bundlePath)
	}

	// bundle 不存在 → 尝试 ca_cert_id
	if p.CACertID != "" {
		caPath, err := r.Provider.GetCACertPath(p.CACertID)
		if err != nil {
			return "", newErr(ErrCertResolve, "获取 CA 证书路径失败", err.Error())
		}
		return caPath, nil
	}

	log.Warn().
		Str("domain", domain).
		Str("bundle", bundlePath).
		Msg("tcps_proxy: 信任域 bundle 不存在，CA 校验降级为不加载")
	return "", nil
}

// Seal 把 confDir 下所有存在的明文私钥就地加密。
func (r *DefaultCertResolver) Seal(p *PlatformTCPSProxy, confDir string) error {
	var sealed int
	for _, name := range keyFileNames {
		path := filepath.Join(confDir, name)
		st, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Warn().Err(err).Str("path", path).
				Msg("tcps: Seal 检查私钥失败，跳过")
			continue
		}
		if st.IsDir() || st.Size() == 0 {
			continue
		}
		if err := r.Provider.SealPrivateKey(path); err != nil {
			log.Error().Err(err).Str("path", path).
				Msg("tcps: Seal 私钥加密失败（nginx 已加载，不阻断）")
			continue
		}
		sealed++
	}
	if sealed > 0 {
		log.Info().
			Str("service_id", p.ServiceID).
			Int("sealed", sealed).
			Msg("tcps: 启动后私钥已重新加密")
	}
	return nil
}

// Cleanup 停止后删除 confDir 下所有私钥。
func (r *DefaultCertResolver) Cleanup(p *PlatformTCPSProxy, confDir string) error {
	for _, name := range keyFileNames {
		_ = os.Remove(filepath.Join(confDir, name))
	}
	return nil
}

// writeKey 写明文私钥到 confDir（0600）。
func writeKey(dir, name string, pem []byte) (string, error) {
	p := filepath.Join(dir, name+".key")
	if err := os.WriteFile(p, pem, 0600); err != nil {
		return "", newErr(ErrPermission, "写入私钥失败", p+": "+err.Error())
	}
	_ = os.Chmod(p, 0600)
	return p, nil
}
