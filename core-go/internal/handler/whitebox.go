package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/yourorg/core-go/internal/config"
)

// whiteboxEncrypt 用白盒工具加密私钥文件。
//
// ★ 密钥集成在二进制内部，无 .pass 文件。
//
// 流程：
//  1. 拷贝明文 plain → outEnc（目标位置）
//  2. 调 whitebox_sm4 <outEnc>  （就地加密）
//  3. 校验 outEnc 是合法白盒密文（magic "WBX1"）
//
// 调用方约定：
//
//	plain  : 明文私钥临时文件路径
//	outEnc : 加密后私钥的最终路径
func whiteboxEncrypt(
	ctx context.Context, cfg *config.Config,
	plain, outEnc string,
) error {
	if cfg.WhiteboxBin == "" {
		return fmt.Errorf("whitebox_bin not configured")
	}
	if _, err := os.Stat(cfg.WhiteboxBin); err != nil {
		return fmt.Errorf("whitebox_sm4 not found: %s: %w", cfg.WhiteboxBin, err)
	}

	// 1. 拷贝明文到目标位置
	if err := copyFileMode(plain, outEnc, 0600); err != nil {
		return fmt.Errorf("copy plain → %s: %w", outEnc, err)
	}

	// 2. 就地加密
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cctx, cfg.WhiteboxBin, outEnc)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = nil
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		_ = os.Remove(outEnc)
		return fmt.Errorf("whitebox encrypt: %v: %s", err, stderr.String())
	}

	// 3. 校验密文 magic
	f, err := os.Open(outEnc)
	if err != nil {
		return fmt.Errorf("open ciphertext: %w", err)
	}
	defer f.Close()
	head := make([]byte, 4)
	if _, err := io.ReadFull(f, head); err != nil {
		return fmt.Errorf("read ciphertext head: %w", err)
	}
	if !bytes.Equal(head, []byte("WBX1")) {
		return fmt.Errorf("ciphertext magic mismatch (got %q)", head)
	}

	// 4. 权限 0600
	_ = os.Chmod(outEnc, 0600)

	return nil
}

// copyFileMode 拷贝文件并设置权限。
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
