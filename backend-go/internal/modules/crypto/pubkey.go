package crypto

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// =============================================================================
// 公钥 SM3 计算
// =============================================================================
//
// 统一规则（与 core-go 侧保持一致）：
//
//	SubjectPublicKeyInfo (DER) --SM3--> 64 字符小写 hex
//
// 证书：openssl x509 -pubkey -noout | openssl pkey -pubin -outform DER
// CSR ：openssl req  -pubkey -noout | openssl pkey -pubin -outform DER

// PubkeySM3FromCertPEM 计算 X.509 证书 SubjectPublicKeyInfo DER 的 SM3 hex。
func PubkeySM3FromCertPEM(opensslBin, certPEM string) (string, error) {
	if strings.TrimSpace(opensslBin) == "" {
		return "", fmt.Errorf("openssl_bin 未配置")
	}
	if strings.TrimSpace(certPEM) == "" {
		return "", fmt.Errorf("certPEM 为空")
	}

	tmpDir, err := os.MkdirTemp("", "cert-pub-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "cert.pem")
	pubDerPath := filepath.Join(tmpDir, "pub.der")

	if err := os.WriteFile(certPath, []byte(certPEM), 0600); err != nil {
		return "", err
	}

	pubPEM, err := exec.Command(opensslBin, "x509", "-in", certPath, "-pubkey", "-noout").Output()
	if err != nil {
		return "", fmt.Errorf("openssl x509 -pubkey: %w", err)
	}

	pkeyCmd := exec.Command(opensslBin, "pkey", "-pubin", "-outform", "DER", "-out", pubDerPath)
	pkeyCmd.Stdin = bytes.NewReader(pubPEM)
	if err := pkeyCmd.Run(); err != nil {
		return "", fmt.Errorf("openssl pkey -outform DER: %w", err)
	}

	return sm3HexOfFile(opensslBin, pubDerPath)
}

// PubkeySM3FromCSRPEM 计算 PKCS#10 CSR SubjectPublicKeyInfo DER 的 SM3 hex。
func PubkeySM3FromCSRPEM(opensslBin, csrPEM string) (string, error) {
	if strings.TrimSpace(opensslBin) == "" {
		return "", fmt.Errorf("openssl_bin 未配置")
	}
	if strings.TrimSpace(csrPEM) == "" {
		return "", fmt.Errorf("csrPEM 为空")
	}

	tmpDir, err := os.MkdirTemp("", "csr-pub-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	csrPath := filepath.Join(tmpDir, "req.csr")
	pubDerPath := filepath.Join(tmpDir, "pub.der")

	if err := os.WriteFile(csrPath, []byte(csrPEM), 0600); err != nil {
		return "", err
	}

	pubPEM, err := exec.Command(opensslBin, "req", "-in", csrPath, "-pubkey", "-noout").Output()
	if err != nil {
		return "", fmt.Errorf("openssl req -pubkey: %w", err)
	}

	pkeyCmd := exec.Command(opensslBin, "pkey", "-pubin", "-outform", "DER", "-out", pubDerPath)
	pkeyCmd.Stdin = bytes.NewReader(pubPEM)
	if err := pkeyCmd.Run(); err != nil {
		return "", fmt.Errorf("openssl pkey -outform DER: %w", err)
	}

	return sm3HexOfFile(opensslBin, pubDerPath)
}

// sm3HexOfFile 用铜锁 openssl 计算文件 SM3，返回小写 hex。
func sm3HexOfFile(opensslBin, path string) (string, error) {
	out, err := exec.Command(opensslBin, "dgst", "-sm3", "-r", path).Output()
	if err != nil {
		return "", fmt.Errorf("openssl dgst -sm3: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 1 {
		return "", fmt.Errorf("openssl dgst 输出异常: %q", string(out))
	}
	return strings.ToLower(parts[0]), nil
}

// =============================================================================
// 随机口令
// =============================================================================

// readRand 读取 n 字节随机数。
var readRand = func(b []byte) (int, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(b)
}

// randomHex 返回 n 字节的随机 hex 字符串（长度 2n）。
func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := readRand(buf); err != nil {
		return fmt.Sprintf("%0*x", n*2, 0)
	}
	return fmt.Sprintf("%x", buf)
}
