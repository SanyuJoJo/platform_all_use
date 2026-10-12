package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/yourorg/backend/contract"
)

// Manager 负责插件的发现、加载、卸载、热重载与回退。
type Manager struct {
	pluginDir string
	backupDir string

	mu        sync.RWMutex
	handles   map[string]Handle
	manifests map[string]contract.Manifest
}

// NewManager 创建插件管理器。
func NewManager(pluginDir, backupDir string) (*Manager, error) {
	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("解析插件目录绝对路径失败: %w", err)
	}
	absBackupDir, err := filepath.Abs(backupDir)
	if err != nil {
		return nil, fmt.Errorf("解析备份目录绝对路径失败: %w", err)
	}
	if err := os.MkdirAll(absPluginDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建插件目录失败: %w", err)
	}
	if err := os.MkdirAll(absBackupDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败: %w", err)
	}
	return &Manager{
		pluginDir: absPluginDir,
		backupDir: absBackupDir,
		handles:   make(map[string]Handle),
		manifests: make(map[string]contract.Manifest),
	}, nil
}

// Discover 扫描插件目录下的 manifest.json。
func (m *Manager) Discover() ([]contract.Manifest, error) {
	entries, err := os.ReadDir(m.pluginDir)
	if err != nil {
		return nil, err
	}
	var out []contract.Manifest
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(m.pluginDir, e.Name(), "manifest.json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var mf contract.Manifest
		if err := json.Unmarshal(data, &mf); err != nil {
			log.Printf("[manager] 解析 %s 失败: %v", path, err)
			continue
		}
		mf.ApplyDefaults()
		out = append(out, mf)
	}
	return out, nil
}

// LoadAll 加载全部已发现的插件。
func (m *Manager) LoadAll() {
	manifests, _ := m.Discover()
	for _, mf := range manifests {
		if _, err := m.Load(mf); err != nil {
			log.Printf("[manager] 加载插件 %s 失败: %v", mf.Name, err)
		}
	}
}

// Load 加载单个插件。
func (m *Manager) Load(mf contract.Manifest) (Handle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.handles[mf.Name]; exists {
		return nil, fmt.Errorf("插件 %s 已加载", mf.Name)
	}

	mf.ApplyDefaults()
	pluginDir := filepath.Join(m.pluginDir, mf.Name)
	config := m.readConfig(mf.Name)

	var (
		h   Handle
		err error
	)

	switch {
	case mf.Runtime == "go-plugin" && mf.Mode == "managed":
		binaryPath, rerr := resolveEntrypoint(pluginDir, mf.Entrypoint)
		if rerr != nil {
			return nil, rerr
		}
		if _, statErr := os.Stat(binaryPath); statErr != nil {
			return nil, fmt.Errorf("go-plugin 插件 %s 入口文件不存在: %s", mf.Name, binaryPath)
		}
		if err = verifyChecksum(binaryPath, mf.Checksum); err != nil {
			return nil, err
		}
		h, err = loadGoPlugin(binaryPath, mf, config)

	case mf.Runtime == "http" && mf.Mode == "managed":
		workDir := pluginDir
		if mf.WorkDir != "" {
			workDir = filepath.Join(pluginDir, mf.WorkDir)
		}
		h, err = loadHTTPManaged(mf, workDir, config)

	case mf.Runtime == "http" && mf.Mode == "standalone":
		h, err = loadHTTPStandalone(mf, config)

	default:
		return nil, fmt.Errorf("不支持的插件配置: mode=%s runtime=%s", mf.Mode, mf.Runtime)
	}

	if err != nil {
		return nil, err
	}

	m.handles[mf.Name] = h
	m.manifests[mf.Name] = mf
	log.Printf("[manager] 插件 %s v%s 已加载 (mode=%s runtime=%s)",
		mf.Name, mf.Version, mf.Mode, mf.Runtime)
	return h, nil
}

// Get 获取插件句柄。
func (m *Manager) Get(name string) (Handle, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.handles[name]
	return h, ok
}

// Stop 停止并卸载插件。
func (m *Manager) Stop(ctx context.Context, name string) error {
	m.mu.Lock()
	h, ok := m.handles[name]
	if ok {
		delete(m.handles, name)
		delete(m.manifests, name)
	}
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("插件 %s 未加载", name)
	}
	return h.Shutdown(ctx)
}

// Reload 热重载插件到新版本；失败自动回退到旧版本。
func (m *Manager) Reload(ctx context.Context, mf contract.Manifest) error {
	m.mu.RLock()
	old, exists := m.handles[mf.Name]
	oldMf := m.manifests[mf.Name]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("插件 %s 未加载", mf.Name)
	}

	if err := m.snapshotPlugin(mf.Name, oldMf); err != nil {
		return fmt.Errorf("备份失败: %w", err)
	}

	_ = old.Shutdown(ctx)
	m.mu.Lock()
	delete(m.handles, mf.Name)
	delete(m.manifests, mf.Name)
	m.mu.Unlock()

	if _, err := m.Load(mf); err != nil {
		log.Printf("[manager] 新版本加载失败，正在回退: %v", err)
		_ = m.restorePlugin(mf.Name, oldMf)
		if _, rerr := m.Load(oldMf); rerr != nil {
			return fmt.Errorf("回退失败: %v (原始错误: %v)", rerr, err)
		}
		return fmt.Errorf("重载失败已回退: %w", err)
	}
	log.Printf("[manager] 插件 %s 重载到 v%s", mf.Name, mf.Version)
	return nil
}

// Rollback 回退到某个历史版本。
func (m *Manager) Rollback(ctx context.Context, name, version string) error {
	m.mu.RLock()
	oldMf, ok := m.manifests[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("插件 %s 未加载", name)
	}

	if err := m.Stop(ctx, name); err != nil {
		log.Printf("[manager] 停止当前版本失败: %v", err)
	}

	if err := m.restorePluginVersion(name, version); err != nil {
		return err
	}
	newMf, err := m.readManifest(name)
	if err != nil {
		return err
	}
	if _, err := m.Load(newMf); err != nil {
		_ = m.restorePlugin(name, oldMf)
		_, _ = m.Load(oldMf)
		return err
	}
	log.Printf("[manager] 插件 %s 已回退到 v%s", name, version)
	return nil
}

// Health 返回插件健康状态。
func (m *Manager) Health() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string)
	for name, h := range m.handles {
		out[name] = fmt.Sprintf("%s@%s (%s/%s)", h.Name(), h.Version(), h.Mode(), h.Runtime())
	}
	return out
}

// Shutdown 停止所有插件。
func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	handles := m.handles
	m.handles = make(map[string]Handle)
	m.manifests = make(map[string]contract.Manifest)
	m.mu.Unlock()
	for name, h := range handles {
		if err := h.Shutdown(ctx); err != nil {
			log.Printf("[manager] 停止插件 %s 失败: %v", name, err)
		}
	}
}

// —— 内部工具 ——

func (m *Manager) readConfig(name string) map[string]any {
	path := filepath.Join(m.pluginDir, name, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		log.Printf("[manager] 插件 %s 配置解析失败: %v", name, err)
		return map[string]any{}
	}
	return out
}

func (m *Manager) readManifest(name string) (contract.Manifest, error) {
	path := filepath.Join(m.pluginDir, name, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return contract.Manifest{}, err
	}
	var mf contract.Manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return contract.Manifest{}, err
	}
	mf.ApplyDefaults()
	return mf, nil
}

func (m *Manager) snapshotPlugin(name string, mf contract.Manifest) error {
	dir := filepath.Join(m.pluginDir, name, "history", mf.Version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if mf.Runtime == "go-plugin" {
		pluginDir := filepath.Join(m.pluginDir, name)
		src, err := resolveEntrypoint(pluginDir, mf.Entrypoint)
		if err != nil {
			return err
		}
		if err := copyFile(src, filepath.Join(dir, filepath.Base(src))); err != nil {
			return err
		}
	}
	srcMf := filepath.Join(m.pluginDir, name, "manifest.json")
	_ = copyFile(srcMf, filepath.Join(dir, "manifest.json"))
	return nil
}

func (m *Manager) restorePlugin(name string, mf contract.Manifest) error {
	return m.restorePluginVersion(name, mf.Version)
}

func (m *Manager) restorePluginVersion(name, version string) error {
	dir := filepath.Join(m.pluginDir, name, "history", version)
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("历史版本 %s 不存在: %w", version, err)
	}
	if err := copyFile(filepath.Join(dir, "manifest.json"),
		filepath.Join(m.pluginDir, name, "manifest.json")); err != nil {
		return err
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() == "manifest.json" {
			continue
		}
		if err := copyFile(filepath.Join(dir, e.Name()),
			filepath.Join(m.pluginDir, name, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
