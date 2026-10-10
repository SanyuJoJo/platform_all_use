package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Warning 表示配置加载过程中的告警/错误，由 main 在 logger 初始化后统一输出。
type Warning struct {
	Level   string // "error" | "warn"
	Message string
}

// Config 应用配置。
//
// 字段分组：
//   - 基础应用信息
//   - 数据库
//   - 服务监听
//   - CORS
//   - 前端静态资源
//   - 模块管理
//   - License
//   - 前端动态菜单（v1.3 已合入）
//   - Core 引擎（兼容字段，第二版已合入）
//   - 证书路径（证书路径改造新增）
type Config struct {
	// ---------- 基础应用信息 ----------
	AppName                  string
	AppEnv                   string
	Debug                    bool
	SecretKey                string
	Algorithm                string
	AccessTokenExpireMinutes int
	LogLevel                 string

	// ---------- 数据库 ----------
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// ---------- 服务监听 ----------
	ServerHost      string
	ServerPort      int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// ---------- CORS ----------
	CORSOrigins   []string
	CORSFromEmpty bool

	// ---------- 前端静态资源 ----------
	FrontendDeployDir string

	// ---------- 模块管理 ----------
	ModulesDir        string
	ModuleUploadDir   string
	ModuleZipMaxSize  int
	ModuleZipMaxTotal int
	ModuleZipMaxFiles int

	// ---------- License ----------
	LicenseSecretKey           string
	LicenseActivationURL       string
	LicenseMachineCodeOverride string

	// ---------- 前端动态菜单（v1.3 已合入） ----------
	// FrontendConfigPath 前端菜单与子应用入口配置文件路径。
	// 为空或文件不存在时，沿用代码内默认菜单。
	FrontendConfigPath string

	// ---------- Core 引擎（第二版已合入，兼容保留） ----------
	// CoreDispatchPath core 调用入口。
	//
	// 语义：
	//   - 指向脚本：旧模式（core/sbin/*.sh）；
	//   - 指向二进制：等价于 CoreGoBin（推荐）。
	//
	// 新代码统一优先使用 CoreGoBin；本字段仅作兼容回退。
	CoreDispatchPath string
	// CoreRoot core 根目录（CoreDispatchPath 指向 bin/core 时的显式根目录）。
	CoreRoot string
	// CoreTimeoutMs core 调用超时（毫秒）。
	CoreTimeoutMs int
	// CoreMaxConcurrency 最大并发调用数。
	CoreMaxConcurrency int

	// ---------- 证书路径（证书路径改造新增） ----------
	// InstallRoot 安装根目录，默认 /opt/plats_tool。
	InstallRoot string
	// CoreGoBin core-go 主二进制（替代 core/sbin/*.sh）。
	// 优先级高于 CoreDispatchPath。
	CoreGoBin string
	// CoreGoWhiteboxBin 白盒 SM4 工具路径。
	CoreGoWhiteboxBin string
	// CoreGoOpensslBin 铜锁 openssl 路径。
	CoreGoOpensslBin string
	// CertRoot 证书数据根目录。
	CertRoot string
	// CARoot CA 根目录。
	CARoot string
	// ServerRoot SERVER 证书根目录。
	ServerRoot string
	// CertTmpDir 证书临时目录（解密私钥用，权限 0700）。
	CertTmpDir string
}

// C 全局配置实例。
var C *Config

// Load 加载配置。
//
// 优先级（从高到低）：
//  1. 显式环境变量（含 .env）
//  2. 派生自 PLATS_INSTALL_ROOT
//  3. 内置默认值
func Load() (*Config, []Warning, error) {
	_ = godotenv.Load()

	v := viper.New()

	// ---------- 基础默认值 ----------
	v.SetDefault("APP_NAME", "Platform Backend")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("DEBUG", true)
	v.SetDefault("SECRET_KEY", "change-this-in-production-please")
	v.SetDefault("ALGORITHM", "HS256")
	v.SetDefault("ACCESS_TOKEN_EXPIRE_MINUTES", 1440)
	v.SetDefault("LOG_LEVEL", "DEBUG")

	v.SetDefault("DATABASE_URL", "sqlite+aiosqlite:///./app.db")
	v.SetDefault("DB_MAX_OPEN_CONNS", 20)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "1h")

	v.SetDefault("SERVER_HOST", "")
	v.SetDefault("HOST", "")
	v.SetDefault("SERVER_PORT", 0)
	v.SetDefault("PORT", 0)
	v.SetDefault("READ_TIMEOUT", "10s")
	v.SetDefault("WRITE_TIMEOUT", "10s")
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")

	v.SetDefault("CORS_ORIGINS", "")
	v.SetDefault("FRONTEND_DEPLOY_DIR", "")

	// ---------- 模块管理默认值 ----------
	v.SetDefault("MODULES_DIR", "src/modules")
	v.SetDefault("MODULE_UPLOAD_DIR", "./uploads/modules")
	v.SetDefault("MODULE_ZIP_MAX_SIZE", 50*1024*1024)
	v.SetDefault("MODULE_ZIP_MAX_TOTAL", 200*1024*1024)
	v.SetDefault("MODULE_ZIP_MAX_FILES", 2000)

	// ---------- License 默认值 ----------
	v.SetDefault("LICENSE_SECRET_KEY", "change-this-license-secret-in-production")
	v.SetDefault("LICENSE_ACTIVATION_URL", "")
	v.SetDefault("LICENSE_MACHINE_CODE_OVERRIDE", "")

	// ---------- 前端动态菜单默认值（v1.3 已合入） ----------
	v.SetDefault("FRONTEND_CONFIG_PATH", "configs/frontend.yaml")

	// ---------- Core 引擎默认值（第二版已合入） ----------
	v.SetDefault("CORE_DISPATCH_PATH", "")
	v.SetDefault("CORE_ROOT", "")
	v.SetDefault("CORE_TIMEOUT_MS", 10000)
	v.SetDefault("CORE_MAX_CONCURRENCY", 8)

	// ---------- 证书路径默认值（证书路径改造新增） ----------
	v.SetDefault("PLATS_INSTALL_ROOT", "/opt/plats_tool")
	v.SetDefault("PLATS_CORE_GO_BIN", "")
	v.SetDefault("PLATS_CORE_GO_WHITEBOX_BIN", "")
	v.SetDefault("PLATS_CORE_GO_OPENSSL_BIN", "")
	v.SetDefault("PLATS_CERT_ROOT", "")
	v.SetDefault("PLATS_CA_ROOT", "")
	v.SetDefault("PLATS_SERVER_ROOT", "")
	v.SetDefault("PLATS_CERT_TMP_DIR", "")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	appEnv := strings.ToLower(v.GetString("APP_ENV"))

	// ---------- 服务监听：SERVER_HOST / SERVER_PORT 优先，回退 HOST / PORT ----------
	serverHost := strings.TrimSpace(v.GetString("SERVER_HOST"))
	if serverHost == "" {
		serverHost = strings.TrimSpace(v.GetString("HOST"))
	}
	if serverHost == "" {
		serverHost = "0.0.0.0"
	}
	serverPort := v.GetInt("SERVER_PORT")
	if serverPort == 0 {
		serverPort = v.GetInt("PORT")
	}
	if serverPort == 0 {
		serverPort = 8000
	}

	// ---------- CORS ----------
	var warnings []Warning
	corsRaw := strings.TrimSpace(v.GetString("CORS_ORIGINS"))
	corsFromEmpty := false
	var cors []string
	appendCORSWarning := func() {
		if appEnv == "production" {
			warnings = append(warnings, Warning{
				Level:   "error",
				Message: "生产环境未配置 CORS_ORIGINS，已回退为 ['*']，这是不安全配置，请立即在环境变量中指定允许的域名",
			})
		} else {
			warnings = append(warnings, Warning{
				Level:   "warn",
				Message: "CORS_ORIGINS 未配置，回退为 ['*']，生产环境请显式指定允许的域名",
			})
		}
	}
	switch {
	case corsRaw == "":
		cors = []string{"*"}
		corsFromEmpty = true
		appendCORSWarning()
	case corsRaw == "*":
		cors = []string{"*"}
	default:
		for _, item := range strings.Split(corsRaw, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				cors = append(cors, item)
			}
		}
		if len(cors) == 0 {
			cors = []string{"*"}
			corsFromEmpty = true
			appendCORSWarning()
		}
	}

	cfg := &Config{
		// 基础
		AppName:                  v.GetString("APP_NAME"),
		AppEnv:                   appEnv,
		Debug:                    v.GetBool("DEBUG"),
		SecretKey:                v.GetString("SECRET_KEY"),
		Algorithm:                v.GetString("ALGORITHM"),
		AccessTokenExpireMinutes: v.GetInt("ACCESS_TOKEN_EXPIRE_MINUTES"),
		LogLevel:                 v.GetString("LOG_LEVEL"),

		// 数据库
		DatabaseURL:       v.GetString("DATABASE_URL"),
		DBMaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		DBConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),

		// 服务监听
		ServerHost:      serverHost,
		ServerPort:      serverPort,
		ReadTimeout:     v.GetDuration("READ_TIMEOUT"),
		WriteTimeout:    v.GetDuration("WRITE_TIMEOUT"),
		ShutdownTimeout: v.GetDuration("SHUTDOWN_TIMEOUT"),

		// CORS
		CORSOrigins:   cors,
		CORSFromEmpty: corsFromEmpty,

		// 前端静态资源
		FrontendDeployDir: v.GetString("FRONTEND_DEPLOY_DIR"),

		// 模块管理
		ModulesDir:        v.GetString("MODULES_DIR"),
		ModuleUploadDir:   v.GetString("MODULE_UPLOAD_DIR"),
		ModuleZipMaxSize:  v.GetInt("MODULE_ZIP_MAX_SIZE"),
		ModuleZipMaxTotal: v.GetInt("MODULE_ZIP_MAX_TOTAL"),
		ModuleZipMaxFiles: v.GetInt("MODULE_ZIP_MAX_FILES"),

		// License
		LicenseSecretKey:           v.GetString("LICENSE_SECRET_KEY"),
		LicenseActivationURL:       v.GetString("LICENSE_ACTIVATION_URL"),
		LicenseMachineCodeOverride: v.GetString("LICENSE_MACHINE_CODE_OVERRIDE"),

		// 前端动态菜单（v1.3 已合入）
		FrontendConfigPath: v.GetString("FRONTEND_CONFIG_PATH"),

		// Core 引擎（第二版已合入）
		CoreDispatchPath:   v.GetString("CORE_DISPATCH_PATH"),
		CoreRoot:           v.GetString("CORE_ROOT"),
		CoreTimeoutMs:      v.GetInt("CORE_TIMEOUT_MS"),
		CoreMaxConcurrency: v.GetInt("CORE_MAX_CONCURRENCY"),
	}

	// ---------- 路径归一化（Core 引擎 + 证书路径） ----------
	applyPathConfig(cfg, v)

	C = cfg
	return cfg, warnings, nil
}

// applyPathConfig 归一化 Core 引擎与证书路径配置。
//
// 覆盖字段：
//   - InstallRoot
//   - CoreGoBin / CoreGoWhiteboxBin / CoreGoOpensslBin
//   - CertRoot / CARoot / ServerRoot / CertTmpDir
//
// 优先级：
//  1. 显式环境变量（PLATS_*）
//  2. 派生自 InstallRoot
//  3. 内置默认值
//
// 兼容性：
//   - 若 PLATS_CORE_GO_BIN 为空，但 CORE_DISPATCH_PATH 非空，
//     则 CoreGoBin 取 CORE_DISPATCH_PATH；
//   - 若两者都为空，则 CoreGoBin 派生自 InstallRoot。
func applyPathConfig(cfg *Config, v *viper.Viper) {
	// InstallRoot
	installRoot := cleanAbs(v.GetString("PLATS_INSTALL_ROOT"), "/opt/plats_tool")
	cfg.InstallRoot = installRoot

	// CoreGoBin：显式配置 > CoreDispatchPath 回退 > 派生自 InstallRoot
	explicitCoreGoBin := strings.TrimSpace(v.GetString("PLATS_CORE_GO_BIN"))
	switch {
	case explicitCoreGoBin != "":
		cfg.CoreGoBin = cleanAbs(explicitCoreGoBin, "")
	case strings.TrimSpace(cfg.CoreDispatchPath) != "":
		cfg.CoreGoBin = cleanAbs(cfg.CoreDispatchPath, "")
	default:
		cfg.CoreGoBin = cleanAbs("", filepath.Join(installRoot, "core", "bin", "core"))
	}

	// 白盒与铜锁 openssl
	cfg.CoreGoWhiteboxBin = cleanAbs(
		v.GetString("PLATS_CORE_GO_WHITEBOX_BIN"),
		filepath.Join(installRoot, "core-go", "bin", "whitebox_sm4"),
	)
	cfg.CoreGoOpensslBin = cleanAbs(
		v.GetString("PLATS_CORE_GO_OPENSSL_BIN"),
		filepath.Join(installRoot, "core-go", "libs", "tongsuo", "bin", "openssl"),
	)

	// 证书路径
	certRoot := cleanAbs(
		v.GetString("PLATS_CERT_ROOT"),
		filepath.Join(installRoot, "user_data", "cert"),
	)
	cfg.CertRoot = certRoot

	cfg.CARoot = cleanAbs(
		v.GetString("PLATS_CA_ROOT"),
		filepath.Join(certRoot, "ca"),
	)
	cfg.ServerRoot = cleanAbs(
		v.GetString("PLATS_SERVER_ROOT"),
		filepath.Join(certRoot, "server"),
	)
	cfg.CertTmpDir = cleanAbs(
		v.GetString("PLATS_CERT_TMP_DIR"),
		filepath.Join(certRoot, "tmp"),
	)
}

// cleanAbs 归一化路径：空值回退默认，返回绝对路径。
//
// 说明：
//   - 若 def 为空且 p 为空，返回空字符串（用于调用方已确认 p 非空场景）；
//   - 相对路径基于当前工作目录转绝对，避免不同 CWD 下行为漂移。
func cleanAbs(p, def string) string {
	if strings.TrimSpace(p) == "" {
		p = def
	}
	if strings.TrimSpace(p) == "" {
		return ""
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	return filepath.Clean(abs)
}

// Atoui 安全字符串转 int。
//
// 用于少量需要手工解析的环境变量场景；
// 一般的配置项推荐使用 viper.GetInt。
func Atoui(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// Exists 判断路径是否存在。
func Exists(p string) bool {
	if p == "" {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}
