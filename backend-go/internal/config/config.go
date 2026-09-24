package config

import (
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
// 覆盖范围：
//   - P0 框架底座：应用、数据库、服务监听、CORS、前端托管
//   - P3 模块管理：模块源码目录、上传目录、ZIP 限制
//   - P5 License 管理：签名密钥、激活服务、机器码覆盖
//   - 密码操作模块：core dispatch.sh 路径、超时、并发
type Config struct {
	// ---- 应用基础 ----
	AppName                  string
	AppEnv                   string
	Debug                    bool
	SecretKey                string
	Algorithm                string
	AccessTokenExpireMinutes int
	LogLevel                 string

	// ---- 数据库 ----
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	// ---- 服务监听 ----
	ServerHost      string
	ServerPort      int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	// ---- CORS ----
	CORSOrigins   []string
	CORSFromEmpty bool

	// ---- 前端产物路径 ----
	FrontendDeployDir string

	// ---- 模块管理（P3） ----
	// ModulesDir 模块源码根目录（相对 backend-go 工作目录）。
	ModulesDir string
	// ModuleUploadDir 模块上传/安装的受控根目录（source_path 白名单）。
	ModuleUploadDir string
	// ModuleZipMaxSize ZIP 单文件大小上限（字节），默认 50MB。
	ModuleZipMaxSize int
	// ModuleZipMaxTotal ZIP 解压后总大小上限（字节），默认 200MB。
	ModuleZipMaxTotal int
	// ModuleZipMaxFiles ZIP 文件数量上限，默认 2000。
	ModuleZipMaxFiles int

	// ---- License 管理（P5） ----
	// LicenseSecretKey License 签名密钥（HMAC-SHA256）。
	LicenseSecretKey string
	// LicenseActivationURL 在线激活服务地址（可选）。
	LicenseActivationURL string
	// LicenseMachineCodeOverride 机器码覆盖（仅开发/测试使用）。
	LicenseMachineCodeOverride string

	// ---- 密码操作模块（P0 平台后端新增） ----
	// CoreDispatchPath core/sbin/dispatch.sh 绝对路径。
	CoreDispatchPath string
	// CoreTimeoutMs core 调用默认超时（毫秒），默认 5000。
	CoreTimeoutMs int
	// CoreMaxConcurrency core 调用最大并发，默认 8。
	CoreMaxConcurrency int
}

// C 全局配置实例，供其他包直接读取（如 license 模块读取 LicenseSecretKey）。
var C *Config

// Load 加载配置。
//
// 读取优先级（从低到高）：
//  1. viper.SetDefault 内置默认值
//  2. .env 文件（godotenv.Load）
//  3. 系统环境变量（v.AutomaticEnv）
//
// 环境变量别名兼容：
//   - SERVER_HOST 优先，回退 HOST
//   - SERVER_PORT 优先，回退 PORT
//   - 与 Python 版无痛升级
//
// 返回：
//   - *Config：配置实例（同时写入全局 C）
//   - []Warning：加载过程中的告警（如生产环境未配置 CORS）
//   - error：加载失败（当前实现不返回 error，预留）
func Load() (*Config, []Warning, error) {
	_ = godotenv.Load()

	v := viper.New()

	// ========================================================================
	// 1. 应用基础默认值
	// ========================================================================
	v.SetDefault("APP_NAME", "Platform Backend")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("DEBUG", true)
	v.SetDefault("SECRET_KEY", "change-this-in-production-please")
	v.SetDefault("ALGORITHM", "HS256")
	v.SetDefault("ACCESS_TOKEN_EXPIRE_MINUTES", 1440)
	v.SetDefault("LOG_LEVEL", "DEBUG")

	// ========================================================================
	// 2. 数据库默认值
	// ========================================================================
	v.SetDefault("DATABASE_URL", "sqlite+aiosqlite:///./app.db")
	v.SetDefault("DB_MAX_OPEN_CONNS", 20)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "1h")

	// ========================================================================
	// 3. 服务监听默认值
	// ========================================================================
	// 同时定义 SERVER_HOST / HOST 与 SERVER_PORT / PORT，
	// 通过下方逻辑决定优先级。
	v.SetDefault("SERVER_HOST", "")
	v.SetDefault("HOST", "")
	v.SetDefault("SERVER_PORT", 0)
	v.SetDefault("PORT", 0)
	v.SetDefault("READ_TIMEOUT", "10s")
	v.SetDefault("WRITE_TIMEOUT", "10s")
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")

	// ========================================================================
	// 4. CORS 默认值
	// ========================================================================
	v.SetDefault("CORS_ORIGINS", "")

	// ========================================================================
	// 5. 前端产物路径默认值
	// ========================================================================
	v.SetDefault("FRONTEND_DEPLOY_DIR", "")

	// ========================================================================
	// 6. 模块管理默认值（P3）
	// ========================================================================
	v.SetDefault("MODULES_DIR", "src/modules")
	v.SetDefault("MODULE_UPLOAD_DIR", "./uploads/modules")
	v.SetDefault("MODULE_ZIP_MAX_SIZE", 50*1024*1024)   // 50MB
	v.SetDefault("MODULE_ZIP_MAX_TOTAL", 200*1024*1024) // 200MB
	v.SetDefault("MODULE_ZIP_MAX_FILES", 2000)

	// ========================================================================
	// 7. License 管理默认值（P5）
	// ========================================================================
	v.SetDefault("LICENSE_SECRET_KEY", "change-this-license-secret-in-production")
	v.SetDefault("LICENSE_ACTIVATION_URL", "")
	v.SetDefault("LICENSE_MACHINE_CODE_OVERRIDE", "")

	// ========================================================================
	// 8. 密码操作模块默认值（P0 平台后端）
	// ========================================================================
	v.SetDefault("CORE_DISPATCH_PATH", "/opt/core/sbin/dispatch.sh")
	v.SetDefault("CORE_TIMEOUT_MS", 5000)
	v.SetDefault("CORE_MAX_CONCURRENCY", 8)

	// ========================================================================
	// 环境变量注入
	// ========================================================================
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ========================================================================
	// 应用环境归一化
	// ========================================================================
	appEnv := strings.ToLower(v.GetString("APP_ENV"))

	// ========================================================================
	// 服务监听：SERVER_HOST / SERVER_PORT 优先，回退 HOST / PORT
	//
	// 别名优先级说明（N-05）：
	//   - SERVER_HOST 优先于 HOST
	//   - SERVER_PORT 优先于 PORT
	//   - 二者均未设置时使用内置默认值
	// ========================================================================
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

	// ========================================================================
	// CORS 解析与告警
	//
	// V11-P2-03 修复：无论 raw 为空、为空白、还是 split 后为空，
	// 都统一产生 warning 并置 CORSFromEmpty=true。
	// ========================================================================
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
		// 场景 1：完全未配置
		cors = []string{"*"}
		corsFromEmpty = true
		appendCORSWarning()
	case corsRaw == "*":
		// 场景 2：显式配置为通配符（正常，不告警）
		cors = []string{"*"}
	default:
		// 场景 3：逗号分隔
		for _, item := range strings.Split(corsRaw, ",") {
			item = strings.TrimSpace(item)
			if item != "" {
				cors = append(cors, item)
			}
		}
		if len(cors) == 0 {
			// 场景 3b：形如 " , " 的输入，split 后全部为空
			cors = []string{"*"}
			corsFromEmpty = true
			appendCORSWarning()
		}
	}

	// ========================================================================
	// 组装 Config
	// ========================================================================
	cfg := &Config{
		// ---- 应用基础 ----
		AppName:                  v.GetString("APP_NAME"),
		AppEnv:                   appEnv,
		Debug:                    v.GetBool("DEBUG"),
		SecretKey:                v.GetString("SECRET_KEY"),
		Algorithm:                v.GetString("ALGORITHM"),
		AccessTokenExpireMinutes: v.GetInt("ACCESS_TOKEN_EXPIRE_MINUTES"),
		LogLevel:                 v.GetString("LOG_LEVEL"),

		// ---- 数据库 ----
		DatabaseURL:       v.GetString("DATABASE_URL"),
		DBMaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		DBConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),

		// ---- 服务监听 ----
		ServerHost:      serverHost,
		ServerPort:      serverPort,
		ReadTimeout:     v.GetDuration("READ_TIMEOUT"),
		WriteTimeout:    v.GetDuration("WRITE_TIMEOUT"),
		ShutdownTimeout: v.GetDuration("SHUTDOWN_TIMEOUT"),

		// ---- CORS ----
		CORSOrigins:   cors,
		CORSFromEmpty: corsFromEmpty,

		// ---- 前端产物路径 ----
		FrontendDeployDir: v.GetString("FRONTEND_DEPLOY_DIR"),

		// ---- 模块管理（P3） ----
		ModulesDir:        v.GetString("MODULES_DIR"),
		ModuleUploadDir:   v.GetString("MODULE_UPLOAD_DIR"),
		ModuleZipMaxSize:  v.GetInt("MODULE_ZIP_MAX_SIZE"),
		ModuleZipMaxTotal: v.GetInt("MODULE_ZIP_MAX_TOTAL"),
		ModuleZipMaxFiles: v.GetInt("MODULE_ZIP_MAX_FILES"),

		// ---- License 管理（P5） ----
		LicenseSecretKey:           v.GetString("LICENSE_SECRET_KEY"),
		LicenseActivationURL:       v.GetString("LICENSE_ACTIVATION_URL"),
		LicenseMachineCodeOverride: v.GetString("LICENSE_MACHINE_CODE_OVERRIDE"),

		// ---- 密码操作模块（P0 平台后端） ----
		CoreDispatchPath:   v.GetString("CORE_DISPATCH_PATH"),
		CoreTimeoutMs:      v.GetInt("CORE_TIMEOUT_MS"),
		CoreMaxConcurrency: v.GetInt("CORE_MAX_CONCURRENCY"),
	}

	// 写入全局实例
	C = cfg

	return cfg, warnings, nil
}
