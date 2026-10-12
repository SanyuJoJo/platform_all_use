package manager

import (
	"sort"
	"strings"

	"github.com/yourorg/backend/contract"
)

// RouteEntry 是路由表中一条编译后的记录。
type RouteEntry struct {
	Route      contract.Route
	PluginName string
}

// Router 根据 URL 路径与方法匹配路由，决定请求归属哪个插件。
type Router struct {
	entries []RouteEntry
}

// NewRouter 从一组 manifest 构建路由表。
func NewRouter(manifests []contract.Manifest) *Router {
	var entries []RouteEntry
	for _, mf := range manifests {
		for _, r := range mf.Routes {
			if r.Path == "" {
				continue
			}
			if !strings.HasPrefix(r.Path, "/") {
				r.Path = "/" + r.Path
			}
			entries = append(entries, RouteEntry{Route: r, PluginName: mf.Name})
		}
	}
	// 按 priority 降序；同 priority 时按 path 长度降序
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Route.Priority != entries[j].Route.Priority {
			return entries[i].Route.Priority > entries[j].Route.Priority
		}
		return len(entries[i].Route.Path) > len(entries[j].Route.Path)
	})
	return &Router{entries: entries}
}

// Match 返回匹配的路由条目；未命中返回 false。
func (r *Router) Match(path, method string) (RouteEntry, bool) {
	for _, e := range r.entries {
		if e.Route.MatchPath(path) && e.Route.MatchMethod(method) {
			return e, true
		}
	}
	return RouteEntry{}, false
}

// Entries 返回路由表快照，便于调试输出。
func (r *Router) Entries() []RouteEntry {
	out := make([]RouteEntry, len(r.entries))
	copy(out, r.entries)
	return out
}
