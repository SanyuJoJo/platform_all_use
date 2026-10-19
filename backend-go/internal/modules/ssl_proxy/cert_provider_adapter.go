// backend-go/internal/modules/ssl_proxy/cert_provider_adapter.go
package ssl_proxy

import (
	"context"
	"fmt"

	"backend-go/internal/modules/crypto"
)

// CryptoCertProvider 基于一期证书管理模块（crypto 包）的适配器。
type CryptoCertProvider struct {
	caSvc   *crypto.CAService
	certSvc *crypto.CertService
}

func NewCryptoCertProvider(caSvc *crypto.CAService, certSvc *crypto.CertService) *CryptoCertProvider {
	return &CryptoCertProvider{caSvc: caSvc, certSvc: certSvc}
}

// ★ GetCertPath 返回证书在一期磁盘上的绝对路径，供 nginx ssl_certificate 直接引用。
//
// 证书是公开数据，无需复制到 SSL 代理目录。
func (p *CryptoCertProvider) GetCertPath(certID string) (string, error) {
	if certID == "" {
		return "", fmt.Errorf("cert_id 为空")
	}
	return p.certSvc.GetCertPath(certID)
}

// ★ GetCACertPath 返回指定 CA 证书在一期磁盘上的绝对路径。
//
// 用于 ca_cert_id 场景（无 ca_domain），直接引用一期磁盘文件，不复制。
func (p *CryptoCertProvider) GetCACertPath(caCertID string) (string, error) {
	if caCertID == "" {
		return "", fmt.Errorf("ca_cert_id 为空")
	}
	return p.caSvc.GetCAPath(caCertID)
}

// GetServerCertPEM 返回服务器证书 PEM 内容。
//
// 保留此方法供部分场景（如需要 PEM 字节而非路径）使用；本模块 Resolve 已改为
// 优先用 GetCertPath，仅路径引用。
func (p *CryptoCertProvider) GetServerCertPEM(certID string) ([]byte, error) {
	return p.fetchCertPEM(certID, "服务器证书")
}

// GetClientCertPEM 返回客户端证书 PEM 内容（保留）。
func (p *CryptoCertProvider) GetClientCertPEM(certID string) ([]byte, error) {
	return p.fetchCertPEM(certID, "客户端证书")
}

// GetPrivateKeyPEM 返回私钥明文 PEM。
//
// 一期 CertService.Export("key") 内部会白盒解密（KeyPath）或从 keystore 导出（KeyRef）。
// 返回值用于 SSL 代理临时落盘到 core/conf/<sid>/server.key；Seal 后重新加密。
func (p *CryptoCertProvider) GetPrivateKeyPEM(certID, keyRef string) ([]byte, error) {
	if certID == "" && keyRef == "" {
		return nil, fmt.Errorf("cert_id 与 key_ref 均为空")
	}
	_ = keyRef
	res, err := p.certSvc.Export(context.Background(), certID, "key", "")
	if err != nil {
		return nil, fmt.Errorf("导出证书 %s 私钥失败: %w", certID, err)
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("证书 %s 未关联可用私钥", certID)
	}
	return res.Data, nil
}

// GetCACertPEM 返回指定 CA 证书的 PEM。
//
// 保留兜底用途；本模块 Resolve 已改为优先用 GetCACertPath。
func (p *CryptoCertProvider) GetCACertPEM(caCertID string) ([]byte, error) {
	if caCertID == "" {
		return nil, fmt.Errorf("ca_cert_id 为空")
	}
	ca, err := p.caSvc.Get(caCertID)
	if err != nil {
		return nil, fmt.Errorf("查询 CA %s 失败: %w", caCertID, err)
	}
	data, err := p.caSvc.ReadCACert(ca)
	if err != nil {
		return nil, fmt.Errorf("读取 CA %s 证书失败: %w", caCertID, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("CA %s 证书内容为空", caCertID)
	}
	return data, nil
}

// GetGMPair 返回国密双证书（签名证书 + 加密证书 PEM）及其私钥明文。
//
// 返回的 signCertPEM / encCertPEM 部分调用方（旧版 Resolve）可能用到；
// 新 Resolve 用 GetCertPath 拿路径，只取私钥部分。
func (p *CryptoCertProvider) GetGMPair(signCertID, encCertID string) (
	signCertPEM, signKeyPEM, encCertPEM, encKeyPEM []byte,
	err error,
) {
	if signCertID == "" {
		return nil, nil, nil, nil, fmt.Errorf("签名证书 ID 为空")
	}
	if encCertID == "" {
		return nil, nil, nil, nil, fmt.Errorf("加密证书 ID 为空")
	}
	signCertPEM, err = p.fetchCertPEM(signCertID, "国密签名证书")
	if err != nil {
		return
	}
	signKeyPEM, err = p.GetPrivateKeyPEM(signCertID, "")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("读取签名私钥失败: %w", err)
	}
	encCertPEM, err = p.fetchCertPEM(encCertID, "国密加密证书")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	encKeyPEM, err = p.GetPrivateKeyPEM(encCertID, "")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("读取加密私钥失败: %w", err)
	}
	return signCertPEM, signKeyPEM, encCertPEM, encKeyPEM, nil
}

// ★ SealPrivateKey 就地加密明文私钥（白盒 SM4）。
func (p *CryptoCertProvider) SealPrivateKey(path string) error {
	return p.certSvc.SealPrivateKey(path)
}

func (p *CryptoCertProvider) fetchCertPEM(certID, label string) ([]byte, error) {
	if certID == "" {
		return nil, fmt.Errorf("%s ID 为空", label)
	}
	res, err := p.certSvc.Export(context.Background(), certID, "cert", "")
	if err != nil {
		return nil, fmt.Errorf("获取%s失败: %w", label, err)
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("%s %s 内容为空", label, certID)
	}
	return res.Data, nil
}
