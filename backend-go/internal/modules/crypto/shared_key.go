package crypto

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// KeyCrypto 私钥加密检测与解密。
//
// ★ 核心保护：
//   - IsEncrypted 改为 Go 层解析 PEM 头部，不依赖 openssl；
//   - DecryptWithPasswordFile 清洗密码文件（去换行）+ 显式 -passin file:；
//   - 所有 exec 调用设置 Stdin = 空，避免 openssl 从 tty 读取挂起。
type KeyCrypto struct{}

func NewKeyCrypto() *KeyCrypto { return &KeyCrypto{} }

// IsEncrypted 判断私钥是否加密。
//
// ★ 直接解析 PEM 内容，不调用 openssl：
//   - "BEGIN ENCRYPTED PRIVATE KEY"  → PKCS#8 加密，返回 true
//   - PEM 头 "Proc-Type: 4,ENCRYPTED" → 传统格式加密，返回 true
//   - "BEGIN PRIVATE KEY" / "BEGIN RSA PRIVATE KEY" / "BEGIN EC PRIVATE KEY" → 明文
//
// 这样避免了 openssl -passin pass: 在部分版本中对空密码处理不一致导致挂起。
func (k *KeyCrypto) IsEncrypted(opensslBin, keyPath string) (bool, error) {
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return false, fmt.Errorf("read key file: %w", err)
	}
	return isEncryptedPEM(data), nil
}

// isEncryptedPEM 通过 PEM 内容判断是否加密。
func isEncryptedPEM(data []byte) bool {
	// 1. PKCS#8 加密（最常见）
	if bytes.Contains(data, []byte("-----BEGIN ENCRYPTED PRIVATE KEY-----")) {
		return true
	}

	// 2. 解析 PEM block，检查 Headers
	rest := data
	for {
		block, next := pem.Decode(rest)
		if block == nil {
			break
		}
		if strings.HasPrefix(block.Type, "ENCRYPTED ") {
			return true
		}
		if pt, ok := block.Headers["Proc-Type"]; ok && strings.Contains(pt, "ENCRYPTED") {
			return true
		}
		rest = next
	}

	return false
}

// DecryptWithPasswordFile 用密码文件解密私钥。
//
// ★ 加固点：
//   - 清理密码文件末尾的换行符（OpenSSL 会把换行当密码一部分）；
//   - 用干净副本传给 -passin file:，避免污染原文件；
//   - 显式 -passin file:<path>，绝不进入交互模式。
func (k *KeyCrypto) DecryptWithPasswordFile(
	opensslBin, keyPath, passwordFile, outputPath string,
) error {
	// 1. 读取原密码文件并去除尾部换行
	raw, err := os.ReadFile(passwordFile)
	if err != nil {
		return fmt.Errorf("read password file: %w", err)
	}
	clean := bytes.TrimRight(raw, "\r\n")

	// 2. 写一份干净副本（同目录，0600）
	cleanFile := passwordFile + ".clean"
	if err := os.WriteFile(cleanFile, clean, 0600); err != nil {
		return fmt.Errorf("write clean password file: %w", err)
	}
	defer os.Remove(cleanFile)

	// 3. 调用 openssl pkey 解密
	_, stderr, err := runOpenSSLWithStdin(
		opensslBin, nil,
		"pkey",
		"-in", keyPath,
		"-passin", "file:"+cleanFile,
		"-out", outputPath,
	)
	if err != nil {
		return fmt.Errorf("decrypt key: %v: %s", err, strings.TrimSpace(stderr))
	}
	return nil
}

// runOpenSSLWithStdin 内部辅助：执行 openssl 命令，stdin 显式置空。
func runOpenSSLWithStdin(bin string, stdin []byte, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	} else {
		cmd.Stdin = bytes.NewReader(nil)
	}

	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
