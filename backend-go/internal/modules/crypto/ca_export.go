package crypto
import (
	"context"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)
type ExportFormat string
const (
	ExportFormatCert   ExportFormat = "cert"
	ExportFormatKey    ExportFormat = "key"
	ExportFormatPKCS12 ExportFormat = "pkcs12"
)
type ExportResult struct {
	Data        []byte
	Filename    string
	ContentType string
}
// Export 导出 CA：证书 / 私钥 / PKCS#12。
func (s *CAService) Export(
	ctx context.Context, caID string, format ExportFormat, password string,
) (*ExportResult, error) {
	ca, err := s.Get(caID)
	if err != nil {
		return nil, err
	}
	switch format {
	case ExportFormatCert:
		return s.exportCert(ca)
	case ExportFormatKey:
		return s.exportKey(ctx, ca)
	case ExportFormatPKCS12:
		return s.exportPKCS12(ctx, ca, password)
	}
	return nil, exception.New(
		exception.CodeParamInvalid,
		"不支持的导出格式: "+string(format), 400, nil,
	)
}
func (s *CAService) exportCert(ca *models.CA) (*ExportResult, error) {
	data, err := s.files.ReadCoreFile(ca.CertPath)
	if err != nil {
		return nil, err
	}
	return &ExportResult{
		Data:        data,
		Filename:    ca.CAID + ".pem",
		ContentType: "application/x-pem-file",
	}, nil
}
func (s *CAService) exportKey(ctx context.Context, ca *models.CA) (*ExportResult, error) {
	if ca.KeyRef == "" {
		return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥", 400, nil)
	}
	data, err := s.caller.ExportKey(ctx, ca.KeyRef)
	if err != nil {
		return nil, err
	}
	return &ExportResult{
		Data:        data,
		Filename:    ca.CAID + ".key.pem",
		ContentType: "application/x-pem-file",
	}, nil
}
func (s *CAService) exportPKCS12(
	ctx context.Context, ca *models.CA, password string,
) (*ExportResult, error) {
	if ca.KeyRef == "" {
		return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥", 400, nil)
	}
	if len(password) < 6 {
		return nil, exception.New(exception.CodeParamInvalid, "PKCS#12 密码至少 6 位", 400, nil)
	}
	data, err := s.caller.ExportPKCS12(ctx, ca.CertPath, ca.KeyRef, password)
	if err != nil {
		return nil, err
	}
	return &ExportResult{
		Data:        data,
		Filename:    ca.CAID + ".p12",
		ContentType: "application/x-pkcs12",
	}, nil
}
