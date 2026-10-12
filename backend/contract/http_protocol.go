package contract

// HTTP 协议 —— 非 Go 插件（Python / Java / Node / Rust 等）遵循的通信契约。
//
// 端点说明：
//
//   GET  {HealthPath}
//        响应 200: {"status":"ok","name":"xxx","version":"1.0.0"}
//
//   POST {InitPath}
//        请求: {"config": {...}}
//        响应: {"error": ""}
//
//   POST {ExecutePath}
//        请求: {"input": <任意 JSON>}
//        响应: {"output": <任意 JSON>, "error": ""}
//
//   POST {ShutdownPath}
//        请求: {}
//        响应: {"error": ""}

// HTTPInitializeRequest Initialize 端点请求体
type HTTPInitializeRequest struct {
	Config map[string]any `json:"config"`
}

// HTTPExecuteRequest Execute 端点请求体
type HTTPExecuteRequest struct {
	Input any `json:"input"`
}

// HTTPExecuteResponse Execute 端点响应体
type HTTPExecuteResponse struct {
	Output any    `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// HTTPHealthResponse Health 端点响应体
type HTTPHealthResponse struct {
	Status  string `json:"status"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// HTTPGenericResponse 通用响应体
type HTTPGenericResponse struct {
	Error string `json:"error,omitempty"`
}

// —— 路由转发协议 ——
//
// host 在转发请求给插件时，使用如下结构。
//
// 对于 go-plugin 插件：
//   handle.Execute(ctx, HTTPForwardRequest{...}) 返回 HTTPForwardResponse
//
// 对于 HTTP 插件（standalone 模式）：
//   host 直接反向代理，插件收到原始 HTTP 请求。
// 对于 HTTP 插件（managed 模式）：
//   host 同样反向代理，插件监听本地端口。

// HTTPForwardRequest 是 host 转发给 go-plugin 插件的 HTTP 请求表示。
type HTTPForwardRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   map[string]string `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// HTTPForwardResponse 是 go-plugin 插件应返回的 HTTP 响应表示。
type HTTPForwardResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}
