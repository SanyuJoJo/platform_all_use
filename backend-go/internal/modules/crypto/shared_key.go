package crypto
import (
	"bytes"
	"context"
	"os/exec"
	"time"
)
// KeyCrypto 私钥加密检测与解密。所有密码通过文件传递，避免出现在进程命令行。
type KeyCrypto struct{}
func NewKeyCrypto() *KeyCrypto { return &KeyCrypto{} }
// IsEncrypted 检测私钥是否加密。
func (k *KeyCrypto) IsEncrypted(opensslBin, keyPath string) (bool, error) {
	_, _, err := k.run(opensslBin, "pkey", "-in", keyPath, "-noout")
	return err != nil, nil
}
// DecryptWithPasswordFile 用密码文件解密私钥。
func (k *KeyCrypto) DecryptWithPasswordFile(
	opensslBin, keyPath, passwordFile, outputPath string,
) error {
	_, _, err := k.run(opensslBin, "pkey",
		"-in", keyPath,
		"-passin", "file:"+passwordFile,
		"-out", outputPath)
	return err
}
func (k *KeyCrypto) run(bin string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), errBuf.String(), err
}
