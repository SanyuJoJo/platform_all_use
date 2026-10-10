package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/yourorg/core-go/internal/platform"
)

// Config 运行配置。
type Config struct {
	CoreRoot          string
	ConfDir           string
	LogDir            string
	DataDir           string
	TmpDir            string
	KeysDir           string
	MasterKeyPath     string
	OpenSSLBin        string
	WhiteboxBin       string // ★ 新增：白盒工具路径
	WhitelistPath     string
	AuditLogPath      string
	TongsuoMinVersion string
	TimeoutMs         int
	MaxConcurrency    int
}

type yamlConf struct {
	Core struct {
		TongsuoOpenSSLBin string `yaml:"tongsuo_openssl_bin"`
		TongsuoMinVersion string `yaml:"tongsuo_min_version"`
		WhiteboxBin       string `yaml:"whitebox_bin"` // ★ 新增
		TimeoutMs         int    `yaml:"timeout_ms"`
		MaxConcurrency    int    `yaml:"max_concurrency"`
		DataDir           string `yaml:"data_dir"`
		LogDir            string `yaml:"log_dir"`
		TmpDir            string `yaml:"tmp_dir"`
		MasterKeyPath     string `yaml:"master_key_path"`
		AuditLogFile      string `yaml:"audit_log_file"`
	} `yaml:"core"`
}

// Load 加载配置。
func Load() (*Config, error) {
	coreRoot := os.Getenv("CORE_ROOT")
	if coreRoot == "" {
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			for _, candidate := range []string{
				filepath.Dir(exeDir),
				exeDir,
			} {
				if _, err := os.Stat(filepath.Join(candidate, "conf")); err == nil {
					coreRoot = candidate
					break
				}
			}
		}
	}
	if coreRoot == "" {
		coreRoot, _ = os.Getwd()
	}
	coreRoot, _ = filepath.Abs(coreRoot)

	confDir := filepath.Join(coreRoot, "conf")
	confFile := filepath.Join(confDir, "core.yaml")

	var yc yamlConf
	if data, err := os.ReadFile(confFile); err == nil {
		_ = yaml.Unmarshal(data, &yc)
	}

	cfg := &Config{
		CoreRoot:          coreRoot,
		ConfDir:           confDir,
		LogDir:            resolvePath(coreRoot, defaultStr(yc.Core.LogDir, "logs")),
		DataDir:           resolvePath(coreRoot, defaultStr(yc.Core.DataDir, "data")),
		TmpDir:            resolvePath(coreRoot, defaultStr(yc.Core.TmpDir, "tmp")),
		WhitelistPath:     filepath.Join(confDir, "algorithm-whitelist.yaml"),
		OpenSSLBin:        yc.Core.TongsuoOpenSSLBin,
		TongsuoMinVersion: defaultStr(yc.Core.TongsuoMinVersion, "8.5.0"),
		TimeoutMs:         defaultInt(yc.Core.TimeoutMs, 5000),
		MaxConcurrency:    defaultInt(yc.Core.MaxConcurrency, 8),
	}

	cfg.KeysDir = filepath.Join(cfg.DataDir, "keys")
	cfg.MasterKeyPath = resolvePath(coreRoot,
		defaultStr(yc.Core.MasterKeyPath, filepath.Join("data", "keys", "master.key")))
	cfg.AuditLogPath = resolvePath(coreRoot,
		defaultStr(yc.Core.AuditLogFile, filepath.Join("logs", "audit.jsonl")))

	// ★ 新增：白盒工具路径
	cfg.WhiteboxBin = resolvePath(coreRoot,
		defaultStr(yc.Core.WhiteboxBin, "bin/whitebox_sm4"))

	// OpenSSLBin 若未配置，按平台自动解析
	if cfg.OpenSSLBin == "" {
		cfg.OpenSSLBin = filepath.Join(coreRoot, "libs", "openssl",
			platform.OS()+"-"+platform.Arch(),
			"openssl"+platform.ExeSuffix())
	}

	for _, d := range []string{cfg.LogDir, cfg.DataDir, cfg.TmpDir, cfg.KeysDir} {
		if err := os.MkdirAll(d, 0750); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", d, err)
		}
	}
	return cfg, nil
}

func resolvePath(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, filepath.FromSlash(p))
}

func defaultStr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func defaultInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
