// Package config 负责加载 host 的运行时配置。
//
// 配置查找顺序（先找到先用）：
//   1. 命令行 --config 指定的路径
//   2. 环境变量 HOST_CONFIG 指定的路径
//   3. 当前工作目录 ./config.json
//   4. 可执行文件所在目录 ./config.json
//   5. 可执行文件所在目录的上级目录 ../config.json
//
// 加载优先级（后者覆盖前者）：
//   内置默认值 → 配置文件 → 环境变量 → 命令行参数
//
// 路径解析规则：
//   配置文件与命令行中出现的相对路径，会基于"配置文件所在目录"解析为绝对路径。
//   这样 config.json 可以跟随二进制一起移动，不受 CWD 影响。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config 是 host 的运行时配置。
type Config struct {
	// —— 网络 ——
	Host string `json:"host"` // 监听地址，如 "0.0.0.0" 或 "127.0.0.1"
	Port int    `json:"port"` // 监听端口

	// —— 目录 ——
	PluginDir string `json:"plugin_dir"` // 插件目录
	BackupDir string `json:"backup_dir"` // 备份目录

	// —— 其他 ——
	LogLevel string `json:"log_level,omitempty"` // 日志级别（预留）
}

// Default 返回内置默认配置。
func Default() Config {
	return Config{
		Host:      "0.0.0.0",
		Port:      8080,
		PluginDir: "./plugins",
		BackupDir: "./backups",
		LogLevel:  "info",
	}
}

// Load 按优先级加载配置。返回的 Config 已完成所有覆盖与路径解析。
//   - configPath: 命令行 --config 指定的路径（可为空）
//   - flagHost/flagPort: 命令行传入的 host/port（空值/0 表示未指定）
//   - flagPluginDir/flagBackupDir: 命令行传入的目录（空表示未指定）
func Load(configPath, flagHost string, flagPort int, flagPluginDir, flagBackupDir string) (Config, string, error) {
	cfg := Default()

	// 1. 定位配置文件
	path, err := locateConfig(configPath)
	if err != nil {
		return cfg, "", err
	}

	// 2. 加载配置文件
	if path != "" {
		if err := loadFile(path, &cfg); err != nil {
			return cfg, "", fmt.Errorf("加载配置文件 %s 失败: %w", path, err)
		}
	}

	// 3. 环境变量覆盖
	applyEnv(&cfg)

	// 4. 命令行参数覆盖（优先级最高）
	if flagHost != "" {
		cfg.Host = flagHost
	}
	if flagPort != 0 {
		cfg.Port = flagPort
	}
	if flagPluginDir != "" {
		cfg.PluginDir = flagPluginDir
	}
	if flagBackupDir != "" {
		cfg.BackupDir = flagBackupDir
	}

	// 5. 相对路径基于配置文件所在目录解析为绝对路径
	//    若没有配置文件，则基于 CWD 解析。
	//    命令行传入的绝对路径保持不变。
	baseDir := ""
	if path != "" {
		baseDir = filepath.Dir(path)
	}
	cfg.PluginDir = resolveRel(baseDir, cfg.PluginDir)
	cfg.BackupDir = resolveRel(baseDir, cfg.BackupDir)

	// 6. 校验
	if err := validate(&cfg); err != nil {
		return cfg, path, err
	}

	return cfg, path, nil
}

// resolveRel 把相对路径 p 基于 base 解析为绝对路径。
//   - p 为空：原样返回
//   - p 已是绝对路径：原样返回
//   - base 为空：基于 CWD 解析
//   - 其他：基于 base 解析
func resolveRel(base, p string) string {
	if p == "" {
		return p
	}
	if filepath.IsAbs(p) {
		return p
	}
	if base == "" {
		if abs, err := filepath.Abs(p); err == nil {
			return abs
		}
		return p
	}
	return filepath.Clean(filepath.Join(base, p))
}

// locateConfig 按顺序查找配置文件，返回找到的绝对路径。
// 未找到时返回空路径（不算错误，使用内置默认即可）。
func locateConfig(explicit string) (string, error) {
	// 优先级 1：显式指定
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("解析 --config 路径失败: %w", err)
		}
		if _, err := os.Stat(abs); err != nil {
			return "", fmt.Errorf("--config 指定的文件不存在: %s", abs)
		}
		return abs, nil
	}

	// 优先级 2：环境变量
	if env := os.Getenv("HOST_CONFIG"); env != "" {
		abs, err := filepath.Abs(env)
		if err != nil {
			return "", fmt.Errorf("解析 HOST_CONFIG 路径失败: %w", err)
		}
		if _, err := os.Stat(abs); err != nil {
			return "", fmt.Errorf("HOST_CONFIG 指定的文件不存在: %s", abs)
		}
		return abs, nil
	}

	// 优先级 3~5：依次尝试若干默认位置
	candidates := []string{"config.json"}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "config.json"),
			filepath.Join(exeDir, "..", "config.json"),
		)
	}

	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}

	return "", nil
}

// loadFile 读取并解析 JSON 配置文件，把字段写入 cfg。
func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// 用中间变量，避免零值覆盖掉默认值
	tmp := *cfg
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*cfg = tmp
	return nil
}

// applyEnv 用环境变量覆盖字段。
//   HOST_BIND       → Host
//   HOST_PORT       → Port
//   HOST_PLUGIN_DIR → PluginDir
//   HOST_BACKUP_DIR → BackupDir
//   HOST_LOG_LEVEL  → LogLevel
func applyEnv(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv("HOST_BIND")); v != "" {
		cfg.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("HOST_PORT")); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := strings.TrimSpace(os.Getenv("HOST_PLUGIN_DIR")); v != "" {
		cfg.PluginDir = v
	}
	if v := strings.TrimSpace(os.Getenv("HOST_BACKUP_DIR")); v != "" {
		cfg.BackupDir = v
	}
	if v := strings.TrimSpace(os.Getenv("HOST_LOG_LEVEL")); v != "" {
		cfg.LogLevel = v
	}
}

// validate 校验配置合法性。
func validate(cfg *Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("配置 host 不能为空")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("配置 port 非法: %d（应为 1~65535）", cfg.Port)
	}
	if cfg.PluginDir == "" {
		return fmt.Errorf("配置 plugin_dir 不能为空")
	}
	if cfg.BackupDir == "" {
		return fmt.Errorf("配置 backup_dir 不能为空")
	}
	return nil
}

// ListenAddr 返回 host:port 格式的监听地址。
func (c Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DisplayAddr 返回便于展示的访问地址。
// 当 Host 为 0.0.0.0 / :: / 空 时，替换为 127.0.0.1。
func (c Config) DisplayAddr() string {
	h := c.Host
	switch h {
	case "", "0.0.0.0", "::", "[::]":
		h = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", h, c.Port)
}

// String 返回便于打印的配置摘要。
func (c Config) String() string {
	return fmt.Sprintf(
		"host=%s port=%d plugin_dir=%s backup_dir=%s log_level=%s",
		c.Host, c.Port, c.PluginDir, c.BackupDir, c.LogLevel,
	)
}
