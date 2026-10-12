package manager

import "context"

// Handle 是主应用对插件的统一视图，屏蔽底层 RPC / HTTP 差异。
type Handle interface {
	Name() string
	Version() string
	Execute(ctx context.Context, input any) (any, error)
	Shutdown(ctx context.Context) error
	Mode() string
	Runtime() string
}
