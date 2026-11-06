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

	// ★ 规范化：用 PEM 推断规范 ID（不再用 detail.PublicKeyAlgorithm 展示名）
	algID := certAlgorithmID(bin, certAbs)
	if algID == "" {
		// 兜底：按展示名反查；再不济留空（前端 algStore.label 会走兜底）
		algID = displayNameToID(detail.PublicKeyAlgorithm, "")
	}

	// ★ cert_type 归一化为枚举
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
		Algorithm:          algID,                     // ★ 规范 ID
		Fingerprint:        detail.Fingerprint,
		PublicKeyAlgorithm: detail.PublicKeyAlgorithm, // ★ 展示名
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

// =============================================================================
// 导入国密双证
// =============================================================================

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
	cert.HasPrivateKey = true

	if err := s.db.Create(cert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("双证加密证书落库失败：%v", err), 500, nil)
	}
	return cert, nil
}

// certTypeFromUsage 把证书扩展归一化为平台业务枚举。
//
// ★ 平台业务枚举（唯一允许值）：
//
//	"signature"           : 签名
//	"encryption"          : 加密
//	"signature,encryption": 双用途
//	"dual_sign"           : 双证-签名（由其他路径直接写入）
//	"dual_enc"            : 双证-加密（同上）
//	"imported"            : 导入且无法推断
//
// 映射规则：
//
//	KeyUsage 含 "digital signature" / "non repudiation" / "contentcommitment"
//	                     → 视为签名
//	EKU 含 "TLS Web Server Authentication" / "clientAuth" / "codeSigning"
//	                     → 视为签名
//	KeyUsage 含 "key encipherment" / "data encipherment" / "key agreement"
//	                     → 视为加密
//
// 兼容旧行为：同时有签名+加密时输出 "signature,encryption"；
// 前端 formatCertType 会拆分为"签名、加密"展示。
func certTypeFromUsage(keyUsage, extendedKeyUsage string) string {
	ku := strings.ToLower(strings.TrimSpace(keyUsage))
	eku := strings.ToLower(strings.TrimSpace(extendedKeyUsage))

	if ku == "" && eku == "" {
		return "imported"
	}

	hasSig := strings.Contains(ku, "digital signature") ||
		strings.Contains(ku, "non repudiation") ||
		strings.Contains(ku, "contentcommitment") ||
		strings.Contains(eku, "server authentication") ||
		strings.Contains(eku, "serverauth") ||
		strings.Contains(eku, "client authentication") ||
		strings.Contains(eku, "clientauth") ||
		strings.Contains(eku, "code signing") ||
		strings.Contains(eku, "codesigning")

	hasEnc := strings.Contains(ku, "key encipherment") ||
		strings.Contains(ku, "data encipherment") ||
		strings.Contains(ku, "key agreement")

	switch {
	case hasSig && hasEnc:
		return "signature,encryption"
	case hasSig:
		return "signature"
	case hasEnc:
		return "encryption"
	}
	return "imported"
}
