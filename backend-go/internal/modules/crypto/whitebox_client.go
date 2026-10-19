package crypto

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"backend-go/internal/exception"
)

// WhiteboxClient 白盒 SM4 工具客户端。
//
// ★ 密钥集成在二进制内部：
//   - 加密：whitebox_sm4 <file>          就地加密
//   - 解密：whitebox_sm4 <file> <out>    解密到指定路径
//   - 不再有 .pass 文件
type WhiteboxClient struct {
	Bin string
}

func NewWhiteboxClient(bin string) *WhiteboxClient {
	return &WhiteboxClient{Bin: bin}
}

func (c *WhiteboxClient) run(ctx context.Context, args ...string) error {
	if _, err := os.Stat(c.Bin); err != nil {
		return exception.New(exception.CodeInternalError, "whitebox_sm4 不存在", 500, nil)
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(nil)
	if err := cmd.Run(); err != nil {
		return exception.New(
			exception.CodeInternalError,
			fmt.Sprintf("whitebox 执行失败: %s", stderr.String()),
			500, nil,
		)
	}
	return nil
}

// EncryptFile 就地加密。
//
// plain 内容会被拷贝到 enc 路径，然后 whitebox_sm4 就地在 enc 上加密。
func (c *WhiteboxClient) EncryptFile(ctx context.Context, plain, enc string) error {
	if err := copyFileMode(plain, enc, 0600); err != nil {
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("copy plain → enc: %v", err), 500, nil)
	}
	if err := c.run(ctx, enc); err != nil {
		_ = os.Remove(enc)
		return err
	}
	return os.Chmod(enc, 0600)
}

// ★ EncryptInPlace 就地加密（明文文件 → 加密后覆盖自身）。
//
// 相比 EncryptFile（拷贝明文到目标再加密），本方法直接在原文件上加密：
//
//	whitebox_sm4 <path>
//
// 适用场景：nginx 启动后立即把临时落盘的明文私钥就地加密，
// 明文生命周期缩短到 nginx 加载私钥的那一瞬间。
func (c *WhiteboxClient) EncryptInPlace(ctx context.Context, path string) error {
	if path == "" {
		return exception.New(exception.CodeParamInvalid, "路径为空", 400, nil)
	}
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return exception.New(exception.CodeNotFound,
				"文件不存在: "+path, 404, nil)
		}
		return exception.New(exception.CodeInternalError,
			fmt.Sprintf("访问文件失败: %v", err), 500, nil)
	}
	if st.IsDir() || st.Size() == 0 {
		return exception.New(exception.CodeParamInvalid,
			"文件非法（空或目录）: "+path, 400, nil)
	}
	if err := c.run(ctx, path); err != nil {
		return err
	}
	return os.Chmod(path, 0600)
}

// DecryptToTemp 解密到临时文件（0600），返回临时明文路径。
func (c *WhiteboxClient) DecryptToTemp(ctx context.Context, enc, tmpDir string) (string, error) {
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(tmpDir, "wbx-plain-*.pem")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	f.Close()
	_ = os.Remove(tmp)
	_ = os.Chmod(tmp, 0600)
	if err := c.run(ctx, enc, tmp); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return tmp, nil
}

// ReEncrypt 就地重新加密（明文临时文件 → 密文原位置）。
func (c *WhiteboxClient) ReEncrypt(ctx context.Context, plainTmp, enc string) error {
	if err := copyFileMode(plainTmp, enc, 0600); err != nil {
		return err
	}
	if err := c.run(ctx, enc); err != nil {
		return err
	}
	return os.Chmod(enc, 0600)
}

func (c *WhiteboxClient) HasPrivateKey(keyPath string) bool {
	if keyPath == "" {
		return false
	}
	st, err := os.Stat(filepath.Clean(keyPath))
	return err == nil && !st.IsDir() && st.Size() > 0
}

// =============================================================================
// 文件辅助
// =============================================================================

func copyFileMode(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		_ = os.Remove(dst)
		return err
	}
	return os.Chmod(dst, mode)
}
