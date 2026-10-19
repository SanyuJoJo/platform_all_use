// backend-go/internal/modules/ssl_proxy/cert_resolver.go
package ssl_proxy

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// CertProvider SSL 代理对一期证书管理模块的抽象接口。
type CertProvider interface {
	// ★ GetCertPath 返回证书在一期磁盘上的绝对路径（不复制）
	GetCertPath(certID string) (string, error)

	// ★ GetCACertPath 返回 CA 证书在一期磁盘上的绝对路径（不复制）
	GetCACertPath(caCertID string) (string, error)

	// GetPrivateKeyPEM 返回私钥明文 PEM（必须解密，nginx 无法读白盒密文）
	GetPrivateKeyPEM(certID, keyRef string) ([]byte, error)

	// GetGMPair 返回国密双证书（签名证书 + 加密证书 PEM 内容）及私钥明文
	GetGMPair(signCertID, encCertID string) (
		signCertPEM, signKeyPEM, encCertPEM, encKeyPEM []byte,
		err error,
	)

	// SealPrivateKey 就地加密明文私钥（白盒 SM4）
	SealPrivateKey(path string) error
}

// ResolvedCertPaths 证书解析结果：全部为本地文件路径。
//
// ★ 路径来源分两类：
//   - 证书（ServerCertPath / ServerSignCert / ... / CAPath）：
//     指向一期 CertRoot 下的原始文件，不复制；
//   - 私钥（ServerKeyPath / ServerSignKey / ...）：
//     指向 core/conf/<service_id>/ 下的临时落盘文件，Seal 后加密。
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
	// Resolve 解析证书：
	//   - 证书：直接引用一期路径，不复制；
	//   - 私钥：解密后落盘到 confDir（0600 明文），供 nginx 加载；
	//   - CA：ca_domain 引用一期 bundle；ca_cert_id 引用一期单文件路径。
	Resolve(p *PlatformSSLProxy, confDir string) (*ResolvedCertPaths, error)

	// Seal 把 confDir 下所有私钥就地加密（白盒 SM4）。
	Seal(p *PlatformSSLProxy, confDir string) error

	// Cleanup 停止时删除 confDir 下的所有私钥。
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
// ★ 优化（本次调整）：
//   - 证书（.crt / .cert.pem）：直接引用一期磁盘路径，不复制；
//   - 私钥：解密落盘到 confDir（0600 明文），nginx 加载后 Seal 加密；
//   - CA：ca_domain 引用一期 bundle.crt；ca_cert_id 引用一期单文件路径。
//
// ★ 为什么私钥需要落盘：
//   nginx 无法读取白盒加密的私钥，必须由一期解密为明文 PEM 后
//   落盘到 SSL 代理自己的目录，nginx 加载到内存后立即 Seal 加密。
//   落盘目录隔离在 core/conf/<sid>/，不污染一期 CertRoot。
func (r *DefaultCertResolver) Resolve(p *PlatformSSLProxy, confDir string) (*ResolvedCertPaths, error) {
	out := &ResolvedCertPaths{}

	if err := os.MkdirAll(confDir, 0700); err != nil {
		return nil, newErr(ErrCertResolve, "创建配置目录失败", err.Error())
	}

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
			// 证书：引用一期路径
			certPath, err := r.Provider.GetCertPath(p.CertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取服务器证书路径失败", err.Error())
			}
			// 私钥：解密落盘
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
			// 证书：引用一期路径
			signCertPath, err := r.Provider.GetCertPath(p.CertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密签名证书路径失败", err.Error())
			}
			encCertPath, err := r.Provider.GetCertPath(p.EncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密加密证书路径失败", err.Error())
			}
			// 私钥：解密落盘
			_, signKeyPEM, _, encKeyPEM, err := r.Provider.GetGMPair(p.CertID, p.EncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密双证书失败", err.Error())
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

		// CA 解析
		switch {
		case p.CADomain != "":
			// 引用一期维护的 bundle.crt
			bundlePath := filepath.Join(CertRoot, "ca", p.CADomain, "bundle.crt")
			st, err := os.Stat(bundlePath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, newErr(ErrCertResolve,
						"CA 信任域 bundle 不存在", bundlePath)
				}
				return nil, newErr(ErrCertResolve,
					"访问 CA 信任域 bundle 失败", bundlePath+": "+err.Error())
			}
			if st.IsDir() || st.Size() == 0 {
				return nil, newErr(ErrCertResolve,
					"CA 信任域 bundle 为空或非法", bundlePath)
			}
			out.CAPath = bundlePath

		case p.CACertID != "":
			// 引用一期单 CA 文件路径（不复制）
			caPath, err := r.Provider.GetCACertPath(p.CACertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取 CA 证书路径失败", err.Error())
			}
			out.CAPath = caPath
		}
	}

	// ---------- 客户端代理 ----------
	if p.Mode == ModeClient {
		if p.CACertID != "" {
			caPath, err := r.Provider.GetCACertPath(p.CACertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取 CA 证书路径失败", err.Error())
			}
			out.CAPath = caPath
		}

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
			_, signKeyPEM, _, encKeyPEM, err := r.Provider.GetGMPair(p.ClientCertID, p.ClientEncCertID)
			if err != nil {
				return nil, newErr(ErrCertResolve, "获取国密客户端双证书失败", err.Error())
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

// Seal 把 confDir 下所有存在的明文私钥就地加密（白盒 SM4）。
//
// ★ 幂等：文件不存在则跳过；加密失败仅日志告警，不阻断启动
// （nginx 已加载私钥到内存，加密失败不影响当前运行）。
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
//
// 证书我们本就没复制到 confDir，所以只需清理私钥。
func (r *DefaultCertResolver) Cleanup(p *PlatformSSLProxy, confDir string) error {
	for _, name := range keyFileNames {
		_ = os.Remove(filepath.Join(confDir, name))
	}
	return nil
}

// ---------- 内部工具 ----------

// writeKey 写明文私钥到 confDir（0600），供 nginx 加载。
//
// 证书不再复制（nginx 直接引用一期 CertRoot 下的证书文件），
// 本函数是 confDir 中唯一的落盘写入点。
func writeKey(dir, name string, pem []byte) (string, error) {
	p := filepath.Join(dir, name+".key")
	if err := os.WriteFile(p, pem, 0600); err != nil {
		return "", newErr(ErrPermission, "写入私钥失败", p+": "+err.Error())
	}
	_ = os.Chmod(p, 0600)
	return p, nil
}
