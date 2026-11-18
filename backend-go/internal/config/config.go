package config

import (
	"fmt"
	"net"
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
//   - 前端动态菜单
//   - Core 引擎（兼容字段）
//   - 证书路径
//   - 媒体管理（音视频模块）
//   - ZLMediaKit（音视频模块）
//   - ZLMediaKit 进程管理（音视频模块）
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

	// ---------- 前端动态菜单 ----------
	FrontendConfigPath string

	// ---------- Core 引擎（兼容字段） ----------
	CoreDispatchPath   string
	CoreRoot           string
	CoreTimeoutMs      int
	CoreMaxConcurrency int

	// ---------- 证书路径 ----------
	InstallRoot       string
	CoreGoBin         string
	CoreGoWhiteboxBin string
	// CoreGoOpensslBin 铜锁 openssl 路径。
	//
	// ★ 实际部署路径为 <install_root>/core/run/bin/openssl，
	//   由 core/Makefile 的 TONSUO_DEPLOY_OPENSSL 保证。
	//   动态库仍在 <install_root>/core/libs/tongsuo/libs（含 ossl-modules/oqsprovider.so）。
	CoreGoOpensslBin string
	CertRoot         string
	CARoot           string
	ServerRoot       string
	CertTmpDir       string

	// ---------- 媒体管理（音视频模块） ----------
	// MediaRoot 媒体文件存储根目录（默认派生自 InstallRoot：<root>/user_data/media）。
	//
	// 目录结构：
	//   <MediaRoot>/
	//   ├── video/                 点播视频文件
	//   │   └── thumbnails/        视频缩略图
	//   ├── audio/                 点播音频文件
	//   └── tmp/uploads/<id>/      分片上传临时目录（2h 过期清理）
	MediaRoot        string
	MediaChunkSize   int64
	MediaSessionTTL  time.Duration
	MediaFFprobePath string
	MediaFFmpegPath  string

	// ---------- ZLMediaKit（音视频模块） ----------
	// ZLMBaseURL ZLM HTTP API 基础地址（后端调用）。
	//
	// 优先级：
	//   1. .env 显式配置（EnvExplicit("ZLM_BASE_URL")）—— 完整 URL，最高优先级；
	//   2. 由 ZLMHTTPPort 自动派生 http://127.0.0.1:<port>。
	//
	// 未显式配置时，后端会从 ZLM 的 config.ini 读 [http] port
	// 并自动派生 ZLMBaseURL，运维无需手工同步。
	ZLMBaseURL string
	// ZLMSecret ZLM HTTP API secret。
	//
	// 未显式配置时，从 ZLM 的 config.ini [api] secret 读取；
	// 若 config.ini 里是 ZLM 官方示例值（035c...），后端会主动
	// 生成随机值写回 config.ini（避免 ZLM 启动时自动改写导致前后端不一致）。
	ZLMSecret string
	// ZLMHTTPHost 对外可达的 HTTP host（拼播放地址用，浏览器须能访问）。
	ZLMHTTPHost string
	// ZLMHTTPPort 对外 HTTP 端口（默认从 config.ini [http] port 读取）。
	ZLMHTTPPort int
	// ZLMRTMPPort RTMP 端口（默认 1935）。
	ZLMRTMPPort int
	// ZLMRTCPort WebRTC 端口（默认 8000；0 表示未启用）。
	ZLMRTCPort int
	// PublicBaseURL 后端对外可达基础 URL（拼接媒体完整地址用）。
	PublicBaseURL string

	// ---------- ZLMediaKit 进程管理（音视频模块） ----------
	// ZLMAutoStart 后端启动时若 ZLM 探活失败则自动托管拉起。
	ZLMAutoStart bool
	// ZLMEngineBin ZLM 可执行文件路径（默认派生 <video>/run/bin/MediaServer）。
	ZLMEngineBin string
	// ZLMEngineConfig ZLM 配置文件路径（默认派生 <video>/conf/zlmediakit/config.ini）。
	ZLMEngineConfig string
	// ZLMEngineLibDir 动态库目录（作为子进程 LD_LIBRARY_PATH）。
	ZLMEngineLibDir string
	// ZLMEngineWorkDir ZLM 子进程工作目录。
	ZLMEngineWorkDir string
	// ZLMEngineLogFile ZLM stdout/stderr 重定向的日志文件。
	ZLMEngineLogFile string
	// ZLMEngineShutdownOnExit 后端退出时是否 kill 托管的 ZLM。
	ZLMEngineShutdownOnExit bool
}

// C 全局配置实例。
var C *Config

// Load 加载配置。
//
// 优先级（从高到低）：
//  1. 显式环境变量（含 .env）
//  2. ZLM config.ini（仅 ZLM_HTTP_PORT / ZLM_SECRET 两项，且未显式配置时）
//  3. 派生自 PLATS_INSTALL_ROOT
//  4. 内置默认值
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

	// ---------- 前端动态菜单默认值 ----------
	v.SetDefault("FRONTEND_CONFIG_PATH", "configs/frontend.yaml")

	// ---------- Core 引擎默认值 ----------
	v.SetDefault("CORE_DISPATCH_PATH", "")
	v.SetDefault("CORE_ROOT", "")
	v.SetDefault("CORE_TIMEOUT_MS", 10000)
	v.SetDefault("CORE_MAX_CONCURRENCY", 8)

	// ---------- 证书路径默认值 ----------
	v.SetDefault("PLATS_INSTALL_ROOT", "/opt/plats_tool")
	v.SetDefault("PLATS_CORE_GO_BIN", "")
	v.SetDefault("PLATS_CORE_GO_WHITEBOX_BIN", "")
	v.SetDefault("PLATS_CORE_GO_OPENSSL_BIN", "")
	v.SetDefault("PLATS_CERT_ROOT", "")
	v.SetDefault("PLATS_CA_ROOT", "")
	v.SetDefault("PLATS_SERVER_ROOT", "")
	v.SetDefault("PLATS_CERT_TMP_DIR", "")

	// ---------- 媒体管理默认值 ----------
	v.SetDefault("MEDIA_ROOT", "")
	v.SetDefault("MEDIA_CHUNK_SIZE", 5*1024*1024)
	v.SetDefault("MEDIA_SESSION_TTL", "2h")
	v.SetDefault("MEDIA_FFPROBE_PATH", "ffprobe")
	v.SetDefault("MEDIA_FFMPEG_PATH", "ffmpeg")

	// ---------- ZLM 默认值 ----------
	v.SetDefault("ZLM_BASE_URL", "")
	v.SetDefault("ZLM_SECRET", "")
	v.SetDefault("ZLM_HTTP_HOST", "127.0.0.1")
	v.SetDefault("ZLM_HTTP_PORT", 0) // 0 → 从 config.ini 读
	v.SetDefault("ZLM_RTMP_PORT", 1935)
	v.SetDefault("ZLM_RTC_PORT", 8000)
	v.SetDefault("PUBLIC_BASE_URL", "")

	// ---------- ZLM 进程管理默认值 ----------
	v.SetDefault("ZLM_AUTO_START", false)
	v.SetDefault("ZLM_ENGINE_BIN", "")
	v.SetDefault("ZLM_ENGINE_CONFIG", "")
	v.SetDefault("ZLM_ENGINE_LIB_DIR", "")
	v.SetDefault("ZLM_ENGINE_WORK_DIR", "")
	v.SetDefault("ZLM_ENGINE_LOG_FILE", "")
	v.SetDefault("ZLM_ENGINE_SHUTDOWN_ON_EXIT", false)

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

		// 前端动态菜单
		FrontendConfigPath: v.GetString("FRONTEND_CONFIG_PATH"),

		// Core 引擎
		CoreDispatchPath:   v.GetString("CORE_DISPATCH_PATH"),
		CoreRoot:           v.GetString("CORE_ROOT"),
		CoreTimeoutMs:      v.GetInt("CORE_TIMEOUT_MS"),
		CoreMaxConcurrency: v.GetInt("CORE_MAX_CONCURRENCY"),

		// 媒体管理
		MediaRoot:        v.GetString("MEDIA_ROOT"),
		MediaChunkSize:   v.GetInt64("MEDIA_CHUNK_SIZE"),
		MediaSessionTTL:  v.GetDuration("MEDIA_SESSION_TTL"),
		MediaFFprobePath: v.GetString("MEDIA_FFPROBE_PATH"),
		MediaFFmpegPath:  v.GetString("MEDIA_FFMPEG_PATH"),

		// ZLMediaKit（ZLM_HTTP_PORT / ZLM_SECRET 留空时由 applyZLMFromINI 从 ini 填充）
		ZLMBaseURL:    v.GetString("ZLM_BASE_URL"),
		ZLMSecret:     v.GetString("ZLM_SECRET"),
		ZLMHTTPHost:   v.GetString("ZLM_HTTP_HOST"),
		ZLMHTTPPort:   v.GetInt("ZLM_HTTP_PORT"),
		ZLMRTMPPort:   v.GetInt("ZLM_RTMP_PORT"),
		ZLMRTCPort:    v.GetInt("ZLM_RTC_PORT"),
		PublicBaseURL: v.GetString("PUBLIC_BASE_URL"),

		// ZLMediaKit 进程管理
		ZLMAutoStart:            v.GetBool("ZLM_AUTO_START"),
		ZLMEngineBin:            v.GetString("ZLM_ENGINE_BIN"),
		ZLMEngineConfig:         v.GetString("ZLM_ENGINE_CONFIG"),
		ZLMEngineLibDir:         v.GetString("ZLM_ENGINE_LIB_DIR"),
		ZLMEngineWorkDir:        v.GetString("ZLM_ENGINE_WORK_DIR"),
		ZLMEngineLogFile:        v.GetString("ZLM_ENGINE_LOG_FILE"),
		ZLMEngineShutdownOnExit: v.GetBool("ZLM_ENGINE_SHUTDOWN_ON_EXIT"),
	}

	// ---------- 路径归一化 ----------
	applyPathConfig(cfg, v)

	// ---------- 从 ZLM config.ini 自动读取端口与 secret ----------
	applyZLMFromINI(cfg)

	C = cfg
	return cfg, warnings, nil
}

// applyPathConfig 归一化 Core 引擎、证书路径、媒体根目录、对外 URL 与 ZLM 引擎路径。
//
// 覆盖字段：
//   - InstallRoot
//   - CoreGoBin / CoreGoWhiteboxBin / CoreGoOpensslBin
//   - CertRoot / CARoot / ServerRoot / CertTmpDir
//   - MediaRoot
//   - PublicBaseURL
//   - ZLMEngineBin / ZLMEngineConfig / ZLMEngineLibDir / ZLMEngineWorkDir / ZLMEngineLogFile
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
		filepath.Join(installRoot, "core", "bin", "whitebox_sm4"),
	)
	cfg.CoreGoOpensslBin = cleanAbs(
		v.GetString("PLATS_CORE_GO_OPENSSL_BIN"),
		filepath.Join(installRoot, "core", "run", "bin", "openssl"),
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

	// 媒体根目录
	cfg.MediaRoot = cleanAbs(
		cfg.MediaRoot,
		filepath.Join(installRoot, "user_data", "media"),
	)

	// 对外基础 URL
	if strings.TrimSpace(cfg.PublicBaseURL) == "" {
		host := cfg.ServerHost
		if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
			host = "127.0.0.1"
		}
		cfg.PublicBaseURL = fmt.Sprintf("http://%s:%d", host, cfg.ServerPort)
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")

	// ZLM 对外可达 host（拼播放地址用，浏览器必须能访问）。
	// 未显式配置 ZLM_HTTP_HOST 时自动取本机第一个非回环 IPv4，
	// 避免返回 127.0.0.1 导致浏览器客户端无法拉流。
	if strings.TrimSpace(cfg.ZLMHTTPHost) == "" || cfg.ZLMHTTPHost == "127.0.0.1" {
		if lan := deriveLANIP(); lan != "" {
			cfg.ZLMHTTPHost = lan
		}
	}

	// ZLM 引擎路径派生
	videoRoot := cleanAbs(
		v.GetString("PLATS_VIDEO_ROOT"),
		filepath.Join(installRoot, "video"),
	)
	cfg.ZLMEngineBin = cleanAbs(cfg.ZLMEngineBin,
		filepath.Join(videoRoot, "run", "bin", "MediaServer"))
	cfg.ZLMEngineConfig = cleanAbs(cfg.ZLMEngineConfig,
		filepath.Join(videoRoot, "conf", "zlmediakit", "config.ini"))
	cfg.ZLMEngineLibDir = cleanAbs(cfg.ZLMEngineLibDir,
		filepath.Join(videoRoot, "libs", "ZLMediaKit"))
	cfg.ZLMEngineWorkDir = cleanAbs(cfg.ZLMEngineWorkDir,
		filepath.Join(videoRoot, "run", "bin"))
	cfg.ZLMEngineLogFile = cleanAbs(cfg.ZLMEngineLogFile,
		filepath.Join(videoRoot, "logs", "MediaServer.log"))
}

// applyZLMFromINI 从 ZLM 的 config.ini 读取 [http] port 与 [api] secret，
// 覆盖到 cfg 上。
//
// 优先级（高到低）：
//  1. .env 显式配置（EnvExplicit("ZLM_BASE_URL" / "ZLM_HTTP_PORT" / "ZLM_SECRET")）
//  2. config.ini 实际值（ZLM 真正监听的才是真相）
//  3. 内置默认值（仅 ZLM_HTTP_PORT=8080 作为 ini 缺失时的兜底）
//
// 语义说明：
//   - 用户显式配置 ZLM_BASE_URL（完整 URL）时，视为"托管外部 ZLM"，
//     完全跳过 ini 读取，尊重显式值；
//   - 否则读 config.ini，以 ini 为准更新 ZLMHTTPPort / ZLMSecret，
//     并派生 ZLMBaseURL = http://127.0.0.1:<port>；
//   - 若 config.ini 的 secret 是 ZLM 官方示例值（035c...），
//     主动生成随机值写回 ini，避免 ZLM 启动时自动改写导致前后端不一致。
func applyZLMFromINI(cfg *Config) {
	// 用户显式配置完整 baseURL → 视为外部 ZLM，不读 ini
	if EnvExplicit("ZLM_BASE_URL") {
		cfg.ZLMBaseURL = strings.TrimRight(cfg.ZLMBaseURL, "/")
		return
	}

	if cfg.ZLMEngineConfig == "" {
		// 无 config.ini 路径（极少见），回退默认值
		if cfg.ZLMHTTPPort <= 0 {
			cfg.ZLMHTTPPort = 8080
		}
		if cfg.ZLMBaseURL == "" {
			cfg.ZLMBaseURL = fmt.Sprintf("http://127.0.0.1:%d", cfg.ZLMHTTPPort)
		}
		return
	}

	iniCfg, err := ReadZLMConfigINI(cfg.ZLMEngineConfig)
	if err != nil {
		// 文件不存在是正常情况（ZLM 尚未部署）
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr,
				"[config] WARN 读取 ZLM config.ini 失败（%s）：%v\n",
				cfg.ZLMEngineConfig, err)
		}
		if cfg.ZLMHTTPPort <= 0 {
			cfg.ZLMHTTPPort = 8080
		}
		if cfg.ZLMBaseURL == "" {
			cfg.ZLMBaseURL = fmt.Sprintf("http://127.0.0.1:%d", cfg.ZLMHTTPPort)
		}
		return
	}

	// ---------- 1. 主动修复 ZLM 官方示例 secret ----------
	iniSecret := iniCfg.APISecret
	if iniSecret == ZLMSampleSecret {
		newSecret := GenerateHexSecret(20)
		if err := UpdateZLMConfigINI(cfg.ZLMEngineConfig, "api", "secret", newSecret); err == nil {
			iniSecret = newSecret
			fmt.Fprintf(os.Stderr,
				"[config] INFO 检测到 ZLM 官方示例 secret，已生成随机 secret 写回 %s\n",
				cfg.ZLMEngineConfig)
		} else {
			fmt.Fprintf(os.Stderr,
				"[config] WARN 写回 ZLM config.ini 失败：%v（将使用 .env ZLM_SECRET）\n", err)
		}
	}

	// ---------- 2. HTTP 端口：以 ini 为准 ----------
	if iniCfg.HTTPPort > 0 {
		if EnvExplicit("ZLM_HTTP_PORT") && cfg.ZLMHTTPPort != iniCfg.HTTPPort {
			fmt.Fprintf(os.Stderr,
				"[config] WARN ZLM_HTTP_PORT=%d 与 config.ini [http] port=%d 不一致，"+
					"以 config.ini 为准（ZLM 实际监听端口）\n",
				cfg.ZLMHTTPPort, iniCfg.HTTPPort)
		}
		cfg.ZLMHTTPPort = iniCfg.HTTPPort
	} else if cfg.ZLMHTTPPort <= 0 {
		cfg.ZLMHTTPPort = 8080
	}

	// ---------- 3. API secret：以 ini 为准 ----------
	if iniSecret != "" {
		if EnvExplicit("ZLM_SECRET") && cfg.ZLMSecret != iniSecret {
			fmt.Fprintf(os.Stderr,
				"[config] WARN ZLM_SECRET 与 config.ini [api] secret 不一致，"+
					"以 config.ini 为准\n")
		}
		cfg.ZLMSecret = iniSecret
	}

	// ---------- 4. 派生 ZLMBaseURL ----------
	cfg.ZLMBaseURL = fmt.Sprintf("http://127.0.0.1:%d", cfg.ZLMHTTPPort)
}

// cleanAbs 归一化路径：空值回退默认，返回绝对路径。
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


// deriveLANIP 返回本机第一个非回环 IPv4 地址。
//
// 用于派生 ZLM_HTTP_HOST 的默认值：播放地址必须能被浏览器客户端访问，
// 不能是 127.0.0.1。取第一个满足条件的地址，失败时返回空字符串。
func deriveLANIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		if v4 := ip.To4(); v4 != nil {
			return v4.String()
		}
	}
	return ""
}
