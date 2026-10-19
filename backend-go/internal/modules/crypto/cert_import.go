package crypto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// =============================================================================
// 导入单证书
// =============================================================================

func (s *CertService) Import(
	ctx context.Context, req *ImportCertRequest,
) (*models.Certificate, error) {
	certPEM := strings.TrimSpace(req.CertPEM)
	if certPEM == "" || !strings.Contains(certPEM, "-----BEGIN CERTIFICATE-----") {
		return nil, exception.New(exception.CodeParamInvalid, "证书不是有效的 PEM 格式", 400, nil)
	}

	dirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("分配 server 目录失败：%v", err), 500, nil)
	}
	dir, _ := s.layout.ServerDir(dirNo)

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}
	pubSM3, err := PubkeySM3FromCertPEM(bin, certPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("计算证书公钥 SM3 失败：%v", err), 400, nil)
	}
	certAbs := filepath.Join(dir, pubSM3+".cert.pem")
	if err := os.WriteFile(certAbs, []byte(certPEM), 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写证书失败：%v", err), 500, nil)
	}
	success := false
	defer func() {
		if !success {
			_ = os.Remove(certAbs)
		}
	}()

	detail, err := s.parser.Parse(certAbs)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("证书解析失败: %v", err), 400, nil)
	}

	var keyRef *string
	var keyPath string

	if strings.TrimSpace(req.KeyPEM) != "" {
		if !strings.Contains(req.KeyPEM, "-----BEGIN") {
			return nil, exception.New(exception.CodeParamInvalid, "私钥不是有效的 PEM 格式", 400, nil)
		}
		plainKeyPEM := []byte(req.KeyPEM)
		if s.keys != nil {
			tmpDir, _ := os.MkdirTemp("", "import-key-")
			defer os.RemoveAll(tmpDir)
			tmpKey := filepath.Join(tmpDir, "key.pem")
			if err := os.WriteFile(tmpKey, []byte(req.KeyPEM), 0600); err != nil {
				return nil, exception.New(exception.CodeInternalError, "写临时私钥失败", 500, nil)
			}
			if encrypted, _ := s.keys.IsEncrypted(bin, tmpKey); encrypted {
				if strings.TrimSpace(req.KeyPassword) == "" {
					return nil, exception.New(exception.CodeParamInvalid,
						"私钥已加密，请提供私钥密码", 400, nil)
				}
				pwPath := filepath.Join(tmpDir, "pw.txt")
				plainPath := filepath.Join(tmpDir, "plain.key")
				if err := os.WriteFile(pwPath, []byte(req.KeyPassword), 0600); err != nil {
					return nil, exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
				}
				if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
					return nil, exception.New(exception.CodeParamInvalid,
						"私钥密码错误或解密失败", 400, nil)
				}
				data, err := os.ReadFile(plainPath)
				if err != nil {
					return nil, exception.New(exception.CodeInternalError, "读解密私钥失败", 500, nil)
				}
				plainKeyPEM = data
			}
		}

		keyPath = filepath.Join(dir, pubSM3+".key.pem")

		plainTmp := filepath.Join(dir, ".import.key.plain.tmp")
		if err := os.WriteFile(plainTmp, plainKeyPEM, 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
		}
		defer os.Remove(plainTmp)

		if err := s.whitebox.EncryptFile(ctx, plainTmp, keyPath); err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
		}
		_ = os.Chmod(keyPath, 0600)
	} else if req.KeyRef != "" {
		keyRef = &req.KeyRef
	}

	// ★ 修复：证书类型不再推断，直接跟随证书真实扩展。
	//   - KeyUsage / ExtendedKeyUsage 都为空 → "unknown"
	//   - 有则保存原始文本
	certType := certTypeFromUsage(detail.KeyUsage, detail.ExtendedKeyUsage)

	now := time.Now().UTC()
	cert := &models.Certificate{
		CertID:             "cert-" + dirNo + "-" + pubSM3[:12],
		CertType:           certType,
		Serial:             detail.Serial,
		Subject:            detail.Subject,
		SubjectCN:          cnFromSubject(detail.Subject),
		Issuer:             detail.Issuer,
		IssuerCN:           cnFromSubject(detail.Issuer),
		Algorithm:          detail.PublicKeyAlgorithm,
		Fingerprint:        detail.Fingerprint,
		PublicKeyAlgorithm: detail.PublicKeyAlgorithm,
		SignatureAlgorithm: detail.SignatureAlgorithm,
		NotBefore:          now,
		NotAfter:           now,
		CertPath:           certAbs,
		KeyRef:             keyRef,
		Status:             "VALID",
		CreatedAt:          now,
	}
	if keyPath != "" {
		cert.KeyPath = &keyPath
	}
	// ★ 运行时填充展示字段
	cert.HasPrivateKey =
		(cert.KeyRef != nil && *cert.KeyRef != "") ||
			(cert.KeyPath != nil && *cert.KeyPath != "")

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("证书元数据落库失败: %v", err), 500, nil)
	}
	success = true
	return cert, nil
}

// ImportDualCert 国密双证导入。
func (s *CertService) ImportDualCert(
	ctx context.Context, req *ImportDualCertRequest,
) (*models.Certificate, error) {
	if req == nil ||
		strings.TrimSpace(req.CSRPubSM3) == "" ||
		strings.TrimSpace(req.SignCertPEM) == "" ||
		strings.TrimSpace(req.EncCertPEM) == "" ||
		strings.TrimSpace(req.EncKeyPEM) == "" {
		return nil, exception.New(exception.CodeParamInvalid, "双证导入参数不完整", 400, nil)
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	csrDir, err := s.layout.FindCSRDirByPubkeySM3(req.CSRPubSM3)
	if err != nil {
		return nil, err
	}

	signPubSM3, err := PubkeySM3FromCertPEM(bin, req.SignCertPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("签名证书公钥 SM3 计算失败：%v", err), 400, nil)
	}
	signCertAbs := filepath.Join(csrDir, signPubSM3+".cert.pem")
	if err := os.WriteFile(signCertAbs, []byte(req.SignCertPEM), 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写签名证书失败：%v", err), 500, nil)
	}

	encDirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("分配加密证书目录失败：%v", err), 500, nil)
	}
	encDir, _ := s.layout.ServerDir(encDirNo)

	encPubSM3, err := PubkeySM3FromCertPEM(bin, req.EncCertPEM)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密证书公钥 SM3 计算失败：%v", err), 400, nil)
	}
	encCertAbs := filepath.Join(encDir, encPubSM3+".cert.pem")
	if err := os.WriteFile(encCertAbs, []byte(req.EncCertPEM), 0640); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写加密证书失败：%v", err), 500, nil)
	}

	plainKeyPEM := []byte(req.EncKeyPEM)
	if strings.TrimSpace(req.EncKeyPassword) != "" && s.keys != nil {
		tmpDir, _ := os.MkdirTemp("", "import-dual-key-")
		defer os.RemoveAll(tmpDir)
		tmpKey := filepath.Join(tmpDir, "key.pem")
		pwPath := filepath.Join(tmpDir, "pw.txt")
		plainPath := filepath.Join(tmpDir, "plain.key")
		if err := os.WriteFile(tmpKey, []byte(req.EncKeyPEM), 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError, "写临时私钥失败", 500, nil)
		}
		if err := os.WriteFile(pwPath, []byte(req.EncKeyPassword), 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError, "写口令文件失败", 500, nil)
		}
		if err := s.keys.DecryptWithPasswordFile(bin, tmpKey, pwPath, plainPath); err != nil {
			return nil, exception.New(exception.CodeParamInvalid,
				"加密私钥口令错误或解密失败", 400, nil)
		}
		data, err := os.ReadFile(plainPath)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError, "读解密私钥失败", 500, nil)
		}
		plainKeyPEM = data
	}

	encKeyAbs := filepath.Join(encDir, encPubSM3+".key.pem")

	plainTmp := filepath.Join(encDir, ".import.dual.key.plain.tmp")
	if err := os.WriteFile(plainTmp, plainKeyPEM, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	if err := s.whitebox.EncryptFile(ctx, plainTmp, encKeyAbs); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
	}
	_ = os.Chmod(encKeyAbs, 0600)

	now := time.Now().UTC()
	encDetail, _ := s.parser.Parse(encCertAbs)
	cert := &models.Certificate{
		CertID:    "dual-enc-" + encDirNo + "-" + encPubSM3[:12],
		CertType:  "dual_enc",
		CertPath:  encCertAbs,
		Algorithm: "SM2",
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now,
		CreatedAt: now,
	}
	if encDetail != nil {
		cert.Serial = encDetail.Serial
		cert.Subject = encDetail.Subject
		cert.SubjectCN = cnFromSubject(encDetail.Subject)
		cert.Issuer = encDetail.Issuer
		cert.IssuerCN = cnFromSubject(encDetail.Issuer)
		cert.Fingerprint = encDetail.Fingerprint
		cert.PublicKeyAlgorithm = encDetail.PublicKeyAlgorithm
		cert.SignatureAlgorithm = encDetail.SignatureAlgorithm
	}
	cert.KeyPath = &encKeyAbs
	// ★ 运行时填充
	cert.HasPrivateKey = true

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("双证加密证书落库失败：%v", err), 500, nil)
	}
	return cert, nil
}

// certTypeFromUsage 直接从证书扩展读取用途，不做任何业务推断。
//
// 规则：
//   - KeyUsage 和 ExtendedKeyUsage 都为空 → "unknown"
//   - 否则返回原始用途文本，两段用 "; " 分隔
//
// 示例：
//   - KeyUsage = "Digital Signature, Key Encipherment"
//     ExtendedKeyUsage = "TLS Web Server Authentication"
//     → "Digital Signature, Key Encipherment; TLS Web Server Authentication"
//   - 两者都为空 → "unknown"
//
// ★ 不再根据公钥算法（SM2/ECC/RSA...）猜测证书用途，
//   避免与第三方工具（按证书真实扩展展示）出现分歧。
func certTypeFromUsage(keyUsage, extendedKeyUsage string) string {
	ku := strings.TrimSpace(keyUsage)
	eku := strings.TrimSpace(extendedKeyUsage)

	if ku == "" && eku == "" {
		return "unknown"
	}

	parts := make([]string, 0, 2)
	if ku != "" {
		parts = append(parts, ku)
	}
	if eku != "" {
		parts = append(parts, eku)
	}
	return strings.Join(parts, "; ")
}
