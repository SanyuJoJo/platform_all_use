package config
import (
	"strings"
	"time"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)
// Warning 配置加载过程中的告警/错误。
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
	// ---- 模块管理相关 ----
	ModulesDir        string
	ModuleUploadDir   string
	ModuleZipMaxSize  int
	ModuleZipMaxTotal int
	ModuleZipMaxFiles int
	// ---- License 管理相关（v1.0 新增）----
	// LicenseSecretKey License 签名密钥（HMAC-SHA256）。
	LicenseSecretKey string
	// LicenseActivationURL 在线激活服务地址（可选）。
	LicenseActivationURL string
	// LicenseMachineCodeOverride 机器码覆盖（仅开发/测试使用）。
	LicenseMachineCodeOverride string
}
var C *Config
// Load 加载配置。
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
	// 模块管理相关默认值
	v.SetDefault("MODULES_DIR", "src/modules")
	v.SetDefault("MODULE_UPLOAD_DIR", "./uploads/modules")
	v.SetDefault("MODULE_ZIP_MAX_SIZE", 50*1024*1024)
	v.SetDefault("MODULE_ZIP_MAX_TOTAL", 200*1024*1024)
	v.SetDefault("MODULE_ZIP_MAX_FILES", 2000)
	// License 管理相关默认值
	v.SetDefault("LICENSE_SECRET_KEY", "change-this-license-secret-in-production")
	v.SetDefault("LICENSE_ACTIVATION_URL", "")
	v.SetDefault("LICENSE_MACHINE_CODE_OVERRIDE", "")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	appEnv := strings.ToLower(v.GetString("APP_ENV"))
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
		ModulesDir:        v.GetString("MODULES_DIR"),
		ModuleUploadDir:   v.GetString("MODULE_UPLOAD_DIR"),
		ModuleZipMaxSize:  v.GetInt("MODULE_ZIP_MAX_SIZE"),
		ModuleZipMaxTotal: v.GetInt("MODULE_ZIP_MAX_TOTAL"),
		ModuleZipMaxFiles: v.GetInt("MODULE_ZIP_MAX_FILES"),
		LicenseSecretKey:           v.GetString("LICENSE_SECRET_KEY"),
		LicenseActivationURL:       v.GetString("LICENSE_ACTIVATION_URL"),
		LicenseMachineCodeOverride: v.GetString("LICENSE_MACHINE_CODE_OVERRIDE"),
	}
	C = cfg
	return cfg, warnings, nil
}
