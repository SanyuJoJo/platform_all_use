package crypto

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// KeyCrypto 私钥加密检测与解密。
//
// 所有密码通过文件或显式参数传递，绝不让 openssl 打开 /dev/tty 交互式读取，
// 避免在服务端进程中出现“等待输入口令”导致请求卡死。
type KeyCrypto struct{}

// NewKeyCrypto 创建 KeyCrypto。
func NewKeyCrypto() *KeyCrypto { return &KeyCrypto{} }

// IsEncrypted 检测私钥是否加密。
//
// 关键修复：
//   - 显式传 `-passin pass:`（空密码），openssl 不再打开 tty 提示输入；
//   - 未加密私钥：空密码能读通 → err == nil → 返回 false；
//   - 加密私钥  ：空密码立即失败 → err != nil → 返回 true。
//
// 旧实现没有传 `-passin`，openssl 会向 /dev/tty 输出 "Enter pass phrase" 并阻塞，
// 直到 context 15 秒超时，表现为导入接口挂起 15 秒后返回 400。
func (k *KeyCrypto) IsEncrypted(opensslBin, keyPath string) (bool, error) {
	_, _, err := k.run(
		opensslBin, "pkey",
		"-in", keyPath,
		"-noout",
		"-passin", "pass:",
	)
	return err != nil, nil
}

// DecryptWithPasswordFile 用密码文件解密私钥。
//
// -passin file:<path> 让 openssl 从文件读取密码，不进入交互模式。
func (k *KeyCrypto) DecryptWithPasswordFile(
	opensslBin, keyPath, passwordFile, outputPath string,
) error {
	_, _, err := k.run(
		opensslBin, "pkey",
		"-in", keyPath,
		"-passin", "file:"+passwordFile,
		"-out", outputPath,
	)
	return err
}

// run 执行 openssl 命令。
//
// 显式设置 cmd.Stdin = bytes.NewReader(nil)，确保任何 openssl 子命令都不会
// 从标准输入或 /dev/tty 读取密码，避免请求被卡住。
func (k *KeyCrypto) run(bin string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Stdin = bytes.NewReader(nil)

	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
