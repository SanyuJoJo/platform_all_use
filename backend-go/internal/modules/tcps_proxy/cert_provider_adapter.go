// backend-go/internal/modules/tcps_proxy/cert_provider_adapter.go
package tcps_proxy
import (
	"context"
	"fmt"
	"backend-go/internal/modules/crypto"
)
type CryptoCertProvider struct {
	caSvc   *crypto.CAService
	certSvc *crypto.CertService
}
func NewCryptoCertProvider(caSvc *crypto.CAService, certSvc *crypto.CertService) *CryptoCertProvider {
	return &CryptoCertProvider{caSvc: caSvc, certSvc: certSvc}
}
func (p *CryptoCertProvider) GetCertPath(certID string) (string, error) {
	if certID == "" {
		return "", fmt.Errorf("cert_id 为空")
	}
	return p.certSvc.GetCertPath(certID)
}
func (p *CryptoCertProvider) GetCACertPath(caCertID string) (string, error) {
	if caCertID == "" {
		return "", fmt.Errorf("ca_cert_id 为空")
	}
	return p.caSvc.GetCAPath(caCertID)
}
func (p *CryptoCertProvider) GetPrivateKeyPEM(certID, keyRef string) ([]byte, error) {
	if certID == "" && keyRef == "" {
		return nil, fmt.Errorf("cert_id 与 key_ref 均为空")
	}
	res, err := p.certSvc.Export(context.Background(), certID, "key", "")
	if err != nil {
		return nil, fmt.Errorf("导出证书 %s 私钥失败: %w", certID, err)
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("证书 %s 未关联可用私钥", certID)
	}
	return res.Data, nil
}
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
func (p *CryptoCertProvider) GetGMPair(signCertID, encCertID string) (
	signCertPEM, signKeyPEM, encCertPEM, encKeyPEM []byte,
	err error,
) {
	if signCertID == "" || encCertID == "" {
		return nil, nil, nil, nil, fmt.Errorf("国密证书 ID 为空")
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
		return
	}
	encKeyPEM, err = p.GetPrivateKeyPEM(encCertID, "")
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("读取加密私钥失败: %w", err)
	}
	return
}
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
