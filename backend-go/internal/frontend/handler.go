package frontend
import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"backend-go/internal/config"
	"backend-go/internal/exception"
)
// Handler 前端静态资源自托管处理器。
//
// 约定：
//   - FRONTEND_DEPLOY_DIR 指向前端 deploy 目录（推荐绝对路径）；
//   - main-app/index.html 作为 SPA 入口；
//   - main-app/assets 挂载到 /assets，禁止目录列表；
//   - sub-apps 由 NoRoute 统一处理，保证缺失子应用时返回统一 JSON 404，
//     绝不回退到主应用 index.html；
//   - /api/v1/* 优先走 API，不进入前端 fallback。
//
// v1.1 修复：
//   - P0-2：/sub-apps/* 缺失时返回 404，不回退主应用 index.html；
//   - P1-1：NewHandler 启动检查 assets / sub-apps 目录并 WARN；
//   - P1-2：删除冗余的 .. 判断，路径穿越防护完全依赖 tryServeFile；
//   - P1-4：新增 noListFileSystem，禁止目录列表。
type Handler struct {
	enabled    bool
	deployDir  string
	mainAppDir string
	subAppsDir string
	indexPath  string
}
// NewHandler 根据配置创建前端托管处理器。
//
// 若未配置 FRONTEND_DEPLOY_DIR 或 main-app/index.html 不存在，
// 则返回未启用状态，保持后端默认 JSON 404 行为。
//
// v1.1（P1-1）：启动时检查 assets / sub-apps 目录，缺失时打印 WARN，
// 便于运维/开发快速定位。
func NewHandler(cfg *config.Config) *Handler {
	deployDir := strings.TrimSpace(cfg.FrontendDeployDir)
	if deployDir == "" {
		return &Handler{enabled: false}
	}
	abs, err := filepath.Abs(deployDir)
	if err != nil {
		log.Warn().Err(err).Str("dir", deployDir).Msg("前端部署目录解析失败，前端静态托管已禁用")
		return &Handler{enabled: false}
	}
	mainAppDir := filepath.Join(abs, "main-app")
	subAppsDir := filepath.Join(abs, "sub-apps")
	indexPath := filepath.Join(mainAppDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		log.Warn().
			Str("index", indexPath).
			Msg("前端主应用 index.html 不存在，前端静态托管已禁用")
		return &Handler{enabled: false}
	}
	// v1.1（P1-1）：启动检查 assets / sub-apps 目录
	assetsDir := filepath.Join(mainAppDir, "assets")
	if _, err := os.Stat(assetsDir); err != nil {
		log.Warn().
			Str("assets", assetsDir).
			Msg("前端主应用 assets 目录不存在，/assets/* 静态资源可能 404")
	}
	if _, err := os.Stat(subAppsDir); err != nil {
		log.Warn().
			Str("sub_apps", subAppsDir).
			Msg("前端子应用目录不存在，/sub-apps/* 子应用可能无法加载")
	}
	log.Info().
		Str("deploy_dir", abs).
		Str("main_app", mainAppDir).
		Str("sub_apps", subAppsDir).
		Msg("前端静态资源自托管已初始化")
	return &Handler{
		enabled:    true,
		deployDir:  abs,
		mainAppDir: mainAppDir,
		subAppsDir: subAppsDir,
		indexPath:  indexPath,
	}
}
// Enabled 返回是否启用前端托管。
func (h *Handler) Enabled() bool {
	return h != nil && h.enabled
}
// Register 注册静态资源路由，并覆盖 NoRoute 实现 SPA fallback。
//
// 调用时机：必须在所有 API 路由注册完成之后调用。
//
// 注意：
//   - NoMethod 由 main.go 统一注册，本方法不重复注册，
//     避免覆盖 main.go 的注册语义；
//   - /sub-apps 不注册 StaticFS，全部交由 NoRoute 处理：
//     ① 子应用入口缺失时返回统一 JSON 404，不回退主应用 HTML；
//     ② 前端资源不存在时返回统一 JSON 404，不泄露目录列表。
func (h *Handler) Register(r *gin.Engine) {
	if !h.Enabled() {
		return
	}
	// 主应用 assets 目录（禁止目录列表）
	assetsDir := filepath.Join(h.mainAppDir, "assets")
	if _, err := os.Stat(assetsDir); err == nil {
		r.StaticFS("/assets", noListFileSystem{root: http.Dir(assetsDir)})
	} else {
		log.Warn().
			Str("assets", assetsDir).
			Msg("前端主应用 assets 目录不存在，/assets/* 静态资源可能 404")
	}
	// 说明：不注册 /sub-apps 的 Static 路由。
	// 子应用入口（含 404）统一由 NoRoute 处理，保证：
	//   1. 缺失子应用时返回统一 JSON 404，不会回退到主应用 index.html；
	//   2. 前端资源不存在时不会泄露目录列表。
	favicon := filepath.Join(h.mainAppDir, "favicon.ico")
	if _, err := os.Stat(favicon); err == nil {
		r.StaticFile("/favicon.ico", favicon)
	}
	r.NoRoute(h.NoRoute)
	log.Info().
		Str("deploy_dir", h.deployDir).
		Msg("前端静态资源自托管已启用")
}
// NoRoute 处理未匹配路由。
//
// 规则：
//   - /api/ 或 /api 开头：返回统一 JSON 404；
//   - 非 GET/HEAD：返回统一 JSON 405；
//   - /sub-apps/*：仅命中真实文件或 {dir}/index.html；
//     缺失时返回统一 JSON 404，绝不 fallback 到主应用 index.html；
//   - 其他路径：尝试主应用真实文件，失败则 fallback 到 main-app/index.html。
//
// v1.1 变更：
//   - P0-2：/sub-apps/* 缺失时返回 404 JSON；
//   - P1-2：删除冗余的 .. 判断（path.Clean 后恒为 false）。
func (h *Handler) NoRoute(c *gin.Context) {
	reqPath := c.Request.URL.Path
	// API 路径返回 JSON 404
	if strings.HasPrefix(reqPath, "/api/") || reqPath == "/api" {
		exception.NoRoute(c)
		return
	}
	// 仅处理 GET/HEAD
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		exception.NoMethod(c)
		return
	}
	cleanPath := path.Clean("/" + reqPath)
	// 子应用路径：严格处理，缺失时返回 404，不 fallback 到主应用
	if cleanPath == "/sub-apps" || strings.HasPrefix(cleanPath, "/sub-apps/") {
		h.serveSubAppOr404(c, cleanPath)
		return
	}
	// 主应用真实文件
	if h.tryServeFile(c, h.mainAppDir, cleanPath) {
		return
	}
	// SPA fallback 到主应用 index.html
	c.File(h.indexPath)
}
// serveSubAppOr404 尝试从 subAppsDir 提供文件，失败时返回统一 JSON 404。
//
// v1.1（P0-2）新增：保证 qiankun 加载子应用入口时，
// 若资源缺失能拿到明确的 JSON 404，而非主应用 HTML。
func (h *Handler) serveSubAppOr404(c *gin.Context, cleanPath string) {
	subPath := strings.TrimPrefix(cleanPath, "/sub-apps")
	subPath = strings.TrimSuffix(subPath, "/")
	if subPath == "" || subPath == "/" {
		exception.NoRoute(c)
		return
	}
	// 尝试精确文件（如 /sub-apps/auth/assets/index.js）
	if h.tryServeFile(c, h.subAppsDir, subPath) {
		return
	}
	// 尝试 {dir}/index.html（如 /sub-apps/auth/index.html）
	if h.tryServeFile(c, h.subAppsDir, subPath+"/index.html") {
		return
	}
	// 子应用资源缺失，返回统一 JSON 404
	exception.NoRoute(c)
}
// tryServeFile 尝试返回 root 下的文件。
//
// 路径穿越防护：使用 filepath.Rel 校验 absFull 是否在 absRoot 内部。
// 若 absFull 在 absRoot 之外（或以 .. 开头），返回 false。
func (h *Handler) tryServeFile(c *gin.Context, root, urlPath string) bool {
	rel := strings.TrimPrefix(urlPath, "/")
	if rel == "" {
		return false
	}
	full := filepath.Join(root, filepath.FromSlash(rel))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return false
	}
	relToRoot, err := filepath.Rel(absRoot, absFull)
	if err != nil {
		return false
	}
	if relToRoot == ".." ||
		strings.HasPrefix(relToRoot, ".."+string(os.PathSeparator)) {
		return false
	}
	info, err := os.Stat(absFull)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(absFull)
	return true
}
// noListFileSystem 包装 http.FileSystem，禁止目录列表。
//
// v1.1（P1-4）新增：当请求为目录时，仅当该目录下存在 index.html
// 才允许；否则返回 os.ErrNotExist，从根本上避免目录信息泄露。
type noListFileSystem struct {
	root http.FileSystem
}
func (n noListFileSystem) Open(name string) (http.File, error) {
	f, err := n.root.Open(name)
	if err != nil {
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if stat.IsDir() {
		// 目录请求：必须有 index.html 才允许
		indexName := path.Join(name, "index.html")
		indexFile, err := n.root.Open(indexName)
		if err != nil {
			_ = f.Close()
			return nil, os.ErrNotExist
		}
		_ = indexFile.Close()
	}
	return f, nil
}
