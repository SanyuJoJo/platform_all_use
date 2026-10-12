package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/yourorg/backend/contract"
)

// Forward 将请求转发给指定插件。响应写入 w。
func (m *Manager) Forward(w http.ResponseWriter, r *http.Request, entry RouteEntry) {
	handle, ok := m.Get(entry.PluginName)
	if !ok {
		http.Error(w, "plugin not loaded: "+entry.PluginName, http.StatusServiceUnavailable)
		return
	}

	switch handle.Runtime() {
	case "http":
		m.forwardHTTP(w, r, entry)
	case "go-plugin":
		m.forwardRPC(w, r, entry, handle)
	default:
		http.Error(w, "unsupported runtime: "+handle.Runtime(), http.StatusInternalServerError)
	}
}

// forwardHTTP 反向代理到插件的 HTTP 端口。
func (m *Manager) forwardHTTP(w http.ResponseWriter, r *http.Request, entry RouteEntry) {
	m.mu.RLock()
	mf, ok := m.manifests[entry.PluginName]
	m.mu.RUnlock()
	if !ok {
		http.Error(w, "manifest not found", http.StatusInternalServerError)
		return
	}

	target, err := url.Parse(mf.HTTPBaseURL())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if entry.Route.StripPrefix {
			stripped := strings.TrimPrefix(req.URL.Path, entry.Route.Path)
			if stripped == "" || stripped[0] != '/' {
				stripped = "/" + stripped
			}
			req.URL.Path = stripped
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "proxy error: "+err.Error(), http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

// forwardRPC 通过 go-plugin RPC 调用插件的 Execute。
func (m *Manager) forwardRPC(w http.ResponseWriter, r *http.Request, entry RouteEntry, handle Handle) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	query := map[string]string{}
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			query[k] = vs[0]
		}
	}
	headers := map[string]string{}
	for k, vs := range r.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}

	path := r.URL.Path
	if entry.Route.StripPrefix {
		path = strings.TrimPrefix(path, entry.Route.Path)
		if path == "" || path[0] != '/' {
			path = "/" + path
		}
	}

	fwdReq := contract.HTTPForwardRequest{
		Method:  r.Method,
		Path:    path,
		Query:   query,
		Headers: headers,
		Body:    string(body),
	}

	out, err := handle.Execute(r.Context(), fwdReq)
	if err != nil {
		http.Error(w, "plugin error: "+err.Error(), http.StatusBadGateway)
		return
	}

	writeForwardResponse(w, out)
}

// writeForwardResponse 将插件返回的 output 转成 HTTP 响应。
// 支持两种 output：
//  1. contract.HTTPForwardResponse 结构
//  2. 任意值，直接作为 200 响应体
func writeForwardResponse(w http.ResponseWriter, out any) {
	data, err := json.Marshal(out)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", out), http.StatusInternalServerError)
		return
	}

	var resp contract.HTTPForwardResponse
	if err := json.Unmarshal(data, &resp); err == nil && resp.Status != 0 {
		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		w.WriteHeader(resp.Status)
		if resp.Body != "" {
			_, _ = w.Write([]byte(resp.Body))
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
