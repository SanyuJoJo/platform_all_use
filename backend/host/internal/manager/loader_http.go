
package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/yourorg/backend/contract"
)

type httpHandle struct {
	mf     contract.Manifest
	client *http.Client
	cmd    *exec.Cmd
}

func (h *httpHandle) Name() string    { return h.mf.Name }
func (h *httpHandle) Version() string { return h.mf.Version }
func (h *httpHandle) Mode() string    { return h.mf.Mode }
func (h *httpHandle) Runtime() string { return "http" }

func (h *httpHandle) Execute(ctx context.Context, input any) (any, error) {
	body := contract.HTTPExecuteRequest{Input: input}
	var reply contract.HTTPExecuteResponse
	if err := h.post(ctx, h.mf.ExecutePath, body, &reply); err != nil {
		return nil, err
	}
	if reply.Error != "" {
		return nil, fmt.Errorf("%s", reply.Error)
	}
	return reply.Output, nil
}

func (h *httpHandle) Shutdown(ctx context.Context) error {
	_ = h.post(ctx, h.mf.ShutdownPath, struct{}{}, &contract.HTTPGenericResponse{})

	if h.cmd != nil && h.cmd.Process != nil {
		_ = h.cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			_, _ = h.cmd.Process.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = h.cmd.Process.Kill()
		}
	}
	return nil
}

func (h *httpHandle) post(ctx context.Context, path string, req any, resp any) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	url := h.mf.HTTPBaseURL() + path
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	res, err := h.client.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", res.StatusCode, string(data))
	}
	if resp != nil {
		return json.Unmarshal(data, resp)
	}
	return nil
}

// loadHTTPStandalone 连接一个已独立启动的 HTTP 插件。
func loadHTTPStandalone(mf contract.Manifest, config map[string]any) (Handle, error) {
	h := &httpHandle{mf: mf, client: &http.Client{Timeout: 30 * time.Second}}
	if err := waitHealthy(h, 15*time.Second); err != nil {
		return nil, fmt.Errorf("独立插件 %s 健康检查失败: %w", mf.Name, err)
	}
	if err := initializeHTTP(h, config); err != nil {
		return nil, err
	}
	return h, nil
}

// loadHTTPManaged 启动 HTTP 插件子进程并等待就绪。
func loadHTTPManaged(mf contract.Manifest, workDir string, config map[string]any) (Handle, error) {
	entry, err := resolveEntrypoint(workDir, mf.Entrypoint)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(entry, mf.ExecArgs...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "PLUGIN_STANDALONE=true")
	for k, v := range mf.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 HTTP 插件失败: %w", err)
	}

	h := &httpHandle{
		mf:     mf,
		client: &http.Client{Timeout: 30 * time.Second},
		cmd:    cmd,
	}
	if err := waitHealthy(h, 20*time.Second); err != nil {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("HTTP 插件 %s 启动超时: %w", mf.Name, err)
	}
	if err := initializeHTTP(h, config); err != nil {
		_ = cmd.Process.Kill()
		return nil, err
	}
	return h, nil
}

func initializeHTTP(h *httpHandle, config map[string]any) error {
	var reply contract.HTTPGenericResponse
	if err := h.post(context.Background(), h.mf.InitPath,
		contract.HTTPInitializeRequest{Config: config}, &reply); err != nil {
		return fmt.Errorf("HTTP 初始化失败: %w", err)
	}
	if reply.Error != "" {
		return fmt.Errorf("HTTP 初始化失败: %s", reply.Error)
	}
	return nil
}

func waitHealthy(h *httpHandle, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := h.mf.HTTPBaseURL() + h.mf.HealthPath
	var lastErr error
	for time.Now().Before(deadline) {
		res, err := h.client.Get(url)
		if err == nil && res.StatusCode == 200 {
			res.Body.Close()
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("status=%d", res.StatusCode)
			res.Body.Close()
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("健康检查超时: %v", lastErr)
}
