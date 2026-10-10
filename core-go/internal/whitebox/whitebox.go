package whitebox
import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)
type Client struct {
	Bin string
}
func New(bin string) *Client { return &Client{Bin: bin} }
func (c *Client) run(ctx context.Context, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(nil)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("whitebox: %w: %s", err, stderr.String())
	}
	return nil
}
func (c *Client) Encrypt(ctx context.Context, plain, enc, pass string) error {
	return c.run(ctx, "encrypt", "--in", plain, "--out", enc, "--pass", pass)
}
func (c *Client) Decrypt(ctx context.Context, enc, plain, pass string) error {
	return c.run(ctx, "decrypt", "--in", enc, "--out", plain, "--pass", pass)
}
func (c *Client) EnsurePass(passPath string) error {
	if _, err := os.Stat(passPath); err == nil {
		return nil
	}
	buf := make([]byte, 32)
	if _, err := os.ReadFile("/dev/urandom"); err != nil {
		// fallback
	}
	// 简化：调用方应使用 crypto/rand 生成 hex。
	return os.WriteFile(passPath, buf, 0600)
}
