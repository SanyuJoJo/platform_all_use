package crypto

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"

	"backend-go/internal/exception"
	"backend-go/internal/models"
)

// =============================================================================
// 类型定义
// =============================================================================

var sm4OID = asn1.ObjectIdentifier{1, 2, 156, 10197, 1, 104}

type algIdentifier struct {
	Algorithm asn1.ObjectIdentifier
}

type gmEnvelope struct {
	KeyAlg       algIdentifier
	SymKeyCipher asn1.RawValue
	EncPubPoint  asn1.BitString
	EncKeyCipher asn1.BitString
}

type legacyEnvelope struct {
	Version             string `json:"version"`
	Algorithm           string `json:"algorithm"`
	SignCertID          string `json:"sign_cert_id,omitempty"`
	EncCertID           string `json:"enc_cert_id,omitempty"`
	SignAlg             string `json:"sign_alg"`
	EncAlg              string `json:"enc_alg"`
	SymmetricKeyCipher  string `json:"symmetric_key_cipher"`
	IV                  string `json:"iv"`
	EncryptedPrivateKey string `json:"encrypted_private_key"`
}

// =============================================================================
// 数字信封：构建
// =============================================================================

func (s *CertService) buildEnvelopeForDual(
	csrPubPEM []byte, encCertPEM []byte, encKeyPEM []byte,
) (string, error) {
	der, err := s.buildEncryptedEnvelopeDER(csrPubPEM, encCertPEM, encKeyPEM)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

func (s *CertService) buildEncryptedEnvelopeDER(
	csrPubPEM []byte, encCertPEM []byte, encKeyPEM []byte,
) ([]byte, error) {
	if len(bytes.TrimSpace(csrPubPEM)) == 0 {
		return nil, fmt.Errorf("加密公钥为空，必须使用 P10 公钥")
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, fmt.Errorf("铜锁 openssl 不可用")
	}

	encPubPoint, err := extractEncCertECPoint(bin, encCertPEM)
	if err != nil {
		return nil, fmt.Errorf("提取加密证书公钥点失败: %w", err)
	}
	if len(encPubPoint) != 65 || encPubPoint[0] != 0x04 {
		return nil, fmt.Errorf("加密证书公钥点格式异常（len=%d, first=%02x）",
			len(encPubPoint), encPubPoint[0])
	}

	rawEncKey, err := extractRawECPrivateKey(bin, encKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("提取加密私钥裸字节失败: %w", err)
	}
	if len(rawEncKey) != 32 {
		return nil, fmt.Errorf("加密私钥裸字节长度异常：%d", len(rawEncKey))
	}

	sm4Key := make([]byte, 16)
	if _, err := rand.Read(sm4Key); err != nil {
		return nil, fmt.Errorf("生成 SM4 密钥失败: %w", err)
	}

	symKeyCipherASN1, err := sm2EncryptWithPubKey(bin, csrPubPEM, sm4Key)
	if err != nil {
		return nil, fmt.Errorf("SM2 加密对称密钥失败: %w", err)
	}
	if len(symKeyCipherASN1) == 0 || symKeyCipherASN1[0] != 0x30 {
		return nil, fmt.Errorf("SM2 密文不是 SEQUENCE（首字节 %02x）", symKeyCipherASN1[0])
	}

	encKeyCipher, err := sm4EncryptECBNoPadding(bin, sm4Key, rawEncKey)
	if err != nil {
		return nil, fmt.Errorf("SM4 加密加密私钥失败: %w", err)
	}

	env := gmEnvelope{
		KeyAlg:       algIdentifier{Algorithm: sm4OID},
		SymKeyCipher: asn1.RawValue{FullBytes: symKeyCipherASN1},
		EncPubPoint: asn1.BitString{
			Bytes: encPubPoint, BitLength: len(encPubPoint) * 8,
		},
		EncKeyCipher: asn1.BitString{
			Bytes: encKeyCipher, BitLength: len(encKeyCipher) * 8,
		},
	}
	der, err := asn1.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("序列化信封失败: %w", err)
	}
	return der, nil
}

// =============================================================================
// 数字信封：解析
// =============================================================================

func (s *CertService) decodeEnvelopeAny(raw string) (*gmEnvelope, *legacyEnvelope, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil, fmt.Errorf("空字符串")
	}
	var decoded []byte
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil {
		decoded = b
	} else {
		decoded = []byte(raw)
	}
	if len(decoded) > 0 && decoded[0] == 0x30 {
		var env gmEnvelope
		if _, err := asn1.Unmarshal(decoded, &env); err == nil &&
			len(env.SymKeyCipher.FullBytes) > 0 &&
			len(env.EncKeyCipher.Bytes) > 0 {
			return &env, nil, nil
		}
	}
	if len(decoded) > 0 && decoded[0] == '{' {
		var leg legacyEnvelope
		if err := json.Unmarshal(decoded, &leg); err == nil &&
			leg.SymmetricKeyCipher != "" && leg.EncryptedPrivateKey != "" {
			return nil, &leg, nil
		}
	}
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil &&
		len(b) > 0 && b[0] == '{' {
		var leg legacyEnvelope
		if err := json.Unmarshal(b, &leg); err == nil &&
			leg.SymmetricKeyCipher != "" && leg.EncryptedPrivateKey != "" {
			return nil, &leg, nil
		}
	}
	return nil, nil, fmt.Errorf("不是有效的加密数字信封")
}

// =============================================================================
// 查询信封信息
// =============================================================================

func (s *CertService) QueryEnvelope(
	req *QueryEnvelopeRequest,
) (*QueryEnvelopeResponse, error) {
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "pkcs10"
	}
	if format != "pkcs10" {
		if format == "cfca" {
			return nil, exception.New(exception.CodeParamInvalid,
				"CFCA 证书格式暂未实现", 400, nil)
		}
		return nil, exception.New(exception.CodeParamInvalid,
			"format 必须是 pkcs10 或 cfca", 400, nil)
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "cert"
	}
	if mode != "cert" && mode != "manual" {
		return nil, exception.New(exception.CodeParamInvalid,
			"mode 必须是 cert 或 manual", 400, nil)
	}
	if strings.TrimSpace(req.EncryptedEnvelope) == "" {
		return nil, exception.New(exception.CodeParamInvalid,
			"请提供加密的数字信封", 400, nil)
	}
	env, leg, err := s.decodeEnvelopeAny(req.EncryptedEnvelope)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密的数字信封解析失败: %v", err), 400, nil)
	}

	var signKeyPEM string
	var signKeyPassword string
	switch mode {
	case "cert":
		if strings.TrimSpace(req.CertID) == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"从签名证书解析需要 cert_id", 400, nil)
		}
		var cert models.Certificate
		if err := s.db.Where("cert_id = ?", req.CertID).First(&cert).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, exception.New(exception.CodeNotFound, "证书不存在", 404, nil)
			}
			return nil, exception.New(exception.CodeInternalError, "查询证书失败", 500, nil)
		}
		if cert.KeyRef == nil || *cert.KeyRef == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"该签名证书未关联私钥（平台不持有客户签名私钥）。"+
					"请改用\"手动输入\"模式，上传您的 CSR 私钥 PEM", 400, nil)
		}
		plain, err := s.caller.ExportKey(context.Background(), *cert.KeyRef)
		if err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("导出签名私钥失败: %v", err), 500, nil)
		}
		signKeyPEM = string(plain)
	case "manual":
		if strings.TrimSpace(req.SignKeyPEM) == "" {
			return nil, exception.New(exception.CodeParamInvalid,
				"手动输入模式需要提供 sign_key_pem（CSR 私钥 PEM）", 400, nil)
		}
		signKeyPEM = req.SignKeyPEM
		signKeyPassword = req.SignKeyPassword
	}

	plainKey, err := s.decryptEnvelopeAuto(req.EncryptedEnvelope, signKeyPEM, signKeyPassword)
	if err != nil {
		return nil, err
	}
	resp := &QueryEnvelopeResponse{
		Format:          format,
		Mode:            mode,
		DecryptedKeyPEM: string(plainKey),
	}
	if env != nil {
		resp.SymmetricKeyCipher = base64.StdEncoding.EncodeToString(env.SymKeyCipher.FullBytes)
		resp.EncryptedPrivateKey = base64.StdEncoding.EncodeToString(env.EncKeyCipher.Bytes)
	} else if leg != nil {
		resp.SymmetricKeyCipher = leg.SymmetricKeyCipher
		resp.IV = leg.IV
		resp.EncryptedPrivateKey = leg.EncryptedPrivateKey
	}
	return resp, nil
}

func (s *CertService) decryptEnvelopeAuto(
	raw string, signKeyPEM, signKeyPassword string,
) ([]byte, error) {
	env, leg, err := s.decodeEnvelopeAny(raw)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("加密的数字信封解析失败: %v", err), 400, nil)
	}
	bin := s.parser.OpensslBin()
	if bin == "" {
		return nil, exception.New(exception.CodeInternalError, "铜锁 openssl 不可用", 500, nil)
	}
	tmpDir, err := os.MkdirTemp("", "envelope-dec-")
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("创建临时目录失败: %v", err), 500, nil)
	}
	defer os.RemoveAll(tmpDir)

	signKeyPath := filepath.Join(tmpDir, "sign.key")
	if err := os.WriteFile(signKeyPath, []byte(signKeyPEM), 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写签名私钥失败: %v", err), 500, nil)
	}
	plainKeyPath := signKeyPath
	if strings.TrimSpace(signKeyPassword) != "" {
		pwPath := filepath.Join(tmpDir, "pw.txt")
		plainPath := filepath.Join(tmpDir, "sign.plain.key")
		if err := os.WriteFile(pwPath, []byte(signKeyPassword), 0600); err != nil {
			return nil, exception.New(exception.CodeInternalError,
				fmt.Sprintf("写口令文件失败: %v", err), 500, nil)
		}
		_, stderr, err := RunOpenSSLFull(bin, "pkey",
			"-in", signKeyPath, "-passin", "file:"+pwPath, "-out", plainPath)
		if err != nil {
			return nil, exception.New(exception.CodeParamInvalid,
				fmt.Sprintf("签名私钥口令错误或解密失败: %s", strings.TrimSpace(stderr)), 400, nil)
		}
		plainKeyPath = plainPath
	}

	switch {
	case env != nil:
		return s.decryptNewEnvelope(bin, tmpDir, plainKeyPath, env)
	case leg != nil:
		return s.decryptLegacyEnvelope(bin, plainKeyPath, leg)
	}
	return nil, exception.New(exception.CodeInternalError, "信封格式未知", 500, nil)
}

func (s *CertService) decryptNewEnvelope(
	bin, tmpDir, plainKeyPath string, env *gmEnvelope,
) ([]byte, error) {
	if len(env.SymKeyCipher.FullBytes) == 0 {
		return nil, exception.New(exception.CodeInternalError, "SM2 密文为空", 500, nil)
	}
	cipherPath := filepath.Join(tmpDir, "sm2.cipher")
	if err := os.WriteFile(cipherPath, env.SymKeyCipher.FullBytes, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写 SM2 密文失败: %v", err), 500, nil)
	}
	sm4KeyPath := filepath.Join(tmpDir, "sm4.key")
	_, stderr, err := RunOpenSSLFull(bin, "pkeyutl", "-decrypt",
		"-inkey", plainKeyPath, "-in", cipherPath, "-out", sm4KeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("SM2 解密对称密钥失败: %s", strings.TrimSpace(stderr)), 400, nil)
	}
	sm4Key, err := os.ReadFile(sm4KeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("读 SM4 密钥失败: %v", err), 500, nil)
	}
	if len(sm4Key) != 16 {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4 密钥长度异常：%d", len(sm4Key)), 500, nil)
	}
	encKeyPath := filepath.Join(tmpDir, "enc.key.cipher")
	if err := os.WriteFile(encKeyPath, env.EncKeyCipher.Bytes, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("写加密私钥密文失败: %v", err), 500, nil)
	}
	decKeyPath := filepath.Join(tmpDir, "enc.key.plain")
	keyHex := hex.EncodeToString(sm4Key)
	_, stderr, err = RunOpenSSLFull(bin, "enc", "-sm4-ecb",
		"-d", "-K", keyHex, "-nopad",
		"-in", encKeyPath, "-out", decKeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4-ECB 解密失败: %s", strings.TrimSpace(stderr)), 500, nil)
	}
	return os.ReadFile(decKeyPath)
}

func (s *CertService) decryptLegacyEnvelope(
	bin, plainKeyPath string, leg *legacyEnvelope,
) ([]byte, error) {
	symCipher, err := base64.StdEncoding.DecodeString(leg.SymmetricKeyCipher)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("对称密钥密文 base64 解码失败: %v", err), 500, nil)
	}
	tmpDir, _ := os.MkdirTemp("", "legacy-dec-")
	defer os.RemoveAll(tmpDir)

	cipherPath := filepath.Join(tmpDir, "sm2.cipher")
	if err := os.WriteFile(cipherPath, symCipher, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写 SM2 密文失败", 500, nil)
	}
	sm4KeyPath := filepath.Join(tmpDir, "sm4.key")
	_, stderr, err := RunOpenSSLFull(bin, "pkeyutl", "-decrypt",
		"-inkey", plainKeyPath, "-in", cipherPath, "-out", sm4KeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeParamInvalid,
			fmt.Sprintf("SM2 解密对称密钥失败: %s", strings.TrimSpace(stderr)), 400, nil)
	}
	sm4Key, _ := os.ReadFile(sm4KeyPath)
	if len(sm4Key) != 16 {
		return nil, exception.New(exception.CodeInternalError, "SM4 密钥长度异常", 500, nil)
	}
	iv, err := base64.StdEncoding.DecodeString(leg.IV)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("IV base64 解码失败: %v", err), 500, nil)
	}
	encKeyCipher, err := base64.StdEncoding.DecodeString(leg.EncryptedPrivateKey)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("加密私钥密文 base64 解码失败: %v", err), 500, nil)
	}
	encKeyPath := filepath.Join(tmpDir, "enc.key.cipher")
	decKeyPath := filepath.Join(tmpDir, "enc.key.plain")
	if err := os.WriteFile(encKeyPath, encKeyCipher, 0600); err != nil {
		return nil, exception.New(exception.CodeInternalError, "写加密私钥密文失败", 500, nil)
	}
	keyHex := hex.EncodeToString(sm4Key)
	ivHex := hex.EncodeToString(iv)
	_, stderr, err = RunOpenSSLFull(bin, "enc", "-sm4-cbc",
		"-d", "-K", keyHex, "-iv", ivHex, "-in", encKeyPath, "-out", decKeyPath)
	if err != nil {
		return nil, exception.New(exception.CodeInternalError,
			fmt.Sprintf("SM4-CBC 解密失败: %s", strings.TrimSpace(stderr)), 500, nil)
	}
	return os.ReadFile(decKeyPath)
}
