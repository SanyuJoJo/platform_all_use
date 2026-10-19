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

// =============================================================================
// 导出
// =============================================================================

func (s *CertService) Export(
	ctx context.Context, certID, exportType, password string,
) (*ExportResult, error) {
	c, err := s.Get(certID)
	if err != nil {
		return nil, err
	}
	switch exportType {
	case "cert":
		data, err := os.ReadFile(c.CertPath)
		if err != nil {
			data, err = s.files.ReadCoreFile(c.CertPath)
			if err != nil {
				return nil, err
			}
		}
		return &ExportResult{
			Data:        data,
			Filename:    certID + ".pem",
			ContentType: "application/x-pem-file",
		}, nil

	case "key":
		keyPEM, err := s.exportKeyFromStorage(ctx, c)
		if err != nil {
			if c.KeyRef != nil && *c.KeyRef != "" {
				keyPEM, err = s.caller.ExportKey(ctx, *c.KeyRef)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		if strings.TrimSpace(password) != "" {
			bin := s.parser.OpensslBin()
			if bin == "" {
				return nil, exception.New(exception.CodeInternalError,
					"铜锁 openssl 不可用，无法加密私钥", 500, nil)
			}
			encrypted, err := encryptPrivateKeyPEM(bin, keyPEM, password)
			if err != nil {
				return nil, exception.New(exception.CodeInternalError,
					fmt.Sprintf("加密私钥失败: %v", err), 500, nil)
			}
			keyPEM = encrypted
		}
		return &ExportResult{
			Data:        keyPEM,
			Filename:    certID + ".key.pem",
			ContentType: "application/x-pem-file",
		}, nil

	case "pkcs12":
		// ★ 修复：支持两种私钥存储模式
		//
		//   1) 旧模式：cert.KeyRef 非空（core keystore），走 core 的 cert.convert；
		//   2) 新模式：cert.KeyPath 非空（白盒密文），本地解密 + openssl pkcs12 -export。
		if len(password) < 6 {
			return nil, exception.New(exception.CodeParamInvalid, "PKCS#12 密码至少 6 位", 400, nil)
		}

		// 模式 1：keystore 引用
		if c.KeyRef != nil && *c.KeyRef != "" {
			data, err := s.caller.ExportPKCS12(ctx, c.CertPath, *c.KeyRef, password)
			if err != nil {
				return nil, err
			}
			return &ExportResult{
				Data:        data,
				Filename:    certID + ".p12",
				ContentType: "application/x-pkcs12",
			}, nil
		}

		// 模式 2：白盒私钥
		keyAbs := s.resolveCertKeyPath(c)
		if keyAbs == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"该证书未关联私钥", 400, nil)
		}
		data, err := s.exportPKCS12FromWhiteboxKey(ctx, c, keyAbs, password)
		if err != nil {
			return nil, err
		}
		return &ExportResult{
			Data:        data,
			Filename:    certID + ".p12",
			ContentType: "application/x-pkcs12",
		}, nil

	default:
		return nil, exception.New(exception.CodeParamInvalid,
			"不支持的导出类型: "+exportType, 400, nil)
	}
}

// exportPKCS12FromWhiteboxKey 用白盒私钥 + 证书本地组装 PKCS#12。
//
// 适用场景：**新模式证书**（cert.KeyPath 非空、cert.KeyRef 空）。
func (s *CertService) exportPKCS12FromWhiteboxKey(
	ctx context.Context, c *models.Certificate, keyAbs, password string,
) ([]byte, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError,
			"铜锁 openssl 不可用", 500, nil)
	}
	certAbs, err := s.resolveCertPath(c.CertPath)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(keyAbs); err != nil {
		return nil, exception.New(exception.CodeNotFound,
			"白盒私钥文件不存在", 404, nil)
	}

	// 1. 白盒解密到临时明文私钥
	plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("解密白盒私钥失败: %v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	// 2. openssl pkcs12 -export
	tmpDir, err := os.MkdirTemp("", "p12-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建临时目录失败: %v", err), 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	pwPath := filepath.Join(tmpDir, "pw.txt")
	p12Path := filepath.Join(tmpDir, "out.p12")
	if err := os.WriteFile(pwPath, []byte(password), 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
	}

	// 说明：
	//   - -name 用 CN 作为 friendlyName，便于客户端展示；
	//   - -passout file: 避免口令出现在 argv 中；
	//   - 铜锁 openssl 默认使用 SM4-CBC 加密 p12 内容，对国密场景兼容。
	name := c.SubjectCN
	if name == "" {
		name = c.CertID
	}
	_, stderr, err := RunOpenSSLFull(bin, "pkcs12", "-export",
		"-inkey", plainTmp,
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

func (s *CertService) exportKeyFromStorage(
	ctx context.Context, c *models.Certificate,
) ([]byte, error) {
	keyAbs := s.resolveCertKeyPath(c)
	if keyAbs == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"该证书未关联白盒私钥", 400, nil)
	}
	if _, err := os.Stat(keyAbs); err != nil {
		return nil, exception.New(exception.CodeNotFound, "白盒私钥文件不存在", 404, nil)
	}
	plainTmp, err := s.whitebox.DecryptToTemp(ctx, keyAbs, s.layout.TmpDir)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = s.whitebox.ReEncrypt(ctx, plainTmp, keyAbs)
		_ = os.Remove(plainTmp)
	}()
	return os.ReadFile(plainTmp)
}

func (s *CertService) resolveCertKeyPath(c *models.Certificate) string {
	if c == nil {
		return ""
	}
	if c.KeyPath != nil && *c.KeyPath != "" {
		p := *c.KeyPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(s.layout.CertRoot, p)
		}
		return p
	}
	if c.ChainPath != nil && *c.ChainPath != "" {
		p := *c.ChainPath
		if !filepath.IsAbs(p) {
			p = filepath.Join(s.layout.CertRoot, p)
		}
		return p
	}
	return ""
}

// =============================================================================
// 私钥导出 / 加密辅助
// =============================================================================

func (s *CertService) exportKeyPEM(
	ctx context.Context, keyRef, password string,
) ([]byte, error) {
	plain, err := s.caller.ExportKey(ctx, keyRef)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(password) == "" {
		return plain, nil
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, fmt.Errorf("铜锁 openssl 不可用，无法加密私钥")
	}
	return encryptPrivateKeyPEM(bin, plain, password)
}

func encryptPrivateKeyPEM(
	opensslBin string, plainPEM []byte, password string,
) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cert-key-enc-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	plainPath := filepath.Join(tmpDir, "plain.key")
	encPath := filepath.Join(tmpDir, "enc.key")
	pwPath := filepath.Join(tmpDir, "pw.txt")
	if err := os.WriteFile(plainPath, plainPEM, 0600); err != nil {
		return nil, fmt.Errorf("写明文私钥失败: %w", err)
	}
	if err := os.WriteFile(pwPath, []byte(password), 0600); err != nil {
		return nil, fmt.Errorf("写口令文件失败: %w", err)
	}
	_, stderr, err := RunOpenSSLFull(opensslBin, "pkey",
		"-in", plainPath, "-aes-256-cbc", "-passout", "file:"+pwPath, "-out", encPath)
	if err != nil {
		return nil, fmt.Errorf("openssl pkey 加密失败: %s", strings.TrimSpace(stderr))
	}
	return os.ReadFile(encPath)
}
