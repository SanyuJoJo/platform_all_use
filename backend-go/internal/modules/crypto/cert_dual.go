package crypto

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// =============================================================================
// 国密双证签发
// =============================================================================

func (s *CertService) signDualCert(
	ctx context.Context, req *SignCertRequest, cleanup *[]string,
) (*SignCertResult, error) {
	if req.Algorithm != "" && req.Algorithm != "SM2" {
		return nil, exception.New(exception.CodeParamInvalid, "国密双证仅支持 SM2", 400, nil)
	}

	params := map[string]interface{}{
		"sign_algorithm": "SM2",
		"enc_algorithm":  "SM2",
		"tlcp_profile":   "GB/T 38636-2020",
		"validity_days":  req.ValidityDays,
	}

	if err := s.resolveCSRSource(req, params); err != nil {
		return nil, err
	}
	csrKeyRef := ""
	if v, ok := params["csr_key_ref"].(string); ok {
		csrKeyRef = v
	}
	if len(req.Subject) > 0 {
		params["subject"] = req.Subject
	}
	if len(req.SAN) > 0 {
		params["san"] = req.SAN
	}
	if err := s.resolveCASource(ctx, req, params, cleanup); err != nil {
		return nil, err
	}

	csrPath := ""
	if v, ok := params["csr_path"].(string); ok {
		csrPath = v
	}
	if csrPath == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"国密双证必须提供 P10（csr_path）", 400, nil)
	}

	// ★ 用 resolveCSRAbsPath，兼容 <CertRoot>/server/... 下的 CSR
	csrAbs, err := s.resolveCSRAbsPath(csrPath)
	if err != nil {
		return nil, err
	}

	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}

	csrPubPEM, err := extractPubKeyFromCSR(bin, csrAbs)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("无法从 P10 提取公钥（%s）：%v", csrAbs, err), 500, nil)
	}
	if len(bytes.TrimSpace(csrPubPEM)) == 0 {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("P10 公钥为空（%s）", csrAbs), 500, nil)
	}

	signDirNo := ""
	if rel, err := filepath.Rel(s.layout.ServerRoot, filepath.Dir(csrAbs)); err == nil {
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		if len(parts) >= 1 && len(parts[0]) == 4 {
			signDirNo = parts[0]
		}
	}
	if signDirNo == "" {
		if err := s.attachOutputLayout(params, req.Domain, ""); err != nil {
			return nil, err
		}
	} else {
		signDir, _ := s.layout.ServerDir(signDirNo)
		params["sign_dir_no"] = signDirNo
		params["sign_output_dir"] = signDir
	}

	encDirNo, err := s.layout.AllocServerDir()
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("分配加密证书目录失败：%v", err), 500, nil)
	}
	encDir, _ := s.layout.ServerDir(encDirNo)
	params["enc_dir_no"] = encDirNo
	params["enc_output_dir"] = encDir
	params["output_layout"] = "server_pubkey_sm3"
	params["cert_root"] = s.layout.CertRoot
	if req.Domain != "" {
		params["domain"] = req.Domain
	}

	resp, err := s.caller.Run(ctx, "dual_cert.create", params)
	if err != nil {
		return nil, err
	}

	signCert, err := s.persistDualCertFromCore(req, resp, csrKeyRef, params)
	if err != nil {
		return nil, err
	}

	result := &SignCertResult{Certificate: signCert}

	signCertPath := getString(resp.Data, "sign_cert_path")
	if signCertPath != "" {
		if data, err := os.ReadFile(signCertPath); err == nil {
			result.SignCertPEM = string(data)
		} else if pem, err := s.files.ReadCoreFile(signCertPath); err == nil {
			result.SignCertPEM = string(pem)
		}
	}
	encCertPath := getString(resp.Data, "enc_cert_path")
	if encCertPath != "" {
		if data, err := os.ReadFile(encCertPath); err == nil {
			result.EncCertPEM = string(data)
		} else if pem, err := s.files.ReadCoreFile(encCertPath); err == nil {
			result.EncCertPEM = string(pem)
		}
	}

	if result.SignCertPEM != "" {
		signCertPubPEM, err := extractPubKeyFromCertPEM(bin, result.SignCertPEM)
		if err != nil {
			log.Warn().Err(err).
				Str("csr_path", csrPath).
				Str("sign_cert_path", signCertPath).
				Msg("从签名证书提取公钥失败（不阻断签发）")
		} else if !pubKeyEqual(csrPubPEM, signCertPubPEM) {
			log.Warn().
				Str("csr_path", csrPath).
				Str("sign_cert_path", signCertPath).
				Str("csr_pub_sha256", pubKeySHA256(csrPubPEM)).
				Str("sign_cert_pub_sha256", pubKeySHA256(signCertPubPEM)).
				Msg("签名证书公钥指纹与 P10 不一致（已忽略，不阻断签发）")
		}
	}

	encKeyRef := getString(resp.Data, "enc_key_ref")
	if encKeyRef == "" {
		return nil, exception.New(exception.CodeInternalError,
			"core 未返回 enc_key_ref", 500, nil)
	}
	defer func() {
		_, _ = s.caller.Run(context.Background(), "key.manage", map[string]interface{}{
			"action":  "delete",
			"key_ref": encKeyRef,
		})
	}()

	encKeyPEM, err := s.caller.ExportKey(ctx, encKeyRef)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("导出加密私钥明文失败：%v", err), 500, nil)
	}

	var encKeyAbsPath string
	if len(encKeyPEM) > 0 && encCertPath != "" {
		encKeyAbsPath, err = s.persistEncKeyWhitebox(ctx, encDirNo, encDir, encCertPath, encKeyPEM)
		if err != nil {
			return nil, err
		}
		var encCert models.Certificate
		if e := s.db.Where("cert_id LIKE ?", "dual-enc-%").
			Order("id DESC").First(&encCert).Error; e == nil {
			encCert.KeyPath = &encKeyAbsPath
			_ = s.db.Save(&encCert).Error
		}
	}

	if len(encKeyPEM) > 0 && result.EncCertPEM != "" {
		envelopeStr, err := s.buildEnvelopeForDual(
			csrPubPEM,
			[]byte(result.EncCertPEM),
			encKeyPEM,
		)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("构建数字信封失败: %v", err), 500, nil)
		}
		result.EncryptedEnvelope = envelopeStr

		if env, _, err := s.decodeEnvelopeAny(envelopeStr); err == nil && env != nil {
			record := &models.EnvelopedKeyRecord{
				SignCertID:          signCert.CertID,
				EncCertID:           getString(resp.Data, "enc_cert_id"),
				Format:              "pkcs10",
				Algorithm:           "SM2+SM4-ECB",
				SignAlg:             "SM3withSM2",
				EncAlg:              "SM4-ECB",
				SymmetricKeyCipher:  base64.StdEncoding.EncodeToString(env.SymKeyCipher.FullBytes),
				IV:                  "",
				EncryptedPrivateKey: base64.StdEncoding.EncodeToString(env.EncKeyCipher.Bytes),
				CreatedAt:           time.Now().UTC(),
			}
			_ = s.db.Create(record).Error
		}
	}

	return result, nil
}

func (s *CertService) persistEncKeyWhitebox(
	ctx context.Context, encDirNo, encDir, encCertPath string, encKeyPEM []byte,
) (string, error) {
	bin := s.parser.OpensslBin()
	if bin == "" {
		return "", exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}
	encCertPEM, err := os.ReadFile(encCertPath)
	if err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("读取加密证书失败：%v", err), 500, nil)
	}
	encPubSM3, err := PubkeySM3FromCertPEM(bin, string(encCertPEM))
	if err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("计算加密证书公钥 SM3 失败：%v", err), 500, nil)
	}

	encKeyPath := filepath.Join(encDir, encPubSM3+".key.pem")

	plainTmp := filepath.Join(encDir, ".enc.key.plain.tmp")
	if err := os.WriteFile(plainTmp, encKeyPEM, 0600); err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("写临时私钥失败：%v", err), 500, nil)
	}
	defer os.Remove(plainTmp)

	if err := s.whitebox.EncryptFile(ctx, plainTmp, encKeyPath); err != nil {
		return "", exception.New(exception.CodeInternalError,
			fmt.Sprintf("白盒加密私钥失败：%v", err), 500, nil)
	}
	_ = os.Chmod(encKeyPath, 0600)

	log.Info().
		Str("dir_no", encDirNo).
		Str("enc_pub_sm3", encPubSM3).
		Str("key_path", encKeyPath).
		Msg("加密私钥白盒加密落盘完成")
	return encKeyPath, nil
}

func (s *CertService) persistDualCertFromCore(
	req *SignCertRequest, resp *CoreResponse, csrKeyRef string,
	params map[string]interface{},
) (*models.Certificate, error) {
	signCertPath := getString(resp.Data, "sign_cert_path")
	encCertPath := getString(resp.Data, "enc_cert_path")
	signPubSM3 := getString(resp.Data, "sign_pubkey_sm3")
	encPubSM3 := getString(resp.Data, "enc_pubkey_sm3")

	signDirNo := getString(resp.Data, "sign_dir_no")
	if signDirNo == "" {
		if v, ok := params["sign_dir_no"].(string); ok {
			signDirNo = v
		}
	}
	encDirNo := getString(resp.Data, "enc_dir_no")
	if encDirNo == "" {
		if v, ok := params["enc_dir_no"].(string); ok {
			encDirNo = v
		}
	}

	if signCertPath == "" && signPubSM3 != "" && signDirNo != "" {
		_, cPath, _, _, _, err := s.layout.ServerPaths(signDirNo, signPubSM3)
		if err == nil {
			signCertPath = cPath
		}
	}
	if encCertPath == "" && encPubSM3 != "" && encDirNo != "" {
		_, cPath, _, _, _, err := s.layout.ServerPaths(encDirNo, encPubSM3)
		if err == nil {
			encCertPath = cPath
		}
	}
	if signCertPath == "" {
		return nil, exception.New(exception.CodeInternalError,
			"core dual_cert.create 未返回 sign_cert_path", 500, nil)
	}

	now := time.Now().UTC()
	baseID := newShortID()

	signAbs, _ := s.resolveCertPath(signCertPath)
	var signDetail *CertDetail
	if signAbs != "" {
		signDetail, _ = s.parser.Parse(signAbs)
	}

	signSubject, signIssuer, signSerial := "", "", ""
	if signDetail != nil {
		signSubject = signDetail.Subject
		signIssuer = signDetail.Issuer
		signSerial = signDetail.Serial
	}
	if signSubject == "" {
		if cn := extractCNFromParams(params); cn != "" {
			signSubject = "CN=" + cn
		}
	}

	signCert := &models.Certificate{
		CertID:    "dual-sign-" + baseID,
		CertType:  "dual_sign",
		Serial:    signSerial,
		Subject:   signSubject,
		SubjectCN: cnFromSubject(signSubject),
		Issuer:    signIssuer,
		IssuerCN:  cnFromSubject(signIssuer),
		CAID:      req.CAID,
		Algorithm: "SM2",
		CertPath:  signCertPath,
		Status:    "VALID",
		NotBefore: now,
		NotAfter:  now.AddDate(0, 0, req.ValidityDays),
		CreatedAt: now,
	}
	if csrKeyRef != "" {
		signCert.KeyRef = &csrKeyRef
	}
	if signDetail != nil {
		signCert.Fingerprint = signDetail.Fingerprint
		signCert.PublicKeyAlgorithm = signDetail.PublicKeyAlgorithm
		signCert.SignatureAlgorithm = signDetail.SignatureAlgorithm
	}
	// ★ 运行时填充
	signCert.HasPrivateKey = signCert.KeyRef != nil && *signCert.KeyRef != ""

	if err := s.db.Create(signCert).Error; err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("双证签名证书落库失败: %v", err), 500, nil)
	}

	if encCertPath != "" {
		encAbs, _ := s.resolveCertPath(encCertPath)
		var encDetail *CertDetail
		if encAbs != "" {
			encDetail, _ = s.parser.Parse(encAbs)
		}

		encSubject, encIssuer, encSerial := "", "", ""
		if encDetail != nil {
			encSubject = encDetail.Subject
			encIssuer = encDetail.Issuer
			encSerial = encDetail.Serial
		}
		if encSubject == "" {
			if cn := extractCNFromParams(params); cn != "" {
				encSubject = "CN=" + cn
			}
		}

		encCert := &models.Certificate{
			CertID:    "dual-enc-" + baseID,
			CertType:  "dual_enc",
			Serial:    encSerial,
			Subject:   encSubject,
			SubjectCN: cnFromSubject(encSubject),
			Issuer:    encIssuer,
			IssuerCN:  cnFromSubject(encIssuer),
			CAID:      req.CAID,
			Algorithm: "SM2",
			CertPath:  encCertPath,
			Status:    "VALID",
			NotBefore: now,
			NotAfter:  now.AddDate(0, 0, req.ValidityDays),
			CreatedAt: now,
		}
		if encDetail != nil {
			encCert.Fingerprint = encDetail.Fingerprint
			encCert.PublicKeyAlgorithm = encDetail.PublicKeyAlgorithm
			encCert.SignatureAlgorithm = encDetail.SignatureAlgorithm
		}
		// 此刻还没有 KeyPath，下一段逻辑（persistEncKeyWhitebox）会填
		_ = s.db.Create(encCert).Error
	}

	return signCert, nil
}
