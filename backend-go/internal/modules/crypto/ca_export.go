package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// exportCert 导出 CA 证书 PEM。
//
// ★ CA 证书存放在 <cert_root>/ca/<domain>/（新路径规范），不在 core/data|tmp，
// 因此不能走 FileStore（其守卫只允许 core/data|tmp），必须用 resolveCAPath。
// 该路径解析同时兼容：绝对路径、相对 cert_root、相对 core_root（旧规范）。
func (s *CAService) exportCert(ca *models.CA) (*ExportResult, error) {
	data, err := s.ReadCACert(ca)
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
	var data []byte
	switch {
	case ca.KeyRef != "":
		d, err := s.caller.ExportKey(ctx, ca.KeyRef)
		if err != nil {
			return nil, err
		}
		data = d
	case ca.KeyPath != "":
		d, err := s.readCAKeyPEM(ctx, ca)
		if err != nil {
			return nil, err
		}
		data = d
	default:
		return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥", 400, nil)
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
	if ca.KeyRef == "" && ca.KeyPath == "" {
		return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥", 400, nil)
	}
	if len(password) < 6 {
		return nil, exception.New(exception.CodeParamInvalid, "PKCS#12 密码至少 6 位", 400, nil)
	}
	certAbs, err := s.resolveCAPath(ca.CertPath)
	if err != nil {
		return nil, err
	}

	var keyPEM []byte
	if ca.KeyRef != "" {
		keyPEM, err = s.caller.ExportKey(ctx, ca.KeyRef)
	} else {
		keyPEM, err = s.readCAKeyPEM(ctx, ca)
	}
	if err != nil {
		return nil, err
	}

	data, err := buildPKCS12(
		s.parser.OpensslBin(), certAbs, keyPEM, ca.SubjectCN, ca.CAID, password,
	)
	if err != nil {
		return nil, err
	}
	return &ExportResult{
		Data:        data,
		Filename:    ca.CAID + ".p12",
		ContentType: "application/x-pkcs12",
	}, nil
}

// readCAKeyPEM 读取 CA 私钥明文。
//
// ★ 新路径规范下 CA 私钥以白盒密文落盘（ca.KeyPath），需要先解密到临时文件；
// 旧规范 CA 私钥在 core keystore（ca.KeyRef），由 CoreCaller.ExportKey 处理。
func (s *CAService) readCAKeyPEM(ctx context.Context, ca *models.CA) ([]byte, error) {
	keyAbs := strings.TrimSpace(ca.KeyPath)
	if keyAbs == "" {
		return nil, exception.New(exception.CodeParamInvalid, "该 CA 未关联私钥", 400, nil)
	}
	if !filepath.IsAbs(keyAbs) {
		if s.layout == nil || s.layout.CertRoot == "" {
			return nil, exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
		}
		keyAbs = filepath.Join(s.layout.CertRoot, keyAbs)
	}
	keyAbs = filepath.Clean(keyAbs)
	if st, err := os.Stat(keyAbs); err != nil || st.IsDir() {
		return nil, exception.New(exception.CodeNotFound, "CA 私钥文件不存在", 404, nil)
	}
	if s.whitebox == nil {
		return nil, exception.New(exception.CodeInternalError, "whitebox 未初始化", 500, nil)
	}
	if s.layout == nil {
		return nil, exception.New(exception.CodeInternalError, "PathLayout 未初始化", 500, nil)
	}

	plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir)
	if err != nil {
		return nil, err
	}
	defer os.Remove(plainTmp)

	data, err := os.ReadFile(plainTmp)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取解密私钥失败: %v", err), 500, nil)
	}
	return data, nil
}

// buildPKCS12 用证书 + 明文私钥 PEM 本地组装 PKCS#12。
//
// 说明：
//   - 不经过 core 的 cert.convert，避免 core 路径守卫只允许 core/data|tmp 的限制；
//   - -passout file: 避免口令出现在命令行；
//   - 铜锁 openssl 默认使用 SM4-CBC 加密 p12 内容，兼容国密场景。
func buildPKCS12(
	opensslBin, certAbs string, keyPEM []byte, name, fallbackName, password string,
) ([]byte, error) {
	if opensslBin == "" {
		return nil, exception.New(exception.CodeInternalError,
			"铜锁 openssl 不可用", 500, nil)
	}
	if _, err := os.Stat(certAbs); err != nil {
		return nil, exception.New(exception.CodeNotFound, "CA 证书文件不存在", 404, nil)
	}

	tmpDir, err := os.MkdirTemp("", "ca-p12-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建临时目录失败: %v", err), 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	keyPath := filepath.Join(tmpDir, "key.pem")
	pwPath := filepath.Join(tmpDir, "pw.txt")
	p12Path := filepath.Join(tmpDir, "out.p12")
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写临时私钥失败", 500, nil)
	}
	if err := os.WriteFile(pwPath, []byte(password), 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
	}

	if strings.TrimSpace(name) == "" {
		name = fallbackName
	}
	_, stderr, err := RunOpenSSLFull(opensslBin, "pkcs12", "-export",
		"-inkey", keyPath,
		"-in", certAbs,
		"-name", name,
		"-out", p12Path,
		"-passout", "file:"+pwPath,
	)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("生成 PKCS#12 失败: %s", strings.TrimSpace(stderr)), 500, nil)
	}
	data, err := os.ReadFile(p12Path)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("读 PKCS#12 文件失败: %v", err), 500, nil)
	}
	return data, nil
}
