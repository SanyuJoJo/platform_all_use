// backend-go/internal/modules/ssl_proxy/cert_resolver.go
package ssl_proxy

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// CertProvider SSL 代理对一期证书管理模块的抽象接口。
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

// ResolvedCertPaths 证书解析结果：全部为本地文件路径。
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

// CertResolver 证书解析器。
type CertResolver interface {
	Resolve(p *PlatformSSLProxy, confDir string) (*ResolvedCertPaths, error)
	Seal(p *PlatformSSLProxy, confDir string) error
	Cleanup(p *PlatformSSLProxy, confDir string) error
}

// DefaultCertResolver 默认实现。
type DefaultCertResolver struct {
	Provider CertProvider
}

func NewCertResolver(provider CertProvider) *DefaultCertResolver {
	return &DefaultCertResolver{Provider: provider}
}

// keyFileNames 所有可能的私钥文件名。
var keyFileNames = []string{
	"server.key",
	"server_sign.key", "server_enc.key",
	"client.key",
	"client_sign.key", "client_enc.key",
}

// Resolve 按 mode / protocol 解析证书并落盘到 confDir。
//
// ★ 本版修复：**客户端模式也解析 ca_domain**（此前只处理 ca_cert_id，
// 导致 ca_domain=safe 时 CAPath 为空，渲染出 `proxy_ssl_trusted_certificate ;`）。
func (r *DefaultCertResolver) Resolve(p *PlatformSSLProxy, confDir string) (*ResolvedCertPaths, error) {
	out := &ResolvedCertPaths{}

	if err := os.MkdirAll(confDir, 0700); err != nil {
		return nil, newErr(ErrCertResolve, "创建配置目录失败", err.Error())
	}

	// 统一解析 CA（服务端 / 客户端都用这个）
	caPath, err := r.resolveCAPath(p)
	if err != nil {
		return nil, err
	}
	out.CAPath = caPath

	// ---------- 服务端代理 ----------
	if p.Mode == ModeServer {
		if !p.ListenSSL {
			return out, nil
		}

		switch p.Protocol {
		case ProtocolHTTP, ProtocolPQC:
			if p.CertID == "" {
				return nil, newErr(ErrCertResolve, "服务端证书缺失", "cert_id 必填")
			}
			certPath, err := r.Provider.GetCertPath(p.CertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取服务器证书路径失败", err.Error())
			}
			keyPEM, err := r.Provider.GetPrivateKeyPEM(p.CertID, "")
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取服务器私钥失败", err.Error())
			}
			kp, err := writeKey(confDir, "server", keyPEM)
			if err != nil {
				return nil, err
			}
			out.ServerCertPath = certPath
			out.ServerKeyPath = kp

		case ProtocolGMSSL:
			if p.CertID == "" {
				return nil, newErr(ErrCertResolve, "国密签名证书缺失", "cert_id 必填")
			}
			if p.EncCertID == "" {
				return nil, newErr(ErrCertResolve, "国密加密证书缺失", "enc_cert_id 必填")
			}
			signCertPath, err := r.Provider.GetCertPath(p.CertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密签名证书路径失败", err.Error())
			}
			encCertPath, err := r.Provider.GetCertPath(p.EncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密加密证书路径失败", err.Error())
			}
			signCertPEM, signKeyPEM, encCertPEM, encKeyPEM, err := r.Provider.GetGMPair(p.CertID, p.EncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密双证书失败", err.Error())
			}
			if err := validateGMCertUsage(signCertPEM, "国密签名证书", true); err != nil {
				return nil, err
			}
			if err := validateGMCertUsage(encCertPEM, "国密加密证书", false); err != nil {
				return nil, err
			}
			sk, err := writeKey(confDir, "server_sign", signKeyPEM)
			if err != nil {
				return nil, err
			}
			ek, err := writeKey(confDir, "server_enc", encKeyPEM)
			if err != nil {
				return nil, err
			}
			out.ServerSignCert, out.ServerSignKey = signCertPath, sk
			out.ServerEncCert, out.ServerEncKey = encCertPath, ek

			// 国际 + 国密双栈
			if p.IntlCertID != "" {
				intlCertPath, err := r.Provider.GetCertPath(p.IntlCertID)
				if err != nil {
					return nil, newErr(ErrCertResolve, "获取国际服务器证书路径失败", err.Error())
				}
				intlKeyPEM, err := r.Provider.GetPrivateKeyPEM(p.IntlCertID, "")
				if err != nil {
					return nil, newErr(ErrCertResolve, "获取国际服务器私钥失败", err.Error())
				}
				intlKeyPath, err := writeKey(confDir, "server", intlKeyPEM)
				if err != nil {
					return nil, err
				}
				out.ServerCertPath, out.ServerKeyPath = intlCertPath, intlKeyPath
			}

		default:
			return nil, newErr(ErrProtocolUnsup, "未知协议", p.Protocol)
		}
	}

	// ---------- 客户端代理 ----------
	if p.Mode == ModeClient {
		switch p.Protocol {
		case ProtocolHTTP, ProtocolPQC:
			if p.ClientCertID != "" {
				certPath, err := r.Provider.GetCertPath(p.ClientCertID)
				if err != nil {
					return nil, newErr(ErrCertResolve, "获取客户端证书路径失败", err.Error())
				}
				keyPEM, err := r.Provider.GetPrivateKeyPEM(p.ClientCertID, p.ClientKeyRef)
				if err != nil {
					return nil, newErr(ErrCertResolve, "获取客户端私钥失败", err.Error())
				}
				kp, err := writeKey(confDir, "client", keyPEM)
				if err != nil {
					return nil, err
				}
				out.ClientCertPath, out.ClientKeyPath = certPath, kp
			}

		case ProtocolGMSSL:
			if p.ClientCertID == "" {
				return nil, newErr(ErrCertResolve, "国密客户端签名证书缺失", "client_cert_id 必填")
			}
			if p.ClientEncCertID == "" {
				return nil, newErr(ErrCertResolve, "国密客户端加密证书缺失", "client_enc_cert_id 必填")
			}
			signCertPath, err := r.Provider.GetCertPath(p.ClientCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密客户端签名证书路径失败", err.Error())
			}
			encCertPath, err := r.Provider.GetCertPath(p.ClientEncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密客户端加密证书路径失败", err.Error())
			}
			signCertPEM, signKeyPEM, encCertPEM, encKeyPEM, err := r.Provider.GetGMPair(p.ClientCertID, p.ClientEncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密客户端双证书失败", err.Error())
			}
			if err := validateGMCertUsage(signCertPEM, "国密客户端签名证书", true); err != nil {
				return nil, err
			}
			if err := validateGMCertUsage(encCertPEM, "国密客户端加密证书", false); err != nil {
				return nil, err
			}
			sk, err := writeKey(confDir, "client_sign", signKeyPEM)
			if err != nil {
				return nil, err
			}
			ek, err := writeKey(confDir, "client_enc", encKeyPEM)
			if err != nil {
				return nil, err
			}
			out.ClientSignCert, out.ClientSignKey = signCertPath, sk
			out.ClientEncCert, out.ClientEncKey = encCertPath, ek
		}
	}

	return out, nil
}

// resolveCAPath 统一解析 CA 路径（服务端 / 客户端共用）。
//
// 优先级：
//  1. ca_domain（空值时兜底 "safe"）→ <CertRoot>/ca/<domain>/bundle.crt
//  2. ca_cert_id → 单 CA 文件路径
//  3. 都没有 → 返回空串（调用方决定是否降级）
//
// ★ 修复要点：此前客户端分支只处理 ca_cert_id，不处理 ca_domain，
// 导致 ca_domain=safe 时 CAPath 为空，模板渲染出空路径的
// `proxy_ssl_trusted_certificate ;`，nginx -t 必然失败。
func (r *DefaultCertResolver) resolveCAPath(p *PlatformSSLProxy) (string, error) {
	domain := p.CADomain
	if domain == "" {
		// 与 Service 层默认值一致：空值视为 safe
		domain = "safe"
	}

	// 1. ca_domain（含默认 safe）
	if domain != "" {
		bundlePath := filepath.Join(CertRoot, "ca", domain, "bundle.crt")
		st, err := os.Stat(bundlePath)
		if err != nil {
			if os.IsNotExist(err) {
				// bundle 不存在：不阻断，返回空让下游降级，
				// 但打 Warn 提示用户去 CA 管理里补
				log.Warn().
					Str("domain", domain).
					Str("bundle", bundlePath).
					Msg("ssl_proxy: 信任域 bundle 不存在，CA 校验将不可用")
				return "", nil
			}
			return "", newErr(ErrCertResolve,
				"访问 CA 信任域 bundle 失败", bundlePath+": "+err.Error())
		}
		if st.IsDir() || st.Size() == 0 {
			return "", newErr(ErrCertResolve,
				"CA 信任域 bundle 为空或非法", bundlePath)
		}
		return bundlePath, nil
	}

	// 2. ca_cert_id 兜底
	if p.CACertID != "" {
		caPath, err := r.Provider.GetCACertPath(p.CACertID)
		if err != nil {
			return "", newErr(ErrCertResolve, "获取 CA 证书路径失败", err.Error())
		}
		return caPath, nil
	}

	return "", nil
}

// Seal 把 confDir 下所有存在的明文私钥就地加密。
func (r *DefaultCertResolver) Seal(p *PlatformSSLProxy, confDir string) error {
	var sealed int
	for _, name := range keyFileNames {
		path := filepath.Join(confDir, name)
		st, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Warn().Err(err).Str("path", path).
				Msg("ssl_proxy: Seal 检查私钥失败，跳过")
			continue
		}
		if st.IsDir() || st.Size() == 0 {
			continue
		}
		if err := r.Provider.SealPrivateKey(path); err != nil {
			log.Error().Err(err).Str("path", path).
				Msg("ssl_proxy: Seal 私钥加密失败（nginx 已加载，不阻断）")
			continue
		}
		sealed++
	}
	if sealed > 0 {
		log.Info().
			Str("service_id", p.ServiceID).
			Int("sealed", sealed).
			Msg("ssl_proxy: 启动后私钥已重新加密")
	}
	return nil
}

// Cleanup 停止后删除 confDir 下所有私钥。
func (r *DefaultCertResolver) Cleanup(p *PlatformSSLProxy, confDir string) error {
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

// validateGMCertUsage 校验国密证书的 X509 KeyUsage 是否满足 tongsuo 的加载要求。
//
// ★ 背景：tongsuo 的 SSL_CTX_use_enc_certificate() 会直接拒绝缺少加密用途
// （keyEncipherment / dataEncipherment）的证书，报错为：
//
//	SSL: error:0A0C0101:SSL routines::called a function you should not call
//
// 从 nginx -t 输出完全看不出根因，因此这里提前给出可读错误。
//
// ★ 说明：国密证书（SM2 曲线）无法用 crypto/x509 完整解析（报
// unsupported elliptic curve），所以这里只按 ASN.1 取出 keyUsage 扩展
// （OID 2.5.29.15）的 BIT STRING 做位判断。
//
// 签名证书与加密证书允许是同一张（tongsuo 不做限制），只要该证书的
// KeyUsage 同时包含两类用途 —— 这正是「申请证书」时同时勾选
// 签名 + 加密 的语义。
func validateGMCertUsage(certPEM []byte, label string, wantSign bool) error {
	const (
		kuDigitalSignature = 0
		kuKeyEncipherment  = 2
		kuDataEncipherment = 3
		kuKeyCertSign      = 5
		kuCRLSign          = 6
	)
	bits, present, err := certKeyUsageBits(certPEM)
	if err != nil {
		return newErr(ErrCertResolve, label+"解析失败", err.Error())
	}
	// ★ 与 OpenSSL 的 X509_get_key_usage() 保持一致：证书未声明
	//   keyUsage 扩展时视为「不限制」（返回全 1），
	//   tongsuo 不会拒绝，因此这里也不做限制。
	if !present {
		return nil
	}
	has := func(idx int) bool { return bits.At(idx) == 1 }

	if wantSign {
		if !has(kuDigitalSignature) && !has(kuKeyCertSign) && !has(kuCRLSign) {
			return newErr(ErrCertResolve, label+"缺少签名用途",
				"证书 KeyUsage 需包含 digitalSignature（当前证书未声明该用途）")
		}
		return nil
	}
	if !has(kuKeyEncipherment) && !has(kuDataEncipherment) {
		return newErr(ErrCertResolve, label+"缺少加密用途",
			"证书 KeyUsage 需包含 keyEncipherment 或 dataEncipherment，否则 nginx(tongsuo) "+
				"加载时报 SSL_CTX_use_enc_certificate failed；"+
				"请在证书管理中重新申请同时勾选「签名 + 加密」的证书")
	}
	return nil
}

// certKeyUsageBits 从 PEM 证书中取出 keyUsage 扩展的 BIT STRING。
//
// 只做最基本的 ASN.1 结构解析，回避 crypto/x509 对 SM2 曲线的不支持问题。
// present=false 表示证书未声明 keyUsage 扩展（此时 tongsuo 的
// X509_get_key_usage() 返回 0，任何用途都不可用）。
func certKeyUsageBits(certPEM []byte) (asn1.BitString, bool, error) {
	var empty asn1.BitString

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return empty, false, fmt.Errorf("证书不是有效的 PEM 格式")
	}

	var cert struct {
		TBSCertificate     asn1.RawValue
		SignatureAlgorithm asn1.RawValue
		SignatureValue     asn1.BitString
	}
	if _, err := asn1.Unmarshal(block.Bytes, &cert); err != nil {
		return empty, false, fmt.Errorf("证书 ASN.1 解析失败: %w", err)
	}

	// TBSCertificate 为 SEQUENCE，按元素拆开：extensions 是 [3] EXPLICIT。
	var fields []asn1.RawValue
	if _, err := asn1.Unmarshal(cert.TBSCertificate.FullBytes, &fields); err != nil {
		return empty, false, fmt.Errorf("证书 TBSCertificate 解析失败: %w", err)
	}

	kuOID := asn1.ObjectIdentifier{2, 5, 29, 15}
	for _, f := range fields {
		if f.Class != asn1.ClassContextSpecific || f.Tag != 3 {
			continue
		}
		var exts []pkix.Extension
		if _, err := asn1.Unmarshal(f.Bytes, &exts); err != nil {
			return empty, false, fmt.Errorf("证书扩展解析失败: %w", err)
		}
		for _, ext := range exts {
			if !ext.Id.Equal(kuOID) {
				continue
			}
			var bits asn1.BitString
			if _, err := asn1.Unmarshal(ext.Value, &bits); err != nil {
				return empty, false, fmt.Errorf("keyUsage 扩展解析失败: %w", err)
			}
			return bits, true, nil
		}
		return empty, false, nil
	}
	return empty, false, nil
}
