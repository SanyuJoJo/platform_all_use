package manager

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/yourorg/backend/contract"
)

type goPluginHandle struct {
	client *plugin.Client
	impl   contract.Plugin
	mf     contract.Manifest
}

func (h *goPluginHandle) Name() string    { return h.impl.Name() }
func (h *goPluginHandle) Version() string { return h.impl.Version() }
func (h *goPluginHandle) Mode() string    { return "managed" }
func (h *goPluginHandle) Runtime() string { return "go-plugin" }

func (h *goPluginHandle) Execute(ctx context.Context, input any) (any, error) {
	return h.impl.Execute(ctx, input)
}

func (h *goPluginHandle) Shutdown(ctx context.Context) error {
	if h.impl != nil {
		_ = h.impl.Shutdown(ctx)
	}
	if h.client != nil {
		h.client.Kill()
	}
	return nil
}

// loadGoPlugin 以 go-plugin 托管模式启动插件子进程。
func loadGoPlugin(binaryPath string, mf contract.Manifest, config map[string]any) (Handle, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  contract.HandshakeConfig,
		Plugins:          contract.PluginMap,
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("go-plugin client 失败: %w", err)
	}

	raw, err := rpcClient.Dispense(contract.PluginKey)
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("go-plugin dispense 失败: %w", err)
	}

	impl, ok := raw.(contract.Plugin)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("插件未实现 contract.Plugin")
	}

	if err := impl.Initialize(context.Background(), config); err != nil {
		client.Kill()
		return nil, fmt.Errorf("插件初始化失败: %w", err)
	}

	return &goPluginHandle{client: client, impl: impl, mf: mf}, nil
}
