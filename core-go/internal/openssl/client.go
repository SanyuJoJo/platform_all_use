package openssl
import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "time"
)
// Client 铜锁 openssl 子进程调用封装。
//
// 安全要点：
//   - 严格 argv 参数列表，禁止 shell 拼接
//   - 显式 cmd.Stdin = nil，防止交互式阻塞
//   - 超时由 context 控制
type Client struct {
    bin string
}
// NewClient 创建 Client。
func NewClient() (*Client, error) {
    bin, err := Resolve()
    if err != nil {
        return nil, err
    }
    return &Client{bin: bin}, nil
}
// BinPath 返回 openssl 路径。
func (c *Client) BinPath() string { return c.bin }
// Run 执行 openssl 命令。
func (c *Client) Run(ctx context.Context, args ...string) ([]byte, error) {
    return c.run(ctx, nil, args...)
}
// RunWithStdin 带 stdin 执行（用于 -passin stdin 场景）。
func (c *Client) RunWithStdin(ctx context.Context, stdin []byte, args ...string) ([]byte, error) {
    return c.run(ctx, stdin, args...)
}
func (c *Client) run(ctx context.Context, stdin []byte, args ...string) ([]byte, error) {
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }
    cmd := exec.CommandContext(ctx, c.bin, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if stdin != nil {
        cmd.Stdin = bytes.NewReader(stdin)
    }
    if err := cmd.Run(); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("openssl timeout")
        }
        return nil, fmt.Errorf("openssl %v failed: %w (stderr: %s)",
            args, err, stderr.String())
    }
    return stdout.Bytes(), nil
}
