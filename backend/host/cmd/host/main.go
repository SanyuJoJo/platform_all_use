package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yourorg/backend/contract"
	"github.com/yourorg/backend/host/internal/config"
	"github.com/yourorg/backend/host/internal/manager"
	"github.com/yourorg/backend/host/internal/updater"
)

var version = "dev"

func main() {
	// —— 命令行参数 ——
	configPath := flag.String("config", "", "配置文件路径（默认按顺序查找 config.json）")
	hostFlag := flag.String("host", "", "覆盖配置中的 host（优先级最高）")
	portFlag := flag.Int("port", 0, "覆盖配置中的 port（优先级最高）")
	pluginDirFlag := flag.String("plugin-dir", "", "覆盖配置中的 plugin_dir")
	backupDirFlag := flag.String("backup-dir", "", "覆盖配置中的 backup_dir")
	showConfig := flag.Bool("show-config", false, "打印最终生效配置后退出")
	flag.Parse()

	// —— 加载配置 ——
	cfg, usedPath, err := config.Load(
		*configPath,
		*hostFlag, *portFlag,
		*pluginDirFlag, *backupDirFlag,
	)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	log.SetPrefix("[host] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *showConfig {
		if usedPath != "" {
			log.Printf("配置文件: %s", usedPath)
		} else {
			log.Printf("配置文件: 未使用（内置默认 + 环境变量/命令行）")
		}
		log.Printf("生效配置: %s", cfg.String())
		return
	}

	log.Printf("启动主应用 version=%s", version)
	if usedPath != "" {
		log.Printf("配置文件: %s", usedPath)
	}
	log.Printf("监听地址 %s（外部访问 http://%s）", cfg.ListenAddr(), cfg.DisplayAddr())

	// —— 初始化插件管理器 ——
	mgr, err := manager.NewManager(cfg.PluginDir, cfg.BackupDir)
	if err != nil {
		log.Fatalf("初始化插件管理器失败: %v", err)
	}
	mgr.LoadAll()

	// —— 初始化升级器 ——
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}
	up := updater.NewUpdater(exePath, cfg.BackupDir)

	// —— 路由表（可在 reload 时重建） ——
	var (
		routerMu sync.RWMutex
		router   *manager.Router
	)
	setRouter := func(r *manager.Router) {
		routerMu.Lock()
		router = r
		routerMu.Unlock()
	}
	getRouter := func() *manager.Router {
		routerMu.RLock()
		defer routerMu.RUnlock()
		return router
	}
	rebuildRouter := func() {
		manifests, err := mgr.Discover()
		if err != nil {
			log.Printf("[host] 重建路由表失败: %v", err)
			return
		}
		r := manager.NewRouter(manifests)
		setRouter(r)
		for _, e := range r.Entries() {
			log.Printf("[host] route %s -> plugin %s (methods=%v priority=%d strip=%v)",
				e.Route.Path, e.PluginName, e.Route.Methods, e.Route.Priority, e.Route.StripPrefix)
		}
	}
	rebuildRouter()

	// —— HTTP 路由 ——
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"version": version,
			"bind":    cfg.ListenAddr(),
			"plugins": mgr.Health(),
		})
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"config_path": usedPath,
			"host":        cfg.Host,
			"port":        cfg.Port,
			"plugin_dir":  cfg.PluginDir,
			"backup_dir":  cfg.BackupDir,
			"log_level":   cfg.LogLevel,
			"listen":      cfg.ListenAddr(),
			"display":     cfg.DisplayAddr(),
		})
	})

	mux.HandleFunc("/plugins", func(w http.ResponseWriter, r *http.Request) {
		list, err := mgr.Discover()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	})

	mux.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		rt := getRouter()
		if rt == nil {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		entries := rt.Entries()
		out := make([]map[string]any, 0, len(entries))
		for _, e := range entries {
			out = append(out, map[string]any{
				"plugin":       e.PluginName,
				"path":         e.Route.Path,
				"methods":      e.Route.Methods,
				"priority":     e.Route.Priority,
				"strip_prefix": e.Route.StripPrefix,
			})
		}
		writeJSON(w, http.StatusOK, out)
	})

	mux.HandleFunc("/plugins/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "missing query param: name", http.StatusBadRequest)
			return
		}
		h, ok := mgr.Get(name)
		if !ok {
			http.Error(w, "plugin not found: "+name, http.StatusNotFound)
			return
		}
		var body struct {
			Input any `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		out, err := h.Execute(ctx, body.Input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"output": out})
	})

	mux.HandleFunc("/plugins/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "missing query param: name", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		if err := mgr.Stop(ctx, name); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/plugins/reload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "missing query param: name", http.StatusBadRequest)
			return
		}
		mf, err := findManifest(mgr, name)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := mgr.Reload(ctx, mf); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		rebuildRouter()
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/plugins/rollback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		ver := r.URL.Query().Get("version")
		if name == "" || ver == "" {
			http.Error(w, "missing query param: name or version", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := mgr.Rollback(ctx, name, ver); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		rebuildRouter()
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/host/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		dir, err := up.Snapshot(version)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"snapshot": dir})
	})

	mux.HandleFunc("/host/snapshots", func(w http.ResponseWriter, r *http.Request) {
		list, err := up.ListSnapshots()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, list)
	})

	mux.HandleFunc("/host/rollback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		dir := r.URL.Query().Get("snapshot")
		if dir == "" {
			http.Error(w, "missing query param: snapshot", http.StatusBadRequest)
			return
		}
		if err := up.Rollback(dir); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok, please restart the host process to take effect",
		})
	})

	// —— 兜底路由：未命中固定 API 的请求走插件路由 ——
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rt := getRouter()
		if rt == nil {
			http.NotFound(w, r)
			return
		}
		entry, ok := rt.Match(r.URL.Path, r.Method)
		if !ok {
			http.NotFound(w, r)
			return
		}
		mgr.Forward(w, r, entry)
	})

	// —— 启动 HTTP 服务 ——
	srv := &http.Server{
		Addr:              cfg.ListenAddr(),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Printf("HTTP 服务监听 %s", cfg.ListenAddr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()

	// —— 优雅退出 ——
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("收到关闭信号，正在优雅退出...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP 服务关闭异常: %v", err)
	}
	mgr.Shutdown(ctx)

	log.Println("已退出")
}

func findManifest(mgr *manager.Manager, name string) (contract.Manifest, error) {
	list, err := mgr.Discover()
	if err != nil {
		return contract.Manifest{}, err
	}
	for _, mf := range list {
		if mf.Name == name {
			return mf, nil
		}
	}
	return contract.Manifest{}, os.ErrNotExist
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}
