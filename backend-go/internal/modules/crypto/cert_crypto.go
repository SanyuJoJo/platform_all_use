package crypto

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// =============================================================================
// SM2 / SM4 底层原语
// =============================================================================

func sm2EncryptWithPubKey(opensslBin string, pubPEM, plaintext []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "sm2-enc-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	pubPath := filepath.Join(tmpDir, "pub.pem")
	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")

	if err := os.WriteFile(pubPath, pubPEM, 0600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(inPath, plaintext, 0600); err != nil {
		return nil, err
	}
	_, stderr, err := RunOpenSSLFull(opensslBin, "pkeyutl", "-encrypt",
		"-pubin", "-inkey", pubPath,
		"-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl pkeyutl -encrypt 失败: %s", strings.TrimSpace(stderr))
	}
	return os.ReadFile(outPath)
}

func sm4EncryptECBNoPadding(opensslBin string, key, plaintext []byte) ([]byte, error) {
	if len(plaintext)%16 != 0 {
		return nil, fmt.Errorf("SM4-ECB/NoPadding 明文长度必须是 16 的倍数，当前 %d", len(plaintext))
	}
	tmpDir, err := os.MkdirTemp("", "sm4-ecb-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	inPath := filepath.Join(tmpDir, "in.bin")
	outPath := filepath.Join(tmpDir, "out.bin")
	if err := os.WriteFile(inPath, plaintext, 0600); err != nil {
		return nil, err
	}
	keyHex := hex.EncodeToString(key)
	_, stderr, err := RunOpenSSLFull(opensslBin, "enc", "-sm4-ecb",
		"-e", "-K", keyHex, "-nopad",
		"-in", inPath, "-out", outPath)
	if err != nil {
		return nil, fmt.Errorf("openssl enc -sm4-ecb 失败: %s", strings.TrimSpace(stderr))
	}
	return os.ReadFile(outPath)
}

func extractEncCertECPoint(opensslBin string, certPEM []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "enc-point-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	certPath := filepath.Join(tmpDir, "cert.pem")
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		return nil, err
	}
	pubPEM, err := RunOpenSSL(opensslBin, "x509", "-in", certPath, "-noout", "-pubkey")
	if err != nil {
		return nil, fmt.Errorf("openssl x509 -pubkey 失败: %w", err)
	}
	pubPath := filepath.Join(tmpDir, "pub.pem")
	if err := os.WriteFile(pubPath, []byte(pubPEM), 0600); err != nil {
		return nil, err
	}
	derOut, stderr, err := RunOpenSSLFull(opensslBin, "ec", "-pubin",
		"-in", pubPath, "-conv_form", "uncompressed", "-outform", "DER")
	if err != nil {
		return nil, fmt.Errorf("openssl ec -conv_form 失败: %s", strings.TrimSpace(stderr))
	}
	var spki struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}
	if _, err := asn1.Unmarshal([]byte(derOut), &spki); err != nil {
		return nil, fmt.Errorf("解析 SPKI 失败: %w", err)
	}
	return spki.PublicKey.Bytes, nil
}

func extractRawECPrivateKey(opensslBin string, keyPEM []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "raw-ec-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)
	keyPath := filepath.Join(tmpDir, "key.pem")
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, err
	}
	out, stderr, err := RunOpenSSLFull(opensslBin, "ec", "-in", keyPath, "-noout", "-text")
	if err != nil {
		out, stderr, err = RunOpenSSLFull(opensslBin, "pkey", "-in", keyPath, "-noout", "-text")
		if err != nil {
			return nil, fmt.Errorf("提取裸私钥失败: %s", strings.TrimSpace(stderr))
		}
	}
	lines := strings.Split(out, "\n")
	var inPriv bool
	var hexBuf strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "priv:") {
			inPriv = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "priv:"))
			if rest != "" {
				hexBuf.WriteString(rest)
			}
			continue
		}
		if inPriv {
			if trimmed == "" || strings.HasPrefix(trimmed, "pub:") {
				break
			}
			hexBuf.WriteString(trimmed)
		}
	}
	hexStr := hexBuf.String()
	for _, r := range []string{":", " ", "\n", "\r", "\t"} {
		hexStr = strings.ReplaceAll(hexStr, r, "")
	}
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("解析私钥 hex 失败: %w", err)
	}
	for len(raw) > 32 && raw[0] == 0 {
		raw = raw[1:]
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("私钥长度异常：%d（应为 32）", len(raw))
	}
	return raw, nil
}

// =============================================================================
// 公钥提取 / 比对
// =============================================================================

func extractPubKeyFromCertPEM(opensslBin, certPEM string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "cert-pub-")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	certPath := filepath.Join(tmpDir, "cert.pem")
	if err := os.WriteFile(certPath, []byte(certPEM), 0600); err != nil {
		return nil, fmt.Errorf("写临时证书失败: %w", err)
	}
	out, err := RunOpenSSL(opensslBin, "x509", "-in", certPath, "-noout", "-pubkey")
	if err != nil {
		return nil, fmt.Errorf("openssl x509 -pubkey 失败: %w", err)
	}
	return []byte(out), nil
}

func extractPubKeyFromCSR(opensslBin, csrPath string) ([]byte, error) {
	if !filepath.IsAbs(csrPath) {
		return nil, fmt.Errorf("csrPath 必须是绝对路径: %s", csrPath)
	}
	if _, err := os.Stat(csrPath); err != nil {
		return nil, fmt.Errorf("P10 文件不存在: %s: %w", csrPath, err)
	}
	out, err := RunOpenSSL(opensslBin, "req", "-in", csrPath, "-noout", "-pubkey")
	if err != nil {
		return nil, fmt.Errorf("openssl req -pubkey 失败: %w", err)
	}
	return []byte(out), nil
}

func pubKeyEqual(a, b []byte) bool {
	da, err1 := decodePEMToDER(a)
	db, err2 := decodePEMToDER(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return bytes.Equal(da, db)
}

func decodePEMToDER(pemBytes []byte) ([]byte, error) {
	s := string(pemBytes)
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	block, _ := pem.Decode([]byte(strings.TrimSpace(s)))
	if block == nil {
		return nil, fmt.Errorf("不是有效的 PEM 块")
	}
	return block.Bytes, nil
}

func pubKeySHA256(pemBytes []byte) string {
	der, err := decodePEMToDER(pemBytes)
	if err != nil {
		return "invalid-pem"
	}
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}
