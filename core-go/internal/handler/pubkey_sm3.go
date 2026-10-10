package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourorg/core-go/internal/openssl"
)

// =============================================================================
// 公钥 SM3 计算
// =============================================================================
//
// 统一规则（与 backend-go 侧 PubkeySM3FromCertPEM / PubkeySM3FromCSRPEM 一致）：
//
//	SubjectPublicKeyInfo (DER) --SM3--> 64 字符小写 hex
//
// 证书：openssl x509 -pubkey -noout | openssl pkey -pubin -outform DER
// CSR ：openssl req  -pubkey -noout | openssl pkey -pubin -outform DER

// pubkeySM3FromCertAbs 计算证书公钥 SPKI DER 的 SM3 十六进制。
func pubkeySM3FromCertAbs(
	ctx context.Context,
	client *openssl.Client,
	certAbs string,
	tmpDir string,
) (string, error) {
	if client == nil {
		return "", fmt.Errorf("openssl client is nil")
	}

	pubPEM, err := client.Run(ctx, "x509", "-in", certAbs, "-pubkey", "-noout")
	if err != nil {
		return "", fmt.Errorf("openssl x509 -pubkey: %w", err)
	}
	pubPEMPath := filepath.Join(tmpDir, "pubkey-sm3.pem")
	if err := os.WriteFile(pubPEMPath, pubPEM, 0600); err != nil {
		return "", fmt.Errorf("write pub pem: %w", err)
	}
	defer os.Remove(pubPEMPath)

	pubDERPath := filepath.Join(tmpDir, "pubkey-sm3.der")
	if _, err := client.Run(ctx, "pkey",
		"-pubin", "-in", pubPEMPath,
		"-outform", "DER",
		"-out", pubDERPath,
	); err != nil {
		return "", fmt.Errorf("openssl pkey -outform DER: %w", err)
	}
	defer os.Remove(pubDERPath)

	out, err := client.Run(ctx, "dgst", "-sm3", "-r", pubDERPath)
	if err != nil {
		return "", fmt.Errorf("openssl dgst -sm3: %w", err)
	}
	parts := strings.Fields(string(out))
	if len(parts) < 1 {
		return "", fmt.Errorf("dgst 输出异常: %q", string(out))
	}
	return strings.ToLower(parts[0]), nil
}

// pubkeySM3FromCSRAbs 计算 CSR 公钥 SPKI DER 的 SM3 十六进制。
func pubkeySM3FromCSRAbs(
	ctx context.Context,
	client *openssl.Client,
	csrAbs string,
	tmpDir string,
) (string, error) {
	if client == nil {
		return "", fmt.Errorf("openssl client is nil")
	}

	pubPEM, err := client.Run(ctx, "req", "-in", csrAbs, "-pubkey", "-noout")
	if err != nil {
		return "", fmt.Errorf("openssl req -pubkey: %w", err)
	}
	pubPEMPath := filepath.Join(tmpDir, "csr-pub-sm3.pem")
	if err := os.WriteFile(pubPEMPath, pubPEM, 0600); err != nil {
		return "", fmt.Errorf("write csr pub pem: %w", err)
	}
	defer os.Remove(pubPEMPath)

	pubDERPath := filepath.Join(tmpDir, "csr-pub-sm3.der")
	if _, err := client.Run(ctx, "pkey",
		"-pubin", "-in", pubPEMPath,
		"-outform", "DER",
		"-out", pubDERPath,
	); err != nil {
		return "", fmt.Errorf("openssl pkey -outform DER: %w", err)
	}
	defer os.Remove(pubDERPath)

	out, err := client.Run(ctx, "dgst", "-sm3", "-r", pubDERPath)
	if err != nil {
		return "", fmt.Errorf("openssl dgst -sm3: %w", err)
	}
	parts := strings.Fields(string(out))
	if len(parts) < 1 {
		return "", fmt.Errorf("dgst 输出异常: %q", string(out))
	}
	return strings.ToLower(parts[0]), nil
}
