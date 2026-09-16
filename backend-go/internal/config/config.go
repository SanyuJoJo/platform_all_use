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
}
var C *Config
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
			// V11-P2-03：形如 " , " 的输入，split 后全部为空
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
	}
	C = cfg
	return cfg, warnings, nil
}
