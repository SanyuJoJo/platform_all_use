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
type Config struct {
	AppName                  string
	AppEnv                   string
	Debug                    bool
	SecretKey                string
	Algorithm                string
	AccessTokenExpireMinutes int
	DatabaseURL              string
	LogLevel                 string
	CORSOrigins              []string
	CORSFromEmpty            bool
	ServerHost               string
	ServerPort               int
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	ShutdownTimeout          time.Duration
	DBMaxOpenConns           int
	DBMaxIdleConns           int
	DBConnMaxLifetime        time.Duration
	FrontendDeployDir        string

	// ---- 模块管理相关（P3 阶段新增） ----
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
}

var C *Config

// Load 加载配置。
//
// v1.3（BUG-03 ~ BUG-07）：新增模块管理相关字段
// （ModulesDir / ModuleUploadDir / ModuleZipMaxSize / ModuleZipMaxTotal /
// ModuleZipMaxFiles），与 Python 版 config.py 完全一致。
func Load() (*Config, []Warning, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetDefault("APP_NAME", "Platform Backend")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("DEBUG", true)
	v.SetDefault("SECRET_KEY", "change-this-in-production-please")
	v.SetDefault("ALGORITHM", "HS256")
	v.SetDefault("ACCESS_TOKEN_EXPIRE_MINUTES", 1440)
	v.SetDefault("DATABASE_URL", "sqlite+aiosqlite:///./app.db")
	v.SetDefault("LOG_LEVEL", "DEBUG")
	v.SetDefault("CORS_ORIGINS", "")
	v.SetDefault("SERVER_HOST", "")
	v.SetDefault("HOST", "")
	v.SetDefault("SERVER_PORT", 0)
	v.SetDefault("PORT", 0)
	v.SetDefault("READ_TIMEOUT", "10s")
	v.SetDefault("WRITE_TIMEOUT", "10s")
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")
	v.SetDefault("DB_MAX_OPEN_CONNS", 20)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "1h")
	v.SetDefault("FRONTEND_DEPLOY_DIR", "")

	// ---- 模块管理相关默认值（与 Python 版 config.py 一致） ----
	v.SetDefault("MODULES_DIR", "src/modules")
	v.SetDefault("MODULE_UPLOAD_DIR", "./uploads/modules")
	v.SetDefault("MODULE_ZIP_MAX_SIZE", 50*1024*1024)      // 50MB
	v.SetDefault("MODULE_ZIP_MAX_TOTAL", 200*1024*1024)    // 200MB
	v.SetDefault("MODULE_ZIP_MAX_FILES", 2000)

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	appEnv := strings.ToLower(v.GetString("APP_ENV"))

	// ---- 服务监听：SERVER_HOST / SERVER_PORT 优先，回退 HOST / PORT ----
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

	// ---- CORS ----
	// V11-P2-03 修复：无论 raw 为空、为空白、还是 split 后为空，
	// 都统一产生 warning 并置 CORSFromEmpty=true。
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
		AppName:                  v.GetString("APP_NAME"),
		AppEnv:                   appEnv,
		Debug:                    v.GetBool("DEBUG"),
		SecretKey:                v.GetString("SECRET_KEY"),
		Algorithm:                v.GetString("ALGORITHM"),
		AccessTokenExpireMinutes: v.GetInt("ACCESS_TOKEN_EXPIRE_MINUTES"),
		DatabaseURL:              v.GetString("DATABASE_URL"),
		LogLevel:                 v.GetString("LOG_LEVEL"),
		CORSOrigins:              cors,
		CORSFromEmpty:            corsFromEmpty,
		ServerHost:               serverHost,
		ServerPort:               serverPort,
		ReadTimeout:              v.GetDuration("READ_TIMEOUT"),
		WriteTimeout:             v.GetDuration("WRITE_TIMEOUT"),
		ShutdownTimeout:          v.GetDuration("SHUTDOWN_TIMEOUT"),
		DBMaxOpenConns:           v.GetInt("DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns:           v.GetInt("DB_MAX_IDLE_CONNS"),
		DBConnMaxLifetime:        v.GetDuration("DB_CONN_MAX_LIFETIME"),
		FrontendDeployDir:        v.GetString("FRONTEND_DEPLOY_DIR"),

		// ---- 模块管理相关 ----
		ModulesDir:        v.GetString("MODULES_DIR"),
		ModuleUploadDir:   v.GetString("MODULE_UPLOAD_DIR"),
		ModuleZipMaxSize:  v.GetInt("MODULE_ZIP_MAX_SIZE"),
		ModuleZipMaxTotal: v.GetInt("MODULE_ZIP_MAX_TOTAL"),
		ModuleZipMaxFiles: v.GetInt("MODULE_ZIP_MAX_FILES"),
	}
	C = cfg
	return cfg, warnings, nil
}
