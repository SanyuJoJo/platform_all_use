package frontend
import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"backend-go/internal/config"
)
// setupTestDeploy 构造测试用前端 deploy 目录。
//
// 返回 (deployDir, mainAppDir, subAppsDir)。
func setupTestDeploy(t *testing.T) (string, string, string) {
	t.Helper()
	tmp := t.TempDir()
	mainApp := filepath.Join(tmp, "main-app")
	subApps := filepath.Join(tmp, "sub-apps")
	if err := os.MkdirAll(filepath.Join(mainApp, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subApps, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(mainApp, "index.html"),
		[]byte("<html>MAIN_INDEX_MARKER</html>"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(mainApp, "assets", "app.js"),
		[]byte("console.log('APP_JS_MARKER');"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	return tmp, mainApp, subApps
}
// setupTestHandler 构造测试用 Handler（跳过 NewHandler 启动检查）。
func setupTestHandler(t *testing.T) (*Handler, string, string) {
	t.Helper()
	_, mainApp, subApps := setupTestDeploy(t)
	indexPath := filepath.Join(mainApp, "index.html")
	return &Handler{
		enabled:    true,
		mainAppDir: mainApp,
		subAppsDir: subApps,
		indexPath:  indexPath,
	}, mainApp, subApps
}
// setupTestRouter 构造仅含 NoRoute 的测试路由。
func setupTestRouter(t *testing.T, h *Handler) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.NoRoute(h.NoRoute)
	return r
}
// ---------------------------------------------------------------------------
// 基础构造 / 启停
// ---------------------------------------------------------------------------
func TestNewHandlerDisabledWhenNoDir(t *testing.T) {
	h := NewHandler(&config.Config{FrontendDeployDir: ""})
	if h.Enabled() {
		t.Fatal("expected disabled when FRONTEND_DEPLOY_DIR empty")
	}
}
func TestNewHandlerDisabledWhenIndexMissing(t *testing.T) {
	tmp := t.TempDir()
	h := NewHandler(&config.Config{FrontendDeployDir: tmp})
	if h.Enabled() {
		t.Fatal("expected disabled when main-app/index.html missing")
	}
}
func TestNewHandlerEnabledWhenIndexExists(t *testing.T) {
	tmp, _, _ := setupTestDeploy(t)
	h := NewHandler(&config.Config{FrontendDeployDir: tmp})
	if !h.Enabled() {
		t.Fatal("expected enabled when main-app/index.html exists")
	}
}
// ---------------------------------------------------------------------------
// NoRoute：API 优先
// ---------------------------------------------------------------------------
func TestNoRouteAPIReturnsJSON404(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/not-exist", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "90002") {
		t.Fatalf("expected JSON code=90002, got: %s", w.Body.String())
	}
}
func TestNoRouteAPIPrefixExact(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "90002") {
		t.Fatalf("expected JSON code=90002, got: %s", w.Body.String())
	}
}
// ---------------------------------------------------------------------------
// NoRoute：SPA fallback
// ---------------------------------------------------------------------------
func TestNoRouteSPAFallback(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "MAIN_INDEX_MARKER") {
		t.Fatalf("expected main index fallback, got: %s", w.Body.String())
	}
}
// ---------------------------------------------------------------------------
// NoRoute：子应用入口
// ---------------------------------------------------------------------------
func TestNoRouteSubAppIndexExists(t *testing.T) {
	h, _, subApps := setupTestHandler(t)
	authDir := filepath.Join(subApps, "auth")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(authDir, "index.html"),
		[]byte("<html>AUTH_INDEX_MARKER</html>"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	r := setupTestRouter(t, h)
	// 带尾斜杠
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sub-apps/auth/", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /sub-apps/auth/, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "AUTH_INDEX_MARKER") {
		t.Fatalf("expected auth index, got: %s", w.Body.String())
	}
	// 不带尾斜杠
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/sub-apps/auth", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /sub-apps/auth, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "AUTH_INDEX_MARKER") {
		t.Fatalf("expected auth index, got: %s", w.Body.String())
	}
}
func TestNoRouteSubAppMissingReturns404(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sub-apps/nonexistent/", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	// P0-2 关键断言：不得回退到主应用 index.html
	if strings.Contains(w.Body.String(), "MAIN_INDEX_MARKER") {
		t.Fatalf("P0-2 regression: sub-app missing should NOT fallback to main index; got: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "90002") {
		t.Fatalf("expected JSON code=90002, got: %s", w.Body.String())
	}
}
func TestNoRouteSubAppExactRootReturns404(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sub-apps/", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /sub-apps/, got %d", w.Code)
	}
}
// ---------------------------------------------------------------------------
// 静态资源
// ---------------------------------------------------------------------------
func TestTryServeFileAssets(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "APP_JS_MARKER") {
		t.Fatalf("expected app.js content, got: %s", w.Body.String())
	}
}
// ---------------------------------------------------------------------------
// 路径穿越
// ---------------------------------------------------------------------------
func TestTryServeFilePathTraversal(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "root")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(tmp, "secret.txt")
	if err := os.WriteFile(secret, []byte("SECRET_CONTENT"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := &Handler{}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ok := h.tryServeFile(c, root, "/../secret.txt")
	if ok {
		t.Fatal("expected path traversal to be blocked")
	}
	if strings.Contains(w.Body.String(), "SECRET_CONTENT") {
		t.Fatalf("secret leaked: %s", w.Body.String())
	}
}
func TestNoRoutePathTraversalViaURL(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	// 用编码形式模拟穿越尝试
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/%2e%2e/secret.txt", nil)
	r.ServeHTTP(w, req)
	// 关键断言：不应泄露 secret；返回主应用 index.html 或 404 均可接受
	if strings.Contains(w.Body.String(), "SECRET_CONTENT") {
		t.Fatalf("path traversal leaked: %s", w.Body.String())
	}
}
// ---------------------------------------------------------------------------
// 非 GET/HEAD → 405 JSON
// ---------------------------------------------------------------------------
func TestNoRouteNonGETReturns405JSON(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	r := setupTestRouter(t, h)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/dashboard", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("[%s] expected 405, got %d", method, w.Code)
		}
		if !strings.Contains(w.Body.String(), "90001") {
			t.Fatalf("[%s] expected JSON code=90001, got: %s", method, w.Body.String())
		}
	}
}
// ---------------------------------------------------------------------------
// 未启用时保持默认行为
// ---------------------------------------------------------------------------
func TestHandlerDisabledRegisterNoop(t *testing.T) {
	h := &Handler{enabled: false}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Register 在未启用时应为 no-op，不 panic
	h.Register(r)
	// 验证 r.NoRoute 仍为 Gin 默认行为（未设置）
	// 通过发起一次请求确认不会命中我们的处理器逻辑
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (Gin default), got %d", w.Code)
	}
}
